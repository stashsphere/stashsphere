package operations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

func GenerateVerificationCode() (string, error) {
	max := big.NewInt(90000000) // 99999999 - 10000000 + 1
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	code := n.Int64() + 10000000
	return fmt.Sprintf("%08d", code), nil
}

func CreateVerificationCode(ctx context.Context, exec boil.ContextExecutor, userId string, email string, validUntil time.Time) (string, error) {
	code, err := GenerateVerificationCode()
	if err != nil {
		return "", err
	}

	// Upsert email_verifications row with verified_at = NULL (pending)
	verification := models.EmailVerification{
		UserID: userId,
		Email:  email,
	}
	err = verification.Upsert(ctx, exec, false, []string{
		models.EmailVerificationColumns.UserID,
		models.EmailVerificationColumns.Email,
	}, boil.None(), boil.Infer())
	if err != nil {
		return "", err
	}

	// Insert the code
	verificationCode := models.EmailVerificationCode{
		UserID:     userId,
		Email:      email,
		DigitCode:  code,
		ValidUntil: validUntil.UTC(),
	}
	return code, verificationCode.Insert(ctx, exec, boil.Infer())
}

func VerifyCode(ctx context.Context, exec boil.ContextExecutor, userId string, email string, code string) error {
	codeRow, err := models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.UserID.EQ(userId),
		models.EmailVerificationCodeWhere.Email.EQ(email),
		models.EmailVerificationCodeWhere.DigitCode.EQ(code),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.InvalidVerificationCodeError{}
		}
		return err
	}

	if time.Now().UTC().After(codeRow.ValidUntil) {
		_, _ = codeRow.Delete(ctx, exec)
		return utils.VerificationCodeExpiredError{}
	}

	// Set verified_at on the email_verifications row
	verification, err := models.EmailVerifications(
		models.EmailVerificationWhere.UserID.EQ(userId),
		models.EmailVerificationWhere.Email.EQ(email),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	verification.VerifiedAt = null.TimeFrom(time.Now())
	_, err = verification.Update(ctx, exec, boil.Whitelist(models.EmailVerificationColumns.VerifiedAt))
	if err != nil {
		return err
	}

	// Delete all codes for this user+email after successful verification
	_, err = models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.UserID.EQ(userId),
		models.EmailVerificationCodeWhere.Email.EQ(email),
	).DeleteAll(ctx, exec)
	return err
}

func GetEmailVerificationStatus(ctx context.Context, exec boil.ContextExecutor, userId string) (*models.EmailVerification, error) {
	verification, err := models.EmailVerifications(
		models.EmailVerificationWhere.UserID.EQ(userId),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return verification, nil
}

func PurgeExpiredVerificationCodes(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	return models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.ValidUntil.LT(time.Now().UTC().Add(-24*time.Hour)),
	).DeleteAll(ctx, exec)
}
