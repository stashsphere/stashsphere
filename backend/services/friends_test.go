package services_test

import (
	"context"
	"testing"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestFriendRequestCreationReject(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	friendService := services.NewFriendService(db, notificationService)
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)
	otherUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	otherUser, err := userService.CreateUser(context.Background(), *otherUserParams)
	assert.NoError(t, err)

	request, err := friendService.CreateFriendRequest(context.Background(), services.CreateFriendRequestParams{
		UserId:     testUser.ID,
		ReceiverId: otherUser.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, request)

	result, err := friendService.GetFriendRequests(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Received, 0)
	assert.Len(t, result.Sent, 1)

	result, err = friendService.GetFriendRequests(context.Background(), otherUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Received, 1)
	assert.Len(t, result.Sent, 0)

	secondRequest, err := friendService.CreateFriendRequest(context.Background(), services.CreateFriendRequestParams{
		UserId:     testUser.ID,
		ReceiverId: otherUser.ID,
	})
	assert.Nil(t, secondRequest)
	assert.ErrorIs(t, err, utils.PendingFriendRequestExistsError{})

	// the other user cannot cancel a request that they did not send
	_, err = friendService.CancelFriendRequest(context.Background(), services.CancelFriendRequestParams{
		UserId:    otherUser.ID,
		RequestId: request.ID,
	})
	assert.ErrorIs(t, err, utils.EntityDoesNotBelongToUserError{})

	count, err := models.FriendRequests(models.FriendRequestWhere.ID.EQ(request.ID)).Count(context.Background(), db)
	assert.NoError(t, err)
	assert.Equal(t, count, int64(1))

	_, err = friendService.CancelFriendRequest(context.Background(), services.CancelFriendRequestParams{
		UserId:    testUser.ID,
		RequestId: request.ID,
	})
	assert.NoError(t, err)

	count, err = models.FriendRequests(models.FriendRequestWhere.ID.EQ(request.ID)).Count(context.Background(), db)
	assert.NoError(t, err)
	assert.Zero(t, count)
}

func TestFriendRequestCreationAccept(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	friendService := services.NewFriendService(db, notificationService)
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	testUser, err := userService.CreateUser(context.Background(), *testUserParams)
	assert.NoError(t, err)
	otherUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	otherUser, err := userService.CreateUser(context.Background(), *otherUserParams)
	assert.NoError(t, err)

	request, err := friendService.CreateFriendRequest(context.Background(), services.CreateFriendRequestParams{
		UserId:     testUser.ID,
		ReceiverId: otherUser.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, request)

	friends, err := friendService.GetFriends(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, friends, 0)

	_, err = friendService.ReactFriendRequest(context.Background(), services.ReactFriendRequestParams{
		FriendRequestId: request.ID,
		UserId:          otherUser.ID,
		Accept:          true,
	})
	assert.NoError(t, err)

	friends, err = friendService.GetFriends(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, friends, 1)

	// friend ship exists, cannot be requested again
	secondRequest, err := friendService.CreateFriendRequest(context.Background(), services.CreateFriendRequestParams{
		UserId:     testUser.ID,
		ReceiverId: otherUser.ID,
	})
	assert.ErrorIs(t, err, utils.FriendShipExistsError{})
	assert.Nil(t, secondRequest)
}

func TestUnfriend(t *testing.T) {
	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	t.Cleanup(tearDownFunc)

	userService := services.NewUserService(db, false, "", 60, nil)
	emailService := services.TestEmailService{}
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &emailService)
	friendService := services.NewFriendService(db, notificationService)
	createUser := func() *models.User {
		params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
		user, err := userService.CreateUser(context.Background(), *params)
		assert.NoError(t, err)
		return user
	}
	testUser := createUser()
	firstFriend := createUser()
	secondFriend := createUser()
	stranger := createUser()

	for _, friend := range []*models.User{firstFriend, secondFriend} {
		request, err := friendService.CreateFriendRequest(context.Background(), services.CreateFriendRequestParams{
			UserId:     testUser.ID,
			ReceiverId: friend.ID,
		})
		assert.NoError(t, err)
		_, err = friendService.ReactFriendRequest(context.Background(), services.ReactFriendRequestParams{
			FriendRequestId: request.ID,
			UserId:          friend.ID,
			Accept:          true,
		})
		assert.NoError(t, err)
	}

	// only the friendship with the given friend is removed, whichever of
	// the two sent the request
	err = friendService.Unfriend(context.Background(), secondFriend.ID, testUser.ID)
	assert.NoError(t, err)
	friends, err := friendService.GetFriends(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, friends, 1)
	assert.ElementsMatch(t, []string{friends[0].Friend1ID, friends[0].Friend2ID}, []string{testUser.ID, firstFriend.ID})

	err = friendService.Unfriend(context.Background(), testUser.ID, stranger.ID)
	assert.ErrorIs(t, err, utils.NotFoundError{EntityName: "Friend"})
	err = friendService.Unfriend(context.Background(), testUser.ID, secondFriend.ID)
	assert.ErrorIs(t, err, utils.NotFoundError{EntityName: "Friend"})
	friends, err = friendService.GetFriends(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, friends, 1)

	err = friendService.Unfriend(context.Background(), testUser.ID, firstFriend.ID)
	assert.NoError(t, err)
	friends, err = friendService.GetFriends(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, friends, 0)
}
