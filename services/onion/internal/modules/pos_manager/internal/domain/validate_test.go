package domain

import "testing"

func TestValidatePIN_RejectsTooShort(t *testing.T) {
	for _, pin := range []string{"", "1", "12", "123"} {
		if err := ValidatePIN(pin); err == nil {
			t.Errorf("expected error for PIN %q (too short)", pin)
		}
	}
}

func TestValidatePIN_RejectsTooLong(t *testing.T) {
	if err := ValidatePIN("123456789"); err == nil {
		t.Fatal("expected error for 9-digit PIN")
	}
}

func TestValidatePIN_RejectsNonDigits(t *testing.T) {
	for _, pin := range []string{"abcd", "12ab", "12!3"} {
		if err := ValidatePIN(pin); err == nil {
			t.Errorf("expected error for non-digit PIN %q", pin)
		}
	}
}

func TestValidatePIN_AcceptsMinLength(t *testing.T) {
	if err := ValidatePIN("1234"); err != nil {
		t.Fatalf("unexpected error for 4-digit PIN: %v", err)
	}
}

func TestValidatePIN_AcceptsMaxLength(t *testing.T) {
	if err := ValidatePIN("12345678"); err != nil {
		t.Fatalf("unexpected error for 8-digit PIN: %v", err)
	}
}

func TestValidatePIN_AcceptsMidLength(t *testing.T) {
	if err := ValidatePIN("123456"); err != nil {
		t.Fatalf("unexpected error for 6-digit PIN: %v", err)
	}
}

func TestValidatePIN_RejectsWithSpacesOnly(t *testing.T) {
	if err := ValidatePIN("    "); err == nil {
		t.Fatal("expected error for whitespace-only PIN")
	}
}
