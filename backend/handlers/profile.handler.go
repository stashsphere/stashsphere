package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
)

type ProfileHandler struct {
	userService *services.UserService
}

func NewProfileHandler(userService *services.UserService) *ProfileHandler {
	return &ProfileHandler{userService}
}

func (ph *ProfileHandler) ProfileHandlerGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.Redirect(http.StatusSeeOther, "/user/login")
	}
	user, err := ph.userService.FindUserByID(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		c.Logger().Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, resources.ProfileFromModel(user))
}

type ProfileUpdateParams struct {
	Name string `json:"name"`
}

func (ph *ProfileHandler) ProfileHandlerPatch(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.Redirect(http.StatusSeeOther, "/user/login")
	}
	params := ProfileUpdateParams{}
	if err := c.Bind(&params); err != nil {
		fmt.Printf("Bind Failed %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	user, err := ph.userService.UpdateUser(c.Request().Context(), authCtx.User.ID, params.Name)
	if err != nil {
		c.Logger().Errorf("Failed to update user %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	return c.JSON(http.StatusOK, resources.ProfileFromModel(user))
}

func (ph *ProfileHandler) ProfileHandlerIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.Redirect(http.StatusSeeOther, "/user/login")
	}
	users, err := ph.userService.GetAllUsers(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to fetch users %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.ProfilesFromModelSlice(users))
}
