package domain

import (
	"errors"
	"testing"
)

func TestUnsubscribeToken_RoundTripsAndRejectsTampering(t *testing.T) {
	secret := []byte("test-secret")
	token := UnsubscribeToken(secret, 7, 42)

	businessID, customerID, err := ParseUnsubscribeToken(secret, token)
	if err != nil || businessID != 7 || customerID != 42 {
		t.Fatalf("ParseUnsubscribeToken = %d, %d, %v; want 7, 42, nil", businessID, customerID, err)
	}

	other := UnsubscribeToken(secret, 7, 43)
	forged := token[:len(token)-len(other)+len(other)/2] + other[len(other)/2:]
	for name, bad := range map[string]string{
		"another customer's signature": forged,
		"truncated":                    token[:len(token)-3],
		"no signature":                 "Ny40Mg",
		"empty":                        "",
	} {
		if _, _, err := ParseUnsubscribeToken(secret, bad); !errors.Is(err, ErrInvalidUnsubscribe) {
			t.Errorf("%s: err = %v, want ErrInvalidUnsubscribe", name, err)
		}
	}
	if _, _, err := ParseUnsubscribeToken([]byte("other-secret"), token); !errors.Is(err, ErrInvalidUnsubscribe) {
		t.Errorf("a token signed with another secret was accepted")
	}
}
