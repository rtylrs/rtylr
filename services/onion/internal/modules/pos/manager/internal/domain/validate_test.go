package domain

import (
	"errors"
	"strings"
	"testing"
)

// NOTE: TestValidatePIN_RejectsTooLong and TestValidatePIN_AcceptsMinLength
// previously asserted the old 4-8 digit policy — that a 4-digit PIN was valid
// and a 9-digit PIN was not. That policy is the defect this module is fixing:
// a 4-digit manager PIN is 10,000 candidates, which is what the brute-force
// budget has to cover. Their expectations are corrected here to the 6-12 digit
// policy shared with pos_users.

func TestValidatePIN_RejectsTooShort(t *testing.T) {
	for _, pin := range []string{"", "1", "12", "123", "12345"} {
		if err := ValidatePIN(pin); err == nil {
			t.Errorf("expected error for PIN %q (too short)", pin)
		}
	}
}

func TestValidatePIN_RejectsTooLong(t *testing.T) {
	for _, pin := range []string{"1234567890123", "12345678901234567890"} {
		if err := ValidatePIN(pin); err == nil {
			t.Errorf("expected error for %d-digit PIN", len(pin))
		}
	}
}

func TestValidatePIN_RejectsNonDigits(t *testing.T) {
	// Long enough to clear the length rule, so each of these is rejected for
	// breaking the charset rule and not incidentally for being too short.
	for _, pin := range []string{"abcdef", "12ab34", "12345!", "12 345", "1234-5"} {
		err := ValidatePIN(pin)
		if err == nil {
			t.Errorf("expected error for non-digit PIN %q", pin)
			continue
		}
		if !errors.Is(err, ErrWeakPIN) {
			t.Errorf("PIN %q returned %v, want ErrWeakPIN so a handler can name the policy", pin, err)
		}
		if !strings.Contains(err.Error(), "digits") {
			t.Errorf("PIN %q was refused with %q, which does not say the PIN must be numeric", pin, err)
		}
	}
}

func TestValidatePIN_AcceptsMinLength(t *testing.T) {
	if err := ValidatePIN("123456"); err != nil {
		t.Fatalf("unexpected error for %d-digit PIN: %v", MinPINLength, err)
	}
}

func TestValidatePIN_AcceptsMaxLength(t *testing.T) {
	if err := ValidatePIN("123456789012"); err != nil {
		t.Fatalf("unexpected error for %d-digit PIN: %v", MaxPINLength, err)
	}
}

func TestValidatePIN_AcceptsMidLength(t *testing.T) {
	if err := ValidatePIN("12345678"); err != nil {
		t.Fatalf("unexpected error for 8-digit PIN: %v", err)
	}
}

func TestValidatePIN_RejectsWithSpacesOnly(t *testing.T) {
	if err := ValidatePIN("    "); err == nil {
		t.Fatal("expected error for whitespace-only PIN")
	}
}

func TestValidatePIN_TrimsSurroundingSpace(t *testing.T) {
	if err := ValidatePIN("  123456  "); err != nil {
		t.Fatalf("a PIN padded with spaces was rejected: %v", err)
	}
}

// The policy must match pos_users exactly. If the two drift, the weaker one
// sets the effective strength of both POS PIN surfaces.
func TestPINPolicyMatchesPOSUsers(t *testing.T) {
	if MinPINLength != 6 || MaxPINLength != 12 {
		t.Fatalf("manager PIN policy is %d-%d digits, want 6-12 to match pos_users", MinPINLength, MaxPINLength)
	}
}

// Every PIN the policy rejects must report ErrWeakPIN, so a write handler can
// name the rule instead of returning a generic invalid input.
func TestValidatePIN_LengthFailuresReportErrWeakPIN(t *testing.T) {
	for _, pin := range []string{"12345", "1234567890123"} {
		if err := ValidatePIN(pin); !errors.Is(err, ErrWeakPIN) {
			t.Errorf("PIN %q returned %v, want ErrWeakPIN", pin, err)
		}
	}
}
