package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/services"
)

type RegisterHandler struct {
	userService *services.UserService
}

func NewRegisterHandler(userService *services.UserService) *RegisterHandler {
	return &RegisterHandler{
		userService,
	}
}

type RegisterPostParams struct {
	Name       string `json:"name" validate:"gt=1"`
	Email      string `json:"email" validate:"email"`
	Password   string `json:"password" validate:"gt=3"`
	InviteCode string `json:"inviteCode"`
}

func (rh *RegisterHandler) RegisterHandlerPost(c echo.Context) error {
	registerParams := RegisterPostParams{}
	if err := c.Bind(&registerParams); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	if err := c.Validate(registerParams); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	_, err := rh.userService.CreateUser(c.Request().Context(), services.CreateUserParams{
		Name:       registerParams.Name,
		Email:      registerParams.Email,
		Password:   registerParams.Password,
		InviteCode: registerParams.InviteCode,
	})
	if err != nil {
		c.Logger().Errorf("error creating user: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	return nil
}
