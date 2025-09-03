package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
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
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	user, err := ph.userService.FindUserByID(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ProfileFromModel(user))
}

type ProfileUpdateParams struct {
	Name        string  `json:"name"`
	FullName    string  `json:"fullName"`
	Information string  `json:"information"`
	ImageId     *string `json:"imageId"`
}

func (p *ProfileUpdateParams) ToUpdateUserParams() services.UpdateUserParams {
	return services.UpdateUserParams{
		Name:        p.Name,
		FullName:    p.FullName,
		Information: p.Information,
		ImageId:     p.ImageId,
	}
}

func (ph *ProfileHandler) ProfileHandlerPatch(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	params := ProfileUpdateParams{}
	if err := c.Bind(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	serviceParams := params.ToUpdateUserParams()
	serviceParams.UserId = authCtx.User.UserId
	user, err := ph.userService.UpdateUser(c.Request().Context(), serviceParams)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ProfileFromModel(user))
}
