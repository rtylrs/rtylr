package domain

import (
	"fmt"
	"strings"
)

// ValidatePOSUserRole returns an error when the role is not a known value.
func ValidatePOSUserRole(role POSUserRole) error {
	switch role {
	case RoleCashier, RoleServer, RoleBartender, RoleKitchen:
		return nil
	default:
		return fmt.Errorf("invalid POS user role %q", role)
	}
}

// ValidateCreatePOSUserInput returns an error when the create input is malformed.
func ValidateCreatePOSUserInput(input CreatePOSUserInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	pin := strings.TrimSpace(input.PIN)
	if len(pin) < 4 {
		return fmt.Errorf("PIN must be at least 4 digits")
	}
	return ValidatePOSUserRole(input.Role)
}

// CanPerformRefund reports whether this user has refund permission.
func (u POSUser) CanPerformRefund() bool {
	return u.CanRefund && u.IsActive
}

// CanPerformVoid reports whether this user has void permission.
func (u POSUser) CanPerformVoid() bool {
	return u.CanVoid && u.IsActive
}

// CanGiveDiscount reports whether this user can apply discounts.
func (u POSUser) CanGiveDiscount() bool {
	return u.CanApplyDiscount && u.IsActive
}
