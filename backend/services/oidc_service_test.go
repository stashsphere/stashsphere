package services_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"testing"

	"github.com/stashsphere/backend/config"
	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stashsphere/backend/utils"
	"github.com/stretchr/testify/assert"
)

func oidcTestSetup(t *testing.T, oidcCfg config.OIDCConfig) (*services.OIDCService, *services.UserService, *sql.DB) {
	t.Helper()

	db, tearDownFunc, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	t.Cleanup(tearDownFunc)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	oidcService := services.NewOIDCService(db, oidcCfg, "http://localhost:8081", privateKey, publicKey)
	userService := services.NewUserService(db, false, "", 60, nil)

	return oidcService, userService, db
}

func defaultOIDCConfig() config.OIDCConfig {
	return config.OIDCConfig{
		Enabled: true,
		Providers: []config.OIDCProviderConfig{
			{
				Name:         "dex",
				DisplayName:  "Dex",
				IssuerURL:    "https://dex.example.com",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
		},
	}
}

func TestFindOrLinkUser_NewUser(t *testing.T) {
	oidcService, _, db := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	result, err := oidcService.FindOrLinkUser(ctx, "dex", &services.OIDCUserInfo{
		Subject: "sub-001",
		Email:   "alice@example.com",
		Name:    "Alice",
	})
	assert.NoError(t, err)
	assert.Equal(t, "authenticated", result.Action)
	assert.NotNil(t, result.User)
	assert.Equal(t, "alice@example.com", result.User.Email)
	assert.Equal(t, "Alice", result.User.Name)
	assert.False(t, result.User.PasswordHash.Valid)

	// Verify external_auth row exists
	extAuth, err := models.ExternalAuths(
		models.ExternalAuthWhere.Provider.EQ("dex"),
		models.ExternalAuthWhere.Subject.EQ("sub-001"),
	).One(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, result.User.ID, extAuth.UserID)
}

func TestFindOrLinkUser_ExistingOIDCLink(t *testing.T) {
	oidcService, _, _ := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	userInfo := &services.OIDCUserInfo{
		Subject: "sub-001",
		Email:   "alice@example.com",
		Name:    "Alice",
	}

	result1, err := oidcService.FindOrLinkUser(ctx, "dex", userInfo)
	assert.NoError(t, err)

	result2, err := oidcService.FindOrLinkUser(ctx, "dex", userInfo)
	assert.NoError(t, err)
	assert.Equal(t, "authenticated", result2.Action)
	assert.Equal(t, result1.User.ID, result2.User.ID)
}

func TestFindOrLinkUser_ExistingPasswordUser(t *testing.T) {
	oidcService, userService, db := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	email := testUserParams.Email
	testUser, err := userService.CreateUser(ctx, *testUserParams)
	assert.NoError(t, err)
	assert.NotNil(t, testUser)

	result, err := oidcService.FindOrLinkUser(ctx, "dex", &services.OIDCUserInfo{
		Subject: "sub-new",
		Email:   email,
		Name:    "OIDC Name",
	})
	assert.NoError(t, err)
	assert.Equal(t, "link_required", result.Action)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, "dex", result.Provider)
	assert.NotEmpty(t, result.ChallengeToken)

	// Verify no external_auth row was created
	count, err := models.ExternalAuths(
		models.ExternalAuthWhere.Provider.EQ("dex"),
		models.ExternalAuthWhere.Subject.EQ("sub-new"),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestFindOrLinkUser_DifferentProviderSameEmail(t *testing.T) {
	cfg := config.OIDCConfig{
		Enabled: true,
		Providers: []config.OIDCProviderConfig{
			{Name: "alpha", DisplayName: "Alpha", IssuerURL: "https://alpha.example.com", ClientID: "c", ClientSecret: "s"},
			{Name: "beta", DisplayName: "Beta", IssuerURL: "https://beta.example.com", ClientID: "c", ClientSecret: "s"},
		},
	}
	oidcService, _, _ := oidcTestSetup(t, cfg)
	ctx := context.Background()

	_, err := oidcService.FindOrLinkUser(ctx, "alpha", &services.OIDCUserInfo{
		Subject: "sub-alpha",
		Email:   "shared@example.com",
		Name:    "Shared",
	})
	assert.NoError(t, err)

	result, err := oidcService.FindOrLinkUser(ctx, "beta", &services.OIDCUserInfo{
		Subject: "sub-beta",
		Email:   "shared@example.com",
		Name:    "Shared",
	})
	assert.NoError(t, err)
	assert.Equal(t, "link_required", result.Action)
}

func TestFindOrLinkUser_EmptyNameFallsBackToEmail(t *testing.T) {
	oidcService, _, _ := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	result, err := oidcService.FindOrLinkUser(ctx, "dex", &services.OIDCUserInfo{
		Subject: "sub-noname",
		Email:   "noname@example.com",
		Name:    "",
	})
	assert.NoError(t, err)
	assert.Equal(t, "authenticated", result.Action)
	assert.Equal(t, "noname@example.com", result.User.Name)
}

func TestVerifyLinkAndAuthenticate_Success(t *testing.T) {
	oidcService, userService, db := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	password := testUserParams.Password
	email := testUserParams.Email
	_, err := userService.CreateUser(ctx, *testUserParams)
	assert.NoError(t, err)

	result, err := oidcService.FindOrLinkUser(ctx, "dex", &services.OIDCUserInfo{
		Subject: "sub-link",
		Email:   email,
		Name:    "Link User",
	})
	assert.NoError(t, err)
	assert.Equal(t, "link_required", result.Action)

	user, err := oidcService.VerifyLinkAndAuthenticate(ctx, "dex", result.ChallengeToken, password)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)

	// Verify external_auth row was created
	count, err := models.ExternalAuths(
		models.ExternalAuthWhere.Provider.EQ("dex"),
		models.ExternalAuthWhere.Subject.EQ("sub-link"),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestVerifyLinkAndAuthenticate_WrongPassword(t *testing.T) {
	oidcService, userService, db := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	email := testUserParams.Email
	_, err := userService.CreateUser(ctx, *testUserParams)
	assert.NoError(t, err)

	result, err := oidcService.FindOrLinkUser(ctx, "dex", &services.OIDCUserInfo{
		Subject: "sub-wrong-pw",
		Email:   email,
		Name:    "Wrong PW",
	})
	assert.NoError(t, err)

	_, err = oidcService.VerifyLinkAndAuthenticate(ctx, "dex", result.ChallengeToken, "wrong-password")
	assert.ErrorIs(t, err, utils.OIDCLinkIncorrectPasswordError{})

	// Verify no external_auth row was created
	count, err := models.ExternalAuths(
		models.ExternalAuthWhere.Provider.EQ("dex"),
		models.ExternalAuthWhere.Subject.EQ("sub-wrong-pw"),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestVerifyLinkAndAuthenticate_InvalidToken(t *testing.T) {
	oidcService, _, _ := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	_, err := oidcService.VerifyLinkAndAuthenticate(ctx, "dex", "garbage-token-string", "password")
	assert.ErrorIs(t, err, utils.OIDCLinkChallengeExpiredError{})
}

func TestVerifyLinkAndAuthenticate_ProviderMismatch(t *testing.T) {
	oidcService, userService, _ := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	email := testUserParams.Email
	_, err := userService.CreateUser(ctx, *testUserParams)
	assert.NoError(t, err)

	result, err := oidcService.FindOrLinkUser(ctx, "dex", &services.OIDCUserInfo{
		Subject: "sub-mismatch",
		Email:   email,
		Name:    "Mismatch",
	})
	assert.NoError(t, err)

	_, err = oidcService.VerifyLinkAndAuthenticate(ctx, "other", result.ChallengeToken, testUserParams.Password)
	assert.ErrorIs(t, err, utils.OIDCLinkChallengeExpiredError{})
}

func TestGetProviderConfigs_Disabled(t *testing.T) {
	cfg := config.OIDCConfig{
		Enabled: false,
		Providers: []config.OIDCProviderConfig{
			{Name: "dex", DisplayName: "Dex"},
		},
	}
	oidcService, _, _ := oidcTestSetup(t, cfg)

	configs := oidcService.GetProviderConfigs()
	assert.Nil(t, configs)
}

func TestGetProviderConfigs_Enabled(t *testing.T) {
	cfg := config.OIDCConfig{
		Enabled: true,
		Providers: []config.OIDCProviderConfig{
			{Name: "alpha", DisplayName: "Alpha", IssuerURL: "https://alpha.example.com", ClientID: "a", ClientSecret: "sa"},
			{Name: "beta", DisplayName: "Beta", IssuerURL: "https://beta.example.com", ClientID: "b", ClientSecret: "sb"},
		},
	}
	oidcService, _, _ := oidcTestSetup(t, cfg)

	configs := oidcService.GetProviderConfigs()
	assert.Len(t, configs, 2)
	assert.Equal(t, "alpha", configs[0].Name)
	assert.Equal(t, "Alpha", configs[0].DisplayName)
	assert.Equal(t, "beta", configs[1].Name)
	assert.Equal(t, "Beta", configs[1].DisplayName)
}

func TestAuthorizationURL_UnknownProvider(t *testing.T) {
	oidcService, _, _ := oidcTestSetup(t, defaultOIDCConfig())
	ctx := context.Background()

	_, _, _, err := oidcService.AuthorizationURL(ctx, "nonexistent")
	assert.ErrorIs(t, err, utils.OIDCProviderNotFoundError{})
}
