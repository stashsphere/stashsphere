package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
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
		return &utils.ParameterError{Err: err}
	}
	if err := c.Validate(registerParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	_, err := rh.userService.CreateUser(c.Request().Context(), services.CreateUserParams{
		Name:       registerParams.Name,
		Email:      registerParams.Email,
		Password:   registerParams.Password,
		InviteCode: registerParams.InviteCode,
	})
	if err != nil {
		return err
	}
	return nil
}
