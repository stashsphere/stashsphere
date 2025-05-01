package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type LoginHandler struct {
	authService *services.AuthService
}

func NewLoginHandler(authService *services.AuthService) *LoginHandler {
	return &LoginHandler{
		authService,
	}
}

type LoginPostParams struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

func (lh *LoginHandler) LoginHandlerPost(c echo.Context) error {
	loginParams := LoginPostParams{}
	if err := c.Bind(&loginParams); err != nil {
		return utils.ParameterError{Err: err}
	}
	_, accessToken, infoToken, err := lh.authService.AuthorizeUser(c.Request().Context(), loginParams.Email, loginParams.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	lh.authService.SetAuthCookies(c, accessToken, infoToken)
	return nil
}

func (lh *LoginHandler) LogoutHandlerDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	lh.authService.ClearAuthCookies(c)
	return nil
}
