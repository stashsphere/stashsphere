package handlers

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type OIDCHandler struct {
	oidcService   *services.OIDCService
	authService   *services.AuthService
	frontendURL   string
	secureCookies bool
}

func NewOIDCHandler(oidcService *services.OIDCService, authService *services.AuthService, frontendURL string, secureCookies bool) *OIDCHandler {
	return &OIDCHandler{
		oidcService:   oidcService,
		authService:   authService,
		frontendURL:   frontendURL,
		secureCookies: secureCookies,
	}
}

// AuthorizeHandlerGet redirects the user to the OIDC provider's authorization endpoint.
func (h *OIDCHandler) AuthorizeHandlerGet(c echo.Context) error {
	providerName := c.Param("provider")

	authURL, state, nonce, err := h.oidcService.AuthorizationURL(c.Request().Context(), providerName)
	if err != nil {
		return err
	}

	// Store state and nonce in cookies for verification in callback
	c.SetCookie(&http.Cookie{
		Name:     "oidc-state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	c.SetCookie(&http.Cookie{
		Name:     "oidc-nonce",
		Value:    nonce,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusFound, authURL)
}

func (h *OIDCHandler) CallbackHandlerGet(c echo.Context) error {
	providerName := c.Param("provider")

	stateCookie, err := c.Cookie("oidc-state")
	if err != nil || stateCookie.Value == "" {
		return utils.OIDCCallbackFailedError{Err: err}
	}
	queryState := c.QueryParam("state")
	if queryState != stateCookie.Value {
		return utils.OIDCCallbackFailedError{Err: err}
	}

	nonceCookie, err := c.Cookie("oidc-nonce")
	if err != nil || nonceCookie.Value == "" {
		return utils.OIDCCallbackFailedError{Err: err}
	}

	// Clear state/nonce cookies
	c.SetCookie(&http.Cookie{Name: "oidc-state", Value: "", Path: "/", MaxAge: -1})
	c.SetCookie(&http.Cookie{Name: "oidc-nonce", Value: "", Path: "/", MaxAge: -1})

	// Check for error from provider
	if errParam := c.QueryParam("error"); errParam != "" {
		q := url.Values{}
		q.Set("error", errParam)
		q.Set("error_description", c.QueryParam("error_description"))
		return c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?"+q.Encode())
	}

	code := c.QueryParam("code")
	if code == "" {
		return utils.OIDCCallbackFailedError{Err: err}
	}

	// Exchange code for ID token
	userInfo, err := h.oidcService.ExchangeCode(c.Request().Context(), providerName, code, nonceCookie.Value)
	if err != nil {
		return err
	}

	result, err := h.oidcService.FindOrLinkUser(c.Request().Context(), providerName, userInfo)
	if err != nil {
		return err
	}

	if result.Action == "link_required" {
		c.SetCookie(&http.Cookie{
			Name:     "oidc-link-challenge",
			Value:    result.ChallengeToken,
			Path:     "/",
			MaxAge:   600,
			HttpOnly: false, // Frontend JS needs to read this
			Secure:   h.secureCookies,
			SameSite: http.SameSiteLaxMode,
		})
		q := url.Values{}
		q.Set("action", "link_required")
		q.Set("email", result.Email)
		q.Set("provider", result.Provider)
		return c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?"+q.Encode())
	}

	// Issue JWT tokens and set cookies
	accessToken, infoToken, refreshToken, refreshInfoToken, err := h.authService.IssueTokensForUser(result.User)
	if err != nil {
		return err
	}
	h.authService.SetAuthCookies(c, accessToken, infoToken, refreshToken, refreshInfoToken)

	return c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?status=success")
}

type LinkPostParams struct {
	Password       string `json:"password" validate:"min=1"`
	ChallengeToken string `json:"challengeToken" validate:"min=1"`
}

func (h *OIDCHandler) LinkHandlerPost(c echo.Context) error {
	providerName := c.Param("provider")

	params := LinkPostParams{}
	if err := c.Bind(&params); err != nil {
		return utils.ParameterError{Err: err}
	}
	if err := c.Validate(params); err != nil {
		return err
	}

	user, err := h.oidcService.VerifyLinkAndAuthenticate(c.Request().Context(), providerName, params.ChallengeToken, params.Password)
	if err != nil {
		return err
	}

	// Clear the challenge cookie
	c.SetCookie(&http.Cookie{Name: "oidc-link-challenge", Value: "", Path: "/", MaxAge: -1})

	// Issue JWT tokens and set cookies
	accessToken, infoToken, refreshToken, refreshInfoToken, err := h.authService.IssueTokensForUser(user)
	if err != nil {
		return err
	}
	h.authService.SetAuthCookies(c, accessToken, infoToken, refreshToken, refreshInfoToken)

	return c.NoContent(http.StatusOK)
}
