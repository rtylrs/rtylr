package passwordutils

import (
	"errors"
	"strings"
	"testing"
)

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "hunter2" {
		t.Fatal("hash must not equal plaintext")
	}
	if !CheckPasswordHash("hunter2", hash) {
		t.Fatal("hash should match plaintext")
	}
	if CheckPasswordHash("wrong", hash) {
		t.Fatal("hash should not match wrong password")
	}
}

func TestHashPassword_DifferentEachCall(t *testing.T) {
	a, _ := HashPassword("same")
	b, _ := HashPassword("same")
	if a == b {
		t.Fatal("two hashes of the same password should differ (salt)")
	}
}

func TestHashPIN_RoundTrip(t *testing.T) {
	hash, err := HashPIN("1234")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !CheckPINHash("1234", hash) {
		t.Fatal("PIN should match")
	}
	if CheckPINHash("0000", hash) {
		t.Fatal("wrong PIN should not match")
	}
	if CheckPINHash("1234", "not-a-bcrypt-hash") {
		t.Fatal("malformed hash should not match")
	}
	if CheckPINHash("1234", "") {
		t.Fatal("empty hash should not match")
	}
}

func TestHashPassword_UsesOWASPCost(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if len(hash) < 7 || hash[:7] != "$2a$12$" {
		t.Errorf("hash cost prefix = %q, want $2a$12$", hash[:min(7, len(hash))])
	}
}

func TestCheckDummyHash_NeverMatches(t *testing.T) {
	for _, password := range []string{"", "hunter2", "correct horse", "password123"} {
		if CheckDummyHash(password) {
			t.Errorf("dummy hash matched %q", password)
		}
	}
}

func TestValidateStrengthRejectsPasswordsBcryptCannotHash(t *testing.T) {
	long := "Aa1" + strings.Repeat("x", MaxPasswordBytes)
	if err := ValidateStrength(long); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("err = %v, want ErrWeakPassword", err)
	}
	atLimit := "Aa1" + strings.Repeat("x", MaxPasswordBytes-3)
	if err := ValidateStrength(atLimit); err != nil {
		t.Fatalf("a %d-byte password was refused: %v", len(atLimit), err)
	}
	if _, err := HashPassword(atLimit); err != nil {
		t.Fatalf("a password the policy accepts must hash: %v", err)
	}
}
