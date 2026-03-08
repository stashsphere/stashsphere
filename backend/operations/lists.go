package operations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

type ListWithReason struct {
	List   models.List
	Reason AccessReasonInformation
}

type ListIdWithReason struct {
	ListId string
	Reason AccessReasonInformation
}

type ListFriendInformation struct {
	Id     string
	Friend string
}

type ListFriendOfFriendInformation struct {
	Id               string
	FriendId         string
	FriendOfFriendId string
}

// DeleteList deletes a list and cleans up related shares.
// The list must be loaded with Things and Shares relations.
func DeleteList(ctx context.Context, exec boil.ContextExecutor, list *models.List) error {
	shareIds := []string{}
	for _, share := range list.R.Shares {
		shareIds = append(shareIds, share.ID)
	}

	thingIds := []string{}
	for _, thing := range list.R.Things {
		thingIds = append(thingIds, thing.ID)
	}

	err := list.RemoveShares(ctx, exec, list.R.Shares...)
	if err != nil {
		return err
	}

	err = list.RemoveThings(ctx, exec, list.R.Things...)
	if err != nil {
		return err
	}

	for _, id := range shareIds {
		share, err := models.Shares(models.ShareWhere.ID.EQ(id),
			qm.Load(qm.Rels(models.ShareRels.Lists)),
			qm.Load(qm.Rels(models.ShareRels.Things)),
		).One(ctx, exec)
		if err != nil {
			return err
		}
		if len(share.R.Lists) == 0 && len(share.R.Things) == 0 {
			_, err = share.Delete(ctx, exec)
			if err != nil {
				return err
			}
		}
	}

	_, err = list.Delete(ctx, exec)
	if err != nil {
		return err
	}

	return RemoveForbiddenThingsFromCarts(ctx, exec, thingIds)
}

func GetListUnchecked(ctx context.Context, exec boil.ContextExecutor, listId string) (*models.List, error) {
	list, err := models.Lists(
		models.ListWhere.ID.EQ(listId),
		qm.Load(models.ListRels.Owner),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.ImagesThings, models.ImagesThingRels.Image)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.QuantityEntries)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Properties)),
		qm.Load(qm.Rels(models.ListRels.Shares, models.ShareRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Shares, models.ShareRels.TargetUser)),
	).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// second order of sharing
func getFriendOfFriendLists(ctx context.Context, exec boil.ContextExecutor, userId string, friendId string) ([]ListFriendOfFriendInformation, error) {
	sharedLists := make([]ListFriendOfFriendInformation, 0)
	type IdRow struct {
		Id               string `boil:"id"`
		FriendOfFriendId string `boil:"friend_of_friend_id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT
			l.id,
			l.owner_id AS friend_of_friend_id
		FROM lists l
		WHERE l.sharing_state='friends-of-friends'
		AND l.owner_id IN (
			SELECT
				CASE WHEN friend1_id=$1 THEN friend2_id ELSE friend1_id END AS other_id
			FROM friendships
			WHERE friend1_id=$1 OR friend2_id=$1
		)`, friendId,
	).Bind(ctx, exec, &idRows)
	if err != nil {
		return nil, err
	}
	for _, idRow := range idRows {
		sharedLists = append(sharedLists, ListFriendOfFriendInformation{
			Id:               idRow.Id,
			FriendId:         friendId,
			FriendOfFriendId: idRow.FriendOfFriendId,
		})
	}
	return sharedLists, nil
}

// first order of sharing
func getFriendLists(ctx context.Context, exec boil.ContextExecutor, userId string) ([]ListFriendInformation, error) {
	sharedLists := make([]ListFriendInformation, 0)
	type IdRow struct {
		Id       string `boil:"id"`
		FriendId string `boil:"friend_id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT
			l.id,
			l.owner_id AS friend_id
		FROM lists l
		WHERE (l.sharing_state='friends' OR l.sharing_state='friends-of-friends')
		AND l.owner_id IN (
			SELECT
				CASE WHEN friend1_id=$1 THEN friend2_id ELSE friend1_id END AS other_id
			FROM friendships
			WHERE friend1_id=$1 OR friend2_id=$1
		)`, userId,
	).Bind(ctx, exec, &idRows)
	if err != nil {
		return nil, err
	}
	for _, idRow := range idRows {
		sharedLists = append(sharedLists, ListFriendInformation{
			Id:     idRow.Id,
			Friend: idRow.FriendId,
		})
	}
	return sharedLists, nil
}

func GetSharedListIdsWithReasonForUser(ctx context.Context, exec boil.ContextExecutor, userId string) ([]ListIdWithReason, error) {
	listsWithReasons := make([]ListIdWithReason, 0)
	seenLists := make(map[string]bool)

	// 1. Lists shared directly with user via shares table (highest priority)
	type ListShareRow struct {
		ListId       string `boil:"list_id"`
		ShareOwnerId string `boil:"owner_id"`
	}
	var listShareRows []ListShareRow
	err := models.NewQuery(
		qm.Distinct("list_id, owner_id"),
		qm.From("shares_lists sl"),
		qm.InnerJoin("shares s on sl.share_id = s.id"),
		qm.Where("s.target_user_id=?", userId),
	).Bind(ctx, exec, &listShareRows)
	if err != nil {
		return nil, err
	}
	for _, row := range listShareRows {
		if !seenLists[row.ListId] {
			listsWithReasons = append(listsWithReasons, ListIdWithReason{
				ListId: row.ListId,
				Reason: AccessReasonSharedDirectly{ShareOwnerId: row.ShareOwnerId},
			})
			seenLists[row.ListId] = true
		}
	}

	// 2. Lists visible via direct friend sharing
	friendListInfos, err := getFriendLists(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	for _, info := range friendListInfos {
		if !seenLists[info.Id] {
			listsWithReasons = append(listsWithReasons, ListIdWithReason{
				ListId: info.Id,
				Reason: AccessReasonFriend{FriendId: info.Friend},
			})
			seenLists[info.Id] = true
		}
	}

	// 3. Lists visible via friend-of-friend sharing
	friendIds, err := GetFriendIds(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	for _, friendId := range friendIds {
		friendOfFriendListInfos, err := getFriendOfFriendLists(ctx, exec, userId, friendId)
		if err != nil {
			return nil, err
		}
		for _, info := range friendOfFriendListInfos {
			if !seenLists[info.Id] {
				listsWithReasons = append(listsWithReasons, ListIdWithReason{
					ListId: info.Id,
					Reason: AccessReasonFriendOfFriend{
						FriendId:         info.FriendId,
						FriendOfFriendId: info.FriendOfFriendId,
					},
				})
				seenLists[info.Id] = true
			}
		}
	}

	return listsWithReasons, nil
}

func GetDirectShareTargetUserIds(ctx context.Context, exec boil.ContextExecutor, listId string) ([]string, error) {
	type UserIdRow struct {
		TargetUserId string `boil:"target_user_id"`
	}
	var userIdRows []UserIdRow
	err := models.NewQuery(
		qm.Distinct("target_user_id"),
		qm.From("shares_lists"),
		qm.InnerJoin("shares on share_id = id"),
		qm.Where("list_id=?", listId),
	).Bind(ctx, exec, &userIdRows)
	if err != nil {
		return nil, err
	}
	userIds := make([]string, len(userIdRows))
	for i, row := range userIdRows {
		userIds[i] = row.TargetUserId
	}
	return userIds, nil
}

func GetListChecked(ctx context.Context, exec boil.ContextExecutor, listId string, userId string) (*ListWithReason, error) {
	list, err := GetListUnchecked(ctx, exec, listId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "List"}
		}
		return nil, err
	}
	sharedListsForUser, err := GetSharedListIdsWithReasonForUser(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	authorized, reason := func() (bool, AccessReasonInformation) {
		for _, listIdWithReason := range sharedListsForUser {
			if listIdWithReason.ListId == listId {
				return true, listIdWithReason.Reason
			}
		}
		if userId == list.OwnerID {
			return true, AccessReasonOwner{}
		}
		return false, nil
	}()
	if !authorized {
		return nil, utils.UserHasNoAccessRightsError{}
	}
	return &ListWithReason{
		List:   *list,
		Reason: reason,
	}, nil
}
