package services

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/config"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

const discoveryTTL = 60 * time.Minute

type cachedProvider struct {
	runtime      *OIDCProviderRuntime
	discoveredAt time.Time
}

type OIDCProviderRuntime struct {
	Config       config.OIDCProviderConfig
	Provider     *oidc.Provider
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
}

type OIDCService struct {
	db         *sql.DB
	oidcConfig config.OIDCConfig
	baseURL    string
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey

	mu    sync.Mutex
	cache map[string]*cachedProvider
}

func NewOIDCService(db *sql.DB, oidcConfig config.OIDCConfig, baseURL string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *OIDCService {
	return &OIDCService{
		db:         db,
		oidcConfig: oidcConfig,
		baseURL:    strings.TrimRight(baseURL, "/"),
		privateKey: privateKey,
		publicKey:  publicKey,
		cache:      make(map[string]*cachedProvider),
	}
}

func (s *OIDCService) redirectURI(providerName string) string {
	return s.baseURL + "/api/auth/oidc/" + providerName + "/callback"
}

func (s *OIDCService) findProviderConfig(name string) (*config.OIDCProviderConfig, error) {
	for i := range s.oidcConfig.Providers {
		if s.oidcConfig.Providers[i].Name == name {
			return &s.oidcConfig.Providers[i], nil
		}
	}
	return nil, utils.OIDCProviderNotFoundError{}
}

// discoverProvider performs OIDC discovery for a single provider.
func (s *OIDCService) discoverProvider(ctx context.Context, pc *config.OIDCProviderConfig) (*OIDCProviderRuntime, error) {
	provider, err := oidc.NewProvider(ctx, pc.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery failed for provider %q: %w", pc.Name, err)
	}

	scopes := pc.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	oauth2Config := oauth2.Config{
		ClientID:     pc.ClientID,
		ClientSecret: pc.ClientSecret,
		RedirectURL:  s.redirectURI(pc.Name),
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: pc.ClientID,
	})

	log.Info().Str("provider", pc.Name).Str("issuer", pc.IssuerURL).Msg("OIDC provider discovered")

	return &OIDCProviderRuntime{
		Config:       *pc,
		Provider:     provider,
		OAuth2Config: oauth2Config,
		Verifier:     verifier,
	}, nil
}

// getProvider returns a discovered provider runtime, using a cached version if
// still valid or performing fresh discovery otherwise.
func (s *OIDCService) getProvider(ctx context.Context, name string) (*OIDCProviderRuntime, error) {
	if !s.oidcConfig.Enabled {
		return nil, utils.OIDCProviderNotFoundError{}
	}

	pc, err := s.findProviderConfig(name)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if cached, ok := s.cache[name]; ok {
		if time.Since(cached.discoveredAt) < discoveryTTL {
			return cached.runtime, nil
		}
	}

	runtime, err := s.discoverProvider(ctx, pc)
	if err != nil {
		return nil, err
	}

	s.cache[name] = &cachedProvider{
		runtime:      runtime,
		discoveredAt: time.Now(),
	}

	return runtime, nil
}

// GetProviderConfigs returns display info for all configured providers.
func (s *OIDCService) GetProviderConfigs() []config.OIDCProviderConfig {
	if !s.oidcConfig.Enabled {
		return nil
	}
	configs := make([]config.OIDCProviderConfig, len(s.oidcConfig.Providers))
	copy(configs, s.oidcConfig.Providers)
	return configs
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// AuthorizationURL builds the authorization URL for a provider and returns it along with state and nonce values.
func (s *OIDCService) AuthorizationURL(ctx context.Context, providerName string) (authURL string, state string, nonce string, err error) {
	p, err := s.getProvider(ctx, providerName)
	if err != nil {
		return "", "", "", err
	}

	state, err = generateRandomString(32)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	nonce, err = generateRandomString(32)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	authURL = p.OAuth2Config.AuthCodeURL(state, oidc.Nonce(nonce))
	return authURL, state, nonce, nil
}

// OIDCUserInfo holds the claims extracted from an ID token.
type OIDCUserInfo struct {
	Subject string
	Email   string
	Name    string
}

// ExchangeCode exchanges an authorization code for tokens and returns user info from the ID token.
func (s *OIDCService) ExchangeCode(ctx context.Context, providerName string, code string, expectedNonce string) (*OIDCUserInfo, error) {
	p, err := s.getProvider(ctx, providerName)
	if err != nil {
		return nil, err
	}

	oauth2Token, err := p.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, utils.OIDCCallbackFailedError{Err: fmt.Errorf("code exchange failed: %w", err)}
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, utils.OIDCCallbackFailedError{Err: fmt.Errorf("no id_token in response")}
	}

	idToken, err := p.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, utils.OIDCCallbackFailedError{Err: fmt.Errorf("id_token verification failed: %w", err)}
	}

	if idToken.Nonce != expectedNonce {
		return nil, utils.OIDCCallbackFailedError{Err: fmt.Errorf("nonce mismatch")}
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, utils.OIDCCallbackFailedError{Err: fmt.Errorf("failed to parse claims: %w", err)}
	}

	if claims.Email == "" {
		return nil, utils.OIDCCallbackFailedError{Err: fmt.Errorf("provider did not return an email claim")}
	}

	return &OIDCUserInfo{
		Subject: idToken.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
	}, nil
}

// CallbackResult represents the outcome of the OIDC callback flow.
type CallbackResult struct {
	// Action is either "authenticated" or "link_required"
	Action string `json:"action"`

	// Set when Action == "authenticated"
	User *models.User `json:"-"`

	// Set when Action == "link_required"
	Email          string `json:"email,omitempty"`
	Provider       string `json:"provider,omitempty"`
	ChallengeToken string `json:"challengeToken,omitempty"`
	ExpiresIn      int    `json:"expiresIn,omitempty"`
}

func (s *OIDCService) FindOrLinkUser(ctx context.Context, providerName string, userInfo *OIDCUserInfo) (*CallbackResult, error) {
	// Step 1: Check external_auth for existing link
	extAuth, err := models.ExternalAuths(
		models.ExternalAuthWhere.Provider.EQ(providerName),
		models.ExternalAuthWhere.Subject.EQ(userInfo.Subject),
	).One(ctx, s.db)
	if err == nil {
		// Found existing link — load user
		user, err := operations.FindUserByID(ctx, s.db, extAuth.UserID)
		if err != nil {
			return nil, err
		}
		return &CallbackResult{Action: "authenticated", User: user}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Step 2: Check users table by email
	_, err = operations.FindUserByEmail(ctx, s.db, userInfo.Email)
	if err != nil {
		if !errors.As(err, &utils.NotFoundError{}) {
			return nil, err
		}

		// Case A: Email not found — create new user
		newUser, err := s.createOIDCUser(ctx, providerName, userInfo)
		if err != nil {
			return nil, err
		}
		return &CallbackResult{Action: "authenticated", User: newUser}, nil
	}

	// Case B: Email found — existing password account, require linking
	challengeToken, err := s.createLinkChallengeToken(providerName, userInfo.Subject, userInfo.Email)
	if err != nil {
		return nil, err
	}

	return &CallbackResult{
		Action:         "link_required",
		Email:          userInfo.Email,
		Provider:       providerName,
		ChallengeToken: challengeToken,
		ExpiresIn:      600,
	}, nil
}

func (s *OIDCService) createOIDCUser(ctx context.Context, providerName string, userInfo *OIDCUserInfo) (*models.User, error) {
	userID, err := gonanoid.New()
	if err != nil {
		return nil, err
	}

	name := userInfo.Name
	if name == "" {
		name = userInfo.Email
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user := models.User{
		ID:    userID,
		Name:  name,
		Email: userInfo.Email,
	}
	if err := user.Insert(ctx, tx, boil.Whitelist(
		models.UserColumns.ID,
		models.UserColumns.Name,
		models.UserColumns.Email,
	)); err != nil {
		return nil, err
	}

	extAuth := models.ExternalAuth{
		UserID:   userID,
		Provider: providerName,
		Subject:  userInfo.Subject,
	}
	if err := extAuth.Insert(ctx, tx, boil.Infer()); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}

// LinkChallengeClaims are the JWT claims for a link challenge token.
type LinkChallengeClaims struct {
	Email    string `json:"email"`
	Provider string `json:"provider"`
	Subject  string `json:"subject"`
	Nonce    string `json:"nonce"`
	jwt.RegisteredClaims
}

func (s *OIDCService) createLinkChallengeToken(providerName string, subject string, email string) (string, error) {
	nonce, err := generateRandomString(16)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := LinkChallengeClaims{
		Email:    email,
		Provider: providerName,
		Subject:  subject,
		Nonce:    nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "inventory",
			Subject:   "oidc-link-challenge",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(s.privateKey)
}

// VerifyLinkAndAuthenticate verifies the challenge token and password, then creates the external_auth link.
func (s *OIDCService) VerifyLinkAndAuthenticate(ctx context.Context, providerName string, challengeTokenStr string, password string) (*models.User, error) {
	// Parse and validate challenge token
	token, err := jwt.ParseWithClaims(challengeTokenStr, &LinkChallengeClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
	}, jwt.WithValidMethods([]string{"EdDSA"}))
	if err != nil {
		return nil, utils.OIDCLinkChallengeExpiredError{}
	}

	claims, ok := token.Claims.(*LinkChallengeClaims)
	if !ok {
		return nil, utils.OIDCLinkChallengeExpiredError{}
	}

	// Verify provider matches
	if claims.Provider != providerName {
		return nil, utils.OIDCLinkChallengeExpiredError{}
	}

	// Look up user by email
	user, err := operations.FindUserByEmail(ctx, s.db, claims.Email)
	if err != nil {
		return nil, err
	}

	// Verify password
	if !user.PasswordHash.Valid {
		return nil, utils.OIDCLinkIncorrectPasswordError{}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)); err != nil {
		return nil, utils.OIDCLinkIncorrectPasswordError{}
	}

	// Create external_auth link
	extAuth := models.ExternalAuth{
		UserID:   user.ID,
		Provider: claims.Provider,
		Subject:  claims.Subject,
	}
	if err := extAuth.Insert(ctx, s.db, boil.Infer()); err != nil {
		return nil, err
	}

	return user, nil
}
