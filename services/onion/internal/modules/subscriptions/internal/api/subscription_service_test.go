package api

import "testing"

func TestIsBillingCapableRole(t *testing.T) {
	for _, role := range []string{"owner", "admin", "billing", "console_admin", " Billing "} {
		if !isBillingCapableRole(role) {
			t.Errorf("isBillingCapableRole(%q) = false, want true", role)
		}
	}

	for _, role := range []string{"operator", "staff", "viewer", "menu_manager", ""} {
		if isBillingCapableRole(role) {
			t.Errorf("isBillingCapableRole(%q) = true, want false", role)
		}
	}
}
