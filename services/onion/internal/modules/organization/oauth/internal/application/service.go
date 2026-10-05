package application

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/voxire/onion/internal/modules/organization/oauth/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

// OAuth client lookups are read-heavy and effectively immutable; the rest
// of the flow (Login, IssueAuthCode, ExchangeCode, RefreshTokens) MUST NOT
// be cached — those are security-sensitive single-use exchanges.
const oauthClientTTL = 10 * time.Minute

type Service struct {
	repo domain.Repository
	jwt  *jwtutils.JWTHelper
}

func NewService(ctx context.Context, repo domain.Repository) *Service {
	secret, _ := ctx.Value(jwtutils.JWTSecretKey).(string)
	return &Service{
		repo: repo,
		jwt:  jwtutils.NewJWTHelperWithDuration(secret, "rtylr", 24*time.Hour),
	}
}

func (s *Service) GetClient(ctx context.Context, clientID string) (*domain.Client, error) {
	key := "cache:oauth:client:" + clientID
	return cacher.Of[*domain.Client]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.Client, error) {
			return s.repo.GetClient(ctx, clientID)
		},
		cacher.WithExpiry(ctx, oauthClientTTL),
	)
}

// ValidateAuthRequest validates client_id, redirect_uri, scope and returns the client.
func (s *Service) ValidateAuthRequest(ctx context.Context, clientID, redirectURI, scope string) (*domain.Client, error) {
	client, err := s.repo.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if !isAllowedRedirectURI(client.RedirectURIs, redirectURI) {
		return nil, domain.ErrInvalidRedirectURI
	}

	if scope != "" && !isScopeAllowed(client.Scopes, scope) {
		return nil, domain.ErrInvalidScope
	}

	return client, nil
}

// Login validates email+password and returns the user ID.
func (s *Service) Login(ctx context.Context, email, password string) (int64, error) {
	userID, err := s.repo.GetUserIDByEmail(ctx, email)
	if err != nil {
		return 0, errors.New("invalid credentials")
	}

	hash, err := s.repo.GetUserPasswordHash(ctx, userID)
	if err != nil {
		return 0, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, errors.New("invalid credentials")
	}

	return userID, nil
}

// IssueAuthCode creates and stores a new authorization code.
func (s *Service) IssueAuthCode(ctx context.Context, clientID string, userID int64, redirectURI, scope, codeChallenge, codeChallengeMethod string) (string, error) {
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		return "", domain.ErrPKCERequired
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	code := hex.EncodeToString(raw)

	authCode := &domain.AuthCode{
		Code:                code,
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	}

	if err := s.repo.SaveAuthCode(ctx, authCode); err != nil {
		return "", err
	}

	return code, nil
}

// ExchangeCode exchanges an authorization code for access+refresh tokens.
func (s *Service) ExchangeCode(ctx context.Context, clientID, code, codeVerifier, redirectURI string) (accessToken, refreshToken string, expiresIn int, err error) {
	if err := s.validateTokenClient(ctx, clientID); err != nil {
		return "", "", 0, err
	}
	authCode, err := s.repo.GetAuthCode(ctx, code)
	if err != nil {
		return "", "", 0, domain.ErrCodeNotFound
	}

	if authCode.ClientID != clientID {
		return "", "", 0, domain.ErrCodeNotFound
	}

	if authCode.Used {
		return "", "", 0, domain.ErrCodeUsed
	}

	if time.Now().After(authCode.ExpiresAt) {
		return "", "", 0, domain.ErrCodeExpired
	}

	if authCode.RedirectURI != redirectURI {
		return "", "", 0, domain.ErrInvalidRedirectURI
	}

	if !verifyPKCE(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
		return "", "", 0, domain.ErrCodeVerifierFailed
	}

	// Consume atomically so two concurrent exchanges cannot both mint tokens.
	if err := s.repo.ConsumeAuthCode(ctx, code); err != nil {
		return "", "", 0, err
	}

	businessID, err := s.repo.GetUserBusinessID(ctx, authCode.UserID)
	if err != nil {
		return "", "", 0, err
	}

	claims := &OAuthTokenClaims{
		UserID:     authCode.UserID,
		BusinessID: businessID,
		Scope:      authCode.Scope,
	}

	at, err := s.jwt.CreateJWT(claims)
	if err != nil {
		return "", "", 0, err
	}

	rt, err := generateRefreshToken()
	if err != nil {
		return "", "", 0, err
	}

	token := &domain.Token{
		AccessToken:  at,
		RefreshToken: rt,
		ClientID:     clientID,
		UserID:       authCode.UserID,
		Scope:        authCode.Scope,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if err := s.repo.SaveToken(ctx, token); err != nil {
		return "", "", 0, err
	}

	return at, rt, int((24 * time.Hour).Seconds()), nil
}

// RefreshTokens exchanges a refresh token for a new access token.
func (s *Service) RefreshTokens(ctx context.Context, clientID, refreshToken string) (accessToken, newRefreshToken string, expiresIn int, err error) {
	if err := s.validateTokenClient(ctx, clientID); err != nil {
		return "", "", 0, err
	}
	token, err := s.repo.GetTokenByRefresh(ctx, refreshToken)
	if err != nil {
		return "", "", 0, domain.ErrTokenNotFound
	}

	if token.Revoked {
		return "", "", 0, domain.ErrTokenRevoked
	}

	if time.Now().After(token.ExpiresAt) {
		return "", "", 0, domain.ErrTokenExpired
	}

	if token.ClientID != clientID {
		return "", "", 0, domain.ErrTokenNotFound
	}

	if err := s.repo.RevokeToken(ctx, refreshToken); err != nil {
		return "", "", 0, err
	}

	businessID, err := s.repo.GetUserBusinessID(ctx, token.UserID)
	if err != nil {
		return "", "", 0, err
	}

	claims := &OAuthTokenClaims{
		UserID:     token.UserID,
		BusinessID: businessID,
		Scope:      token.Scope,
	}

	at, err := s.jwt.CreateJWT(claims)
	if err != nil {
		return "", "", 0, err
	}

	rt, err := generateRefreshToken()
	if err != nil {
		return "", "", 0, err
	}

	newToken := &domain.Token{
		AccessToken:  at,
		RefreshToken: rt,
		ClientID:     clientID,
		UserID:       token.UserID,
		Scope:        token.Scope,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if err := s.repo.SaveToken(ctx, newToken); err != nil {
		return "", "", 0, err
	}

	return at, rt, int((24 * time.Hour).Seconds()), nil
}

// RevokeToken marks a refresh token as revoked.
func (s *Service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeToken(ctx, refreshToken)
}

// isAllowedRedirectURI requires an exact match (after normalization) so a
// prefix like "https://app.example.com" cannot be satisfied by
// "https://app.example.com.evil.com/cb".
func isAllowedRedirectURI(allowed []string, requested string) bool {
	for _, uri := range allowed {
		if uri == requested {
			return true
		}
		if redirectURIsEqual(uri, requested) {
			return true
		}
	}
	return false
}

func redirectURIsEqual(a, b string) bool {
	pa, errA := url.Parse(a)
	pb, errB := url.Parse(b)
	if errA != nil || errB != nil {
		return false
	}
	return pa.Scheme == pb.Scheme &&
		strings.EqualFold(pa.Host, pb.Host) &&
		pa.Path == pb.Path &&
		pa.RawQuery == pb.RawQuery
}

func isScopeAllowed(clientScopes []string, requested string) bool {
	clientScopeSet := make(map[string]struct{}, len(clientScopes))
	for _, s := range clientScopes {
		clientScopeSet[s] = struct{}{}
	}
	for _, s := range strings.Fields(requested) {
		if _, ok := clientScopeSet[s]; !ok {
			return false
		}
	}
	return true
}

func verifyPKCE(verifier, challenge, method string) bool {
	if verifier == "" || challenge == "" || method != "S256" {
		return false
	}
	h := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	// The code challenge is the only secret in this flow: `==` leaks how many
	// leading bytes matched, which turns code_verifier guessing into a byte at
	// a time. hmac.Equal keeps the comparison constant-time, matching the
	// internal token, CDN signature and webhook signature checks.
	return hmac.Equal([]byte(computed), []byte(challenge))
}

func (s *Service) validateTokenClient(ctx context.Context, clientID string) error {
	client, err := s.repo.GetClient(ctx, clientID)
	if err != nil {
		return err
	}
	// The schema has no client-secret credential yet. Refuse private clients
	// instead of treating an unauthenticated client_id as authentication.
	if client == nil || !client.IsPublic {
		return domain.ErrClientAuthenticationRequired
	}
	return nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
