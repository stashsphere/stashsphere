package services

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
)

type AuthService struct {
	db                   *sql.DB
	privateKey           ed25519.PrivateKey
	publicKey            ed25519.PublicKey
	accessTokenLifeTime  time.Duration
	refreshTokenLifeTime time.Duration
	cookieDomain         string
}

func NewAuthService(db *sql.DB, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, accessTokenLifeTime time.Duration, refreshTokenLifetime time.Duration, cookieDomain string) *AuthService {
	return &AuthService{
		db,
		privateKey,
		publicKey,
		accessTokenLifeTime,
		refreshTokenLifetime,
		cookieDomain,
	}
}

func (as *AuthService) AuthorizeUser(ctx context.Context, email string, password string) (*models.User, string, string, string, string, error) {
	user, err := operations.AuthenticateUser(as.db, ctx, email, password)
	if err != nil {
		return nil, "", "", "", "", err
	}
	accessToken, infoToken, err := operations.CreateJWTAccessTokenForUser(user, as.privateKey, time.Now(), as.accessTokenLifeTime)
	if err != nil {
		return nil, "", "", "", "", err
	}
	refreshToken, refreshInfoToken, err := operations.CreateJWTRefreshTokenForUser(user, as.privateKey, time.Now(), as.refreshTokenLifeTime)
	if err != nil {
		return nil, "", "", "", "", err
	}
	return user, accessToken, infoToken, refreshToken, refreshInfoToken, nil
}

func (as *AuthService) SetAuthCookies(ctx echo.Context, accessToken string, infoToken string, refreshToken string, refreshInfoToken string) {
	operations.SetAccessTokenCookie(ctx, as.cookieDomain, accessToken, int(as.accessTokenLifeTime.Seconds()), false)
	operations.SetInfoTokenCookie(ctx, as.cookieDomain, infoToken, int(as.accessTokenLifeTime.Seconds()), false)
	operations.SetRefreshTokenCookie(ctx, as.cookieDomain, refreshToken, int(as.refreshTokenLifeTime.Seconds()), false)
	operations.SetRefreshIntoTokenCookie(ctx, as.cookieDomain, refreshInfoToken, int(as.refreshTokenLifeTime.Seconds()), false)
}

func (as *AuthService) ClearAuthCookies(ctx echo.Context) {
	operations.SetAccessTokenCookie(ctx, as.cookieDomain, "", -1, false)
	operations.SetInfoTokenCookie(ctx, as.cookieDomain, "", -1, false)
	operations.SetRefreshTokenCookie(ctx, as.cookieDomain, "", -1, false)
	operations.SetRefreshIntoTokenCookie(ctx, as.cookieDomain, "", -1, false)
}

func (as *AuthService) AuthorizeUserWithRefreshToken(ctx context.Context, value string) (*models.User, string, string, string, string, error) {
	token, err := jwt.ParseWithClaims(value, &operations.RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		return as.publicKey, nil
	})
	if err != nil {
		return nil, "", "", "", "", err
	}
	claims := token.Claims.(*operations.RefreshClaims)
	user, err := operations.FindUserByID(ctx, as.db, claims.UserId)
	if err != nil {
		return nil, "", "", "", "", err
	}
	accessToken, infoToken, err := operations.CreateJWTAccessTokenForUser(user, as.privateKey, time.Now(), as.accessTokenLifeTime)
	if err != nil {
		return nil, "", "", "", "", err
	}
	refreshToken, refreshInfoToken, err := operations.CreateJWTRefreshTokenForUser(user, as.privateKey, time.Now(), as.refreshTokenLifeTime)
	if err != nil {
		return nil, "", "", "", "", err
	}
	return user, accessToken, infoToken, refreshToken, refreshInfoToken, nil
}
