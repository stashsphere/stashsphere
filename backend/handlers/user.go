package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService}
}

func (ph *UserHandler) Index(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	users, err := ph.userService.GetAllUsers(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.UserProfilesFromModelSlice(users))
}

func (ph *UserHandler) Get(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	userId := c.Param("userId")
	user, err := ph.userService.FindUserByID(c.Request().Context(), userId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.UserProfileFromModel(user))
}
