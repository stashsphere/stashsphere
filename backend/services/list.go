package services

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
)

type ListService struct {
	db *sql.DB
	ns *NotificationService
}

func NewListService(db *sql.DB, ns *NotificationService) *ListService {
	return &ListService{db, ns}
}

type CreateListParams struct {
	Name         string
	ThingIds     []string
	OwnerId      string
	SharingState string
}

func (ls *ListService) CreateList(ctx context.Context, params CreateListParams) (*models.List, error) {
	var outerList *models.List
	targetUsersIds := []string{}

	err := utils.Tx(ctx, ls.db, func(tx *sql.Tx) error {
		listID, err := gonanoid.New()
		if err != nil {
			return err
		}

		var sharingState models.SharingState
		if params.SharingState == "" {
			sharingState, err = operations.GetUserDefaultSharingState(ctx, tx, params.OwnerId)
			if err != nil {
				return err
			}
		} else {
			sharingState = models.SharingStatePrivate
			switch params.SharingState {
			case "friends":
				sharingState = models.SharingStateFriends
			case "friends-of-friends":
				sharingState = models.SharingStateFriendsOfFriends
			}
		}

		switch sharingState {
		case models.SharingStateFriends:
			targetUsersIds, err = operations.GetFriendIds(ctx, tx, params.OwnerId)
			if err != nil {
				return err
			}
		case models.SharingStateFriendsOfFriends:
			ownerFriendIds, err := operations.GetFriendIds(ctx, tx, params.OwnerId)
			if err != nil {
				return err
			}
			for _, friendId := range ownerFriendIds {
				friendOfFriendIds, err := operations.GetFriendIds(ctx, tx, friendId)
				if err != nil {
					return err
				}
				targetUsersIds = append(targetUsersIds, friendOfFriendIds...)
				targetUsersIds = append(targetUsersIds, friendId)
			}
		}

		list := models.List{
			ID:           listID,
			Name:         params.Name,
			OwnerID:      params.OwnerId,
			SharingState: sharingState,
		}

		err = list.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		for _, thingId := range params.ThingIds {
			thing, err := models.Things(models.ThingWhere.ID.EQ(thingId)).One(ctx, tx)
			if err != nil {
				return err
			}
			if thing.OwnerID != params.OwnerId {
				return utils.EntityDoesNotBelongToUserError{}
			}
			err = thing.AddLists(ctx, tx, false, &list)
			if err != nil {
				return err
			}
		}

		outerList = &list
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, targetUserId := range targetUsersIds {
		ls.ns.ListShared(ctx, ListSharedParams{
			ListId:       outerList.ID,
			SharedId:     outerList.OwnerID,
			TargetUserId: targetUserId,
		})
	}
	listWithReason, err := ls.GetList(ctx, outerList.ID, outerList.OwnerID)
	if err != nil {
		return nil, err
	}
	return &listWithReason.List, nil
}

type UpdateListParams struct {
	Name         string
	ThingIds     []string
	SharingState string
}

func (ls *ListService) UpdateList(ctx context.Context, listId string, userId string, params UpdateListParams) (*models.List, error) {
	var outerList *models.List
	targetUsersIds := []string{}
	thingsAddedTargetUserIds := []string{}
	err := utils.Tx(ctx, ls.db, func(tx *sql.Tx) error {
		list, err := models.Lists(qm.Load(models.ListRels.Things), models.ListWhere.ID.EQ(listId)).One(ctx, tx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return utils.NotFoundError{EntityName: "List"}
			}
			return err
		}
		if list.OwnerID != userId {
			return utils.EntityDoesNotBelongToUserError{}
		}

		originalState := list.SharingState

		var sharingState models.SharingState
		if params.SharingState == "" {
			sharingState = originalState
		} else {
			sharingState = models.SharingStatePrivate
			switch params.SharingState {
			case "friends":
				sharingState = models.SharingStateFriends
			case "friends-of-friends":
				sharingState = models.SharingStateFriendsOfFriends
			}

			if sharingState == models.SharingStateFriends && originalState == models.SharingStatePrivate {
				targetUsersIds, err = operations.GetFriendIds(ctx, tx, userId)
				if err != nil {
					return err
				}
			} else if sharingState == models.SharingStateFriendsOfFriends && originalState != models.SharingStateFriendsOfFriends {
				ownerFriendIds, err := operations.GetFriendIds(ctx, tx, userId)
				if err != nil {
					return err
				}
				for _, friendId := range ownerFriendIds {
					friendOfFriendIds, err := operations.GetFriendIds(ctx, tx, friendId)
					if err != nil {
						return err
					}
					targetUsersIds = append(targetUsersIds, friendOfFriendIds...)
					if originalState == models.SharingStatePrivate {
						targetUsersIds = append(targetUsersIds, friendId)
					}
				}
			}
		}

		list.Name = params.Name
		list.SharingState = sharingState

		_, err = list.Update(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}

		allThingIds := []string{}
		oldThingIds := make(map[string]bool)
		for _, oldThing := range list.R.Things {
			oldThingIds[oldThing.ID] = true
			allThingIds = append(allThingIds, oldThing.ID)
		}
		hasNewThings := false
		for _, newId := range params.ThingIds {
			if _, ok := oldThingIds[newId]; !ok {
				allThingIds = append(allThingIds, newId)
				hasNewThings = true
			}
		}

		// Notify users if new things were added to an already-shared list
		if hasNewThings {
			userIdSet := make(map[string]bool)

			switch originalState {
			case models.SharingStateFriends:
				friendIds, err := operations.GetFriendIds(ctx, tx, userId)
				if err != nil {
					return err
				}
				for _, id := range friendIds {
					userIdSet[id] = true
				}
			case models.SharingStateFriendsOfFriends:
				ownerFriendIds, err := operations.GetFriendIds(ctx, tx, userId)
				if err != nil {
					return err
				}
				for _, friendId := range ownerFriendIds {
					friendOfFriendIds, err := operations.GetFriendIds(ctx, tx, friendId)
					if err != nil {
						return err
					}
					for _, id := range friendOfFriendIds {
						userIdSet[id] = true
					}
					userIdSet[friendId] = true
				}
			}

			// Also notify users with direct shares
			directShareUserIds, err := operations.GetDirectShareTargetUserIds(ctx, tx, list.ID)
			if err != nil {
				return err
			}
			for _, id := range directShareUserIds {
				userIdSet[id] = true
			}

			for id := range userIdSet {
				thingsAddedTargetUserIds = append(thingsAddedTargetUserIds, id)
			}
		}

		err = list.RemoveThings(ctx, tx, list.R.Things...)
		if err != nil {
			return err
		}

		for _, thingId := range params.ThingIds {
			thing, err := models.Things(models.ThingWhere.ID.EQ(thingId)).One(ctx, tx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return utils.NotFoundError{EntityName: "Thing"}
				}
				return err
			}
			if thing.OwnerID != userId {
				return utils.EntityDoesNotBelongToUserError{}
			}
			err = thing.AddLists(ctx, tx, false, list)
			if err != nil {
				return err
			}
		}
		err = operations.RemoveForbiddenThingsFromCarts(ctx, tx, allThingIds)
		if err != nil {
			return err
		}

		outerList = list
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, targetUserId := range targetUsersIds {
		ls.ns.ListShared(ctx, ListSharedParams{
			ListId:       outerList.ID,
			SharedId:     outerList.OwnerID,
			TargetUserId: targetUserId,
		})
	}
	for _, targetUserId := range thingsAddedTargetUserIds {
		ls.ns.ThingsAddedToList(ctx, ThingsAddedToListParams{
			ListId:       outerList.ID,
			OwnerId:      outerList.OwnerID,
			TargetUserId: targetUserId,
		})
	}
	listWithReason, err := ls.GetList(ctx, outerList.ID, outerList.OwnerID)
	if err != nil {
		return nil, err
	}
	return &listWithReason.List, nil
}

type GetListsForUserParams struct {
	UserId         string
	PerPage        uint64
	Page           uint64
	Paginate       bool
	FilterOwnerIds []string
}

type GetListsForUserResult struct {
	TotalCount     uint64
	TotalPages     uint64
	Lists          models.ListSlice
	ListReasonMap  map[string]operations.AccessReasonInformation
	ThingReasonMap map[string]operations.AccessReasonInformation
}

func (ls *ListService) GetListsForUser(ctx context.Context, params GetListsForUserParams) (*GetListsForUserResult, error) {
	userId, perPage, page, paginate, filterUserIds := params.UserId, params.PerPage, params.Page, params.Paginate, params.FilterOwnerIds

	tx, err := ls.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	sharedListsWithReasons, err := operations.GetSharedListIdsWithReasonForUser(ctx, tx, userId)
	if err != nil {
		return nil, err
	}
	interfaceIds := make([]interface{}, len(sharedListsWithReasons))
	listReasonMap := make(map[string]operations.AccessReasonInformation)
	for i, s := range sharedListsWithReasons {
		interfaceIds[i] = s.ListId
		listReasonMap[s.ListId] = s.Reason
	}

	sharedThingsWithReasons, err := operations.GetSharedThingsIdWithReasonForUser(ctx, tx, userId)
	if err != nil {
		return nil, err
	}
	thingReasonMap := make(map[string]operations.AccessReasonInformation)
	for _, s := range sharedThingsWithReasons {
		thingReasonMap[s.ThingId] = s.Reason
	}

	searchCond := qm.Expr(
		models.ListWhere.OwnerID.EQ(userId),
		qm.OrIn("id in ?", interfaceIds...),
	)

	if len(filterUserIds) > 0 {
		filterUserInterfaceIds := make([]interface{}, len(filterUserIds))
		for i, u := range filterUserIds {
			filterUserInterfaceIds[i] = u
		}
		searchCond = qm.Expr(searchCond, qm.AndIn("owner_id in ?", filterUserInterfaceIds...))
	}

	listCount, err := models.Lists(searchCond).Count(ctx, tx)
	if err != nil {
		return nil, err
	}

	// empty expr for no pagination
	listQuery := []qm.QueryMod{}
	if paginate {
		listQuery = append(listQuery, qm.Offset(int(perPage*page)), qm.Limit(int(perPage)))
	}

	sortCond := qm.OrderBy(models.ThingColumns.CreatedAt)

	listQuery = append(listQuery,
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.ImagesThings, models.ImagesThingRels.Image)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.QuantityEntries)),
		qm.Load(models.ListRels.Owner),
		qm.Load(qm.Rels(models.ListRels.Shares, models.ShareRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Shares, models.ShareRels.TargetUser)),
		searchCond,
		sortCond,
	)

	lists, err := models.Lists(listQuery...).All(ctx, tx)
	if err != nil {
		return nil, err
	}

	// Add Owner reasons for lists that belong to the user
	for _, list := range lists {
		if list.OwnerID == userId {
			listReasonMap[list.ID] = operations.AccessReasonOwner{}
		}
	}

	totalPages := uint64(math.Ceil(float64(listCount) / float64(perPage)))

	return &GetListsForUserResult{
		TotalCount:     uint64(listCount),
		TotalPages:     totalPages,
		Lists:          lists,
		ListReasonMap:  listReasonMap,
		ThingReasonMap: thingReasonMap,
	}, nil
}

func (ls *ListService) GetListsWhereThingIsPartOf(ctx context.Context, thingId string) (models.ListSlice, error) {
	return models.Lists(qm.InnerJoin("lists_things on lists.id = list_things.list_id", qm.Where("list_things.thingId = ?", thingId))).All(ctx, ls.db)
}

func (ls *ListService) GetList(ctx context.Context, listId string, userId string) (*operations.ListWithReason, error) {
	return operations.GetListChecked(ctx, ls.db, listId, userId)
}

func (ls *ListService) GetSharedListIdsForUser(ctx context.Context, userId string) ([]string, error) {
	listsWithReasons, err := operations.GetSharedListIdsWithReasonForUser(ctx, ls.db, userId)
	if err != nil {
		return nil, err
	}
	listIds := make([]string, len(listsWithReasons))
	for i, listWithReason := range listsWithReasons {
		listIds[i] = listWithReason.ListId
	}
	return listIds, nil
}

func (ts *ListService) DeleteList(ctx context.Context, listId string, userId string) error {
	err := utils.Tx(ctx, ts.db, func(tx *sql.Tx) error {
		list, err := operations.GetListUnchecked(ctx, tx, listId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return utils.NotFoundError{EntityName: "List"}
			}
			return err
		}
		if list.OwnerID != userId {
			return utils.EntityDoesNotBelongToUserError{}
		}

		return operations.DeleteList(ctx, tx, list)
	})
	return err
}
