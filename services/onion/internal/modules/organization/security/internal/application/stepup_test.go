package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" // #nosec G505 -- RFC 6238 TOTP interoperability requires SHA-1.
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/security/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/security/internal/infrastructure"
)

// fakeSecurityRepo records whether a guarded write was reached.
type fakeSecurityRepo struct {
	factorState    domain.MFAFactorState
	factorStateErr error
	pendingSecret  []byte
	markedVerified bool

	setTwoFactorCalls  int
	beginCalls         int
	completeCalls      int
	invalidateCalls    int
	recordedTwoFactor  bool
	recordedVerifyFlag bool
}

func (f *fakeSecurityRepo) GetSecuritySettings(context.Context, int64) (domain.SecuritySettings, error) {
	return domain.SecuritySettings{}, nil
}

func (f *fakeSecurityRepo) SetTwoFactor(_ context.Context, _ int64, enabled bool) error {
	f.setTwoFactorCalls++
	f.recordedTwoFactor = enabled
	return nil
}

func (f *fakeSecurityRepo) GetLoginHistory(context.Context, int64, int, int) ([]domain.LoginEntry, int64, error) {
	return nil, 0, nil
}

func (f *fakeSecurityRepo) InvalidateAllSessions(context.Context, int64) error {
	f.invalidateCalls++
	return nil
}

func (f *fakeSecurityRepo) RecordLogin(context.Context, int64, string, string, string, string, bool) error {
	return nil
}

func (f *fakeSecurityRepo) BeginMFAEnrollment(context.Context, int64, []byte) error {
	f.beginCalls++
	return nil
}

func (f *fakeSecurityRepo) GetMFASecret(context.Context, int64, bool) ([]byte, error) {
	if f.pendingSecret == nil {
		return nil, errors.New("no rows")
	}
	return f.pendingSecret, nil
}

func (f *fakeSecurityRepo) CompleteMFAEnrollment(context.Context, int64, [][]byte) error {
	f.completeCalls++
	return nil
}

func (f *fakeSecurityRepo) MarkMFAVerified(context.Context, int64) error {
	f.markedVerified = true
	f.recordedVerifyFlag = true
	return nil
}

func (f *fakeSecurityRepo) GetMFAFactorState(context.Context, int64) (domain.MFAFactorState, error) {
	if f.factorStateErr != nil {
		return domain.MFAFactorState{}, f.factorStateErr
	}
	return f.factorState, nil
}

func (f *fakeSecurityRepo) ListRecoveryCodes(context.Context, int64) ([]infrastructure.RecoveryCodeRow, error) {
	return nil, nil
}

func (f *fakeSecurityRepo) ConsumeRecoveryCode(context.Context, int64, int64) error {
	return nil
}

func (f *fakeSecurityRepo) GetOrganizationMFAEnforced(context.Context, int64) (bool, error) {
	return false, nil
}

func (f *fakeSecurityRepo) SetOrganizationMFAEnforced(context.Context, int64, bool) error {
	return nil
}

func serviceWithRepo(repo repository) *Service {
	return &Service{repo: repo}
}

func enrolledAgo(d time.Duration) domain.MFAFactorState {
	verified := time.Now().Add(-d)
	return domain.MFAFactorState{FactorEnrolled: true, LastVerifiedAt: &verified}
}

// A stolen bearer token alone must not be enough to turn two-factor off.
func TestSetTwoFactorDeniedWithoutStepUp(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state domain.MFAFactorState
	}{
		{name: "never verified", state: domain.MFAFactorState{FactorEnrolled: true}},
		{name: "verified just outside the window", state: enrolledAgo(stepUpWindow + time.Minute)},
		{name: "timestamp in the future", state: domain.MFAFactorState{
			FactorEnrolled: true,
			LastVerifiedAt: ptrTime(time.Now().Add(10 * time.Minute)),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeSecurityRepo{factorState: tc.state}
			service := serviceWithRepo(repo)
			err := service.SetTwoFactor(context.Background(), 42, false)
			if !errors.Is(err, ErrStepUpRequired) {
				t.Fatalf("SetTwoFactor(enabled=false) error = %v, want ErrStepUpRequired", err)
			}
			if repo.setTwoFactorCalls != 0 {
				t.Fatalf("two-factor state was written %d time(s) without step-up", repo.setTwoFactorCalls)
			}
		})
	}
}

func TestSetTwoFactorAllowedWithFreshStepUp(t *testing.T) {
	repo := &fakeSecurityRepo{factorState: enrolledAgo(30 * time.Second)}
	if err := serviceWithRepo(repo).SetTwoFactor(context.Background(), 42, false); err != nil {
		t.Fatalf("SetTwoFactor after a fresh verification: %v", err)
	}
	if repo.setTwoFactorCalls != 1 || repo.recordedTwoFactor {
		t.Fatalf("expected the disable to be written once, calls=%d enabled=%v", repo.setTwoFactorCalls, repo.recordedTwoFactor)
	}
}

// A user with no factor must still be able to opt in, otherwise no account can
// ever secure itself.
func TestSetTwoFactorAllowedWithoutEnrolledFactor(t *testing.T) {
	repo := &fakeSecurityRepo{factorState: domain.MFAFactorState{}}
	if err := serviceWithRepo(repo).SetTwoFactor(context.Background(), 42, true); err != nil {
		t.Fatalf("SetTwoFactor(enabled=true) without an enrolled factor: %v", err)
	}
	if repo.setTwoFactorCalls != 1 || !repo.recordedTwoFactor {
		t.Fatal("expected the enable to be written")
	}
}

// A state read that cannot be completed denies; it never reads as "no factor".
func TestStepUpDeniesWhenFactorStateCannotBeRead(t *testing.T) {
	t.Setenv("MFA_ENCRYPTION_KEY", "test-step-up-key")
	repo := &fakeSecurityRepo{factorStateErr: errors.New("connection reset")}
	service := serviceWithRepo(repo)

	checks := map[string]func() error{
		"SetTwoFactor": func() error { return service.SetTwoFactor(context.Background(), 42, false) },
		"InvalidateAllSessions": func() error {
			return service.InvalidateAllSessions(context.Background(), 42)
		},
		"BeginMFAEnrollment": func() error {
			_, _, err := service.BeginMFAEnrollment(context.Background(), 42)
			return err
		},
		"CompleteMFAEnrollment": func() error {
			_, err := service.CompleteMFAEnrollment(context.Background(), 42, "123456")
			return err
		},
	}
	for name, call := range checks {
		if err := call(); !errors.Is(err, ErrStepUpUnavailable) {
			t.Errorf("%s error = %v, want ErrStepUpUnavailable", name, err)
		}
	}
	if repo.setTwoFactorCalls != 0 || repo.invalidateCalls != 0 || repo.beginCalls != 0 || repo.completeCalls != 0 {
		t.Fatal("a state read failure still reached a write")
	}
}

// Rotating the second factor is the core of the exploit: enroll, confirm with
// the returned secret, and the attacker's TOTP becomes the active factor.
func TestBeginMFAEnrollmentDeniedWithoutStepUpWhenEnrolled(t *testing.T) {
	t.Setenv("MFA_ENCRYPTION_KEY", "test-step-up-key")
	repo := &fakeSecurityRepo{factorState: domain.MFAFactorState{FactorEnrolled: true}}
	secret, uri, err := serviceWithRepo(repo).BeginMFAEnrollment(context.Background(), 42)
	if !errors.Is(err, ErrStepUpRequired) {
		t.Fatalf("BeginMFAEnrollment error = %v, want ErrStepUpRequired", err)
	}
	if secret != "" || uri != "" {
		t.Fatal("a TOTP secret was returned without step-up")
	}
	if repo.beginCalls != 0 {
		t.Fatal("a pending enrollment was written without step-up")
	}
}

func TestCompleteMFAEnrollmentDeniedWithoutStepUpWhenEnrolled(t *testing.T) {
	t.Setenv("MFA_ENCRYPTION_KEY", "test-step-up-key")
	repo := &fakeSecurityRepo{factorState: domain.MFAFactorState{FactorEnrolled: true}, pendingSecret: []byte("unused")}
	if _, err := serviceWithRepo(repo).CompleteMFAEnrollment(context.Background(), 42, "123456"); !errors.Is(err, ErrStepUpRequired) {
		t.Fatalf("CompleteMFAEnrollment error = %v, want ErrStepUpRequired", err)
	}
	if repo.completeCalls != 0 {
		t.Fatal("the factor was activated without step-up")
	}
}

func TestInvalidateSessionsDeniedWithoutStepUp(t *testing.T) {
	repo := &fakeSecurityRepo{factorState: domain.MFAFactorState{FactorEnrolled: true}}
	if err := serviceWithRepo(repo).InvalidateAllSessions(context.Background(), 42); !errors.Is(err, ErrStepUpRequired) {
		t.Fatalf("InvalidateAllSessions error = %v, want ErrStepUpRequired", err)
	}
	if repo.invalidateCalls != 0 {
		t.Fatal("sessions were revoked without step-up")
	}
}

func TestInvalidateSessionsAllowedWithFreshStepUp(t *testing.T) {
	repo := &fakeSecurityRepo{factorState: enrolledAgo(time.Minute)}
	if err := serviceWithRepo(repo).InvalidateAllSessions(context.Background(), 42); err != nil {
		t.Fatalf("InvalidateAllSessions after a fresh verification: %v", err)
	}
	if repo.invalidateCalls != 1 {
		t.Fatalf("invalidate calls = %d, want 1", repo.invalidateCalls)
	}
}

// First enrollment stays open, and the secret it returns only becomes a factor
// when the caller proves possession of it.
func TestFirstEnrollmentCompletesWithACodeFromTheReturnedSecret(t *testing.T) {
	t.Setenv("MFA_ENCRYPTION_KEY", "test-step-up-key")
	repo := &fakeSecurityRepo{factorState: domain.MFAFactorState{}}
	service := serviceWithRepo(repo)

	secret, uri, err := service.BeginMFAEnrollment(context.Background(), 42)
	if err != nil {
		t.Fatalf("BeginMFAEnrollment: %v", err)
	}
	if secret == "" || !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("unexpected enrollment response: secret=%q uri=%q", secret, uri)
	}

	stored, err := sealMFASecret(secret)
	if err != nil {
		t.Fatalf("sealMFASecret: %v", err)
	}
	repo.pendingSecret = []byte(stored)

	codes, err := service.CompleteMFAEnrollment(context.Background(), 42, totpForTest(secret, time.Now()))
	if err != nil {
		t.Fatalf("CompleteMFAEnrollment: %v", err)
	}
	if len(codes) != 8 {
		t.Fatalf("recovery codes = %d, want 8", len(codes))
	}
	if repo.completeCalls != 1 {
		t.Fatalf("complete calls = %d, want 1", repo.completeCalls)
	}
}

// A 6-digit code on an unrated route is a few percent per 30 seconds, and a
// success opens the five-minute verified window that step-up trusts.
func TestVerifyMFAStopsAcceptingCodesAfterRepeatedFailures(t *testing.T) {
	t.Setenv("MFA_ENCRYPTION_KEY", "test-step-up-key")
	repo := &fakeSecurityRepo{factorState: enrolledAgo(time.Minute)}
	service := serviceWithRepo(repo)
	mfaAttempts.reset(4242)

	for i := 0; i < mfaAttemptLimit; i++ {
		if err := service.VerifyMFA(context.Background(), 4242, "000000"); !errors.Is(err, ErrMFAInvalidCode) {
			t.Fatalf("attempt %d: error = %v, want ErrMFAInvalidCode", i+1, err)
		}
	}
	err := service.VerifyMFA(context.Background(), 4242, "000000")
	if !errors.Is(err, ErrMFAAttemptsExceeded) {
		t.Fatalf("error after %d failures = %v, want ErrMFAAttemptsExceeded", mfaAttemptLimit, err)
	}
}

func TestVerifyMFASuccessClearsAttemptHistory(t *testing.T) {
	t.Setenv("MFA_ENCRYPTION_KEY", "test-step-up-key")
	secret, err := randomTestSecret()
	if err != nil {
		t.Fatalf("randomTestSecret: %v", err)
	}
	stored, err := sealMFASecret(secret)
	if err != nil {
		t.Fatalf("sealMFASecret: %v", err)
	}
	repo := &fakeSecurityRepo{factorState: enrolledAgo(time.Minute), pendingSecret: []byte(stored)}
	service := serviceWithRepo(repo)
	mfaAttempts.reset(4243)

	for i := 0; i < mfaAttemptLimit-1; i++ {
		_ = service.VerifyMFA(context.Background(), 4243, "000000")
	}
	if err := service.VerifyMFA(context.Background(), 4243, totpForTest(secret, time.Now())); err != nil {
		t.Fatalf("VerifyMFA with a valid code: %v", err)
	}
	if !repo.recordedVerifyFlag {
		t.Fatal("a valid code did not mark the session verified")
	}
	if err := mfaAttempts.check(4243, time.Now()); err != nil {
		t.Fatalf("a successful verification did not reset the attempt history: %v", err)
	}
}

func TestMFAAttemptLimiterIsolatesUsersAndExpires(t *testing.T) {
	limiter := &mfaAttemptLimiter{failures: make(map[int64][]time.Time)}
	now := time.Now()
	for i := 0; i < mfaAttemptLimit; i++ {
		limiter.recordFailure(1, now)
	}
	if err := limiter.check(1, now); !errors.Is(err, ErrMFAAttemptsExceeded) {
		t.Fatalf("locked out user 1: %v", err)
	}
	if err := limiter.check(2, now); err != nil {
		t.Fatalf("user 2 must have its own budget, got %v", err)
	}
	if err := limiter.check(1, now.Add(mfaAttemptWindow+time.Second)); err != nil {
		t.Fatalf("the window did not expire, got %v", err)
	}
	limiter.reset(1)
	if err := limiter.check(1, now); err != nil {
		t.Fatalf("reset did not clear the budget, got %v", err)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func randomTestSecret() (string, error) {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("01234567890123456789")), nil
}

func totpForTest(secret string, now time.Time) string {
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return ""
	}
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], uint64(now.Unix()/30)) // #nosec G115 -- test counter.
	// #nosec G505 -- RFC 6238 TOTP interoperability requires SHA-1.
	digest := hmac.New(sha1.New, decoded)
	_, _ = digest.Write(message[:])
	sum := digest.Sum(nil)
	index := sum[len(sum)-1] & 0x0f
	value := (uint32(sum[index])&0x7f)<<24 |
		uint32(sum[index+1])<<16 |
		uint32(sum[index+2])<<8 |
		uint32(sum[index+3])
	return fmt.Sprintf("%06d", value%1000000)
}
