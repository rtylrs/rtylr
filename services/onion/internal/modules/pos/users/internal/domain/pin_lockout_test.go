package domain

import (
	"testing"
	"time"
)

func TestPinBackoff_NeverNegativeAndMonotonic(t *testing.T) {
	previous := time.Duration(0)
	for failures := 0; failures <= 20; failures++ {
		got := PinBackoff(failures)
		if got < 0 {
			t.Fatalf("PinBackoff(%d) = %s, want >= 0", failures, got)
		}
		if got < previous {
			t.Errorf("PinBackoff(%d) = %s is shorter than PinBackoff(%d) = %s; an attacker could pace around a decreasing window",
				failures, got, failures-1, previous)
		}
		previous = got
	}
}

func TestPinBackoff_LockoutOpensAtMaxFailures(t *testing.T) {
	if got := PinBackoff(PinVerifyMaxFailures - 1); got >= PinLockoutDuration {
		t.Errorf("PinBackoff(%d) = %s already equals the lockout, so the ramp is skipped", PinVerifyMaxFailures-1, got)
	}
	if got := PinBackoff(PinVerifyMaxFailures); got != PinLockoutDuration {
		t.Errorf("PinBackoff(%d) = %s, want the full lockout %s", PinVerifyMaxFailures, got, PinLockoutDuration)
	}
	if got := PinBackoff(PinVerifyMaxFailures + 50); got != PinLockoutDuration {
		t.Errorf("PinBackoff(%d) = %s, want the full lockout %s", PinVerifyMaxFailures+50, got, PinLockoutDuration)
	}
}

// The size of the space an attacker has to cover, times the per-window budget,
// is the actual security property. Both the legacy four-digit space and the
// current six-digit one are checked: the lockout is what protects the PINs
// already issued under the old policy.
func TestPinBackoff_ExhaustiveGuessIsImpractical(t *testing.T) {
	for _, tc := range []struct {
		name       string
		candidates int
		minDays    int
	}{
		{"legacy 4-digit PIN", 10000, 14},
		{"current 6-digit PIN", 1000000, 1400},
	} {
		windows := tc.candidates / PinVerifyMaxFailures
		elapsed := time.Duration(windows) * PinLockoutDuration
		if days := elapsed / (24 * time.Hour); days < time.Duration(tc.minDays) {
			t.Errorf("%s: exhausting %d candidates takes %d days, want at least %d",
				tc.name, tc.candidates, days, tc.minDays)
		}
	}
}

func TestPinBackoff_FirstFailureIsFree(t *testing.T) {
	if got := PinBackoff(1); got != 0 {
		t.Errorf("PinBackoff(1) = %s, want 0 so a mistyped PIN stays cheap for the operator", got)
	}
}

func TestPinRetryAt(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	if got := PinRetryAt(0, base); !got.IsZero() {
		t.Errorf("PinRetryAt(0, base) = %s, want the zero time", got)
	}
	if got := PinRetryAt(3, time.Time{}); !got.IsZero() {
		t.Errorf("PinRetryAt(3, zero) = %s, want the zero time: without a last-failure stamp there is no window", got)
	}
	if got, want := PinRetryAt(3, base), base.Add(PinBackoff(3)); !got.Equal(want) {
		t.Errorf("PinRetryAt(3, base) = %s, want %s", got, want)
	}
	if got, want := PinRetryAt(PinVerifyMaxFailures, base), base.Add(PinLockoutDuration); !got.Equal(want) {
		t.Errorf("PinRetryAt(%d, base) = %s, want %s", PinVerifyMaxFailures, got, want)
	}
}

func TestValidatePIN_RejectsFourDigits(t *testing.T) {
	// Four digits is the policy the brute-force finding was built on: 10,000
	// candidates. It must no longer be settable.
	if err := ValidatePIN("1234"); err == nil {
		t.Fatal("ValidatePIN accepted a 4-digit PIN")
	}
}

func TestValidatePIN_AcceptsSixToTwelveDigits(t *testing.T) {
	for _, pin := range []string{"123456", "000000", "123456789012"} {
		if err := ValidatePIN(pin); err != nil {
			t.Errorf("ValidatePIN(%q) = %v, want nil", pin, err)
		}
	}
}

func TestValidatePIN_RejectsOverMaxLength(t *testing.T) {
	// bcrypt ignores bytes past 72, so an unbounded PIN is also a way to make the
	// comparison cheap. The maximum is what closes that.
	long := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	if err := ValidatePIN(long); err == nil {
		t.Error("ValidatePIN accepted a 127-digit PIN")
	}
}

func TestValidatePIN_RejectsNonDigits(t *testing.T) {
	for _, pin := range []string{"12345a", "12 456", "12-456", "abcdef", "１２３４５６"} {
		if err := ValidatePIN(pin); err == nil {
			t.Errorf("ValidatePIN(%q) = nil, want an error: the POS keypad only produces digits", pin)
		}
	}
}

func TestValidatePIN_TrimsSurroundingWhitespace(t *testing.T) {
	if err := ValidatePIN("  123456  "); err != nil {
		t.Errorf("ValidatePIN with padding = %v, want nil", err)
	}
}

func TestValidateCreatePOSUserInput_UsesTheCurrentPINPolicy(t *testing.T) {
	if err := ValidateCreatePOSUserInput(CreatePOSUserInput{Name: "Alex", PIN: "1234", Role: RoleCashier}); err == nil {
		t.Error("create accepted a 4-digit PIN; the policy must be enforced through the create path too")
	}
	if err := ValidateCreatePOSUserInput(CreatePOSUserInput{Name: "Alex", PIN: "123456", Role: RoleCashier}); err != nil {
		t.Errorf("create rejected a compliant PIN: %v", err)
	}
}
