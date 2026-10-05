package hr_core

import "testing"

// Verifying and accepting an employee invite act on a record the token only
// names, and accepting one binds the caller's own account to it. Both routes
// used to be registered as anonymous, which skipped the bearer gate entirely:
// an anonymous request never reaches a handler with an auth context, so the
// acceptance had no caller to check and took the account from the body instead.
func TestEmployeeInviteRoutesAreNotAnonymous(t *testing.T) {
	anonymous := (&Module{}).GetAnonymousRoutes()
	for _, route := range anonymous {
		switch route {
		case "/hr/employees/invite/accept", "/hr/employees/invite/verify":
			t.Fatalf("%s is registered as an anonymous route; invite acceptance must run behind the bearer gate", route)
		}
	}
}
