package auth

import "testing"

func TestHasAnyRole(t *testing.T) {
	user := &AuthUser{Role: " Admin "}
	if !HasAnyRole(user, "owner", "admin") {
		t.Fatal("expected normalized admin role to match")
	}
	if HasAnyRole(user, "billing", "staff") {
		t.Fatal("unexpected role match")
	}
}

func TestIsBusinessAdmin(t *testing.T) {
	for _, role := range []string{"owner", "admin", "console_admin"} {
		if !IsBusinessAdmin(&AuthUser{Role: role}) {
			t.Errorf("role %q should be a business admin", role)
		}
	}
	if IsBusinessAdmin(&AuthUser{Role: "operator"}) {
		t.Fatal("operator should not be a business admin")
	}
}
