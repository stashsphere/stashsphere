package services_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestImageCreation(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tearDownFunc)
	imageService, err := services.NewTmpImageService(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	if err != nil {
		t.Fatal(err)
	}
	pngFile, err := testcommon.Assets.Open("assets/test.png")
	if err != nil {
		t.Fatal(err)
	}
	pngImage, err := imageService.CreateImage(context.Background(), testUser.ID, "test.png", pngFile)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, pngImage.Mime, "image/png", "expected mime type to be png")
	jpgFile, err := testcommon.Assets.Open("assets/test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	jpgImage, err := imageService.CreateImage(context.Background(), testUser.ID, "test.jpg", jpgFile)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, jpgImage.Mime, "image/jpeg", "expected mime type to be jpg")
}

func TestImageAccess(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	if err != nil {
		t.Fatal(err)
	}
	malloryParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	mallory, err := userService.CreateUser(context.Background(), *malloryParams)
	if err != nil {
		t.Fatal(err)
	}
	imageService, err := services.NewTmpImageService(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	pngFile, err := testcommon.Assets.Open("assets/test.png")
	if err != nil {
		t.Fatal(err)
	}
	pngImage, err := imageService.CreateImage(context.Background(), alice.ID, "test.png", pngFile)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = imageService.ImageGet(context.Background(), alice.ID, pngImage.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = imageService.ImageGet(context.Background(), mallory.ID, pngImage.ID)
	assert.ErrorIs(t, err, utils.ErrUserHasNoAccessRights)
}

func TestImageAccessSharedThing(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tearDownFunc)
	userService := services.NewUserService(db, false, "")
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	if err != nil {
		t.Fatal(err)
	}
	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	if err != nil {
		t.Fatal(err)
	}
	imageService, err := services.NewTmpImageService(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	pngFile, err := testcommon.Assets.Open("assets/test.png")
	if err != nil {
		t.Fatal(err)
	}
	pngImage, err := imageService.CreateImage(context.Background(), alice.ID, "test.png", pngFile)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = imageService.ImageGet(context.Background(), bob.ID, pngImage.ID)
	assert.ErrorIs(t, err, utils.ErrUserHasNoAccessRights, "bob does not have access yet")

	thingService := services.NewThingService(db, imageService)
	shareService := services.NewShareService(db)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thingParams.ImagesIds = []string{pngImage.ID}
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.Nil(t, err)
	share, err := shareService.CreateThingShare(context.Background(), services.CreateThingShareParams{
		ThingId:      thing.ID,
		OwnerId:      alice.ID,
		TargetUserId: bob.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, share)
	_, _, err = imageService.ImageGet(context.Background(), bob.ID, pngImage.ID)
	assert.Nil(t, err, "bob has access through thing share")
}

func TestDeleteImage(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tearDownFunc)
	userService := services.NewUserService(db, false, "")
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	if err != nil {
		t.Fatal(err)
	}
	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	if err != nil {
		t.Fatal(err)
	}
	imageService, err := services.NewTmpImageService(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	pngFile, err := testcommon.Assets.Open("assets/test.png")
	if err != nil {
		t.Fatal(err)
	}
	pngImage, err := imageService.CreateImage(context.Background(), alice.ID, "test.png", pngFile)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(imageService.StorePath(), pngImage.Hash)
	assert.FileExists(t, path)
	// bob should not be able to delete the image
	_, err = imageService.DeleteImage(context.Background(), bob.ID, pngImage.ID)
	assert.ErrorIs(t, err, utils.ErrEntityDoesNotBelongToUser)
	// alice should be able to delete the image
	deletedImage, err := imageService.DeleteImage(context.Background(), alice.ID, pngImage.ID)
	assert.NoError(t, err)
	assert.NotNil(t, deletedImage)

	path = filepath.Join(imageService.StorePath(), pngImage.Hash)
	assert.NoFileExists(t, path)
}

func TestDeleteImageInUse(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)

	t.Cleanup(tearDownFunc)
	userService := services.NewUserService(db, false, "")
	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	imageService, err := services.NewTmpImageService(db)
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(imageService.StorePath())
	})
	thingService := services.NewThingService(db, imageService)
	pngFile, err := testcommon.Assets.Open("assets/test.png")
	assert.NoError(t, err)

	pngImage, err := imageService.CreateImage(context.Background(), alice.ID, "test.png", pngFile)
	assert.NoError(t, err)

	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = alice.ID
	thingParams.ImagesIds = []string{pngImage.ID}
	_, err = thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	path := filepath.Join(imageService.StorePath(), pngImage.Hash)
	assert.FileExists(t, path)

	deletedImage, err := imageService.DeleteImage(context.Background(), alice.ID, pngImage.ID)
	assert.ErrorIs(t, err, utils.ErrEntityInUse)
	assert.Nil(t, deletedImage)

	assert.FileExists(t, path)
}
