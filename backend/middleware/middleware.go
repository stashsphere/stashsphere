package middleware

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/operations"
)

type UserContext struct {
	UserId string
	Email  string
	Name   string
}

type AuthContext struct {
	// whether a user has been authenticated
	Authenticated bool
	User          *UserContext
	AccessToken   *jwt.Token
}

func ExtractClaims(tokenContextKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			anonymousContext := AuthContext{
				Authenticated: false,
				User:          nil,
				AccessToken:   nil,
			}
			token, ok := c.Get(tokenContextKey).(*jwt.Token)
			if !ok {
				c.Set("auth", &anonymousContext)
				return next(c)
			}
			claims, ok := token.Claims.(*operations.AccessClaims)
			if !ok {
				return errors.New("failed to cast claims as ApplicationClaims")
			}
			authenticatedContext := AuthContext{
				Authenticated: true,
				User: &UserContext{
					UserId: claims.UserId,
					Email:  claims.Email,
					Name:   claims.Name,
				},
				AccessToken: token,
			}
			c.Set("auth", &authenticatedContext)
			return next(c)
		}
	}
}

// https://github.com/labstack/echo/issues/654#issuecomment-2448192207
func HeadToGetMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == http.MethodHead {
			// Set the method to GET temporarily to reuse the handler
			c.Request().Method = http.MethodGet

			defer func() {
				c.Request().Method = http.MethodHead
			}() // Restore method after

			// Call the next handler and then clear the response body
			if err := next(c); err != nil {
				if err.Error() == echo.ErrMethodNotAllowed.Error() {
					return c.NoContent(http.StatusOK) //nolint:errcheck
				}
				return err
			}
		}

		return next(c)
	}
}
