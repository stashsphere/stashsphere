package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type FriendHandler struct {
	friend_service *services.FriendService
}

func NewFriendHandler(friend_service *services.FriendService) *FriendHandler {
	return &FriendHandler{
		friend_service: friend_service,
	}
}

type NewFriendRequestParams struct {
	ReceiverId string `json:"receiverId"`
}

func (fh *FriendHandler) FriendRequestPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.ErrNoAuthContext
	}
	if !authCtx.Authenticated {
		return utils.ErrNotAuthenticated
	}
	friendRequestParams := NewFriendRequestParams{}
	if err := c.Bind(&friendRequestParams); err != nil {
		return &utils.ErrParameterError{Err: err}
	}
	request, err := fh.friend_service.CreateFriendRequest(c.Request().Context(), services.CreateFriendRequestParams{
		UserId:     authCtx.User.ID,
		ReceiverId: friendRequestParams.ReceiverId,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resources.FriendRequestFromModel(request, authCtx.User.ID))
}

func (fh *FriendHandler) FriendRequestIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.ErrNoAuthContext
	}
	if !authCtx.Authenticated {
		return utils.ErrNotAuthenticated
	}
	requests, err := fh.friend_service.GetFriendRequests(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.FriendRequestsResponseFromResult(requests, authCtx.User.ID))
}

func (fh *FriendHandler) FriendRequestDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.ErrNoAuthContext
	}
	if !authCtx.Authenticated {
		return utils.ErrNotAuthenticated
	}
	requestId := c.Param("requestId")
	_, err := fh.friend_service.CancelFriendRequest(c.Request().Context(), services.CancelFriendRequestParams{
		UserId:    authCtx.User.ID,
		RequestId: requestId,
	})
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

type UpdateFriendRequestParams struct {
	Accept bool `json:"accept"`
}

func (fh *FriendHandler) FriendRequestUpdate(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.ErrNoAuthContext
	}
	if !authCtx.Authenticated {
		return utils.ErrNotAuthenticated
	}
	friendRequestParams := UpdateFriendRequestParams{}
	if err := c.Bind(&friendRequestParams); err != nil {
		return &utils.ErrParameterError{Err: err}
	}
	requestId := c.Param("requestId")
	request, err := fh.friend_service.ReactFriendRequest(c.Request().Context(), services.ReactFriendRequestParams{
		FriendRequestId: requestId,
		UserId:          authCtx.User.ID,
		Accept:          friendRequestParams.Accept,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.FriendRequestFromModel(request, authCtx.User.ID))
}

func (fh *FriendHandler) FriendsIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.ErrNoAuthContext
	}
	if !authCtx.Authenticated {
		return utils.ErrNotAuthenticated
	}
	friends, err := fh.friend_service.GetFriends(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.FriendShipsResponseFromModel(friends, authCtx.User.ID))
}

func (fh *FriendHandler) FriendDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.ErrNoAuthContext
	}
	if !authCtx.Authenticated {
		return utils.ErrNotAuthenticated
	}
	friendId := c.Param("friendId")
	err := fh.friend_service.Unfriend(c.Request().Context(), authCtx.User.ID, friendId)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
