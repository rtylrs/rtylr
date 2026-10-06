package user

import "testing"

func TestAnonymousRoutesMatchGeneratedAuthRoutes(t *testing.T) {
	routes := (&Module{}).GetAnonymousRoutes()
	want := map[string]bool{
		"/common/auth/v1/register":               true,
		"/common/auth/v1/login":                  true,
		"/common/auth/v1/register/google":        true,
		"/common/auth/v1/refresh":                true,
		"/common/auth/v1/revoke":                 true,
		"/common/auth/v1/password-reset/request": true,
		"/common/auth/v1/password-reset/confirm": true,
		"/common/auth/v1/verify-email":           true,
		"/common/auth/v1/resend-verification":    true,
		"/common/auth/v1/invite/accept":          true,
		"/common/auth/v1/email-exists":           true,
		"/common/auth/v1/google/authorize":       true,
	}

	if len(routes) != len(want) {
		t.Fatalf("got %d anonymous routes, want %d", len(routes), len(want))
	}
	for _, route := range routes {
		if !want[route] {
			t.Fatalf("unexpected anonymous route %q", route)
		}
	}
}
