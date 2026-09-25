package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
)

func TestCreate_RejectsBlankName(t *testing.T) {
	s := &Service{}
	for _, name := range []string{"", "   ", "\t\n"} {
		_, err := s.Create(context.Background(), 1, 1, domain.CreatePOSUserInput{Name: name, PIN: "1234"})
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput for name %q, got %v", name, err)
		}
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
