package passwordutils

import (
	"golang.org/x/crypto/bcrypt"
)

const cost = bcrypt.DefaultCost

// pinHashCost is the bcrypt work factor used for short numeric secrets such
// as POS PINs. It is intentionally higher than the default (10) because a
// 4–10 digit PIN has far less entropy than a full password and therefore
// needs a slower KDF to compensate against offline brute-force attacks.
const pinHashCost = 12

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)

	return string(bytes), err
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
