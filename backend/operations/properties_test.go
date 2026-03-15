package operations_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func setupPropertyTest(t *testing.T) (*sql.DB, string) {
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

	// Create a user and a thing to attach properties to
	userService := services.NewUserService(db, false, "", 60, notificationService)
	userParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *userParams)
	assert.NoError(t, err)

	thingService := services.NewThingService(db, imageService, notificationService)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = user.ID
	thingParams.Properties = []operations.CreatePropertyParams{}
	thing, err := thingService.CreateThing(context.Background(), *thingParams)
	assert.NoError(t, err)

	return db, thing.ID
}

func TestCreatePropertyString(t *testing.T) {
	db, thingID := setupPropertyTest(t)

	var property *models.Property
	err := utils.Tx(context.Background(), db, func(tx *sql.Tx) error {
		var err error
		property, err = operations.CreateProperty(context.Background(), tx, thingID, operations.CreatePropertyStringParams{
			Name:  "color",
			Value: "red",
		})
		return err
	})

	assert.NoError(t, err)
	assert.NotNil(t, property)
	assert.NotEmpty(t, property.ID)
	assert.Equal(t, thingID, property.ThingID)
	assert.Equal(t, "color", property.Name)
	assert.Equal(t, models.PropertyTypeString, property.Type)
	assert.True(t, property.ValueString.Valid)
	assert.Equal(t, "red", property.ValueString.String)
	assert.False(t, property.ValueFloat.Valid)
	assert.False(t, property.ValueDatetime.Valid)
	assert.False(t, property.ValueBoolean.Valid)
	assert.False(t, property.Unit.Valid)
}

func TestCreatePropertyFloat(t *testing.T) {
	db, thingID := setupPropertyTest(t)

	var property *models.Property
	err := utils.Tx(context.Background(), db, func(tx *sql.Tx) error {
		var err error
		property, err = operations.CreateProperty(context.Background(), tx, thingID, operations.CreatePropertyFloatParams{
			Name:  "weight",
			Value: 42.5,
			Unit:  nil,
		})
		return err
	})

	assert.NoError(t, err)
	assert.NotNil(t, property)
	assert.NotEmpty(t, property.ID)
	assert.Equal(t, thingID, property.ThingID)
	assert.Equal(t, "weight", property.Name)
	assert.Equal(t, models.PropertyTypeFloat, property.Type)
	assert.True(t, property.ValueFloat.Valid)
	assert.Equal(t, 42.5, property.ValueFloat.Float64)
	assert.False(t, property.ValueString.Valid)
	assert.False(t, property.ValueDatetime.Valid)
	assert.False(t, property.ValueBoolean.Valid)
	assert.False(t, property.Unit.Valid)
}

func TestCreatePropertyFloatWithUnit(t *testing.T) {
	db, thingID := setupPropertyTest(t)

	unit := "kg"
	var property *models.Property
	err := utils.Tx(context.Background(), db, func(tx *sql.Tx) error {
		var err error
		property, err = operations.CreateProperty(context.Background(), tx, thingID, operations.CreatePropertyFloatParams{
			Name:  "weight",
			Value: 42.5,
			Unit:  &unit,
		})
		return err
	})

	assert.NoError(t, err)
	assert.NotNil(t, property)
	assert.Equal(t, models.PropertyTypeFloat, property.Type)
	assert.True(t, property.ValueFloat.Valid)
	assert.Equal(t, 42.5, property.ValueFloat.Float64)
	assert.True(t, property.Unit.Valid)
	assert.Equal(t, "kg", property.Unit.String)
}

func TestCreatePropertyDatetime(t *testing.T) {
	db, thingID := setupPropertyTest(t)

	ts := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	var property *models.Property
	err := utils.Tx(context.Background(), db, func(tx *sql.Tx) error {
		var err error
		property, err = operations.CreateProperty(context.Background(), tx, thingID, operations.CreatePropertyDatetimeParams{
			Name:  "purchase_date",
			Value: ts,
		})
		return err
	})

	assert.NoError(t, err)
	assert.NotNil(t, property)
	assert.NotEmpty(t, property.ID)
	assert.Equal(t, thingID, property.ThingID)
	assert.Equal(t, "purchase_date", property.Name)
	assert.Equal(t, models.PropertyTypeDatetime, property.Type)
	assert.True(t, property.ValueDatetime.Valid)
	assert.True(t, ts.Equal(property.ValueDatetime.Time))
	assert.False(t, property.ValueString.Valid)
	assert.False(t, property.ValueFloat.Valid)
	assert.False(t, property.ValueBoolean.Valid)
}

func TestCreatePropertyBoolean(t *testing.T) {
	db, thingID := setupPropertyTest(t)

	var property *models.Property
	err := utils.Tx(context.Background(), db, func(tx *sql.Tx) error {
		var err error
		property, err = operations.CreateProperty(context.Background(), tx, thingID, operations.CreatePropertyBooleanParams{
			Name:  "in_stock",
			Value: true,
		})
		return err
	})

	assert.NoError(t, err)
	assert.NotNil(t, property)
	assert.NotEmpty(t, property.ID)
	assert.Equal(t, thingID, property.ThingID)
	assert.Equal(t, "in_stock", property.Name)
	assert.Equal(t, models.PropertyTypeBoolean, property.Type)
	assert.True(t, property.ValueBoolean.Valid)
	assert.True(t, property.ValueBoolean.Bool)
	assert.False(t, property.ValueString.Valid)
	assert.False(t, property.ValueFloat.Valid)
	assert.False(t, property.ValueDatetime.Valid)
}

func TestCreatePropertyBooleanFalse(t *testing.T) {
	db, thingID := setupPropertyTest(t)

	var property *models.Property
	err := utils.Tx(context.Background(), db, func(tx *sql.Tx) error {
		var err error
		property, err = operations.CreateProperty(context.Background(), tx, thingID, operations.CreatePropertyBooleanParams{
			Name:  "discontinued",
			Value: false,
		})
		return err
	})

	assert.NoError(t, err)
	assert.NotNil(t, property)
	assert.Equal(t, models.PropertyTypeBoolean, property.Type)
	assert.True(t, property.ValueBoolean.Valid)
	assert.False(t, property.ValueBoolean.Bool)
}
