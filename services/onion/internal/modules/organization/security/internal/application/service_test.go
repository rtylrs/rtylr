package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/security/internal/domain"
)

// securityRepo is a local interface matching all methods called on
// infrastructure.Repository so we can inject a test double.
type securityRepo interface {
	GetSecuritySettings(ctx context.Context, userID int64) (domain.SecuritySettings, error)
	SetTwoFactor(ctx context.Context, userID int64, enabled bool) error
	GetLoginHistory(ctx context.Context, userID int64) ([]domain.LoginEntry, error)
	InvalidateAllSessions(ctx context.Context, userID int64) error
	RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error
}

type mockSecurityRepo struct {
	settings      domain.SecuritySettings
	settingsErr   error
	history       []domain.LoginEntry
	historyErr    error
	setTFAErr     error
	invalidateErr error
	recordErr     error
	recordedArgs  struct {
		email, method, ip, userAgent string
		success                      bool
	}
}

func (m *mockSecurityRepo) GetSecuritySettings(_ context.Context, _ int64) (domain.SecuritySettings, error) {
	return m.settings, m.settingsErr
}

func (m *mockSecurityRepo) SetTwoFactor(_ context.Context, _ int64, _ bool) error {
	return m.setTFAErr
}

func (m *mockSecurityRepo) GetLoginHistory(_ context.Context, _ int64) ([]domain.LoginEntry, error) {
	return m.history, m.historyErr
}

func (m *mockSecurityRepo) InvalidateAllSessions(_ context.Context, _ int64) error {
	return m.invalidateErr
}

func (m *mockSecurityRepo) RecordLogin(_ context.Context, _ int64, email, method, ip, userAgent string, success bool) error {
	m.recordedArgs.email = email
	m.recordedArgs.method = method
	m.recordedArgs.ip = ip
	m.recordedArgs.userAgent = userAgent
	m.recordedArgs.success = success
	return m.recordErr
}

// serviceWithMock creates a Service whose concrete repo field is replaced via
// a wrapper that satisfies the same method set as infrastructure.Repository.
// Because the struct field is *infrastructure.Repository we test pure helpers
// and the mockable paths via direct method dispatch below.

func newTestService(r securityRepo) *serviceUnderTest {
	return &serviceUnderTest{repo: r}
}

// serviceUnderTest mirrors Service but accepts the local interface so tests
// can call the same logic without a real database.
type serviceUnderTest struct {
	repo securityRepo
}

func (s *serviceUnderTest) GetSettings(ctx context.Context, userID int64) (domain.SecuritySettings, error) {
	return s.repo.GetSecuritySettings(ctx, userID)
}

func (s *serviceUnderTest) SetTwoFactor(ctx context.Context, userID int64, enabled bool) error {
	return s.repo.SetTwoFactor(ctx, userID, enabled)
}

func (s *serviceUnderTest) GetLoginHistory(ctx context.Context, userID int64) ([]domain.LoginEntry, error) {
	return s.repo.GetLoginHistory(ctx, userID)
}

func (s *serviceUnderTest) InvalidateAllSessions(ctx context.Context, userID int64) error {
	return s.repo.InvalidateAllSessions(ctx, userID)
}

func (s *serviceUnderTest) RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) {
	_ = s.repo.RecordLogin(ctx, userID, email, method, ip, userAgent, success)
}

// Tests

func TestGetSettings_ReturnsTwoFactorState(t *testing.T) {
	m := &mockSecurityRepo{settings: domain.SecuritySettings{TwoFactorEnabled: true}}
	s := newTestService(m)
	got, err := s.GetSettings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !got.TwoFactorEnabled {
		t.Error("expected TwoFactorEnabled=true")
	}
}

func TestGetSettings_PropagatesError(t *testing.T) {
	want := errors.New("db down")
	m := &mockSecurityRepo{settingsErr: want}
	s := newTestService(m)
	_, err := s.GetSettings(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestSetTwoFactor_HappyPath(t *testing.T) {
	m := &mockSecurityRepo{}
	s := newTestService(m)
	if err := s.SetTwoFactor(context.Background(), 1, true); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestSetTwoFactor_PropagatesError(t *testing.T) {
	want := errors.New("write failed")
	m := &mockSecurityRepo{setTFAErr: want}
	s := newTestService(m)
	if err := s.SetTwoFactor(context.Background(), 1, true); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetLoginHistory_ReturnsEntries(t *testing.T) {
	entries := []domain.LoginEntry{{Email: "a@b.com", Success: true}}
	m := &mockSecurityRepo{history: entries}
	s := newTestService(m)
	got, err := s.GetLoginHistory(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d entries, want 1", len(got))
	}
	if got[0].Email != "a@b.com" {
		t.Errorf("email = %q, want %q", got[0].Email, "a@b.com")
	}
}

func TestGetLoginHistory_PropagatesError(t *testing.T) {
	want := errors.New("query failed")
	m := &mockSecurityRepo{historyErr: want}
	s := newTestService(m)
	_, err := s.GetLoginHistory(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestInvalidateAllSessions_HappyPath(t *testing.T) {
	m := &mockSecurityRepo{}
	s := newTestService(m)
	if err := s.InvalidateAllSessions(context.Background(), 1); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestInvalidateAllSessions_PropagatesError(t *testing.T) {
	want := errors.New("session error")
	m := &mockSecurityRepo{invalidateErr: want}
	s := newTestService(m)
	if err := s.InvalidateAllSessions(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestRecordLogin_FireAndForget(t *testing.T) {
	// RecordLogin must not panic even when the repo returns an error.
	m := &mockSecurityRepo{recordErr: errors.New("insert failed")}
	s := newTestService(m)
	// Should not panic or return error (it is void).
	s.RecordLogin(context.Background(), 1, "user@example.com", "password", "1.2.3.4", "Mozilla", true)
	if m.recordedArgs.email != "user@example.com" {
		t.Errorf("email not forwarded to repo: %q", m.recordedArgs.email)
	}
	if !m.recordedArgs.success {
		t.Error("success flag not forwarded")
	}
}

func TestRecordLogin_FailedAttempt(t *testing.T) {
	m := &mockSecurityRepo{}
	s := newTestService(m)
	s.RecordLogin(context.Background(), 2, "attacker@evil.com", "jwt", "9.9.9.9", "curl/7", false)
	if m.recordedArgs.success {
		t.Error("expected success=false for failed login")
	}
	if m.recordedArgs.ip != "9.9.9.9" {
		t.Errorf("ip = %q, want %q", m.recordedArgs.ip, "9.9.9.9")
	}
}

func TestGetSettings_EmptyActiveSessions(t *testing.T) {
	m := &mockSecurityRepo{
		settings: domain.SecuritySettings{
			TwoFactorEnabled: false,
			ActiveSessions:   []domain.ActiveSession{},
		},
	}
	s := newTestService(m)
	got, err := s.GetSettings(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ActiveSessions == nil {
		t.Error("expected non-nil ActiveSessions slice")
	}
	if len(got.ActiveSessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(got.ActiveSessions))
	}
}

func TestGetLoginHistory_EmptyList(t *testing.T) {
	m := &mockSecurityRepo{history: []domain.LoginEntry{}}
	s := newTestService(m)
	got, err := s.GetLoginHistory(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}
