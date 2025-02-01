package operations

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stashsphere/backend/models"
	"golang.org/x/crypto/bcrypt"
)

func AuthenticateUser(db *sql.DB, ctx context.Context, email string, password string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.Email.EQ(email)).One(ctx, db)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

type ApplicationClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

func CreateJWTTokenForUser(user *models.User, privateKey ed25519.PrivateKey, issuedAt time.Time, lifetime time.Duration) (string, string, error) {
	claims := ApplicationClaims{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			// TODO
			Issuer:  "inventory",
			Subject: "main",
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

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 0)
}
