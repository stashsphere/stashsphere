package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stashsphere/backend/cmd"
	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stretchr/testify/assert"
)

func TestFriendRequestToFriendConflicts(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcBaseConfig(t), false, false, "")
	assert.NoError(t, err)

	ctx := context.Background()
	userService := services.NewUserService(db, false, "", 60, nil)
	notificationService := services.NewNotificationService(db, services.NotificationData{
		FrontendUrl:  "https://example.com",
		InstanceName: "StashsphereTest",
	}, &services.TestEmailService{})
	friendService := services.NewFriendService(db, notificationService)
	senderParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	sender, err := userService.CreateUser(ctx, *senderParams)
	assert.NoError(t, err)
	receiverParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	receiver, err := userService.CreateUser(ctx, *receiverParams)
	assert.NoError(t, err)

	request, err := friendService.CreateFriendRequest(ctx, services.CreateFriendRequestParams{
		UserId:     sender.ID,
		ReceiverId: receiver.ID,
	})
	assert.NoError(t, err)
	_, err = friendService.ReactFriendRequest(ctx, services.ReactFriendRequestParams{
		FriendRequestId: request.ID,
		UserId:          receiver.ID,
		Accept:          true,
	})
	assert.NoError(t, err)

	login := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"email":"`+senderParams.Email+`","password":"`+senderParams.Password+`"}`))
	login.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, login)
	assert.Equal(t, http.StatusOK, loginRec.Code)

	req := httptest.NewRequest(http.MethodPost, "/api/friend_requests",
		strings.NewReader(`{"receiverId":"`+receiver.ID+`"}`))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range loginRec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var body struct {
		Message string `json:"message"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Friendship already exists", body.Message)
}
