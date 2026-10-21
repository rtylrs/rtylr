package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
)

func TestCreate_RejectsBlankName(t *testing.T) {
	s := &Service{}
	for _, name := range []string{"", "   ", "\t\n"} {
		_, err := s.Create(context.Background(), 1, 1, domain.CreatePOSUserInput{Name: name, PIN: "123456"})
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput for name %q, got %v", name, err)
		}
	}
}

// A 4-digit PIN must no longer be settable. The pre-fix floor was 4, which is a
// 10,000-candidate space an attacker can walk against a cost-12 bcrypt endpoint.
func TestCreate_RejectsShortPIN(t *testing.T) {
	s := &Service{}
	for _, pin := range []string{"1", "12", "123", "12345", "1234"} {
		_, err := s.Create(context.Background(), 1, 1, domain.CreatePOSUserInput{Name: "Bob", PIN: pin})
		if !errors.Is(err, domain.ErrWeakPIN) {
			t.Errorf("PIN %q: expected ErrWeakPIN, got %v", pin, err)
		}
	}
}

func TestCreate_RejectsOverlongAndNonDigitPIN(t *testing.T) {
	s := &Service{}
	for _, pin := range []string{"1234567890123", "12345a", "12 456"} {
		_, err := s.Create(context.Background(), 1, 1, domain.CreatePOSUserInput{Name: "Bob", PIN: pin})
		if !errors.Is(err, domain.ErrWeakPIN) {
			t.Errorf("PIN %q: expected ErrWeakPIN, got %v", pin, err)
		}
	}
}

func TestUpdate_RejectsShortPINBeforeWriting(t *testing.T) {
	// The policy must be checked before the row is mutated, or a rejected PIN
	// would still have applied the name and active-flag changes.
	s := &Service{}
	err := s.Update(context.Background(), 1, 1, domain.UpdatePOSUserInput{Name: "Renamed", PIN: "1234"})
	if !errors.Is(err, domain.ErrWeakPIN) {
		t.Fatalf("expected ErrWeakPIN for a 4-digit rotation, got %v", err)
	}
}

func TestSetPIN_RejectsShortPIN(t *testing.T) {
	// SetPIN is onboarding: a weak PIN must be refused rather than hashed. The
	// Service is the zero value, so reaching the repository at all would panic —
	// which is itself the proof that the policy is checked first.
	s := &Service{}
	err := s.SetPIN(context.Background(), 1,
		SetPINActor{UserID: 1},
		domain.SetPINInput{UserID: 1, LocationID: 1, Name: "Bob", PIN: "1234"},
	)
	if !errors.Is(err, domain.ErrWeakPIN) {
		t.Fatalf("expected ErrWeakPIN, got %v", err)
	}
}

func TestCreate_RejectsBlankPIN(t *testing.T) {
	s := &Service{}
	_, err := s.Create(context.Background(), 1, 1, domain.CreatePOSUserInput{Name: "Bob", PIN: ""})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestVerifyPIN_RejectsBlankInputs(t *testing.T) {
	s := &Service{}
	for _, tc := range []struct {
		locationID, userID int64
		pin                string
	}{
		{0, 0, ""},
		{1, 1, ""},
		{0, 1, "1234"},
		{1, 0, "1234"},
	} {
		_, err := s.VerifyPIN(context.Background(), 1, tc.locationID, tc.userID, tc.pin)
		if !errors.Is(err, domain.ErrInvalidPIN) {
			t.Errorf("expected ErrInvalidPIN for (loc=%d, user=%d, pin=%q), got %v", tc.locationID, tc.userID, tc.pin, err)
		}
	}
}
