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

func GetThingUnchecked(ctx context.Context, exec boil.ContextExecutor, thingId string) (*models.Thing, error) {
	thing, err := models.Things(
		qm.Load(models.ThingRels.Properties),
		qm.Load(qm.Rels(models.ThingRels.Lists, models.ListRels.Owner)),
		qm.Load(models.ThingRels.Owner),
		qm.Load(models.ThingRels.QuantityEntries),
		qm.Load(qm.Rels(models.ThingRels.ImagesThings, models.ImagesThingRels.Image)),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.Owner)),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.TargetUser)),
		models.ThingWhere.ID.EQ(thingId)).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	return thing, nil
}

// second order of sharing
func getFriendOfFriendThings(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	sharedThingIds := make([]string, 0)
	type IdRow struct {
		Id string `boil:"id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT id from things where sharing_state='friends-of-friends' and owner_id in (
		SELECT 
		CASE WHEN friend1_id=$1 THEN friend2_id ELSE friend1_id END AS other_id
		FROM friendships
		WHERE friend1_id=$1 OR friend2_id=$1)`, userId,
	).Bind(ctx, exec, &idRows)
	if err != nil {
		return nil, err
	}
	for _, idRow := range idRows {
		sharedThingIds = append(sharedThingIds, idRow.Id)
	}
	return sharedThingIds, nil
}

// first order of sharing
func getFriendThings(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	sharedThingIds := make([]string, 0)
	type IdRow struct {
		Id string `boil:"id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT id from things where (sharing_state='friends' or sharing_state='friends-of-friends') and owner_id in (
		SELECT 
		CASE WHEN friend1_id=$1 THEN friend2_id ELSE friend1_id END AS other_id
		FROM friendships
		WHERE friend1_id=$1 OR friend2_id=$1)`, userId,
	).Bind(ctx, exec, &idRows)
	if err != nil {
		return nil, err
	}
	for _, idRow := range idRows {
		sharedThingIds = append(sharedThingIds, idRow.Id)
	}
	return sharedThingIds, nil
}

func GetSharedThingIdsForUser(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	sharedThingIds := make([]string, 0)

	friendThingIds, err := getFriendThings(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	for _, id := range friendThingIds {
		sharedThingIds = append(sharedThingIds, id)
	}

	friendIds, err := GetFriendIds(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	// get all things that are shared by the friend of the friend
	for _, friendId := range friendIds {
		friendOfFriendThings, err := getFriendOfFriendThings(ctx, exec, friendId)
		if err != nil {
			return nil, err
		}
		for _, id := range friendOfFriendThings {
			sharedThingIds = append(sharedThingIds, id)
		}
	}

	type ThingIdRow struct {
		ThingId string `boil:"thing_id"`
	}
	var sharedThingIdRows []ThingIdRow
	// SELECT DISTINCT thing_id from shares_things JOIN shares ON share_id = id WHERE target_user_id=?;
	err = models.NewQuery(
		qm.Distinct("thing_id"),
		qm.From("shares_things"),
		qm.InnerJoin("shares on share_id = id"),
		qm.Where("target_user_id=?", userId),
	).Bind(ctx, exec, &sharedThingIdRows)
	if err != nil {
		return nil, err
	}
	for _, thingIdRow := range sharedThingIdRows {
		sharedThingIds = append(sharedThingIds, thingIdRow.ThingId)
	}

	//SELECT DISTINCT lt.thing_id FROM public.lists_things lt
	//JOIN public.shares_lists sl ON lt.list_id = sl.list_id
	//JOIN public.shares s ON sl.share_id = s.id
	//WHERE s.target_user_id = '?';
	err = models.NewQuery(
		qm.Distinct("thing_id"),
		qm.From("lists_things lt"),
		qm.InnerJoin("shares_lists sl on lt.list_id = sl.list_id"),
		qm.InnerJoin("shares s on sl.share_id = s.id"),
		qm.Where("s.target_user_id=?", userId),
	).Bind(ctx, exec, &sharedThingIdRows)
	if err != nil {
		return nil, err
	}
	for _, thingIdRow := range sharedThingIdRows {
		sharedThingIds = append(sharedThingIds, thingIdRow.ThingId)
	}

	// fetch all shared lists
	listIds, err := GetSharedListIdsForUser(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	args := make([]interface{}, len(listIds))
	for i, id := range listIds {
		args[i] = id
	}
	err = models.NewQuery(
		qm.Distinct("thing_id"),
		qm.From("lists_things lt"),
		qm.WhereIn("lt.list_id in ?", args...),
	).Bind(ctx, exec, &sharedThingIdRows)
	if err != nil {
		return nil, err
	}
	for _, thingIdRow := range sharedThingIdRows {
		sharedThingIds = append(sharedThingIds, thingIdRow.ThingId)
	}
	return sharedThingIds, nil
}

func SumQuantity(thing *models.Thing) int64 {
	currentQuantity := int64(0)
	for _, x := range thing.R.QuantityEntries {
		currentQuantity += int64(x.DeltaValue)
	}
	return currentQuantity
}

func DeltaQuantity(thing *models.Thing, target uint64) int64 {
	return int64(target) - SumQuantity(thing)
}

func GetThingChecked(ctx context.Context, exec boil.ContextExecutor, thingId string, userId string) (*models.Thing, error) {
	thing, err := GetThingUnchecked(ctx, exec, thingId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "Thing"}
		}
		return nil, err
	}
	sharedThingsForUser, err := GetSharedThingIdsForUser(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	authorized := func() bool {
		for _, id := range sharedThingsForUser {
			if id == thingId {
				return true
			}
		}
		if userId == thing.OwnerID {
			return true
		}
		return false
	}()
	if !authorized {
		return nil, utils.UserHasNoAccessRightsError{}
	}
	return thing, nil
}
