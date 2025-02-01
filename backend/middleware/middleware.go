package middleware

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/operations"
)

type UserContext struct {
	ID    string
	Email string
	Name  string
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
			claims, ok := token.Claims.(*operations.ApplicationClaims)
			if !ok {
				return errors.New("failed to cast claims as ApplicationClaims")
			}
			authenticatedContext := AuthContext{
				Authenticated: true,
				User: &UserContext{
					ID:    claims.ID,
					Email: claims.Email,
					Name:  claims.Name,
				},
				AccessToken: token,
			}
			c.Set("auth", &authenticatedContext)
			return next(c)
		}
	}
}
