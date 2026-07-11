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

type AccessReason uint32

const (
	Owner AccessReason = iota
	SharedDirectly
	Friend
	FriendOfFriend
	ListSharedDirectly
	ListFriend
	ListFriendOfFriend
)

type AccessReasonInformation interface {
	Reason() AccessReason
}

type AccessReasonOwner struct {
}

type AccessReasonSharedDirectly struct {
	ShareOwnerId string
}

type AccessReasonFriend struct {
	FriendId string
}

type AccessReasonFriendOfFriend struct {
	FriendId         string
	FriendOfFriendId string
}

type AccessReasonListSharedDirectly struct {
	ShareOwnerId string
	ListId       string
}

type AccessReasonListFriend struct {
	FriendId string
	ListId   string
}

type AccessReasonListFriendOfFriend struct {
	FriendId         string
	FriendOfFriendId string
	ListId           string
}

func (ar AccessReasonOwner) Reason() AccessReason {
	return Owner
}

func (ar AccessReasonSharedDirectly) Reason() AccessReason {
	return SharedDirectly
}

func (ar AccessReasonFriend) Reason() AccessReason {
	return Friend
}

func (ar AccessReasonFriendOfFriend) Reason() AccessReason {
	return FriendOfFriend
}

func (ar AccessReasonListSharedDirectly) Reason() AccessReason {
	return ListSharedDirectly
}

func (ar AccessReasonListFriend) Reason() AccessReason {
	return ListFriend
}

func (ar AccessReasonListFriendOfFriend) Reason() AccessReason {
	return ListFriendOfFriend
}

type ThingWithReason struct {
	Thing  models.Thing
	Reason AccessReasonInformation
}

type ThingIdWithReason struct {
	ThingId string
	Reason  AccessReasonInformation
}

func GetThingUnchecked(ctx context.Context, exec boil.ContextExecutor, thingId string) (*models.Thing, error) {
	thing, err := models.Things(
		qm.Load(models.ThingRels.Properties),
		qm.Load(qm.Rels(models.ThingRels.Lists, models.ListRels.Owner)),
		qm.Load(models.ThingRels.Owner),
		qm.Load(models.ThingRels.QuantityEntries),
		qm.Load(qm.Rels(models.ThingRels.ImagesThings, models.ImagesThingRels.Image)),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.Owner)),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.TargetUser)),
		qm.Load(models.ThingRels.PublicShares),
		models.ThingWhere.ID.EQ(thingId)).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	return thing, nil
}

// DeleteThing deletes a thing and cleans up related data.
// The thing must be loaded with QuantityEntries, Properties, ImagesThings, Shares, and Lists relations.
func DeleteThing(ctx context.Context, exec boil.ContextExecutor, thing *models.Thing) error {
	shareIds := []string{}
	for _, share := range thing.R.Shares {
		shareIds = append(shareIds, share.ID)
	}

	_, err := thing.R.QuantityEntries.DeleteAll(ctx, exec)
	if err != nil {
		return err
	}

	_, err = thing.R.Properties.DeleteAll(ctx, exec)
	if err != nil {
		return err
	}

	_, err = thing.R.ImagesThings.DeleteAll(ctx, exec)
	if err != nil {
		return err
	}

	err = thing.RemoveShares(ctx, exec, thing.R.Shares...)
	if err != nil {
		return err
	}

	err = thing.RemoveLists(ctx, exec, thing.R.Lists...)
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

	_, err = thing.Delete(ctx, exec)
	return err
}

type FriendOfFriendInformation struct {
	Id               string
	FriendId         string
	FriendOfFriendId string
}

// second order of sharing
func getFriendOfFriendThings(ctx context.Context, exec boil.ContextExecutor, userId string, friendId string) ([]FriendOfFriendInformation, error) {
	sharedThings := make([]FriendOfFriendInformation, 0)
	type IdRow struct {
		Id               string `boil:"id"`
		FriendOfFriendId string `boil:"friend_of_friend_id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT
			t.id,
			t.owner_id AS friend_of_friend_id
		FROM things t
		WHERE t.sharing_state='friends-of-friends'
		AND t.owner_id IN (
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
		sharedThings = append(sharedThings, FriendOfFriendInformation{
			Id:               idRow.Id,
			FriendId:         friendId,
			FriendOfFriendId: idRow.FriendOfFriendId,
		})
	}
	return sharedThings, nil
}

type FriendInformation struct {
	Id     string
	Friend string
}

// first order of sharing
func getFriendThings(ctx context.Context, exec boil.ContextExecutor, userId string) ([]FriendInformation, error) {
	sharedThings := make([]FriendInformation, 0)
	type IdRow struct {
		Id       string `boil:"id"`
		FriendId string `boil:"friend_id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT DISTINCT
			t.id,
			t.owner_id AS friend_id
		FROM things t
		WHERE (t.sharing_state='friends' OR t.sharing_state='friends-of-friends')
		AND t.owner_id IN (
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
		sharedThings = append(sharedThings, FriendInformation{
			Id:     idRow.Id,
			Friend: idRow.FriendId,
		})
	}
	return sharedThings, nil
}

func GetSharedThingsIdWithReasonForUser(ctx context.Context, exec boil.ContextExecutor, userId string) ([]ThingIdWithReason, error) {
	thingsWithReasons := make([]ThingIdWithReason, 0)
	seenThings := make(map[string]bool)

	// 1. Things shared directly with user via shares table (highest priority)
	type ThingShareRow struct {
		ThingId      string `boil:"thing_id"`
		ShareOwnerId string `boil:"owner_id"`
	}
	var thingShareRows []ThingShareRow
	err := models.NewQuery(
		qm.Distinct("thing_id, owner_id"),
		qm.From("shares_things st"),
		qm.InnerJoin("shares s on st.share_id = s.id"),
		qm.Where("s.target_user_id=?", userId),
	).Bind(ctx, exec, &thingShareRows)
	if err != nil {
		return nil, err
	}
	for _, row := range thingShareRows {
		if !seenThings[row.ThingId] {
			thingsWithReasons = append(thingsWithReasons, ThingIdWithReason{
				ThingId: row.ThingId,
				Reason:  AccessReasonSharedDirectly{ShareOwnerId: row.ShareOwnerId},
			})
			seenThings[row.ThingId] = true
		}
	}

	// 2. Things visible via direct friend sharing
	friendThingInfos, err := getFriendThings(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	for _, info := range friendThingInfos {
		if !seenThings[info.Id] {
			thingsWithReasons = append(thingsWithReasons, ThingIdWithReason{
				ThingId: info.Id,
				Reason:  AccessReasonFriend{FriendId: info.Friend},
			})
			seenThings[info.Id] = true
		}
	}

	// 3. Things visible via friend-of-friend sharing
	friendIds, err := GetFriendIds(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	for _, friendId := range friendIds {
		friendOfFriendThingInfos, err := getFriendOfFriendThings(ctx, exec, userId, friendId)
		if err != nil {
			return nil, err
		}
		for _, info := range friendOfFriendThingInfos {
			if !seenThings[info.Id] {
				thingsWithReasons = append(thingsWithReasons, ThingIdWithReason{
					ThingId: info.Id,
					Reason: AccessReasonFriendOfFriend{
						FriendId:         info.FriendId,
						FriendOfFriendId: info.FriendOfFriendId,
					},
				})
				seenThings[info.Id] = true
			}
		}
	}

	// 4. Things in lists shared directly with user
	type ListThingShareRow struct {
		ThingId      string `boil:"thing_id"`
		ListId       string `boil:"list_id"`
		ShareOwnerId string `boil:"owner_id"`
	}
	var listThingShareRows []ListThingShareRow
	err = models.NewQuery(
		qm.Distinct("lt.thing_id, sl.list_id, s.owner_id"),
		qm.From("lists_things lt"),
		qm.InnerJoin("shares_lists sl on lt.list_id = sl.list_id"),
		qm.InnerJoin("shares s on sl.share_id = s.id"),
		qm.Where("s.target_user_id=?", userId),
	).Bind(ctx, exec, &listThingShareRows)
	if err != nil {
		return nil, err
	}
	for _, row := range listThingShareRows {
		if !seenThings[row.ThingId] {
			thingsWithReasons = append(thingsWithReasons, ThingIdWithReason{
				ThingId: row.ThingId,
				Reason: AccessReasonListSharedDirectly{
					ShareOwnerId: row.ShareOwnerId,
					ListId:       row.ListId,
				},
			})
			seenThings[row.ThingId] = true
		}
	}

	// 5. Things in lists visible via friend/friend-of-friend sharing
	listIdsWithReasons, err := GetSharedListIdsWithReasonForUser(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	for _, listWithReason := range listIdsWithReasons {
		type ThingIdRow struct {
			ThingId string `boil:"thing_id"`
		}
		var thingIdRows []ThingIdRow
		err = models.NewQuery(
			qm.Distinct("thing_id"),
			qm.From("lists_things"),
			qm.Where("list_id=?", listWithReason.ListId),
		).Bind(ctx, exec, &thingIdRows)
		if err != nil {
			return nil, err
		}
		for _, thingRow := range thingIdRows {
			if !seenThings[thingRow.ThingId] {
				var reason AccessReasonInformation
				switch listReason := listWithReason.Reason.(type) {
				case AccessReasonFriend:
					reason = AccessReasonListFriend{
						FriendId: listReason.FriendId,
						ListId:   listWithReason.ListId,
					}
				case AccessReasonFriendOfFriend:
					reason = AccessReasonListFriendOfFriend{
						FriendId:         listReason.FriendId,
						FriendOfFriendId: listReason.FriendOfFriendId,
						ListId:           listWithReason.ListId,
					}
				default:
					// If list was shared directly, we already handled it above
					continue
				}
				thingsWithReasons = append(thingsWithReasons, ThingIdWithReason{
					ThingId: thingRow.ThingId,
					Reason:  reason,
				})
				seenThings[thingRow.ThingId] = true
			}
		}
	}

	return thingsWithReasons, nil
}

func GetOwnedThingIds(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	type ThingIdRow struct {
		ThingId string `boil:"thing_id"`
	}
	var rows []ThingIdRow
	err := models.NewQuery(
		qm.Distinct("id as thing_id"),
		qm.From("things"),
		qm.Where("owner_id = ?", userId),
	).Bind(ctx, exec, &rows)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ThingId
	}
	return ids, nil
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

func GetThingChecked(ctx context.Context, exec boil.ContextExecutor, thingId string, userId string) (*ThingWithReason, error) {
	thing, err := GetThingUnchecked(ctx, exec, thingId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "Thing"}
		}
		return nil, err
	}
	sharedThingsForUser, err := GetSharedThingsIdWithReasonForUser(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	authorized, reason := func() (bool, AccessReasonInformation) {
		for _, thingIdWithReason := range sharedThingsForUser {
			if thingIdWithReason.ThingId == thingId {
				return true, thingIdWithReason.Reason
			}
		}
		if userId == thing.OwnerID {
			return true, AccessReasonOwner{}
		}
		return false, nil
	}()
	if !authorized {
		return nil, utils.UserHasNoAccessRightsError{}
	}
	return &ThingWithReason{
		Thing:  *thing,
		Reason: reason,
	}, nil
}
