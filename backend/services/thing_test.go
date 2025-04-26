package services_test

import (
	"context"
	"os"
	"testing"

	"github.com/stashsphere/backend/factories"
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
	assert.Nil(t, err, nil)
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
	shareService := services.NewShareService(db)
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
