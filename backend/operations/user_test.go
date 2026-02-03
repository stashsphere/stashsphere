package operations_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestPurgeUser(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(imageService.StorePath())
	})

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)

	userService := services.NewUserService(db, false, "", 60, notificationService)
	thingService := services.NewThingService(db, imageService, notificationService)
	listService := services.NewListService(db, notificationService)
	friendService := services.NewFriendService(db, notificationService)

	// Create user to be purged
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	// Create another user (bob) for friendship testing
	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	// Create an image and set it as alice's profile image
	pngFile, err := testcommon.Assets.Open("assets/test.png")
	assert.NoError(t, err)
	profileImage, err := imageService.CreateImage(context.Background(), alice.ID, "test.png", pngFile)
	assert.NoError(t, err)

	_, err = userService.UpdateUser(context.Background(), services.UpdateUserParams{
		UserId:  alice.ID,
		Name:    alice.Name,
		ImageId: &profileImage.ID,
	})
	assert.NoError(t, err)

	// Create a thing for alice
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thingParams.SharingState = "private"
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	// Create a list for alice containing the thing
	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	listParams.ThingIds = []string{thing.ID}
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	// Create friendship between alice and bob
	request, err := friendService.CreateFriendRequest(context.Background(), services.CreateFriendRequestParams{
		UserId:     alice.ID,
		ReceiverId: bob.ID,
	})
	assert.NoError(t, err)
	_, err = friendService.ReactFriendRequest(context.Background(), services.ReactFriendRequestParams{
		FriendRequestId: request.ID,
		UserId:          bob.ID,
		Accept:          true,
	})
	assert.NoError(t, err)

	// Verify data exists before purge
	userExists, err := models.UserExists(context.Background(), db, alice.ID)
	assert.NoError(t, err)
	assert.True(t, userExists, "user should exist before purge")

	thingExists, err := models.ThingExists(context.Background(), db, thing.ID)
	assert.NoError(t, err)
	assert.True(t, thingExists, "thing should exist before purge")

	listExists, err := models.ListExists(context.Background(), db, list.ID)
	assert.NoError(t, err)
	assert.True(t, listExists, "list should exist before purge")

	friendships, err := models.Friendships().All(context.Background(), db)
	assert.NoError(t, err)
	assert.Len(t, friendships, 1, "friendship should exist before purge")

	imageFilePath := filepath.Join(imageService.StorePath(), profileImage.Hash)
	_, err = os.Stat(imageFilePath)
	assert.NoError(t, err, "image file should exist before purge")

	// Purge the user
	err = utils.Tx(context.Background(), db, func(tx *sql.Tx) error {
		return operations.PurgeUser(context.Background(), tx, alice.ID, imageService.StorePath())
	})
	assert.NoError(t, err)

	// Verify all data is deleted
	userExists, err = models.UserExists(context.Background(), db, alice.ID)
	assert.NoError(t, err)
	assert.False(t, userExists, "user should not exist after purge")

	thingExists, err = models.ThingExists(context.Background(), db, thing.ID)
	assert.NoError(t, err)
	assert.False(t, thingExists, "thing should not exist after purge")

	listExists, err = models.ListExists(context.Background(), db, list.ID)
	assert.NoError(t, err)
	assert.False(t, listExists, "list should not exist after purge")

	// Friendship should be deleted (alice was part of it)
	friendships, err = models.Friendships().All(context.Background(), db)
	assert.NoError(t, err)
	assert.Len(t, friendships, 0, "friendship should be deleted after purge")

	// Bob should still exist
	bobExists, err := models.UserExists(context.Background(), db, bob.ID)
	assert.NoError(t, err)
	assert.True(t, bobExists, "bob should still exist after alice is purged")

	// Image file should be deleted
	_, err = os.Stat(imageFilePath)
	assert.True(t, os.IsNotExist(err), "image file should be deleted after purge")
}

func TestPurgeUserNotFound(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	err = operations.PurgeUser(context.Background(), db, "non-existent-user-id", "/tmp")
	assert.Error(t, err)
	assert.IsType(t, utils.NotFoundError{}, err)
}
