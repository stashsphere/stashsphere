package services_test

import (
	"context"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePassword(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
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

	userService := services.NewUserService(db, false, "", 60, nil)
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

	userService := services.NewUserService(db, false, "", 60, nil)

	err = userService.UpdatePassword(context.Background(), services.UpdatePasswordParams{
		UserId:      "non-existent-user-id",
		OldPassword: "somePassword",
		NewPassword: "newPassword123",
	})
	assert.Error(t, err, "should fail for non-existent user")
}

func TestScheduleDeletion(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	userService := services.NewUserService(db, false, "", 60, notificationService)

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)
	assert.NotNil(t, testUser)

	updatedUser, err := userService.ScheduleDeletion(context.Background(), testUser.ID, testUserParams.Password)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.True(t, updatedUser.PurgeAt.Valid, "PurgeAt should be set")

	assert.Len(t, emailService.Mails, 1)
	assert.Equal(t, testUserParams.Email, emailService.Mails[0].To)
	assert.Contains(t, emailService.Mails[0].Subject, "scheduled for deletion")
}

func TestScheduleDeletionWrongPassword(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	userService := services.NewUserService(db, false, "", 60, notificationService)

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)

	_, err = userService.ScheduleDeletion(context.Background(), testUser.ID, "wrongpassword")
	assert.Error(t, err)
	assert.IsType(t, utils.ParameterError{}, err)
	assert.Len(t, emailService.Mails, 0)
}

func TestScheduleDeletionIdempotent(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	userService := services.NewUserService(db, false, "", 60, notificationService)

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)

	first, err := userService.ScheduleDeletion(context.Background(), testUser.ID, testUserParams.Password)
	assert.NoError(t, err)
	assert.True(t, first.PurgeAt.Valid)

	second, err := userService.ScheduleDeletion(context.Background(), testUser.ID, testUserParams.Password)
	assert.NoError(t, err)
	assert.Equal(t, first.PurgeAt, second.PurgeAt, "PurgeAt should not change on second call")
	assert.Len(t, emailService.Mails, 1, "only one email should be sent")
}

func TestCancelDeletionNotScheduled(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)

	_, err = userService.CancelDeletion(context.Background(), testUser.ID)
	assert.Error(t, err)
	assert.IsType(t, utils.NotFoundError{}, err)
}
