package services_test

import (
	"context"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePassword(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	originalPassword := testUserParams.Password
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)
	assert.NotNil(t, testUser)

	newPassword := "newSecurePassword123"
	err = userService.UpdatePassword(context.Background(), services.UpdatePasswordParams{
		UserId:      testUser.ID,
		OldPassword: originalPassword,
		NewPassword: newPassword,
	})
	assert.NoError(t, err)

	authenticatedUser, err := operations.AuthenticateUserByID(context.Background(), db, testUser.ID, newPassword)
	assert.NoError(t, err)
	assert.NotNil(t, authenticatedUser)
	assert.Equal(t, testUser.ID, authenticatedUser.ID)
}

func TestUpdatePasswordWithWrongOldPassword(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	originalPassword := testUserParams.Password
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)
	assert.NotNil(t, testUser)

	err = userService.UpdatePassword(context.Background(), services.UpdatePasswordParams{
		UserId:      testUser.ID,
		OldPassword: "wrongOldPassword",
		NewPassword: "newPassword123",
	})
	assert.Error(t, err, "should fail with wrong old password")

	authenticatedUser, err := operations.AuthenticateUserByID(context.Background(), db, testUser.ID, originalPassword)
	assert.NoError(t, err)
	assert.NotNil(t, authenticatedUser)
}

func TestUpdatePasswordForNonExistentUser(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "")

	err = userService.UpdatePassword(context.Background(), services.UpdatePasswordParams{
		UserId:      "non-existent-user-id",
		OldPassword: "somePassword",
		NewPassword: "newPassword123",
	})
	assert.Error(t, err, "should fail for non-existent user")
}
