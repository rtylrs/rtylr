package domain

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// MinPINLength is the shortest manager PIN that may be set or rotated. It
	// matches the pos_users POS PIN floor: a 4-digit PIN is 10,000 candidates,
	// which a bcrypt-cost-12 verifier can walk quickly once a per-route rate
	// limit is the only brake.
	MinPINLength = 6
	// MaxPINLength bounds the plaintext handed to bcrypt. bcrypt ignores
	// secrets longer than 72 bytes instead of deriving a key, so an unbounded
	// PIN would reintroduce a fast timing oracle on the verify path.
	MaxPINLength = 12
)

// ErrWeakPIN reports a PIN outside the length or charset policy. It is distinct
// from a generic invalid input so handlers can tell an operator which rule their
// PIN broke.
var ErrWeakPIN = errors.New("pin does not meet the policy")

// ValidatePIN enforces the manager PIN policy: MinPINLength..MaxPINLength ASCII
// digits. It is applied when a PIN is set or rotated, never when one is
// verified, so PINs issued under the older 4-8 digit policy keep working.
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
