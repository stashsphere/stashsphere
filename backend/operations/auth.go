package operations

import (
	"context"
	"crypto/ed25519"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stashsphere/backend/models"
	"golang.org/x/crypto/bcrypt"
)

func AuthenticateUser(user *models.User, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return err
	}

	return nil
}

func AuthenticateUserByEmail(exec boil.ContextExecutor, ctx context.Context, email string, password string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.Email.EQ(email)).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	err = AuthenticateUser(user, password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func AuthenticateUserByID(ctx context.Context, exec boil.ContextExecutor, userId string, password string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	err = AuthenticateUser(user, password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

type AccessClaims struct {
	UserId string `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func CreateJWTAccessTokenForUser(user *models.User, privateKey ed25519.PrivateKey, issuedAt time.Time, lifetime time.Duration) (string, string, error) {
	claims := AccessClaims{
		UserId: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			Issuer:    "inventory",
			Subject:   "access",
		},
	}
	jwtAccess := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	accessToken, err := jwtAccess.SignedString(privateKey)
	if err != nil {
		return "", "", err
	}
	jwtInfo := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	// create an unsigned jwt token which is handed to the
	// user as regular cookie
	infoToken, err := jwtInfo.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", "", err
	}
	return accessToken, infoToken, err
}

type RefreshClaims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

func CreateJWTRefreshTokenForUser(user *models.User, privateKey ed25519.PrivateKey, issuedAt time.Time, lifetime time.Duration) (string, string, error) {
	claims := RefreshClaims{
		UserId: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			Issuer:    "inventory",
			Subject:   "refresh",
		},
	}
	jwtAccess := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	refreshToken, err := jwtAccess.SignedString(privateKey)
	if err != nil {
		return "", "", err
	}
	jwtInfo := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	// create an unsigned jwt token which is handed to the
	// user as regular cookie
	infoToken, err := jwtInfo.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", "", err
	}
	return refreshToken, infoToken, err
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 0)
}
