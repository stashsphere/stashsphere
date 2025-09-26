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

func TestCartCreation(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)

	t.Cleanup(tearDownFunc)

	is, err := services.NewTmpImageService(db)
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})
	userService := services.NewUserService(db, false, "")
	thingService := services.NewThingService(db, is)
	cartService := services.NewCartService(db)
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, emailService)
	shareService := services.NewShareService(db, notificationService)

	aliceParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	alice, err := userService.CreateUser(context.Background(), *aliceParams)
	assert.NoError(t, err)

	bobParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	bob, err := userService.CreateUser(context.Background(), *bobParams)
	assert.NoError(t, err)

	aliceThings := []string{}
	for range 3 {
		thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
		thingParams.OwnerId = alice.ID
		thing, err := thingService.CreateThing(context.Background(), *thingParams)
		assert.NoError(t, err)
		aliceThings = append(aliceThings, thing.ID)
	}

	bobThings := []string{}
	shares := []*models.Share{}
	for range 2 {
		thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
		thingParams.OwnerId = bob.ID
		thing, err := thingService.CreateThing(context.Background(), *thingParams)
		assert.NoError(t, err)
		bobThings = append(bobThings, thing.ID)
		// share thing to bob
		share, err := shareService.CreateThingShare(context.Background(), services.CreateThingShareParams{
			ThingId:      thing.ID,
			OwnerId:      bob.ID,
			TargetUserId: alice.ID,
		})
		assert.Nil(t, err)
		shares = append(shares, share)
	}

	// private bob item, not shared
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = bob.ID
	privateThing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	entries, err := cartService.UpdateCart(context.Background(), services.UpdateCartParams{
		UserId:   alice.ID,
		ThingIds: aliceThings,
	})
	assert.NoError(t, err)
	assert.Len(t, entries, len(aliceThings))

	for _, entry := range entries {
		assert.Contains(t, aliceThings, entry.ThingID)
	}

	combinedThings := []string{}
	combinedThings = append(combinedThings, aliceThings...)
	combinedThings = append(combinedThings, bobThings...)

	entries, err = cartService.UpdateCart(context.Background(), services.UpdateCartParams{
		UserId:   alice.ID,
		ThingIds: combinedThings,
	})
	assert.NoError(t, err)
	assert.Len(t, entries, len(combinedThings))

	for _, entry := range entries {
		assert.Contains(t, combinedThings, entry.ThingID)
	}

	combinedThingsWithPrivate := []string{}
	combinedThingsWithPrivate = append(combinedThings, privateThing.ID)
	entries, err = cartService.UpdateCart(context.Background(), services.UpdateCartParams{
		UserId:   alice.ID,
		ThingIds: combinedThingsWithPrivate,
	})
	assert.ErrorIs(t, err, utils.UserHasNoAccessRightsError{})
	for _, entry := range entries {
		assert.NotEqual(t, entry.ThingID, privateThing.ID)
	}

	err = shareService.DeleteShare(context.Background(), shares[0].ID, bob.ID)
	assert.NoError(t, err)

	entries, err = cartService.GetCart(context.Background(), alice.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, len(combinedThings)-1)
}
