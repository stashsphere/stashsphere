package operations

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

// PurgeUser deletes a user and all their related data.
func PurgeUser(ctx context.Context, exec boil.ContextExecutor, userId string, imageStorePath string) error {
	user, err := models.Users(
		models.UserWhere.ID.EQ(userId),
		qm.Load(models.UserRels.CartEntries),
		qm.Load(models.UserRels.RecipientNotifications),
		qm.Load(models.UserRels.SenderFriendRequests),
		qm.Load(models.UserRels.ReceiverFriendRequests),
		qm.Load(models.UserRels.Friend1Friendships),
		qm.Load(models.UserRels.Friend2Friendships),
		qm.Load(models.UserRels.OwnerShares),
		qm.Load(models.UserRels.TargetUserShares),
		qm.Load(models.UserRels.OwnerImages),
		qm.Load(models.UserRels.Profile),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.NotFoundError{EntityName: "User"}
		}
		return err
	}

	// Delete cart entries
	if _, err := user.R.CartEntries.DeleteAll(ctx, exec); err != nil {
		return err
	}

	// Delete notifications
	if _, err := user.R.RecipientNotifications.DeleteAll(ctx, exec); err != nil {
		return err
	}

	// Delete friendships first (they have FK to friend_requests)
	if _, err := user.R.Friend1Friendships.DeleteAll(ctx, exec); err != nil {
		return err
	}
	if _, err := user.R.Friend2Friendships.DeleteAll(ctx, exec); err != nil {
		return err
	}

	// Delete friend requests (sent and received)
	if _, err := user.R.SenderFriendRequests.DeleteAll(ctx, exec); err != nil {
		return err
	}
	if _, err := user.R.ReceiverFriendRequests.DeleteAll(ctx, exec); err != nil {
		return err
	}

	// Delete shares owned by user (load relations and remove them first)
	for _, share := range user.R.OwnerShares {
		shareWithRels, err := models.Shares(
			models.ShareWhere.ID.EQ(share.ID),
			qm.Load(models.ShareRels.Things),
			qm.Load(models.ShareRels.Lists),
		).One(ctx, exec)
		if err != nil {
			return err
		}
		if err := shareWithRels.RemoveThings(ctx, exec, shareWithRels.R.Things...); err != nil {
			return err
		}
		if err := shareWithRels.RemoveLists(ctx, exec, shareWithRels.R.Lists...); err != nil {
			return err
		}
	}
	if _, err := user.R.OwnerShares.DeleteAll(ctx, exec); err != nil {
		return err
	}

	// Delete shares targeting user (load relations and remove them first)
	for _, share := range user.R.TargetUserShares {
		shareWithRels, err := models.Shares(
			models.ShareWhere.ID.EQ(share.ID),
			qm.Load(models.ShareRels.Things),
			qm.Load(models.ShareRels.Lists),
		).One(ctx, exec)
		if err != nil {
			return err
		}
		if err := shareWithRels.RemoveThings(ctx, exec, shareWithRels.R.Things...); err != nil {
			return err
		}
		if err := shareWithRels.RemoveLists(ctx, exec, shareWithRels.R.Lists...); err != nil {
			return err
		}
	}
	if _, err := user.R.TargetUserShares.DeleteAll(ctx, exec); err != nil {
		return err
	}

	// Delete things (load each with relations and use DeleteThing)
	things, err := models.Things(
		models.ThingWhere.OwnerID.EQ(userId),
		qm.Load(models.ThingRels.Properties),
		qm.Load(models.ThingRels.QuantityEntries),
		qm.Load(models.ThingRels.ImagesThings),
		qm.Load(models.ThingRels.Shares),
		qm.Load(models.ThingRels.Lists),
	).All(ctx, exec)
	if err != nil {
		return err
	}
	for _, thing := range things {
		if err := DeleteThing(ctx, exec, thing); err != nil {
			return err
		}
	}

	// Delete lists (load each with relations and use DeleteList)
	lists, err := models.Lists(
		models.ListWhere.OwnerID.EQ(userId),
		qm.Load(models.ListRels.Things),
		qm.Load(models.ListRels.Shares),
	).All(ctx, exec)
	if err != nil {
		return err
	}
	for _, list := range lists {
		if err := DeleteList(ctx, exec, list); err != nil {
			return err
		}
	}

	// Delete profile first (it may reference images via image_id FK)
	if user.R.Profile != nil {
		if _, err := user.R.Profile.Delete(ctx, exec); err != nil {
			return err
		}
	}

	// Delete images and their files
	// First clean up images_things entries and collect hashes
	imageHashes := make([]string, 0, len(user.R.OwnerImages))
	for _, image := range user.R.OwnerImages {
		if _, err := models.ImagesThings(models.ImagesThingWhere.ImageID.EQ(image.ID)).DeleteAll(ctx, exec); err != nil {
			return err
		}
		imageHashes = append(imageHashes, image.Hash)
	}
	// Delete image records from database
	if _, err := user.R.OwnerImages.DeleteAll(ctx, exec); err != nil {
		return err
	}
	// Now delete the files (after records are gone, so DeleteContent's count check passes)
	for _, hash := range imageHashes {
		if err := DeleteContent(ctx, exec, imageStorePath, hash); err != nil {
			if !errors.Is(err, utils.EntityInUseError{}) {
				return err
			}
		}
	}

	// Finally delete the user
	if _, err := user.Delete(ctx, exec); err != nil {
		return err
	}

	return nil
}

func ScheduleUserDeletion(ctx context.Context, exec boil.ContextExecutor, userId string, purgeAt time.Time) (*models.User, error) {
	user, err := FindUserByID(ctx, exec, userId)
	if err != nil {
		return nil, err
	}

	user.PurgeAt = null.TimeFrom(purgeAt)
	_, err = user.Update(ctx, exec, boil.Whitelist(models.UserColumns.PurgeAt))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func FindUserByID(ctx context.Context, exec boil.ContextExecutor, userId string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "user"}
		}
		return nil, err
	}
	return user, nil
}

func FindUserByEmail(ctx context.Context, exec boil.ContextExecutor, email string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.Email.EQ(email)).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "user"}
		}
		return nil, err
	}
	return user, nil
}

func FindUserWithProfileByID(ctx context.Context, exec boil.ContextExecutor, userId string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.ID.EQ(userId),
		qm.Load(models.UserRels.Profile),
		qm.Load(qm.Rels(models.UserRels.Profile, models.ProfileRels.Image)),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "user"}
		}
		return nil, err
	}
	return user, nil
}
