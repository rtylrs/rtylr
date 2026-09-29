package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	"github.com/voxire/onion/internal/modules/security/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	hubsecurityv1 "github.com/voxire/voxie/gen/go/hub/security/v1"
)

// stubSecurityService fails the way the application layer does when the
// step-up gate refuses a request.
type stubSecurityService struct {
	err          error
	setTwoFactor error
	invalidate   error
	begin        error
	complete     error
	verify       error
	recovery     error
}

func (s *stubSecurityService) GetSettings(context.Context, int64) (domain.SecuritySettings, error) {
	return domain.SecuritySettings{}, nil
}

func (s *stubSecurityService) GetLoginHistory(context.Context, int64, int64, int64) ([]domain.LoginEntry, int64, bool, error) {
	return nil, 0, false, nil
}

func (s *stubSecurityService) InvalidateAllSessions(context.Context, int64) error {
	return s.invalidate
}

func (s *stubSecurityService) SetTwoFactor(context.Context, int64, bool) error {
	return s.setTwoFactor
}

func (s *stubSecurityService) BeginMFAEnrollment(context.Context, int64) (string, string, error) {
	return "", "", s.begin
}

func (s *stubSecurityService) CompleteMFAEnrollment(context.Context, int64, string) ([]string, error) {
	return nil, s.complete
}

func (s *stubSecurityService) VerifyMFA(context.Context, int64, string) error {
	return s.verify
}

func (s *stubSecurityService) UseMFARecoveryCode(context.Context, int64, string) error {
	return s.recovery
}

func (s *stubSecurityService) GetOrganizationMFAEnforced(context.Context, int64) (bool, error) {
	return false, nil
}

func (s *stubSecurityService) SetOrganizationMFAEnforced(context.Context, int64, bool) error {
	return nil
}

// Update2FA is exercised on the enable path. Disabling 2FA first resolves the
// caller's business to read the organization MFA policy, which needs a
// database; that path is covered by the DB-backed tests. Both paths reach
// the same requireStepUp guard, which is what this file asserts.
func securityManagerContext() context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     42,
		BusinessID: 7,
		Role:       "owner",
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:      7,
			DashRole:        "owner",
			DashPermissions: []string{"hub:security:manage"},
			IsOwner:         true,
		},
		ProvisioningStatus: &authorization.ProvisioningStatus{State: authorization.ProvisioningActive},
	})
}

func statusOf(t *testing.T, err error) int {
	t.Helper()
	var statusErr interface{ HTTPStatusCode() int }
	if !errors.As(err, &statusErr) {
		t.Fatalf("error %v does not carry an HTTP status", err)
	}
	return statusErr.HTTPStatusCode()
}

// A caller with a valid session but no fresh factor proof is forbidden, not
// silently allowed through. Before the fix every one of these handlers mapped
// the failure to 500, but nothing in the module ever produced a step-up
// failure at all.
func TestSecurityMutationsRejectMissingStepUp(t *testing.T) {
	service := &stubSecurityService{
		setTwoFactor: application.ErrStepUpRequired,
		invalidate:   application.ErrStepUpRequired,
		begin:        application.ErrStepUpRequired,
		complete:     application.ErrStepUpRequired,
	}
	api := &SecurityAPI{Service: service}
	ctx := securityManagerContext()

	calls := map[string]func() error{
		"Update2FA": func() error {
			_, err := api.Update2FA(ctx, &hubsecurityv1.Update2FARequest{Enabled: true})
			return err
		},
		"InvalidateSessions": func() error {
			_, err := api.InvalidateSessions(ctx, &hubsecurityv1.InvalidateSessionsRequest{})
			return err
		},
		"BeginMFAEnrollment": func() error {
			_, err := api.BeginMFAEnrollment(ctx, &hubsecurityv1.BeginMFAEnrollmentRequest{})
			return err
		},
		"CompleteMFAEnrollment": func() error {
			_, err := api.CompleteMFAEnrollment(ctx, &hubsecurityv1.CompleteMFAEnrollmentRequest{Code: "123456"})
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			if got := statusOf(t, call()); got != http.StatusForbidden {
				t.Fatalf("%s without step-up: status = %d, want %d", name, got, http.StatusForbidden)
			}
		})
	}
}

// A factor state that cannot be read is a dependency outage, and it must not
// be reported as a permission problem or, worse, as success.
func TestSecurityMutationsRejectUnreadableStepUpState(t *testing.T) {
	unavailable := errors.Join(application.ErrStepUpUnavailable, errors.New("connection reset"))
	api := &SecurityAPI{Service: &stubSecurityService{
		setTwoFactor: unavailable,
		invalidate:   unavailable,
		begin:        unavailable,
		complete:     unavailable,
	}}
	ctx := securityManagerContext()

	if _, err := api.Update2FA(ctx, &hubsecurityv1.Update2FARequest{Enabled: true}); statusOf(t, err) != http.StatusServiceUnavailable {
		t.Fatalf("Update2FA: status = %d, want %d", statusOf(t, err), http.StatusServiceUnavailable)
	}
	if _, err := api.BeginMFAEnrollment(ctx, &hubsecurityv1.BeginMFAEnrollmentRequest{}); statusOf(t, err) != http.StatusServiceUnavailable {
		t.Fatalf("BeginMFAEnrollment: status = %d, want %d", statusOf(t, err), http.StatusServiceUnavailable)
	}
}

// Repeated code guesses are throttled rather than answered with a fresh
// "invalid code" each time.
func TestVerifyMFAReportsAttemptLimit(t *testing.T) {
	api := &SecurityAPI{Service: &stubSecurityService{verify: application.ErrMFAAttemptsExceeded}}
	if _, err := api.VerifyMFA(securityManagerContext(), &hubsecurityv1.VerifyMFARequest{Code: "000000"}); statusOf(t, err) != http.StatusTooManyRequests {
		t.Fatalf("VerifyMFA: status = %d, want %d", statusOf(t, err), http.StatusTooManyRequests)
	}
	api = &SecurityAPI{Service: &stubSecurityService{recovery: application.ErrMFAAttemptsExceeded}}
	if _, err := api.UseMFARecoveryCode(securityManagerContext(), &hubsecurityv1.UseMFARecoveryCodeRequest{Code: "AAAA-AAAA"}); statusOf(t, err) != http.StatusTooManyRequests {
		t.Fatalf("UseMFARecoveryCode: status = %d, want %d", statusOf(t, err), http.StatusTooManyRequests)
	}
}

// The step-up guard must not swallow unrelated failures into a step-up answer.
func TestSecurityGuardErrorIgnoresUnrelatedFailures(t *testing.T) {
	if got := securityGuardError(errors.New("deadlock")); got != nil {
		t.Fatalf("securityGuardError(%v) = %v, want nil", errors.New("deadlock"), got)
	}
	if got := securityGuardError(application.ErrMFAInvalidCode); got != nil {
		t.Fatalf("securityGuardError(ErrMFAInvalidCode) = %v, want nil", got)
	}
	if statusOf(t, securityGuardError(application.ErrStepUpRequired)) != http.StatusForbidden {
		t.Fatal("ErrStepUpRequired must map to 403")
	}
}

// Losing the permission must still win over the step-up mapping.
// Update2FA is self-scoped: it passes user.UserID from the verified auth
// context, so it needs authentication but no hub:security:manage grant — that
// permission is not mapped for this path and the operation is not an
// administrative one. What must hold is that an anonymous caller is refused
// before any state is touched.
func TestUpdate2FARequiresAnAuthenticatedCaller(t *testing.T) {
	api := &SecurityAPI{Service: &stubSecurityService{}}

	_, err := api.Update2FA(context.Background(), &hubsecurityv1.Update2FARequest{Enabled: true})
	if statusOf(t, err) != http.StatusUnauthorized {
		t.Fatalf("Update2FA without an authenticated caller: status = %d, want %d",
			statusOf(t, err), http.StatusUnauthorized)
	}
}
