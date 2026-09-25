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

func TestIsConsoleAdminRole(t *testing.T) {
	for _, role := range []string{"console_admin", " Console_Admin "} {
		if !isConsoleAdminRole(role) {
			t.Errorf("isConsoleAdminRole(%q) = false, want true", role)
		}
	}

	for _, role := range []string{"owner", "admin", "billing", "operator", ""} {
		if isConsoleAdminRole(role) {
			t.Errorf("isConsoleAdminRole(%q) = true, want false", role)
		}
	}
}
