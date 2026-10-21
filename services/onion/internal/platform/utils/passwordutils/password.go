package passwordutils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// passwordHashCost is the bcrypt work factor for user passwords. It follows
// the OWASP recommendation (12+) rather than the library default so offline
// guesses against a leaked hash stay expensive; existing hashes verify
// unchanged and re-hash at the new cost on their next update.
const passwordHashCost = 12

// pinHashCost is the bcrypt work factor used for short numeric secrets such
// as POS PINs. It is intentionally higher than the default (10) because a
// 4–10 digit PIN has far less entropy than a full password and therefore
// needs a slower KDF to compensate against offline brute-force attacks.
const pinHashCost = 12

// MaxPasswordBytes is bcrypt's input limit. Longer passwords make hashing fail,
// so they are rejected by the policy with a clear error instead.
const MaxPasswordBytes = 72

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)

	return string(bytes), err
}

// dummyPasswordHash is a bcrypt hash of an unrelated secret at
// passwordHashCost. Unknown-email logins compare against it so both
// credential branches burn the same KDF time and response latency cannot
// reveal whether an account exists. Generated once out-of-band; never
// matches a real login.
// #nosec G101 -- This is a deliberately fixed bcrypt timing-equalization hash, not a credential.
// nosemgrep: generic.secrets.security.detected-bcrypt-hash.detected-bcrypt-hash
const dummyPasswordHash = "$2a$12$9mHPQjAD8j.dUXOXOABTqeClV/Y5CECeAWtmq91Wsgc0NQM.eUksK"

// CheckDummyHash runs a bcrypt comparison against the dummy hash to equalize
// timing on the unknown-email path of credential checks.
func CheckDummyHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))

	return err == nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}

// HashPIN bcrypt-hashes a short secret (e.g. a POS user's numeric PIN) at
// work factor 12. Callers must never log the plaintext PIN or the returned
// hash.
func HashPIN(pin string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), pinHashCost)

	return string(bytes), err
}

// CheckPINHash reports whether the supplied plaintext PIN matches the
// bcrypt hash stored for a user. It returns false on any mismatch or
// malformed hash, never returning an error so that timing-sensitive
// callers can treat the boolean uniformly.
func CheckPINHash(pin, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))

	return err == nil
}

// ErrWeakPassword reports a password that fails the shared strength policy:
// at least 8 characters with an upper-case letter, a lower-case letter and a
// digit.
var ErrWeakPassword = errors.New("password does not meet the strength policy")

// ValidateStrength enforces the account-password policy shared by register,
// password change/reset and invite acceptance. PINs are validated separately
// by their owners and intentionally not subject to this rule.
func ValidateStrength(password string) error {
	if len(password) < 8 || len(password) > MaxPasswordBytes {
		return ErrWeakPassword
	}
	hasUpper, hasLower, hasDigit := false, false, false
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return ErrWeakPassword
	}
	return nil
}
