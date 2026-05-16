package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/voxire/onion/internal/modules/oauth/internal/domain"
	"github.com/voxire/onion/internal/modules/oauth/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

type Service struct {
	repo domain.Repository
	jwt  *jwtutils.JWTHelper
}

func NewService(ctx context.Context) *Service {
	secret, _ := ctx.Value(jwtutils.JWTSecretKey).(string)
	return &Service{
		repo: infrastructure.NewRepository(ctx),
		jwt:  jwtutils.NewJWTHelperWithDuration(secret, "rtylr", 24*time.Hour),
	}
}

func (s *Service) GetClient(_ context.Context, clientID string) (*domain.Client, error) {
	return s.repo.GetClient(clientID)
}

// ValidateAuthRequest validates client_id, redirect_uri, scope and returns the client.
func (s *Service) ValidateAuthRequest(_ context.Context, clientID, redirectURI, scope string) (*domain.Client, error) {
	client, err := s.repo.GetClient(clientID)
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
func (s *Service) Login(_ context.Context, email, password string) (int64, error) {
	userID, err := s.repo.GetUserIDByEmail(email)
	if err != nil {
		return 0, errors.New("invalid credentials")
	}

	hash, err := s.repo.GetUserPasswordHash(userID)
	if err != nil {
		return 0, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, errors.New("invalid credentials")
	}

	return userID, nil
}

// IssueAuthCode creates and stores a new authorization code.
func (s *Service) IssueAuthCode(_ context.Context, clientID string, userID int64, redirectURI, scope, codeChallenge, codeChallengeMethod string) (string, error) {
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

	if err := s.repo.SaveAuthCode(authCode); err != nil {
		return "", err
	}

	return code, nil
}

// ExchangeCode exchanges an authorization code for access+refresh tokens.
func (s *Service) ExchangeCode(_ context.Context, clientID, code, codeVerifier, redirectURI string) (accessToken, refreshToken string, expiresIn int, err error) {
	authCode, err := s.repo.GetAuthCode(code)
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

	if err := s.repo.MarkAuthCodeUsed(code); err != nil {
		return "", "", 0, err
	}

	businessID, role, _ := s.repo.GetUserBusinessIDAndRole(authCode.UserID)

	claims := &OAuthTokenClaims{
		UserID:     authCode.UserID,
		BusinessID: businessID,
		Role:       role,
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

	if err := s.repo.SaveToken(token); err != nil {
		return "", "", 0, err
	}

	return at, rt, int((24 * time.Hour).Seconds()), nil
}

// RefreshTokens exchanges a refresh token for a new access token.
func (s *Service) RefreshTokens(ctx context.Context, clientID, refreshToken string) (accessToken, newRefreshToken string, expiresIn int, err error) {
	token, err := s.repo.GetTokenByRefresh(refreshToken)
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

	if err := s.repo.RevokeToken(refreshToken); err != nil {
		return "", "", 0, err
	}

	businessID, role, _ := s.repo.GetUserBusinessIDAndRole(token.UserID)

	claims := &OAuthTokenClaims{
		UserID:     token.UserID,
		BusinessID: businessID,
		Role:       role,
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

	if err := s.repo.SaveToken(newToken); err != nil {
		return "", "", 0, err
	}

	return at, rt, int((24 * time.Hour).Seconds()), nil
}

// RevokeToken marks a refresh token as revoked.
func (s *Service) RevokeToken(_ context.Context, refreshToken string) error {
	return s.repo.RevokeToken(refreshToken)
}

func isAllowedRedirectURI(allowed []string, requested string) bool {
	for _, uri := range allowed {
		if strings.HasPrefix(requested, uri) {
			return true
		}
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

func verifyPKCE(verifier, challenge, method string) bool {
	if method == "S256" {
		h := sha256.Sum256([]byte(verifier))
		computed := base64.RawURLEncoding.EncodeToString(h[:])
		return computed == challenge
	}
	// plain method (not recommended, but support it)
	return verifier == challenge
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
