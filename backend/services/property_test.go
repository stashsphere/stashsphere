package services_test

import (
	"context"
	"database/sql"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stretchr/testify/assert"
)

type testEnv struct {
	db              *sql.DB
	imageService    *services.ImageService
	propertyService *services.PropertyService
	ctx             context.Context
}

func setupTestEnv(t *testing.T) *testEnv {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDownFunc)

	is, err := services.NewTmpImageService(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(is.StorePath())
	})

	propertyService := services.NewPropertyService(db)

	return &testEnv{
		db:              db,
		imageService:    is,
		propertyService: propertyService,
		ctx:             context.Background(),
	}
}

func createTestUser(t *testing.T, ctx context.Context, db *sql.DB) *models.User {
	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(ctx, *testUserParams)
	assert.NoError(t, err)
	return testUser
}

func createTestThing(t *testing.T, ctx context.Context, db *sql.DB, is *services.ImageService, ownerId string) *models.Thing {
	thingService := services.NewThingService(db, is)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = ownerId
	thingParams.Properties = []operations.CreatePropertyParams{}
	thing, err := thingService.CreateThing(ctx, *thingParams)
	assert.NoError(t, err)
	return thing
}

func createThingWithProperties(t *testing.T, ctx context.Context, db *sql.DB, is *services.ImageService, ownerId string, properties []operations.CreatePropertyParams) *models.Thing {
	thingService := services.NewThingService(db, is)
	thingParams := factories.ThingFactory.MustCreate().(*services.CreateThingParams)
	thingParams.OwnerId = ownerId
	thingParams.Properties = properties
	thing, err := thingService.CreateThing(ctx, *thingParams)
	assert.NoError(t, err)
	return thing
}

func createDirectShare(t *testing.T, ctx context.Context, db *sql.DB, thingId, ownerId, targetUserId string) *models.Share {
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, emailService)
	shareService := services.NewShareService(db, notificationService)

	share, err := shareService.CreateThingShare(ctx, services.CreateThingShareParams{
		ThingId:      thingId,
		OwnerId:      ownerId,
		TargetUserId: targetUserId,
	})
	assert.NoError(t, err)
	return share
}

func assertContainsString(t *testing.T, slice []string, item string, msgAndArgs ...interface{}) {
	for _, s := range slice {
		if s == item {
			return
		}
	}
	assert.Fail(t, "slice does not contain item", msgAndArgs...)
}

func assertNotContainsString(t *testing.T, slice []string, item string, msgAndArgs ...interface{}) {
	for _, s := range slice {
		if s == item {
			assert.Fail(t, "slice contains item but should not", msgAndArgs...)
			return
		}
	}
}

func assertStringSliceEqual(t *testing.T, expected, actual []string, msgAndArgs ...interface{}) {
	sortedExpected := make([]string, len(expected))
	sortedActual := make([]string, len(actual))
	copy(sortedExpected, expected)
	copy(sortedActual, actual)
	sort.Strings(sortedExpected)
	sort.Strings(sortedActual)
	assert.Equal(t, sortedExpected, sortedActual, msgAndArgs...)
}

func TestPropertyAutoComplete_NameCompletion_WithMatches(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
		operations.CreatePropertyStringParams{Name: "Condition", Value: "Good"},
		operations.CreatePropertyStringParams{Name: "Material", Value: "Wood"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Co",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 2)
	assertContainsString(t, result.Values, "Color")
	assertContainsString(t, result.Values, "Condition")
	assertNotContainsString(t, result.Values, "Material")
}

func TestPropertyAutoComplete_NameCompletion_NoMatches(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
		operations.CreatePropertyStringParams{Name: "Size", Value: "Large"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Weight",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 0)
}

func TestPropertyAutoComplete_ValueCompletion_StringType_WithMatches(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties1 := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties1)

	properties2 := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Royal Blue"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties2)

	valuePrefix := "R"
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Color",
		Value:  &valuePrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result.CompletionType)
	assert.Len(t, result.Values, 2)
	assertContainsString(t, result.Values, "Red")
	assertContainsString(t, result.Values, "Royal Blue")
}

func TestPropertyAutoComplete_ValueCompletion_StringType_NoMatches(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Blue"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties)

	valuePrefix := "Red"
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Color",
		Value:  &valuePrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result.CompletionType)
	assert.Len(t, result.Values, 0)
}

// Test Cases - Access Control

func TestPropertyAutoComplete_OnlyUserOwnedThings(t *testing.T) {
	env := setupTestEnv(t)
	alice := createTestUser(t, env.ctx, env.db)
	bob := createTestUser(t, env.ctx, env.db)

	aliceProperties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, alice.ID, aliceProperties)

	bobProperties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Blue"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, bob.ID, bobProperties)

	valuePrefix := ""
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: alice.ID,
		Name:   "Color",
		Value:  &valuePrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result.CompletionType)
	assert.Len(t, result.Values, 1)
	assertContainsString(t, result.Values, "Red")
	assertNotContainsString(t, result.Values, "Blue")
}

func TestPropertyAutoComplete_CannotSeeOtherUsersPrivateThings(t *testing.T) {
	env := setupTestEnv(t)
	alice := createTestUser(t, env.ctx, env.db)
	bob := createTestUser(t, env.ctx, env.db)

	aliceProperties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Material", Value: "Wood"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, alice.ID, aliceProperties)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: bob.ID,
		Name:   "Material",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 0)
}

func TestPropertyAutoComplete_DirectThingShare(t *testing.T) {
	env := setupTestEnv(t)
	alice := createTestUser(t, env.ctx, env.db)
	bob := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Brand", Value: "Nike"},
	}
	thing := createThingWithProperties(t, env.ctx, env.db, env.imageService, alice.ID, properties)

	createDirectShare(t, env.ctx, env.db, thing.ID, alice.ID, bob.ID)

	valuePrefix := ""
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: bob.ID,
		Name:   "Brand",
		Value:  &valuePrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result.CompletionType)
	assert.Len(t, result.Values, 1)
	assertContainsString(t, result.Values, "Nike")
}

func TestPropertyAutoComplete_SharedAndOwnedThingsCombined(t *testing.T) {
	env := setupTestEnv(t)
	alice := createTestUser(t, env.ctx, env.db)
	bob := createTestUser(t, env.ctx, env.db)

	aliceProperties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Type", Value: "Book"},
	}
	aliceThing := createThingWithProperties(t, env.ctx, env.db, env.imageService, alice.ID, aliceProperties)

	bobProperties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Type", Value: "Magazine"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, bob.ID, bobProperties)

	createDirectShare(t, env.ctx, env.db, aliceThing.ID, alice.ID, bob.ID)

	valuePrefix := ""
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: bob.ID,
		Name:   "Type",
		Value:  &valuePrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result.CompletionType)
	assert.Len(t, result.Values, 2)
	assertContainsString(t, result.Values, "Book")
	assertContainsString(t, result.Values, "Magazine")
}

func TestPropertyAutoComplete_NoAccessAfterShareRemoved(t *testing.T) {
	env := setupTestEnv(t)
	alice := createTestUser(t, env.ctx, env.db)
	bob := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "SecretData", Value: "ConfidentialValue"},
	}
	thing := createThingWithProperties(t, env.ctx, env.db, env.imageService, alice.ID, properties)

	share := createDirectShare(t, env.ctx, env.db, thing.ID, alice.ID, bob.ID)

	valuePrefix := ""
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: bob.ID,
		Name:   "SecretData",
		Value:  &valuePrefix,
	})
	assert.NoError(t, err)
	assert.Len(t, result.Values, 1)

	// Delete the share
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(env.db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, emailService)
	shareService := services.NewShareService(env.db, notificationService)
	err = shareService.DeleteShare(env.ctx, share.ID, alice.ID)
	assert.NoError(t, err)

	result, err = env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: bob.ID,
		Name:   "SecretData",
		Value:  &valuePrefix,
	})
	assert.NoError(t, err)
	assert.Len(t, result.Values, 0)
}

func TestPropertyAutoComplete_DeduplicatesValues(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties1 := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties1)

	properties2 := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties2)

	properties3 := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties3)

	valuePrefix := "R"
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Color",
		Value:  &valuePrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result.CompletionType)
	assert.Len(t, result.Values, 1)
	assertContainsString(t, result.Values, "Red")
}

func TestPropertyAutoComplete_DeduplicatesNames(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	unit := "kg"
	properties1 := []operations.CreatePropertyParams{
		operations.CreatePropertyFloatParams{Name: "Weight", Value: 10.5, Unit: &unit},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties1)

	properties2 := []operations.CreatePropertyParams{
		operations.CreatePropertyFloatParams{Name: "Weight", Value: 20.3, Unit: &unit},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties2)

	properties3 := []operations.CreatePropertyParams{
		operations.CreatePropertyFloatParams{Name: "Weight", Value: 5.0, Unit: &unit},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties3)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "W",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 1)
	assertContainsString(t, result.Values, "Weight")
}

func TestPropertyAutoComplete_EmptyNameQuery(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
		operations.CreatePropertyStringParams{Name: "Size", Value: "Large"},
		operations.CreatePropertyFloatParams{Name: "Weight", Value: 10.0, Unit: nil},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 3)
	assertContainsString(t, result.Values, "Color")
	assertContainsString(t, result.Values, "Size")
	assertContainsString(t, result.Values, "Weight")
}

func TestPropertyAutoComplete_EmptyValueQuery(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties1 := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Blue"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties1)

	properties2 := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Color", Value: "Red"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties2)

	valuePrefix := ""
	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Color",
		Value:  &valuePrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, "value", result.CompletionType)
	assert.Len(t, result.Values, 2)
	assertContainsString(t, result.Values, "Blue")
	assertContainsString(t, result.Values, "Red")
}

func TestPropertyAutoComplete_CaseSensitiveMatching(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "color", Value: "Red"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Color",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 0, "LIKE operator is case-sensitive")
}

func TestPropertyAutoComplete_SpecialCharactersInName(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyStringParams{Name: "Weight_%_test", Value: "100"},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "Weight_",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 1)
	assertContainsString(t, result.Values, "Weight_%_test")
}

func TestPropertyAutoComplete_MultiplePropertyTypes(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	unit := "kg"
	properties := []operations.CreatePropertyParams{
		operations.CreatePropertyFloatParams{Name: "Weight", Value: 10.5, Unit: &unit},
		operations.CreatePropertyStringParams{Name: "Color", Value: "Blue"},
		operations.CreatePropertyDatetimeParams{Name: "Date", Value: time.Now()},
	}
	createThingWithProperties(t, env.ctx, env.db, env.imageService, user.ID, properties)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 3)
	assertContainsString(t, result.Values, "Weight")
	assertContainsString(t, result.Values, "Color")
	assertContainsString(t, result.Values, "Date")
}

func TestPropertyAutoComplete_NoThingsExist(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "AnyProperty",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 0)
}

func TestPropertyAutoComplete_NoPropertiesOnThings(t *testing.T) {
	env := setupTestEnv(t)
	user := createTestUser(t, env.ctx, env.db)

	createTestThing(t, env.ctx, env.db, env.imageService, user.ID)

	result, err := env.propertyService.AutoComplete(env.ctx, services.PropertyAutoCompleteParams{
		UserId: user.ID,
		Name:   "AnyProperty",
		Value:  nil,
	})

	assert.NoError(t, err)
	assert.Equal(t, "name", result.CompletionType)
	assert.Len(t, result.Values, 0)
}
