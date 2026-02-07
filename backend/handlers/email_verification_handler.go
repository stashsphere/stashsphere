package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type EmailVerificationHandler struct {
	userService *services.UserService
}

func NewEmailVerificationHandler(userService *services.UserService) *EmailVerificationHandler {
	return &EmailVerificationHandler{userService}
}

func (h *EmailVerificationHandler) RequestVerification(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}

	err := h.userService.RequestEmailVerification(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

type VerifyEmailParams struct {
	Code string `json:"code" validate:"required,len=8"`
}

func (h *EmailVerificationHandler) VerifyEmail(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}

	var params VerifyEmailParams
	if err := c.Bind(&params); err != nil {
		return err
	}
	if err := c.Validate(params); err != nil {
		return err
	}

	err := h.userService.VerifyEmail(c.Request().Context(), authCtx.User.UserId, authCtx.User.Email, params.Code)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
