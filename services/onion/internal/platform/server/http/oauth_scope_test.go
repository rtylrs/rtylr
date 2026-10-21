package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

type fixedClaimsValidator struct{ claims *authPkg.TokenClaims }

func (v fixedClaimsValidator) ValidateToken(context.Context, string) (int64, error) {
	return v.claims.UserID, nil
}

func (v fixedClaimsValidator) ValidateTokenFull(context.Context, string) (*authPkg.TokenClaims, error) {
	return v.claims, nil
}

func oauthRequest(t *testing.T, claims *authPkg.TokenClaims, method, path string) (*httptest.ResponseRecorder, *authPkg.AuthUser) {
	t.Helper()
	resolver := authPkg.NewStaticTokenValidatorResolver(fixedClaimsValidator{claims: claims}, nil)
	var seen *authPkg.AuthUser
	handler := GetAuthMiddlewareFunc(NewRouteAccessPolicy(), resolver)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, _ = authPkg.GetAuthUser(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set(AuthHeaderKey, BearerPrefix+"token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec, seen
}

func TestOAuthTokenIsLimitedToGrantedScopes(t *testing.T) {
	cases := []struct {
		name   string
		scope  string
		method string
		path   string
		want   int
	}{
		{"read scope reads its app", "crm:read", http.MethodGet, "/crm/customer/v1/customers", http.StatusNoContent},
		{"read scope cannot write", "crm:read", http.MethodPost, "/crm/customer/v1/customers", http.StatusForbidden},
		{"write scope implies read", "crm:write", http.MethodGet, "/crm/customer/v1/customers", http.StatusNoContent},
		{"write scope writes", "crm:write", http.MethodDelete, "/crm/customer/v1/customers/7", http.StatusNoContent},
		{"scope for another app is refused", "flow:write", http.MethodGet, "/crm/customer/v1/customers", http.StatusForbidden},
		{"pos scope covers pos routes", "pos:write", http.MethodPost, "/pos/orders/compat", http.StatusNoContent},
		{"first party routes are unreachable", "crm:write flow:write", http.MethodGet, "/hub/user/v1/users/me", http.StatusForbidden},
		{"security settings are unreachable", "crm:write", http.MethodPost, "/hub/security/v1/mfa/disable", http.StatusForbidden},
		{"empty scope reaches nothing", "", http.MethodGet, "/crm/customer/v1/customers", http.StatusForbidden},
		{"empty scope still reads its identity", "", http.MethodGet, "/common/oauth/v1/me", http.StatusNoContent},
		{"traversal cannot borrow a scoped prefix", "crm:read", http.MethodGet, "/crm/../hub/user/v1/users/me", http.StatusForbidden},
		{"case does not bypass the app prefix", "crm:read", http.MethodGet, "/CRM/customer/v1/customers", http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := &authPkg.TokenClaims{UserID: 7, BusinessID: 3, OAuthClientID: "rtylr-mcp", Scope: tc.scope}
			rec, _ := oauthRequest(t, claims, tc.method, tc.path)
			if rec.Code != tc.want {
				t.Fatalf("%s %s with %q = %d, want %d", tc.method, tc.path, tc.scope, rec.Code, tc.want)
			}
		})
	}
}

func TestInsufficientScopeAdvertisesRequiredScope(t *testing.T) {
	claims := &authPkg.TokenClaims{UserID: 7, OAuthClientID: "rtylr-mcp", Scope: "crm:read"}
	rec, _ := oauthRequest(t, claims, http.MethodPost, "/crm/customer/v1/customers")
	if got := rec.Header().Get("WWW-Authenticate"); got != `Bearer error="insufficient_scope", scope="crm:write"` {
		t.Fatalf("WWW-Authenticate = %q", got)
	}
}

func TestSessionTokensAreNotScopeLimited(t *testing.T) {
	claims := &authPkg.TokenClaims{UserID: 7, BusinessID: 3}
	for _, path := range []string{"/hub/user/v1/users/me", "/crm/customer/v1/customers", "/finance/invoices/v1/invoices"} {
		rec, user := oauthRequest(t, claims, http.MethodPost, path)
		if rec.Code != http.StatusNoContent || user.IsOAuthToken() {
			t.Fatalf("session token on %s = %d (oauth=%v), want unrestricted", path, rec.Code, user.IsOAuthToken())
		}
	}
}

func TestOAuthScopesReachTheAuthUser(t *testing.T) {
	claims := &authPkg.TokenClaims{UserID: 7, OAuthClientID: "rtylr-mcp", Scope: "flow:read flow:read pos:write"}
	_, user := oauthRequest(t, claims, http.MethodGet, "/flow/task/v2/tasks")
	if user == nil || len(user.OAuthScopes) != 2 || !user.HasOAuthScope("pos:read") || user.HasOAuthScope("flow:write") {
		t.Fatalf("auth user scopes = %+v", user)
	}
}
