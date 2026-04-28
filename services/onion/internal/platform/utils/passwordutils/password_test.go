package passwordutils

import "testing"

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
