package services

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
)

type AuthService struct {
	db                  *sql.DB
	privateKey          ed25519.PrivateKey
	publicKey           ed25519.PublicKey
	accessTokenLifeTime time.Duration
	cookieDomain        string
}

func NewAuthService(db *sql.DB, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, accessTokenLifeTime time.Duration, cookieDomain string) *AuthService {
	return &AuthService{
		db,
		privateKey,
		publicKey,
		accessTokenLifeTime,
		cookieDomain,
	}
}

func (as *AuthService) AuthorizeUser(ctx context.Context, email string, password string) (*models.User, string, string, error) {
	user, err := operations.AuthenticateUser(as.db, ctx, email, password)
	if err != nil {
		return nil, "", "", err
	}
	accessToken, infoToken, err := operations.CreateJWTTokenForUser(user, as.privateKey, time.Now(), as.accessTokenLifeTime)
	if err != nil {
		return nil, "", "", err
	}
	return user, accessToken, infoToken, nil
}

func (as *AuthService) SetAuthCookies(ctx echo.Context, accessToken string, infoToken string) {
	// TODO SECURE depending on dev mode
	operations.SetAccessTokenCookie(ctx, as.cookieDomain, accessToken, int(as.accessTokenLifeTime.Seconds()), false)
	operations.SetInfoTokenCookie(ctx, as.cookieDomain, infoToken, int(as.accessTokenLifeTime.Seconds()), false)
}

func (as *AuthService) ClearAuthCookies(ctx echo.Context) {
	operations.SetAccessTokenCookie(ctx, as.cookieDomain, "", -1, false)
	operations.SetInfoTokenCookie(ctx, as.cookieDomain, "", -1, false)
}
