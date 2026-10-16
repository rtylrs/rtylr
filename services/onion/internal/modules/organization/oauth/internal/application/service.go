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

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/voxire/onion/internal/modules/organization/oauth/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

// OAuth client lookups are read-heavy and effectively immutable; the rest
// of the flow (Login, IssueAuthCode, ExchangeCode, RefreshTokens) MUST NOT
// be cached — those are security-sensitive single-use exchanges.
const oauthClientTTL = 10 * time.Minute

const accessTokenTTL = 24 * time.Hour

type Service struct {
	repo domain.Repository
	jwt  *jwtutils.JWTHelper
}

func NewService(ctx context.Context, repo domain.Repository) *Service {
	secret, _ := ctx.Value(jwtutils.JWTSecretKey).(string)
	return &Service{
		repo: repo,
		jwt:  jwtutils.NewJWTHelperWithDuration(secret, "rtylr", accessTokenTTL),
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

	if len(strings.Fields(scope)) == 0 || !isScopeAllowed(client.Scopes, scope) {
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

// TokenGrant is the result of a successful token exchange.
type TokenGrant struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Scope        string
}

// ExchangeCode exchanges an authorization code for access+refresh tokens.
func (s *Service) ExchangeCode(ctx context.Context, clientID, code, codeVerifier, redirectURI string) (*TokenGrant, error) {
	if _, err := s.validateTokenClient(ctx, clientID); err != nil {
		return nil, err
	}
	authCode, err := s.repo.GetAuthCode(ctx, code)
	if err != nil {
		return nil, domain.ErrCodeNotFound
	}

	if authCode.ClientID != clientID {
		return nil, domain.ErrCodeNotFound
	}

	if authCode.Used {
		return nil, domain.ErrCodeUsed
	}

	if time.Now().After(authCode.ExpiresAt) {
		return nil, domain.ErrCodeExpired
	}

	if authCode.RedirectURI != redirectURI {
		return nil, domain.ErrInvalidRedirectURI
	}

	if !verifyPKCE(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
		return nil, domain.ErrCodeVerifierFailed
	}

	// Consume atomically so two concurrent exchanges cannot both mint tokens.
	if err := s.repo.ConsumeAuthCode(ctx, code); err != nil {
		return nil, err
	}

	tokenEpoch, err := s.repo.GetUserTokenEpoch(ctx, authCode.UserID)
	if err != nil {
		return nil, err
	}

	return s.mintTokens(ctx, clientID, authCode.UserID, authCode.Scope, tokenEpoch)
}

// RefreshTokens exchanges a refresh token for a new access token.
func (s *Service) RefreshTokens(ctx context.Context, clientID, refreshToken string) (*TokenGrant, error) {
	client, err := s.validateTokenClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	token, err := s.repo.GetTokenByRefresh(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrTokenNotFound
	}

	if token.Revoked {
		return nil, domain.ErrTokenRevoked
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}

	if token.ClientID != clientID {
		return nil, domain.ErrTokenNotFound
	}

	// A password change or reset bumps the epoch; the access token the
	// refresh token was issued with records the epoch it was minted under.
	currentEpoch, err := s.repo.GetUserTokenEpoch(ctx, token.UserID)
	if err != nil {
		return nil, err
	}
	if issuedEpoch(token.AccessToken) != currentEpoch {
		_ = s.repo.RevokeToken(ctx, refreshToken)
		return nil, domain.ErrTokenRevoked
	}

	// Scopes the client no longer offers are dropped, never kept alive.
	scope := narrowScope(token.Scope, client.Scopes)
	if scope == "" {
		return nil, domain.ErrInvalidScope
	}

	// Atomic so two concurrent refreshes cannot both rotate the same token.
	if err := s.repo.ConsumeRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return s.mintTokens(ctx, clientID, token.UserID, scope, currentEpoch)
}

func (s *Service) mintTokens(ctx context.Context, clientID string, userID int64, scope string, tokenEpoch int64) (*TokenGrant, error) {
	businessID, err := s.repo.GetUserBusinessID(ctx, userID)
	if err != nil {
		return nil, err
	}

	claims := &OAuthTokenClaims{
		UserID:     userID,
		BusinessID: businessID,
		Scope:      scope,
		ClientID:   clientID,
		TokenEpoch: tokenEpoch,
	}

	at, err := s.jwt.CreateJWT(claims)
	if err != nil {
		return nil, err
	}

	rt, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveToken(ctx, &domain.Token{
		AccessToken:  at,
		RefreshToken: rt,
		ClientID:     clientID,
		UserID:       userID,
		Scope:        scope,
		ExpiresAt:    time.Now().Add(accessTokenTTL),
	}); err != nil {
		return nil, err
	}

	return &TokenGrant{
		AccessToken:  at,
		RefreshToken: rt,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		Scope:        scope,
	}, nil
}

// issuedEpoch reads the epoch embedded in an access token this service stored.
// The signature is not checked: the value comes from our own table, and an
// unreadable token yields zero, which fails closed once any epoch was bumped.
func issuedEpoch(accessToken string) int64 {
	claims := &OAuthTokenClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(accessToken, claims); err != nil {
		return 0
	}
	return claims.TokenEpoch
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
	if pa.Scheme != pb.Scheme || pa.Path != pb.Path || pa.RawQuery != pb.RawQuery {
		return false
	}
	// Native clients listen on an ephemeral loopback port (RFC 8252 7.3).
	if pa.Scheme == "http" && isLoopbackHost(pa.Hostname()) && pa.Port() == "" {
		return strings.EqualFold(pa.Hostname(), pb.Hostname()) && pb.User == nil
	}
	return strings.EqualFold(pa.Host, pb.Host)
}

func isLoopbackHost(host string) bool {
	switch strings.ToLower(host) {
	case "127.0.0.1", "::1", "localhost":
		return true
	}
	return false
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

func narrowScope(granted string, clientScopes []string) string {
	allowed := make(map[string]struct{}, len(clientScopes))
	for _, scope := range clientScopes {
		allowed[scope] = struct{}{}
	}
	kept := make([]string, 0)
	for _, scope := range strings.Fields(granted) {
		if _, ok := allowed[scope]; ok {
			kept = append(kept, scope)
		}
	}
	return strings.Join(kept, " ")
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

func (s *Service) validateTokenClient(ctx context.Context, clientID string) (*domain.Client, error) {
	client, err := s.repo.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	// The schema has no client-secret credential yet. Refuse private clients
	// instead of treating an unauthenticated client_id as authentication.
	if client == nil || !client.IsPublic {
		return nil, domain.ErrClientAuthenticationRequired
	}
	return client, nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
