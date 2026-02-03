package operations

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stashsphere/backend/models"
)

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
func getFriendOfFriendLists(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	sharedListIds := make([]string, 0)
	type IdRow struct {
		Id string `boil:"id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT id from lists where sharing_state='friends-of-friends' and owner_id in (
		SELECT 
		CASE WHEN friend1_id=$1 THEN friend2_id ELSE friend1_id END AS other_id
		FROM friendships
		WHERE friend1_id=$1 OR friend2_id=$1)`, userId,
	).Bind(ctx, exec, &idRows)
	if err != nil {
		return nil, err
	}
	for _, idRow := range idRows {
		sharedListIds = append(sharedListIds, idRow.Id)
	}
	return sharedListIds, nil
}

// first order of sharing
func getFriendLists(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	sharedListIds := make([]string, 0)
	type IdRow struct {
		Id string `boil:"id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT id from lists where (sharing_state='friends' or sharing_state='friends-of-friends') and owner_id in (
		SELECT 
		CASE WHEN friend1_id=$1 THEN friend2_id ELSE friend1_id END AS other_id
		FROM friendships
		WHERE friend1_id=$1 OR friend2_id=$1)`, userId,
	).Bind(ctx, exec, &idRows)
	if err != nil {
		return nil, err
	}
	for _, idRow := range idRows {
		sharedListIds = append(sharedListIds, idRow.Id)
	}
	return sharedListIds, nil
}

func GetSharedListIdsForUser(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	sharedListIds := make([]string, 0)

	friendListIds, err := getFriendLists(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	for _, id := range friendListIds {
		sharedListIds = append(sharedListIds, id)
	}

	friendIds, err := GetFriendIds(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	// get all lists that are shared by the friend of the friend
	for _, friendId := range friendIds {
		friendOfFriendLists, err := getFriendOfFriendLists(ctx, exec, friendId)
		if err != nil {
			return nil, err
		}
		for _, id := range friendOfFriendLists {
			sharedListIds = append(sharedListIds, id)
		}
	}

	type ListIdRow struct {
		ListId string `boil:"list_id"`
	}
	var sharedListIdRows []ListIdRow
	err = models.NewQuery(
		qm.Distinct("list_id"),
		qm.From("shares_lists"),
		qm.InnerJoin("shares on share_id = id"),
		qm.Where("target_user_id=?", userId),
	).Bind(ctx, exec, &sharedListIdRows)
	if err != nil {
		return nil, err
	}
	for _, listIdRow := range sharedListIdRows {
		sharedListIds = append(sharedListIds, listIdRow.ListId)
	}
	return sharedListIds, nil
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
