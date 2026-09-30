package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

// stubMFAValidator issues a token for one fixed user.
type stubMFAValidator struct {
	userID int64
}

func (s stubMFAValidator) ValidateToken(context.Context, string) (int64, error) {
	return s.userID, nil
}

func (s stubMFAValidator) ValidateTokenFull(context.Context, string) (*authPkg.TokenClaims, error) {
	return &authPkg.TokenClaims{UserID: s.userID, BusinessID: 7, Role: "owner"}, nil
}

type stubMFAValidatorResolver struct {
	validator authPkg.TokenValidator
}

func (r stubMFAValidatorResolver) ResolveTokenValidator(context.Context, authPkg.ProductPlatform) authPkg.TokenValidator {
	return r.validator
}

// stubMFAAccessResolver carries the provisioning state the MFA gate reads.
type stubMFAAccessResolver struct {
	status    *authorization.ProvisioningStatus
	verified  bool
	verifyErr error
}

func (s *stubMFAAccessResolver) GetEffectiveAccess(context.Context, int64, int64) (*authorization.EffectiveAccess, error) {
	return &authorization.EffectiveAccess{
		BusinessID:      7,
		DashRole:        "owner",
		DashPermissions: []string{"hub:security:manage"},
		IsOwner:         true,
	}, nil
}

func (s *stubMFAAccessResolver) GetProvisioningStatus(context.Context, int64) (*authorization.ProvisioningStatus, error) {
	return s.status, nil
}

func (s *stubMFAAccessResolver) IsMFAVerified(context.Context, int64) (bool, error) {
	return s.verified, s.verifyErr
}

func mfaGateRequest(t *testing.T, method, path string, resolver *stubMFAAccessResolver) *httptest.ResponseRecorder {
	t.Helper()
	handler := GetAuthMiddlewareFunc(
		nil,
		stubMFAValidatorResolver{validator: stubMFAValidator{userID: 42}},
		resolver,
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(method, path, nil)
	request.Header.Set(AuthHeaderKey, "Bearer token")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func enrolledMFAStatus() *authorization.ProvisioningStatus {
	return &authorization.ProvisioningStatus{
		State:         authorization.ProvisioningActive,
		EmailVerified: true,
		MFARequired:   true,
		MFAEnrolled:   true,
	}
}

// A stolen bearer token must not reach a security-setting mutation while the
// factor is enrolled but unproved. Before the fix isMFAPath exempted
// PATCH /hub/security/v1/2fa and the two enrollment endpoints from this gate,
// so all three returned 200 with nothing but the token.
func TestMFAGateBlocksSecuritySettingMutationsWhenEnrolled(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{name: "disable two-factor", method: http.MethodPatch, path: "/hub/security/v1/2fa"},
		{name: "enable two-factor", method: http.MethodPatch, path: "/hub/security/v1/2fa"},
		{name: "invalidate sessions", method: http.MethodPost, path: "/hub/security/v1/invalidate-sessions"},
		{name: "begin enrollment", method: http.MethodPost, path: "/hub/security/v1/mfa/enroll"},
		{name: "complete enrollment", method: http.MethodPost, path: "/hub/security/v1/mfa/enroll/confirm"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolver := &stubMFAAccessResolver{status: enrolledMFAStatus(), verified: false}
			if got := mfaGateRequest(t, tc.method, tc.path, resolver).Code; got != http.StatusUnauthorized {
				t.Fatalf("%s %s with an unproved factor: status = %d, want %d", tc.method, tc.path, got, http.StatusUnauthorized)
			}
		})
	}
}

// The proof endpoints and the settings read stay reachable, otherwise a user
// with an unproved factor could never satisfy the gate.
func TestMFAGateKeepsProofAndReadPathsReachable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{name: "security state read", method: http.MethodGet, path: "/hub/security/v1/state"},
		{name: "verify code", method: http.MethodPost, path: "/hub/security/v1/mfa/verify"},
		{name: "use recovery code", method: http.MethodPost, path: "/hub/security/v1/mfa/recovery"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolver := &stubMFAAccessResolver{status: enrolledMFAStatus(), verified: false}
			if got := mfaGateRequest(t, tc.method, tc.path, resolver).Code; got != http.StatusOK {
				t.Fatalf("%s %s: status = %d, want %d", tc.method, tc.path, got, http.StatusOK)
			}
		})
	}
}

// A user who is required to enroll but has not yet is the one case where the
// enrollment endpoints must bypass the gate, otherwise the account can never
// leave that state.
func TestMFAGateKeepsEnrollmentReachableWhenNotEnrolled(t *testing.T) {
	for _, path := range []string{"/hub/security/v1/mfa/enroll", "/hub/security/v1/mfa/enroll/confirm"} {
		t.Run(path, func(t *testing.T) {
			resolver := &stubMFAAccessResolver{status: &authorization.ProvisioningStatus{
				State:         authorization.ProvisioningActive,
				EmailVerified: true,
				MFARequired:   true,
				MFAEnrolled:   false,
			}}
			if got := mfaGateRequest(t, http.MethodPost, path, resolver).Code; got != http.StatusOK {
				t.Fatalf("POST %s before enrollment: status = %d, want %d", path, got, http.StatusOK)
			}
		})
	}
	// The setting mutation is still refused: there is nothing to change yet.
	resolver := &stubMFAAccessResolver{status: &authorization.ProvisioningStatus{
		State:         authorization.ProvisioningActive,
		EmailVerified: true,
		MFARequired:   true,
		MFAEnrolled:   false,
	}}
	if got := mfaGateRequest(t, http.MethodPatch, "/hub/security/v1/2fa", resolver).Code; got != http.StatusUnauthorized {
		t.Fatalf("PATCH /hub/security/v1/2fa before enrollment: status = %d, want %d", got, http.StatusUnauthorized)
	}
}

// Once the factor is proved the same mutations are reachable again.
func TestMFAGateAllowsMutationsAfterVerification(t *testing.T) {
	for _, path := range []string{"/hub/security/v1/2fa", "/hub/security/v1/mfa/enroll", "/hub/security/v1/invalidate-sessions"} {
		t.Run(path, func(t *testing.T) {
			resolver := &stubMFAAccessResolver{status: enrolledMFAStatus(), verified: true}
			method := http.MethodPost
			if path == "/hub/security/v1/2fa" {
				method = http.MethodPatch
			}
			if got := mfaGateRequest(t, method, path, resolver).Code; got != http.StatusOK {
				t.Fatalf("%s %s after verification: status = %d, want %d", method, path, got, http.StatusOK)
			}
		})
	}
}

// An unreadable MFA-verified state denies rather than permitting.
func TestMFAGateDeniesWhenMFAStateCannotBeRead(t *testing.T) {
	resolver := &stubMFAAccessResolver{status: enrolledMFAStatus(), verifyErr: errors.New("database down")}
	if got := mfaGateRequest(t, http.MethodGet, "/hub/security/v1/state", resolver).Code; got != http.StatusOK {
		t.Fatalf("reads must not depend on MFA verification state, got %d", got)
	}
	if got := mfaGateRequest(t, http.MethodPatch, "/hub/security/v1/2fa", resolver).Code; got != http.StatusUnauthorized {
		t.Fatalf("PATCH /hub/security/v1/2fa with unreadable MFA state: status = %d, want %d", got, http.StatusUnauthorized)
	}
}

func TestIsMFAPathSplitsReadsFromMutations(t *testing.T) {
	cases := []struct {
		path        string
		mfaEnrolled bool
		want        bool
	}{
		{path: "/hub/security/v1/state", mfaEnrolled: false, want: true},
		{path: "/hub/security/v1/state", mfaEnrolled: true, want: true},
		{path: "/hub/security/v1/mfa/verify", mfaEnrolled: true, want: true},
		{path: "/hub/security/v1/mfa/recovery", mfaEnrolled: true, want: true},
		{path: "/hub/security/v1/mfa/enroll", mfaEnrolled: false, want: true},
		{path: "/hub/security/v1/mfa/enroll/confirm", mfaEnrolled: false, want: true},
		{path: "/hub/security/v1/mfa/enroll", mfaEnrolled: true, want: false},
		{path: "/hub/security/v1/mfa/enroll/confirm", mfaEnrolled: true, want: false},
		{path: "/hub/security/v1/2fa", mfaEnrolled: false, want: false},
		{path: "/hub/security/v1/2fa", mfaEnrolled: true, want: false},
		{path: "/hub/security/v1/invalidate-sessions", mfaEnrolled: true, want: false},
		{path: "/hub/security/v1/organization/mfa", mfaEnrolled: true, want: false},
		{path: "/hub/security/v1/login-history", mfaEnrolled: true, want: false},
		{path: "/hub/team/v1/users", mfaEnrolled: true, want: false},
	}
	for _, tc := range cases {
		request := httptest.NewRequest(http.MethodPost, tc.path, nil)
		if got := isMFAPath(request, tc.mfaEnrolled); got != tc.want {
			t.Errorf("isMFAPath(%q, enrolled=%v) = %v, want %v", tc.path, tc.mfaEnrolled, got, tc.want)
		}
	}
}
