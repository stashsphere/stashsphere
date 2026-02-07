package operations_test

import (
	"context"
	"testing"
	"time"

	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateVerificationCode(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()
	validUntil := time.Now().Add(30 * time.Minute)

	code, err := operations.CreateVerificationCode(ctx, db, user.ID, user.Email, validUntil)
	assert.NoError(t, err)
	assert.Len(t, code, 8)

	// Verify the code row exists
	codeRow, err := models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.UserID.EQ(user.ID),
		models.EmailVerificationCodeWhere.Email.EQ(user.Email),
		models.EmailVerificationCodeWhere.DigitCode.EQ(code),
	).One(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, code, codeRow.DigitCode)

	// Verify the email_verifications row was created with NULL verified_at
	verification, err := models.EmailVerifications(
		models.EmailVerificationWhere.UserID.EQ(user.ID),
		models.EmailVerificationWhere.Email.EQ(user.Email),
	).One(ctx, db)
	assert.NoError(t, err)
	assert.False(t, verification.VerifiedAt.Valid, "verified_at should be NULL")
}

func TestVerifyCodeSuccess(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()
	validUntil := time.Now().Add(30 * time.Minute)

	code, err := operations.CreateVerificationCode(ctx, db, user.ID, user.Email, validUntil)
	assert.NoError(t, err)

	err = operations.VerifyCode(ctx, db, user.ID, user.Email, code)
	assert.NoError(t, err)

	// Verify email_verifications row has verified_at set
	verification, err := models.EmailVerifications(
		models.EmailVerificationWhere.UserID.EQ(user.ID),
		models.EmailVerificationWhere.Email.EQ(user.Email),
	).One(ctx, db)
	assert.NoError(t, err)
	assert.True(t, verification.VerifiedAt.Valid, "verified_at should be set")

	// All codes for user+email should be deleted
	codeCount, err := models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.UserID.EQ(user.ID),
		models.EmailVerificationCodeWhere.Email.EQ(user.Email),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), codeCount, "all codes should be deleted after verification")
}

func TestVerifyCodeDeletesAllCodes(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()
	validUntil := time.Now().Add(30 * time.Minute)

	// Create two codes
	code1, err := operations.CreateVerificationCode(ctx, db, user.ID, user.Email, validUntil)
	assert.NoError(t, err)
	_, err = operations.CreateVerificationCode(ctx, db, user.ID, user.Email, validUntil)
	assert.NoError(t, err)

	// Verify with the first code
	err = operations.VerifyCode(ctx, db, user.ID, user.Email, code1)
	assert.NoError(t, err)

	// Both codes should be deleted
	codeCount, err := models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.UserID.EQ(user.ID),
		models.EmailVerificationCodeWhere.Email.EQ(user.Email),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), codeCount)
}

func TestVerifyCodeInvalidCode(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()

	err = operations.VerifyCode(ctx, db, user.ID, user.Email, "99999999")
	assert.Error(t, err)
	assert.IsType(t, utils.InvalidVerificationCodeError{}, err)
}

func TestVerifyCodeExpired(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()
	// Create a code that's already expired
	validUntil := time.Now().Add(-1 * time.Minute)

	code, err := operations.CreateVerificationCode(ctx, db, user.ID, user.Email, validUntil)
	assert.NoError(t, err)

	err = operations.VerifyCode(ctx, db, user.ID, user.Email, code)
	assert.Error(t, err)
	assert.IsType(t, utils.VerificationCodeExpiredError{}, err)

	// Expired code should be deleted
	codeCount, err := models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.UserID.EQ(user.ID),
		models.EmailVerificationCodeWhere.Email.EQ(user.Email),
		models.EmailVerificationCodeWhere.DigitCode.EQ(code),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), codeCount, "expired code should be deleted")
}

func TestGetEmailVerificationStatusPending(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()
	validUntil := time.Now().Add(30 * time.Minute)
	_, err = operations.CreateVerificationCode(ctx, db, user.ID, user.Email, validUntil)
	assert.NoError(t, err)

	verification, err := operations.GetEmailVerificationStatus(ctx, db, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, verification)
	assert.False(t, verification.VerifiedAt.Valid, "should be pending (verified_at NULL)")
}

func TestGetEmailVerificationStatusVerified(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()
	validUntil := time.Now().Add(30 * time.Minute)
	code, err := operations.CreateVerificationCode(ctx, db, user.ID, user.Email, validUntil)
	assert.NoError(t, err)

	err = operations.VerifyCode(ctx, db, user.ID, user.Email, code)
	assert.NoError(t, err)

	verification, err := operations.GetEmailVerificationStatus(ctx, db, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, verification)
	assert.True(t, verification.VerifiedAt.Valid, "should be verified")
	assert.Equal(t, user.Email, verification.Email)
}

func TestPurgeExpiredVerificationCodes(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	userService := services.NewUserService(db, false, "", 60, nil)
	params := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	user, err := userService.CreateUser(context.Background(), *params)
	assert.NoError(t, err)

	ctx := context.Background()

	// Create a code expired more than 24 hours ago (should be purged)
	_, err = operations.CreateVerificationCode(ctx, db, user.ID, user.Email, time.Now().UTC().Add(-48*time.Hour))
	assert.NoError(t, err)

	// Create a code expired less than 24 hours ago (should NOT be purged)
	_, err = operations.CreateVerificationCode(ctx, db, user.ID, user.Email, time.Now().UTC().Add(-12*time.Hour))
	assert.NoError(t, err)

	// Create a valid code (should NOT be purged)
	_, err = operations.CreateVerificationCode(ctx, db, user.ID, user.Email, time.Now().Add(30*time.Minute))
	assert.NoError(t, err)

	count, err := operations.PurgeExpiredVerificationCodes(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "only the code expired >24h ago should be purged")

	// 2 codes should remain
	remaining, err := models.EmailVerificationCodes(
		models.EmailVerificationCodeWhere.UserID.EQ(user.ID),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), remaining)
}

func TestPurgeExpiredVerificationCodesNoneExpired(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDown)

	ctx := context.Background()
	count, err := operations.PurgeExpiredVerificationCodes(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
