package utils

import "github.com/labstack/echo/v4"

func RedirectToReferrer(c echo.Context, code int, defaultURL string) error {
	referer := c.Request().Header.Get("Referer")
	if referer == "" {
		return c.Redirect(code, defaultURL)
	} else {
		return c.Redirect(code, referer)
	}
}
