package services_test

import (
	"context"
	"os"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestPublicShareThing(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	thingService := services.NewThingService(db, imageService, notificationService)
	publicShareService := services.NewPublicShareService(db)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	share, err := publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: thing.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, share.ID)
	assert.Equal(t, thing.ID, share.ThingID.String)

	fetched, err := publicShareService.GetPublicShare(context.Background(), share.ID)
	assert.NoError(t, err)
	resource := resources.PublicShareFromModel(fetched)
	assert.Equal(t, "thing", resource.Type)
	assert.NotNil(t, resource.Thing)
	assert.Nil(t, resource.List)
	assert.Equal(t, thing.Name, resource.Thing.Name)
	assert.Equal(t, alice.Name, resource.Thing.OwnerName)
}

func TestPublicShareOnlyOwnerCanCreateAndDelete(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
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
	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	thingService := services.NewThingService(db, imageService, notificationService)
	publicShareService := services.NewPublicShareService(db)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	_, err = publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: thing.ID,
		OwnerId:  mallory.ID,
	})
	assert.ErrorIs(t, err, utils.EntityDoesNotBelongToUserError{})

	share, err := publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: thing.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)

	err = publicShareService.DeletePublicShare(context.Background(), share.ID, mallory.ID)
	assert.ErrorIs(t, err, utils.EntityDoesNotBelongToUserError{})

	err = publicShareService.DeletePublicShare(context.Background(), share.ID, alice.ID)
	assert.NoError(t, err)

	_, err = publicShareService.GetPublicShare(context.Background(), share.ID)
	assert.ErrorIs(t, err, utils.NotFoundError{EntityName: "PublicShare"})
}

func TestPublicShareList(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	thingService := services.NewThingService(db, imageService, notificationService)
	listService := services.NewListService(db, notificationService)
	publicShareService := services.NewPublicShareService(db)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	listParams.ThingIds = []string{thing.ID}
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	share, err := publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: list.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, list.ID, share.ListID.String)

	fetched, err := publicShareService.GetPublicShare(context.Background(), share.ID)
	assert.NoError(t, err)
	resource := resources.PublicShareFromModel(fetched)
	assert.Equal(t, "list", resource.Type)
	assert.NotNil(t, resource.List)
	assert.Nil(t, resource.Thing)
	assert.Len(t, resource.List.Things, 1)
	assert.Equal(t, thing.Name, resource.List.Things[0].Name)
}

func TestPublicShareIndex(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
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
	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	thingService := services.NewThingService(db, imageService, notificationService)
	listService := services.NewListService(db, notificationService)
	publicShareService := services.NewPublicShareService(db)

	aliceThingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	aliceThingParams.OwnerId = alice.ID
	aliceThing, err := thingService.CreateThing(context.Background(), *aliceThingParams)
	assert.NoError(t, err)

	aliceListParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	aliceListParams.OwnerId = alice.ID
	aliceList, err := listService.CreateList(context.Background(), *aliceListParams)
	assert.NoError(t, err)

	bobThingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	bobThingParams.OwnerId = bob.ID
	bobThing, err := thingService.CreateThing(context.Background(), *bobThingParams)
	assert.NoError(t, err)

	_, err = publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: aliceThing.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)
	_, err = publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: aliceList.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)
	_, err = publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: bobThing.ID,
		OwnerId:  bob.ID,
	})
	assert.NoError(t, err)

	aliceShares, err := publicShareService.GetPublicSharesForUser(context.Background(), alice.ID)
	assert.NoError(t, err)
	assert.Len(t, aliceShares, 2)

	entries := resources.PublicShareIndexEntriesFromModelSlice(aliceShares)
	namesByType := map[string]string{}
	for _, entry := range entries {
		namesByType[entry.Type] = entry.ObjectName
	}
	assert.Equal(t, aliceThing.Name, namesByType["thing"])
	assert.Equal(t, aliceList.Name, namesByType["list"])

	bobShares, err := publicShareService.GetPublicSharesForUser(context.Background(), bob.ID)
	assert.NoError(t, err)
	assert.Len(t, bobShares, 1)
}

func TestPublicShareCascadeOnThingDeletion(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	thingService := services.NewThingService(db, imageService, notificationService)
	publicShareService := services.NewPublicShareService(db)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	share, err := publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: thing.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)

	err = thingService.DeleteThing(context.Background(), thing.ID, alice.ID)
	assert.NoError(t, err)

	_, err = publicShareService.GetPublicShare(context.Background(), share.ID)
	assert.ErrorIs(t, err, utils.NotFoundError{EntityName: "PublicShare"}, "deleting the thing revokes the public share")
}

func TestPublicShareImageAccess(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	thingService := services.NewThingService(db, imageService, notificationService)
	listService := services.NewListService(db, notificationService)
	publicShareService := services.NewPublicShareService(db)

	pngFile, err := testcommon.Assets.Open("assets/test.png")
	assert.NoError(t, err)
	sharedImage, err := imageService.CreateImage(context.Background(), alice.ID, "test.png", pngFile)
	assert.NoError(t, err)

	jpgFile, err := testcommon.Assets.Open("assets/test.jpg")
	assert.NoError(t, err)
	unrelatedImage, err := imageService.CreateImage(context.Background(), alice.ID, "test.jpg", jpgFile)
	assert.NoError(t, err)

	sharedThingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	sharedThingParams.OwnerId = alice.ID
	sharedThingParams.ImagesIds = []string{sharedImage.ID}
	sharedThing, err := thingService.CreateThing(context.Background(), *sharedThingParams)
	assert.NoError(t, err)

	unrelatedThingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	unrelatedThingParams.OwnerId = alice.ID
	unrelatedThingParams.ImagesIds = []string{unrelatedImage.ID}
	_, err = thingService.CreateThing(context.Background(), *unrelatedThingParams)
	assert.NoError(t, err)

	thingShare, err := publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: sharedThing.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)

	_, _, err = imageService.ImageGetViaPublicShare(context.Background(), thingShare.ID, sharedImage.Hash)
	assert.NoError(t, err, "image of the shared thing is accessible via the token")

	_, _, err = imageService.ImageGetViaPublicShare(context.Background(), thingShare.ID, unrelatedImage.Hash)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{}, "image of an unshared thing is not accessible")

	_, _, err = imageService.ImageGetViaPublicShare(context.Background(), "bogus-token", sharedImage.Hash)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{}, "bogus token grants no access")

	// list share grants access to images of things in the list
	listParams := factories.ListFactory.MustCreate().(*services.CreateListParams)
	listParams.OwnerId = alice.ID
	listParams.ThingIds = []string{sharedThing.ID}
	list, err := listService.CreateList(context.Background(), *listParams)
	assert.NoError(t, err)

	listShare, err := publicShareService.CreatePublicShare(context.Background(), services.CreatePublicShareParams{
		ObjectId: list.ID,
		OwnerId:  alice.ID,
	})
	assert.NoError(t, err)

	_, _, err = imageService.ImageGetViaPublicShare(context.Background(), listShare.ID, sharedImage.Hash)
	assert.NoError(t, err, "image is accessible via the list share")

	// revoking the share revokes image access
	err = publicShareService.DeletePublicShare(context.Background(), thingShare.ID, alice.ID)
	assert.NoError(t, err)
	_, _, err = imageService.ImageGetViaPublicShare(context.Background(), thingShare.ID, sharedImage.Hash)
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{}, "revoked token grants no access")
}
