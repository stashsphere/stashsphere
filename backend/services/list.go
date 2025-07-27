package services

import (
	"context"
	"database/sql"
	"errors"
	"math"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	err := utils.Tx(ctx, ls.db, func(tx *sql.Tx) error {
		listID, err := gonanoid.New()
		if err != nil {
			return err
		}

		sharingState := models.SharingStatePrivate
		switch params.SharingState {
		case "friends":
			sharingState = models.SharingStateFriends
		case "friends-of-friends":
			sharingState = models.SharingStateFriendsOfFriends
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
	return ls.GetList(ctx, outerList.ID, outerList.OwnerID)
}

type UpdateListParams struct {
	Name         string
	ThingIds     []string
	SharingState string
}

func (ls *ListService) UpdateList(ctx context.Context, listId string, userId string, params UpdateListParams) (*models.List, error) {
	var outerList *models.List
	var newIdsInParameters []string
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

		sharingState := models.SharingStatePrivate
		switch params.SharingState {
		case "friends":
			sharingState = models.SharingStateFriends
		case "friends-of-friends":
			sharingState = models.SharingStateFriendsOfFriends
		}

		list.Name = params.Name
		list.SharingState = sharingState

		_, err = list.Update(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}

		oldThingIds := make(map[string]bool)
		for _, oldThing := range list.R.Things {
			oldThingIds[oldThing.ID] = true
		}
		for _, newId := range params.ThingIds {
			if _, ok := oldThingIds[newId]; !ok {
				// the newId does not exist yet
				newIdsInParameters = append(newIdsInParameters, newId)
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
		outerList = list
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ls.GetList(ctx, outerList.ID, outerList.OwnerID)
}

type GetListsForUserParams struct {
	UserId   string
	PerPage  uint64
	Page     uint64
	Paginate bool
}

func (ls *ListService) GetListsForUser(ctx context.Context, params GetListsForUserParams) (uint64, uint64, models.ListSlice, error) {
	userId, perPage, page, paginate := params.UserId, params.PerPage, params.Page, params.Paginate

	tx, err := ls.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return 0, 0, nil, err
	}

	sharedListIds, err := operations.GetSharedListIdsForUser(ctx, tx, userId)
	if err != nil {
		return 0, 0, nil, err
	}
	interfaceIds := make([]interface{}, len(sharedListIds))
	for i, s := range sharedListIds {
		interfaceIds[i] = s
	}

	searchCond := qm.Expr(
		models.ListWhere.OwnerID.EQ(userId),
		qm.OrIn("id in ?", interfaceIds...),
	)

	listCount, err := models.Lists(searchCond).Count(ctx, tx)
	if err != nil {
		return 0, 0, nil, err
	}

	// empty expr for no pagination
	listQuery := []qm.QueryMod{}
	if paginate {
		listQuery = append(listQuery, qm.Offset(int(perPage*page)), qm.Limit(int(perPage)))
	}

	sortCond := qm.OrderBy(models.ThingColumns.CreatedAt)

	listQuery = append(listQuery,
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Images)),
		qm.Load(models.ListRels.Owner),
		searchCond,
		sortCond,
	)

	lists, err := models.Lists(listQuery...).All(ctx, tx)
	if err != nil {
		return 0, 0, nil, err
	}

	totalPages := uint64(math.Ceil(float64(listCount) / float64(perPage)))

	return uint64(listCount), totalPages, lists, nil
}

func (ls *ListService) GetListsWhereThingIsPartOf(ctx context.Context, thingId string) (models.ListSlice, error) {
	return models.Lists(qm.InnerJoin("lists_things on lists.id = list_things.list_id", qm.Where("list_things.thingId = ?", thingId))).All(ctx, ls.db)
}

func (ls *ListService) GetList(ctx context.Context, listId string, userId string) (*models.List, error) {
	list, err := operations.GetListUnchecked(ctx, ls.db, listId)
	if err != nil {
		return nil, err
	}
	sharedListsForUsers, err := operations.GetSharedListIdsForUser(ctx, ls.db, userId)
	if err != nil {
		return nil, err
	}
	authorized := func() bool {
		for _, id := range sharedListsForUsers {
			if id == listId {
				return true
			}
		}
		if userId == list.OwnerID {
			return true
		}
		return false
	}()
	if !authorized {
		return nil, utils.UserHasNoAccessRightsError{}
	}
	return list, nil
}

func (ls *ListService) GetSharedListIdsForUser(ctx context.Context, userId string) ([]string, error) {
	return operations.GetSharedListIdsForUser(ctx, ls.db, userId)
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

		shareIds := []string{}
		for _, share := range list.R.Shares {
			shareIds = append(shareIds, share.ID)
		}

		err = list.RemoveShares(ctx, tx, list.R.Shares...)
		if err != nil {
			return err
		}

		err = list.RemoveThings(ctx, tx, list.R.Things...)
		if err != nil {
			return err
		}

		for _, id := range shareIds {
			share, err := models.Shares(models.ShareWhere.ID.EQ(id),
				qm.Load(qm.Rels(models.ShareRels.Lists)),
				qm.Load(qm.Rels(models.ShareRels.Things)),
			).One(ctx, tx)
			if err != nil {
				return err
			}
			if len(share.R.Lists) == 0 && len(share.R.Things) == 0 {
				_, err = share.Delete(ctx, tx)
				if err != nil {
					return err
				}
			}
		}

		_, err = list.Delete(ctx, tx)

		return err
	})
	return err
}
