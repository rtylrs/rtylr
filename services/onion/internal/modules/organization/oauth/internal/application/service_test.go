package application

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/voxire/onion/internal/modules/organization/oauth/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
)

type mockRepo struct {
	client          *domain.Client
	authCode        *domain.AuthCode
	token           *domain.Token
	userID          int64
	passwordHash    string
	businessID      int64
	tokenEpoch      int64
	saved           *domain.Token
	epochErr        error
	businessErr     error
	clientErr       error
	authCodeErr     error
	tokenErr        error
	consumeTokenErr error
	err             error
}

func (m *mockRepo) GetClient(_ context.Context, _ string) (*domain.Client, error) {
	if m.clientErr != nil {
		return m.client, m.clientErr
	}
	return m.client, m.err
}
func (m *mockRepo) SaveAuthCode(_ context.Context, _ *domain.AuthCode) error { return m.err }
func (m *mockRepo) GetAuthCode(_ context.Context, _ string) (*domain.AuthCode, error) {
	if m.authCodeErr != nil {
		return m.authCode, m.authCodeErr
	}
	return m.authCode, m.err
}
func (m *mockRepo) MarkAuthCodeUsed(_ context.Context, _ string) error { return m.err }
func (m *mockRepo) ConsumeAuthCode(_ context.Context, _ string) error  { return m.err }
func (m *mockRepo) SaveToken(_ context.Context, token *domain.Token) error {
	m.saved = token
	return m.err
}
func (m *mockRepo) GetTokenByRefresh(_ context.Context, _ string) (*domain.Token, error) {
	if m.tokenErr != nil {
		return m.token, m.tokenErr
	}
	return m.token, m.err
}
func (m *mockRepo) RevokeToken(_ context.Context, _ string) error { return m.err }
func (m *mockRepo) ConsumeRefreshToken(_ context.Context, _ string) error {
	if m.consumeTokenErr != nil {
		return m.consumeTokenErr
	}
	return m.err
}
func (m *mockRepo) GetUserPasswordHash(_ context.Context, _ int64) (string, error) {
	return m.passwordHash, m.err
}

func (m *mockRepo) GetUserIDByEmail(_ context.Context, _ string) (int64, error) {
	return m.userID, m.err
}

func (m *mockRepo) GetUserTokenEpoch(_ context.Context, _ int64) (int64, error) {
	return m.tokenEpoch, m.epochErr
}

func (m *mockRepo) GetUserBusinessID(_ context.Context, _ int64) (int64, error) {
	return m.businessID, m.businessErr
}

func newTestService(repo domain.Repository) *Service {
	return &Service{
		repo: repo,
		jwt:  jwtutils.NewJWTHelperWithDuration("test-secret", "rtylr", 24*time.Hour),
	}
}

// ---- GetClient ----

func TestGetClient_ReturnsDelegatedClient(t *testing.T) {
	client := &domain.Client{ID: "app1", Name: "Test App", RedirectURIs: []string{"http://localhost"}}
	s := newTestService(&mockRepo{client: client})
	got, err := s.GetClient(context.Background(), "app1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "app1" {
		t.Errorf("expected client ID app1, got %s", got.ID)
	}
}

func TestGetClient_PropagatesNotFoundError(t *testing.T) {
	want := domain.ErrClientNotFound
	s := newTestService(&mockRepo{err: want})
	_, err := s.GetClient(context.Background(), "missing")
	if !errors.Is(err, want) {
		t.Errorf("expected ErrClientNotFound, got %v", err)
	}
}

// ---- ValidateAuthRequest ----

func TestValidateAuthRequest_ValidRedirectURIAndScope(t *testing.T) {
	client := &domain.Client{
		ID:           "app1",
		RedirectURIs: []string{"https://example.com/callback"},
		Scopes:       []string{"read", "write"},
	}
	s := newTestService(&mockRepo{client: client})
	got, err := s.ValidateAuthRequest(context.Background(), "app1", "https://example.com/callback", "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "app1" {
		t.Errorf("expected app1, got %s", got.ID)
	}
}

func TestValidateAuthRequest_RejectsPrefixMatches(t *testing.T) {
	client := &domain.Client{
		ID:           "app1",
		RedirectURIs: []string{"https://app.example.com/callback"},
		Scopes:       []string{"read"},
	}
	s := newTestService(&mockRepo{client: client})
	for _, uri := range []string{
		"https://app.example.com.evil.com/callback",
		"https://app.example.comcallback",
		"https://app.example.com/callback/extra",
	} {
		if _, err := s.ValidateAuthRequest(context.Background(), "app1", uri, "read"); !errors.Is(err, domain.ErrInvalidRedirectURI) {
			t.Errorf("expected ErrInvalidRedirectURI for %q, got %v", uri, err)
		}
	}
}

func TestValidateAuthRequest_InvalidRedirectURIReturnsError(t *testing.T) {
	client := &domain.Client{
		ID:           "app1",
		RedirectURIs: []string{"https://allowed.com/callback"},
		Scopes:       []string{"read"},
	}
	s := newTestService(&mockRepo{client: client})
	_, err := s.ValidateAuthRequest(context.Background(), "app1", "https://evil.com/callback", "read")
	if !errors.Is(err, domain.ErrInvalidRedirectURI) {
		t.Errorf("expected ErrInvalidRedirectURI, got %v", err)
	}
}

func TestValidateAuthRequest_InvalidScopeReturnsError(t *testing.T) {
	client := &domain.Client{
		ID:           "app1",
		RedirectURIs: []string{"https://example.com"},
		Scopes:       []string{"read"},
	}
	s := newTestService(&mockRepo{client: client})
	_, err := s.ValidateAuthRequest(context.Background(), "app1", "https://example.com", "write")
	if !errors.Is(err, domain.ErrInvalidScope) {
		t.Errorf("expected ErrInvalidScope, got %v", err)
	}
}

// ---- Login ----

func TestLogin_PropagatesInvalidCredentialsWhenEmailNotFound(t *testing.T) {
	s := newTestService(&mockRepo{err: errors.New("not found")})
	_, err := s.Login(context.Background(), "unknown@test.com", "password")
	if err == nil || err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials', got %v", err)
	}
}

func TestLogin_PropagatesInvalidCredentialsOnBadPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.MinCost)
	s := newTestService(&mockRepo{userID: 1, passwordHash: string(hash)})
	_, err := s.Login(context.Background(), "user@test.com", "wrong-pass")
	if err == nil || err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials', got %v", err)
	}
}

func TestLogin_SucceedsWithCorrectCredentials(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	s := newTestService(&mockRepo{userID: 42, passwordHash: string(hash)})
	id, err := s.Login(context.Background(), "user@test.com", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected userID=42, got %d", id)
	}
}

// ---- IssueAuthCode ----

func TestIssueAuthCode_ReturnsNonEmptyCode(t *testing.T) {
	s := newTestService(&mockRepo{})
	code, err := s.IssueAuthCode(context.Background(), "app1", 1, "https://example.com", "read", "challenge", "S256")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "" {
		t.Error("expected non-empty code")
	}
}

func TestIssueAuthCode_PropagatesSaveError(t *testing.T) {
	want := errors.New("save error")
	s := newTestService(&mockRepo{err: want})
	_, err := s.IssueAuthCode(context.Background(), "app1", 1, "https://example.com", "read", "challenge", "S256")
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestIssueAuthCode_RequiresS256PKCE(t *testing.T) {
	s := newTestService(&mockRepo{})
	for _, tc := range []struct {
		challenge string
		method    string
	}{
		{challenge: "", method: "S256"},
		{challenge: "challenge", method: "plain"},
		{challenge: "challenge", method: ""},
	} {
		if _, err := s.IssueAuthCode(context.Background(), "app1", 1, "https://example.com", "read", tc.challenge, tc.method); !errors.Is(err, domain.ErrPKCERequired) {
			t.Errorf("IssueAuthCode(%q, %q) error = %v, want ErrPKCERequired", tc.challenge, tc.method, err)
		}
	}
}

// ---- ExchangeCode ----

func TestExchangeCode_CodeNotFoundReturnsError(t *testing.T) {
	s := newTestService(&mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, authCodeErr: errors.New("not found")})
	_, _, _, err := s.ExchangeCode(context.Background(), "app1", "bad-code", "verifier", "https://example.com")
	if !errors.Is(err, domain.ErrCodeNotFound) {
		t.Errorf("expected ErrCodeNotFound, got %v", err)
	}
}

func TestExchangeCode_UsedCodeReturnsError(t *testing.T) {
	code := &domain.AuthCode{
		Code:      "used-code",
		ClientID:  "app1",
		Used:      true,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	s := newTestService(&mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, authCode: code})
	_, _, _, err := s.ExchangeCode(context.Background(), "app1", "used-code", "", "https://example.com")
	if !errors.Is(err, domain.ErrCodeUsed) {
		t.Errorf("expected ErrCodeUsed, got %v", err)
	}
}

func TestExchangeCode_ExpiredCodeReturnsError(t *testing.T) {
	code := &domain.AuthCode{
		Code:      "expired",
		ClientID:  "app1",
		Used:      false,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	s := newTestService(&mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, authCode: code})
	_, _, _, err := s.ExchangeCode(context.Background(), "app1", "expired", "", "https://example.com")
	if !errors.Is(err, domain.ErrCodeExpired) {
		t.Errorf("expected ErrCodeExpired, got %v", err)
	}
}

func TestExchangeCode_RejectsPrivateClientWithoutAuthentication(t *testing.T) {
	s := newTestService(&mockRepo{client: &domain.Client{ID: "app1", IsPublic: false}})
	_, _, _, err := s.ExchangeCode(context.Background(), "app1", "code", "verifier", "https://example.com")
	if !errors.Is(err, domain.ErrClientAuthenticationRequired) {
		t.Fatalf("error = %v, want ErrClientAuthenticationRequired", err)
	}
}

func TestExchangeCode_PropagatesIdentityLookupError(t *testing.T) {
	challenge := "ruXwYqQ5u3k9jv9xYz7vM7v4V6rZqW6J6P9f0JbJ3S0"
	code := &domain.AuthCode{
		Code:                "code",
		ClientID:            "app1",
		UserID:              7,
		RedirectURI:         "https://example.com/callback",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(time.Minute),
	}
	s := newTestService(&mockRepo{
		client:      &domain.Client{ID: "app1", IsPublic: true},
		authCode:    code,
		businessErr: errors.New("identity lookup failed"),
	})
	verifier := "verifier"
	hash := sha256.Sum256([]byte(verifier))
	code.CodeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
	_, _, _, err := s.ExchangeCode(context.Background(), "app1", code.Code, verifier, code.RedirectURI)
	if err == nil || err.Error() != "identity lookup failed" {
		t.Fatalf("error = %v, want identity lookup failure", err)
	}
}

func TestExchangeCode_EmbedsTheUsersTokenEpoch(t *testing.T) {
	verifier := "verifier"
	hash := sha256.Sum256([]byte(verifier))
	code := &domain.AuthCode{
		Code:                "code",
		ClientID:            "app1",
		UserID:              7,
		RedirectURI:         "https://example.com/callback",
		CodeChallenge:       base64.RawURLEncoding.EncodeToString(hash[:]),
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(time.Minute),
	}
	repo := &mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, authCode: code, tokenEpoch: 3}
	s := newTestService(repo)
	at, _, _, err := s.ExchangeCode(context.Background(), "app1", code.Code, verifier, code.RedirectURI)
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if got := issuedEpoch(at); got != 3 {
		t.Fatalf("access token epoch = %d, want 3: the auth middleware rejects tokens whose epoch differs from the user's", got)
	}
}

// ---- RefreshTokens ----

func TestRefreshTokens_TokenNotFoundReturnsError(t *testing.T) {
	s := newTestService(&mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, tokenErr: errors.New("not found")})
	_, _, _, err := s.RefreshTokens(context.Background(), "app1", "bad-refresh")
	if !errors.Is(err, domain.ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound, got %v", err)
	}
}

func TestRefreshTokens_RevokedTokenReturnsError(t *testing.T) {
	token := &domain.Token{
		ClientID:  "app1",
		Revoked:   true,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	s := newTestService(&mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, token: token})
	_, _, _, err := s.RefreshTokens(context.Background(), "app1", "refresh-token")
	if !errors.Is(err, domain.ErrTokenRevoked) {
		t.Errorf("expected ErrTokenRevoked, got %v", err)
	}
}

func TestRefreshTokens_LostRotationRaceReturnsRevoked(t *testing.T) {
	token := &domain.Token{
		ClientID:  "app1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	s := newTestService(&mockRepo{
		client:          &domain.Client{ID: "app1", IsPublic: true},
		token:           token,
		consumeTokenErr: domain.ErrTokenRevoked,
	})
	_, _, _, err := s.RefreshTokens(context.Background(), "app1", "refresh-token")
	if !errors.Is(err, domain.ErrTokenRevoked) {
		t.Errorf("expected ErrTokenRevoked, got %v", err)
	}
}

func epochToken(t *testing.T, s *Service, epoch int64) string {
	t.Helper()
	at, err := s.jwt.CreateJWT(&OAuthTokenClaims{UserID: 7, TokenEpoch: epoch})
	if err != nil {
		t.Fatalf("CreateJWT: %v", err)
	}
	return at
}

func TestRefreshTokens_RejectsTokenIssuedBeforeAPasswordChange(t *testing.T) {
	repo := &mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, tokenEpoch: 2}
	s := newTestService(repo)
	repo.token = &domain.Token{ClientID: "app1", UserID: 7, AccessToken: epochToken(t, s, 1), ExpiresAt: time.Now().Add(time.Hour)}
	_, _, _, err := s.RefreshTokens(context.Background(), "app1", "refresh-token")
	if !errors.Is(err, domain.ErrTokenRevoked) {
		t.Fatalf("expected ErrTokenRevoked, got %v", err)
	}
	if repo.saved != nil {
		t.Fatal("a refresh token from before the password change minted a new token")
	}
}

func TestRefreshTokens_CarriesTheCurrentEpochForward(t *testing.T) {
	repo := &mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, tokenEpoch: 2}
	s := newTestService(repo)
	repo.token = &domain.Token{ClientID: "app1", UserID: 7, AccessToken: epochToken(t, s, 2), ExpiresAt: time.Now().Add(time.Hour)}
	at, _, _, err := s.RefreshTokens(context.Background(), "app1", "refresh-token")
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}
	if got := issuedEpoch(at); got != 2 {
		t.Fatalf("refreshed access token epoch = %d, want 2", got)
	}
}

func TestRefreshTokens_ExpiredTokenReturnsError(t *testing.T) {
	token := &domain.Token{
		ClientID:  "app1",
		Revoked:   false,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	s := newTestService(&mockRepo{client: &domain.Client{ID: "app1", IsPublic: true}, token: token})
	_, _, _, err := s.RefreshTokens(context.Background(), "app1", "refresh-token")
	if !errors.Is(err, domain.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

// ---- RevokeToken ----

func TestRevokeToken_DelegatesToRepo(t *testing.T) {
	s := newTestService(&mockRepo{})
	err := s.RevokeToken(context.Background(), "some-refresh-token")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRevokeToken_PropagatesRepoError(t *testing.T) {
	want := errors.New("revoke error")
	s := newTestService(&mockRepo{err: want})
	err := s.RevokeToken(context.Background(), "token")
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- verifyPKCE (pure function) ----

func TestVerifyPKCE_S256MismatchReturnsFalse(t *testing.T) {
	if verifyPKCE("bad-verifier", "challenge-hash", "S256") {
		t.Error("expected false for mismatched S256 PKCE")
	}
}

func TestVerifyPKCE_RejectsPlainMethod(t *testing.T) {
	if verifyPKCE("my-verifier", "my-verifier", "plain") {
		t.Error("expected plain PKCE to be rejected")
	}
}

func TestVerifyPKCE_PlainMethodReturnsFalseOnMismatch(t *testing.T) {
	if verifyPKCE("verifier", "other", "plain") {
		t.Error("expected false for plain method with mismatching values")
	}
}

func TestVerifyPKCE_AcceptsMatchingS256Verifier(t *testing.T) {
	verifier := "sBQnbGvZ_lfS3E1AeSfL0jnKL0b1lZKwPMQVzY5XnQU"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	if !verifyPKCE(verifier, challenge, "S256") {
		t.Fatal("a verifier whose S256 digest is the stored challenge must be accepted")
	}
}

// The challenge is the only secret in the PKCE exchange, so the comparison has
// to be constant-time. A `==` here leaks the length of the matching prefix,
// which is what turns verifier guessing into a byte-at-a-time search. A partial
// match must be rejected, and the rejection must not depend on where the
// strings first differ.
func TestVerifyPKCE_RejectsPartialChallengePrefix(t *testing.T) {
	verifier := "verifier-value-that-is-long-enough"
	sum := sha256.Sum256([]byte(verifier))
	full := base64.RawURLEncoding.EncodeToString(sum[:])

	for cut := 1; cut < len(full); cut++ {
		if verifyPKCE(verifier, full[:cut], "S256") {
			t.Fatalf("a %d-byte prefix of the challenge must not verify", cut)
		}
	}
	if !verifyPKCE(verifier, full, "S256") {
		t.Fatal("the complete challenge must verify")
	}
}
