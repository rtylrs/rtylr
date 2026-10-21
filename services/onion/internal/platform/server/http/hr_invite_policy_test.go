package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
)

// hrMemberResolver is an ordinary invited employee: HR is licensed and assigned,
// so appForPath's HasAppAccess check passes, but no employee-management
// permission is granted. It deliberately does not implement
// ProvisioningStatusResolver, so the MFA/provisioning gate is skipped and the
// test isolates the feature-permission decision.
type hrMemberResolver struct {
	permissions []string
}

func (r hrMemberResolver) GetEffectiveAccess(_ context.Context, businessID, _ int64) (*authorization.EffectiveAccess, error) {
	return &authorization.EffectiveAccess{
		BusinessID: businessID,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppHR,
			Licensed:    true,
			HasAccess:   true,
			Permissions: r.permissions,
		}},
	}, nil
}

func hrMemberRequest(t *testing.T, method, path string, permissions []string, sendToken bool) *httptest.ResponseRecorder {
	t.Helper()

	handler := GetAuthMiddlewareFunc(
		nil,
		stubMFAValidatorResolver{validator: stubMFAValidator{userID: 42}},
		hrMemberResolver{permissions: permissions},
	)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(method, path, nil)
	if sendToken {
		request.Header.Set(AuthHeaderKey, "Bearer token")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	return recorder
}

// An invited employee holds no hr:employees:* permission. Both invite routes
// used to be gated on one, so the middleware refused the person the invite was
// addressed to before the handler's own identity and email checks could run.
func TestHRInviteAcceptAndVerifyReachAUserWithoutEmployeePermissions(t *testing.T) {

	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/hr/employees/invite/accept"},
		{http.MethodGet, "/hr/employees/invite/verify"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			if appID, permission, ok := featurePermissionForRequest(tc.method, tc.path); ok {
				t.Fatalf("%q still maps to %q:%q", tc.path, appID, permission)
			}
			if got := hrMemberRequest(t, tc.method, tc.path, []string{"hr:attendance:self"}, true).Code; got != http.StatusOK {
				t.Fatalf("%s %s as an invited employee: status = %d, want %d", tc.method, tc.path, got, http.StatusOK)
			}
		})
	}
}

// The exemption is scoped to accepting and verifying. Issuing an invite names
// another person's account and must stay behind hr:employees:manage, as must
// every other /hr/employees route and any longer path under the exempt ones.
func TestHRInviteExemptionDoesNotCoverIssuingOrLongerPaths(t *testing.T) {

	for _, tc := range []struct {
		name       string
		method     string
		path       string
		permission string
	}{
		{name: "issue invite", method: http.MethodPost, path: "/hr/employees/invite", permission: "hr:employees:manage"},
		{name: "accept sub-path", method: http.MethodPost, path: "/hr/employees/invite/accept/anything", permission: "hr:employees:manage"},
		{name: "verify sub-path", method: http.MethodGet, path: "/hr/employees/invite/verify/anything", permission: "hr:employees:view"},
		{name: "accept on GET", method: http.MethodGet, path: "/hr/employees/invite/accept", permission: "hr:employees:view"},
		{name: "verify on POST", method: http.MethodPost, path: "/hr/employees/invite/verify", permission: "hr:employees:manage"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appID, permission, ok := featurePermissionForRequest(tc.method, tc.path)
			if !ok || appID != authorization.AppHR || permission != tc.permission {
				t.Fatalf("mapping = %q:%q,%t, want hr:%s,true", appID, permission, ok, tc.permission)
			}
			if got := hrMemberRequest(t, tc.method, tc.path, []string{"hr:attendance:self"}, true).Code; got != http.StatusForbidden {
				t.Fatalf("%s %s without %s: status = %d, want %d", tc.method, tc.path, tc.permission, got, http.StatusForbidden)
			}
		})
	}
}

// The exemption only removes the permission check; both routes still run behind
// the bearer gate, so an anonymous caller never reaches the handler at all.
func TestHRInviteRoutesStillRequireAToken(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/hr/employees/invite/accept"},
		{http.MethodGet, "/hr/employees/invite/verify"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			if got := hrMemberRequest(t, tc.method, tc.path, nil, false).Code; got != http.StatusUnauthorized {
				t.Fatalf("anonymous %s %s: status = %d, want %d", tc.method, tc.path, got, http.StatusUnauthorized)
			}
		})
	}
}

// The two invite routes take a token as their only credential, so they are
// still a guessing surface now that they sit behind the bearer gate.
func TestHRInviteRoutesAreRateLimited(t *testing.T) {
	for _, path := range []string{"/hr/employees/invite/accept", "/hr/employees/invite/verify"} {
		t.Run(path, func(t *testing.T) {
			allowed, throttledAt := hitBurst(t, path, 10)
			if allowed != 5 || throttledAt != 6 {
				t.Fatalf("%s: allowed %d, first 429 at %d; want 5 allowed and 429 at 6", path, allowed, throttledAt)
			}
		})
	}
}
