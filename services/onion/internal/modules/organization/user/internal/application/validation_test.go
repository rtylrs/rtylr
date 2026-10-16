package application

import (
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
)

func TestValidatePasswordRejectsWhatBcryptCannotHash(t *testing.T) {
	v := NewValidator()
	if err := v.ValidatePassword("Aa1" + strings.Repeat("x", 80)); !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("err = %v, want ErrInvalidPassword", err)
	}
	if err := v.ValidatePassword("Aa1" + strings.Repeat("x", 69)); err != nil {
		t.Fatalf("a 72-byte password was refused: %v", err)
	}
}

func TestValidateEmailAndNameBounds(t *testing.T) {
	v := NewValidator()
	long := strings.Repeat("a", 250) + "@example.com"
	if err := v.ValidateEmail(long); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("a %d-character email was accepted: %v", len(long), err)
	}
	if err := v.ValidateFullName(strings.Repeat("é", 256)); !errors.Is(err, domain.ErrInvalidFullName) {
		t.Fatalf("a 256-character name was accepted: %v", err)
	}
	if err := v.ValidateFullName(strings.Repeat("é", 255)); err != nil {
		t.Fatalf("a 255-character name was refused: %v", err)
	}
}
