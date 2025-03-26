package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
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
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	friendRequestParams := NewFriendRequestParams{}
	if err := c.Bind(&friendRequestParams); err != nil {
		c.Logger().Errorf("Bind error: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	request, err := fh.friend_service.CreateFriendRequest(c.Request().Context(), services.CreateFriendRequestParams{
		UserId:     authCtx.User.ID,
		ReceiverId: friendRequestParams.ReceiverId,
	})
	if err != nil {
		c.Logger().Errorf("Could not create friend request: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	return c.JSON(http.StatusCreated, resources.FriendRequestFromModel(request, authCtx.User.ID))
}

func (fh *FriendHandler) FriendRequestIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	requests, err := fh.friend_service.GetFriendRequests(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		c.Logger().Errorf("Could not fetch friend request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.FriendRequestsResponseFromResult(requests, authCtx.User.ID))
}

func (fh *FriendHandler) FriendRequestDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	requestId := c.Param("requestId")
	_, err := fh.friend_service.CancelFriendRequest(c.Request().Context(), services.CancelFriendRequestParams{
		UserId:    authCtx.User.ID,
		RequestId: requestId,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

type UpdateFriendRequestParams struct {
	Accept bool `json:"accept"`
}

func (fh *FriendHandler) FriendRequestUpdate(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	friendRequestParams := UpdateFriendRequestParams{}
	if err := c.Bind(&friendRequestParams); err != nil {
		c.Logger().Errorf("Bind error: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	requestId := c.Param("requestId")
	request, err := fh.friend_service.ReactFriendRequest(c.Request().Context(), services.ReactFriendRequestParams{
		FriendRequestId: requestId,
		UserId:          authCtx.User.ID,
		Accept:          friendRequestParams.Accept,
	})
	if err != nil {
		c.Logger().Errorf("Could not update friend request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.FriendRequestFromModel(request, authCtx.User.ID))
}

func (fh *FriendHandler) FriendsIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	friends, err := fh.friend_service.GetFriends(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		c.Logger().Errorf("Could not fetch friends: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.FriendShipsResponseFromModel(friends, authCtx.User.ID))
}

func (fh *FriendHandler) FriendDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	friendId := c.Param("friendId")
	err := fh.friend_service.Unfriend(c.Request().Context(), authCtx.User.ID, friendId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
