package operations

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetAccessTokenCookie(c echo.Context, domain string, accessToken string, maxAge int, secure bool) {
	cookie := http.Cookie{
		Name:     "stashsphere-access",
		Value:    accessToken,
		Path:     "/",
		Domain:   domain,
		Secure:   secure,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(&cookie)
}

func SetInfoTokenCookie(c echo.Context, domain string, infoToken string, maxAge int, secure bool) {
	cookie := http.Cookie{
		Name:     "stashsphere-info",
		Value:    infoToken,
		Path:     "/",
		Domain:   domain,
		Secure:   secure,
		MaxAge:   maxAge,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(&cookie)
}
