package domain

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// MinPINLength is the shortest PIN that may be set or rotated. A 4-digit
	// PIN is 10,000 candidates, which a bcrypt-cost-12 verifier can walk in
	// minutes once a per-route rate limit is the only brake.
	MinPINLength = 6
	// MaxPINLength bounds the plaintext handed to bcrypt. bcrypt rejects
	// secrets longer than 72 bytes without doing any key derivation, so an
	// unbounded PIN would reintroduce a fast timing oracle on the verify path.
	MaxPINLength = 12
)

// ErrWeakPIN reports a PIN outside the length or charset policy. It is
// distinct from a generic invalid input so handlers can tell an operator
// which rule their PIN broke.
var ErrWeakPIN = errors.New("pin does not meet the policy")

// ValidatePIN enforces the POS PIN policy: MinPINLength..MaxPINLength ASCII
// digits. Applied when a PIN is set or rotated, never when one is verified, so
// PINs issued under the old 4-digit policy keep working.
func ValidatePIN(pin string) error {
	pin = strings.TrimSpace(pin)
	if len(pin) < MinPINLength || len(pin) > MaxPINLength {
		return fmt.Errorf("%w: PIN must be between %d and %d digits", ErrWeakPIN, MinPINLength, MaxPINLength)
	}
	for _, r := range pin {
		if r < '0' || r > '9' {
			return fmt.Errorf("%w: PIN must contain only digits", ErrWeakPIN)
		}
	}
	return nil
}

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
	if err := ValidatePIN(input.PIN); err != nil {
		return err
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
