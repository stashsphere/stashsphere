package services_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestThingCreation(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)

	t.Cleanup(tearDownFunc)
	is, err := services.NewTmpImageService(db)
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})
	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)

	thingService := services.NewThingService(db, is)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = testUser.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)
	assert.NotNil(t, thing)
	assert.NotEmpty(t, thing.ID)
}

func TestThingAccess(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)

	t.Cleanup(tearDownFunc)
	is, err := services.NewTmpImageService(db)
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})
	userService := services.NewUserService(db, false, "")
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	malloryParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	mallory, err := userService.CreateUser(context.Background(), *malloryParams)
	assert.NoError(t, err)

	thingService := services.NewThingService(db, is)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	_, err = thingService.GetThing(context.Background(), thing.ID, mallory.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})
}

func TestThingAccessShareThing(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.Nil(t, err)
	t.Cleanup(tearDownFunc)

	is, err := services.NewTmpImageService(db)
	assert.Nil(t, err)
	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})

	userService := services.NewUserService(db, false, "")
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, emailService)
	shareService := services.NewShareService(db, notificationService)
	thingService := services.NewThingService(db, is)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.Nil(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.Nil(t, err)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.Nil(t, err)

	_, err = thingService.GetThing(context.Background(), thing.ID, bob.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})

	share, err := shareService.CreateThingShare(context.Background(), services.CreateThingShareParams{
		ThingId:      thing.ID,
		OwnerId:      alice.ID,
		TargetUserId: bob.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, share)

	_, err = thingService.GetThing(context.Background(), thing.ID, bob.ID)
	assert.Nil(t, err, "bob has access through thing share")
}

// Test whether quantity is properly saved
func TestThingQuantity(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.Nil(t, err)
	t.Cleanup(tearDownFunc)

	is, err := services.NewTmpImageService(db)
	assert.Nil(t, err)
	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})

	userService := services.NewUserService(db, false, "")
	thingService := services.NewThingService(db, is)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.Nil(t, err)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.Nil(t, err)

	assert.Equal(t, thing.QuantityUnit, "pcs")
	assert.Equal(t, operations.SumQuantity(thing), int64(0))

	updatedThing, err := thingService.EditThing(context.Background(), thing.ID, alice.ID, services.UpdateThingParams{
		Quantity: 123, QuantityUnit: "kg",
	})
	assert.Nil(t, err)

	assert.Equal(t, updatedThing.QuantityUnit, "kg")
	assert.Equal(t, operations.SumQuantity(updatedThing), int64(123))

	updatedThing, err = thingService.EditThing(context.Background(), thing.ID, alice.ID, services.UpdateThingParams{
		Quantity: 1337, QuantityUnit: "meters",
	})
	assert.Nil(t, err)

	assert.Equal(t, updatedThing.QuantityUnit, "meters")
	assert.Equal(t, operations.SumQuantity(updatedThing), int64(1337))
}

func createFriendShip(t *testing.T, db *sql.DB, userId1 string, userId2 string) {
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, emailService)
	friendService := services.NewFriendService(db, notificationService)

	request, err := friendService.CreateFriendRequest(context.Background(), services.CreateFriendRequestParams{
		UserId:     userId1,
		ReceiverId: userId2,
	})
	assert.NoError(t, err)
	assert.NotNil(t, request)
	_, err = friendService.ReactFriendRequest(context.Background(), services.ReactFriendRequestParams{
		FriendRequestId: request.ID,
		UserId:          userId2,
		Accept:          true,
	})
	assert.NoError(t, err)
}

func TestSharingState(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDownFunc)

	is, err := services.NewTmpImageService(db)
	assert.Nil(t, err)
	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})

	userService := services.NewUserService(db, false, "")
	thingService := services.NewThingService(db, is)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	charlieParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	charlie, err := userService.CreateUser(context.Background(), *charlieParams)
	assert.NoError(t, err)

	// bob is a friend of alice
	createFriendShip(t, db, alice.ID, bob.ID)
	// charlie is a friend of bob, but not of alice
	createFriendShip(t, db, charlie.ID, bob.ID)

	privateThingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	privateThingParams.OwnerId = alice.ID
	privateThing, err := thingService.CreateThing(context.Background(), *privateThingParams)
	assert.NoError(t, err)

	friendThingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	friendThingParams.OwnerId = alice.ID
	friendThingParams.SharingState = models.SharingStateFriends.String()

	friendThing, err := thingService.CreateThing(context.Background(), *friendThingParams)
	assert.NoError(t, err)

	friendsOfFriendsThingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	friendsOfFriendsThingParams.OwnerId = alice.ID
	friendsOfFriendsThingParams.SharingState = models.SharingStateFriendsOfFriends.String()

	friendsOfFriendsThing, err := thingService.CreateThing(context.Background(), *friendsOfFriendsThingParams)
	assert.NoError(t, err)

	_, err = thingService.GetThing(context.Background(), privateThing.ID, bob.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})
	_, err = thingService.GetThing(context.Background(), privateThing.ID, charlie.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})

	res, err := thingService.GetThing(context.Background(), friendThing.ID, bob.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	_, err = thingService.GetThing(context.Background(), friendThing.ID, charlie.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})

	res, err = thingService.GetThing(context.Background(), friendsOfFriendsThing.ID, bob.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	res, err = thingService.GetThing(context.Background(), friendsOfFriendsThing.ID, charlie.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestDeletion(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)

	t.Cleanup(tearDownFunc)
	is, err := services.NewTmpImageService(db)
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})
	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	anotherUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	anotherUser, err := userService.CreateUser(context.Background(), *anotherUserParams)
	assert.NoError(t, err)

	thingService := services.NewThingService(db, is)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = testUser.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.Nil(t, err, nil)
	assert.NotNil(t, thing)
	assert.NotEmpty(t, thing.ID)
	// TODO add to list, add image, add properties

	err = thingService.DeleteThing(context.Background(), thing.ID, anotherUser.ID)
	assert.ErrorIs(t, err, utils.EntityDoesNotBelongToUserError{})
	err = thingService.DeleteThing(context.Background(), thing.ID, testUser.ID)
	assert.NoError(t, err)
}
