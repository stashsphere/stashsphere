package services_test

import (
	"context"
	"os"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestListCreation(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	listService := services.NewListService(db, notificationService)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = testUser.ID
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.NotEmpty(t, list.ID)
}

func TestListAccess(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	malloryParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	mallory, err := userService.CreateUser(context.Background(), *malloryParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	listService := services.NewListService(db, notificationService)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	_, err = listService.GetList(context.Background(), list.ID, mallory.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})
}

func TestListAccessShareList(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	shareService := services.NewShareService(db, notificationService)
	listService := services.NewListService(db, notificationService)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	_, err = listService.GetList(context.Background(), list.ID, bob.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})

	share, err := shareService.CreateListShare(context.Background(), services.CreateListShareParams{
		ListId:       list.ID,
		OwnerId:      alice.ID,
		TargetUserId: bob.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, share)

	_, err = listService.GetList(context.Background(), list.ID, bob.ID)
	assert.NoError(t, err, "bob has access through list share")
}

func TestListSharingState(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	listService := services.NewListService(db, notificationService)

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

	privateListParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	privateListParams.OwnerId = alice.ID
	privateList, err := listService.CreateList(context.Background(), *privateListParams)
	assert.NoError(t, err)

	friendListParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	friendListParams.OwnerId = alice.ID
	friendListParams.SharingState = models.SharingStateFriends.String()
	friendList, err := listService.CreateList(context.Background(), *friendListParams)
	assert.NoError(t, err)

	friendsOfFriendsListParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	friendsOfFriendsListParams.OwnerId = alice.ID
	friendsOfFriendsListParams.SharingState = models.SharingStateFriendsOfFriends.String()
	friendsOfFriendsList, err := listService.CreateList(context.Background(), *friendsOfFriendsListParams)
	assert.NoError(t, err)

	_, err = listService.GetList(context.Background(), privateList.ID, bob.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})
	_, err = listService.GetList(context.Background(), privateList.ID, charlie.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})

	res, err := listService.GetList(context.Background(), friendList.ID, bob.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	_, err = listService.GetList(context.Background(), friendList.ID, charlie.ID)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})

	res, err = listService.GetList(context.Background(), friendsOfFriendsList.ID, bob.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	res, err = listService.GetList(context.Background(), friendsOfFriendsList.ID, charlie.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestListSharedWithFriendsNotification(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	listService := services.NewListService(db, notificationService)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	// bob is a friend of alice
	createFriendShip(t, db, alice.ID, bob.ID)

	// alice creates a list shared with friends
	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	listParams.SharingState = models.SharingStateFriends.String()
	_, err = listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	// bob should receive an email notification
	assert.Len(t, emailService.Mails, 1, "bob should receive a notification email")
	assert.Equal(t, bobParams.Email, emailService.Mails[0].To)
}

func TestListUpdateToSharedNotification(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	listService := services.NewListService(db, notificationService)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	// bob is a friend of alice
	createFriendShip(t, db, alice.ID, bob.ID)

	// alice creates a private list
	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	listParams.SharingState = "private"
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	// no emails should be sent for a private list
	assert.Len(t, emailService.Mails, 0, "no notification for private list")

	// alice updates the list to share with friends
	_, err = listService.UpdateList(context.Background(), list.ID, alice.ID, services.UpdateListParams{
		Name:         list.Name,
		ThingIds:     []string{},
		SharingState: models.SharingStateFriends.String(),
	})
	assert.NoError(t, err)

	// bob should now receive an email notification
	assert.Len(t, emailService.Mails, 1, "bob should receive a notification email when list is shared")
	assert.Equal(t, bobParams.Email, emailService.Mails[0].To)
}

func TestListDeletion(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	anotherUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)
	anotherUser, err := userService.CreateUser(context.Background(), *anotherUserParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	listService := services.NewListService(db, notificationService)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = testUser.ID
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.NotEmpty(t, list.ID)

	err = listService.DeleteList(context.Background(), list.ID, anotherUser.ID)
	assert.ErrorIs(t, err, utils.EntityDoesNotBelongToUserError{})
	err = listService.DeleteList(context.Background(), list.ID, testUser.ID)
	assert.NoError(t, err)
}

func TestListCreationWithThings(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
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

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	thingService := services.NewThingService(db, is, notificationService)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = testUser.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	listService := services.NewListService(db, notificationService)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = testUser.ID
	listParams.ThingIds = []string{thing.ID}
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.NotEmpty(t, list.ID)
	assert.Len(t, list.R.Things, 1)
	assert.Equal(t, thing.ID, list.R.Things[0].ID)
}

func TestListUpdateWithThings(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
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

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	thingService := services.NewThingService(db, is, notificationService)
	thing1Params := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thing1Params.OwnerId = testUser.ID
	thing1, err := thingService.CreateThing(context.Background(), *thing1Params)
	assert.NoError(t, err)

	thing2Params := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thing2Params.OwnerId = testUser.ID
	thing2, err := thingService.CreateThing(context.Background(), *thing2Params)
	assert.NoError(t, err)

	listService := services.NewListService(db, notificationService)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = testUser.ID
	listParams.ThingIds = []string{thing1.ID}
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)
	assert.Len(t, list.R.Things, 1)

	updatedList, err := listService.UpdateList(context.Background(), list.ID, testUser.ID, services.UpdateListParams{
		Name:         "Updated List",
		ThingIds:     []string{thing1.ID, thing2.ID},
		SharingState: "private",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Updated List", updatedList.Name)
	assert.Len(t, updatedList.R.Things, 2)
}

func TestListCannotAddOthersThings(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
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

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	thingService := services.NewThingService(db, is, notificationService)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = bob.ID
	bobsThing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	listService := services.NewListService(db, notificationService)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	listParams.ThingIds = []string{bobsThing.ID}
	_, err = listService.CreateList(context.Background(), *listParams)
	assert.ErrorIs(t, err, utils.EntityDoesNotBelongToUserError{})
}

func TestThingsAddedToSharedListNotification(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	is, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})

	userService := services.NewUserService(db, false, "")
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	thingService := services.NewThingService(db, is, notificationService)
	listService := services.NewListService(db, notificationService)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	// bob is a friend of alice
	createFriendShip(t, db, alice.ID, bob.ID)

	// alice creates a thing
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing1, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	// alice creates a shared list with thing1
	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	listParams.ThingIds = []string{thing1.ID}
	listParams.SharingState = models.SharingStateFriends.String()
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	// bob should receive a ListShared notification
	assert.Len(t, emailService.Mails, 1, "bob should receive a ListShared notification")
	assert.Equal(t, bobParams.Email, emailService.Mails[0].To)
	assert.Contains(t, emailService.Mails[0].Subject, "A list has been shared with you")
	emailService.Clear()

	// alice creates another thing and adds it to the list
	thing2Params := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thing2Params.OwnerId = alice.ID
	thing2, err := thingService.CreateThing(context.Background(), *thing2Params)
	assert.NoError(t, err)

	// alice updates the list to add thing2
	_, err = listService.UpdateList(context.Background(), list.ID, alice.ID, services.UpdateListParams{
		Name:         list.Name,
		ThingIds:     []string{thing1.ID, thing2.ID},
		SharingState: models.SharingStateFriends.String(),
	})
	assert.NoError(t, err)

	// bob should receive a ThingsAddedToList notification (not ListShared)
	assert.Len(t, emailService.Mails, 1, "bob should receive a ThingsAddedToList notification")
	assert.Equal(t, bobParams.Email, emailService.Mails[0].To)
	assert.Contains(t, emailService.Mails[0].Subject, "added things to a list")
}

func TestRemoveThingFromListRemovesFromCart(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	is, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})

	userService := services.NewUserService(db, false, "")
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	thingService := services.NewThingService(db, is, notificationService)
	listService := services.NewListService(db, notificationService)
	cartService := services.NewCartService(db)

	// Create alice and bob
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	// bob and alice are friends
	createFriendShip(t, db, alice.ID, bob.ID)

	// bob creates a thing
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = bob.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	// bob creates a list with the thing, shared with friends
	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = bob.ID
	listParams.ThingIds = []string{thing.ID}
	listParams.SharingState = models.SharingStateFriends.String()
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)
	assert.Len(t, list.R.Things, 1)

	// alice can access the thing through the list
	_, err = thingService.GetThing(context.Background(), thing.ID, alice.ID)
	assert.NoError(t, err, "alice should have access to thing through friend's list")

	// alice adds the thing to her cart
	cartEntries, err := cartService.UpdateCart(context.Background(), services.UpdateCartParams{
		UserId:   alice.ID,
		ThingIds: []string{thing.ID},
	})
	assert.NoError(t, err)
	assert.Len(t, cartEntries, 1, "alice's cart should contain the thing")

	// bob removes the thing from the list
	_, err = listService.UpdateList(context.Background(), list.ID, bob.ID, services.UpdateListParams{
		Name:         list.Name,
		ThingIds:     []string{},
		SharingState: models.SharingStateFriends.String(),
	})
	assert.NoError(t, err)

	// alice's cart should now be empty
	aliceCart, err := cartService.GetCart(context.Background(), alice.ID)
	assert.NoError(t, err)
	assert.Len(t, aliceCart, 0, "alice's cart should be empty after thing removed from list")
}
