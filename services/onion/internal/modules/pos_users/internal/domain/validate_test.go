package domain

import "testing"

func TestValidatePOSUserRole_AcceptsAllKnown(t *testing.T) {
	for _, role := range []POSUserRole{RoleCashier, RoleServer, RoleBartender, RoleKitchen} {
		if err := ValidatePOSUserRole(role); err != nil {
			t.Errorf("unexpected error for role %q: %v", role, err)
		}
	}
}

func TestValidatePOSUserRole_RejectsUnknown(t *testing.T) {
	if err := ValidatePOSUserRole("manager"); err == nil {
		t.Fatal("expected error for unknown role 'manager'")
	}
}

func TestValidateCreatePOSUserInput_RejectsBlankName(t *testing.T) {
	if err := ValidateCreatePOSUserInput(CreatePOSUserInput{PIN: "1234", Role: RoleCashier}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateCreatePOSUserInput_RejectsShortPIN(t *testing.T) {
	if err := ValidateCreatePOSUserInput(CreatePOSUserInput{Name: "Alex", PIN: "123", Role: RoleCashier}); err == nil {
		t.Fatal("expected error for 3-digit PIN")
	}
}

func TestValidateCreatePOSUserInput_RejectsInvalidRole(t *testing.T) {
	if err := ValidateCreatePOSUserInput(CreatePOSUserInput{Name: "Alex", PIN: "1234", Role: "admin"}); err == nil {
		t.Fatal("expected error for invalid role")
	}
}

func TestValidateCreatePOSUserInput_AcceptsValid(t *testing.T) {
	if err := ValidateCreatePOSUserInput(CreatePOSUserInput{Name: "Alex", PIN: "4321", Role: RoleCashier}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPOSUser_CanPerformRefund_ActiveWithPermission(t *testing.T) {
	u := POSUser{IsActive: true, CanRefund: true}
	if !u.CanPerformRefund() {
		t.Error("expected CanPerformRefund=true")
	}
}

func TestPOSUser_CanPerformRefund_InactiveWithPermission(t *testing.T) {
	u := POSUser{IsActive: false, CanRefund: true}
	if u.CanPerformRefund() {
		t.Error("expected CanPerformRefund=false for inactive user")
	}
}

func TestPOSUser_CanPerformVoid_ActiveWithPermission(t *testing.T) {
	u := POSUser{IsActive: true, CanVoid: true}
	if !u.CanPerformVoid() {
		t.Error("expected CanPerformVoid=true")
	}
}

func TestPOSUser_CanGiveDiscount_ActiveWithPermission(t *testing.T) {
	u := POSUser{IsActive: true, CanApplyDiscount: true}
	if !u.CanGiveDiscount() {
		t.Error("expected CanGiveDiscount=true")
	}
}

func TestPOSUser_CanGiveDiscount_WithoutPermission(t *testing.T) {
	u := POSUser{IsActive: true, CanApplyDiscount: false}
	if u.CanGiveDiscount() {
		t.Error("expected CanGiveDiscount=false without permission")
	}
}
