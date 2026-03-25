package application

import (
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/modules/user/internal/infrastructure"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return infrastructure.ErrInvalidEmail
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !emailRegex.MatchString(email) {
		return infrastructure.ErrInvalidEmail
	}

	return nil
}

func (v *Validator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return infrastructure.ErrInvalidPassword
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return infrastructure.ErrInvalidPassword
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return infrastructure.ErrInvalidPassword
	}

	if !regexp.MustCompile(`\d`).MatchString(password) {
		return infrastructure.ErrInvalidPassword
	}

	return nil
}

func (v *Validator) ValidateFullName(fullName string) error {
	if fullName == "" {
		return infrastructure.ErrInvalidFullName
	}

	if len(strings.TrimSpace(fullName)) < 2 {
		return infrastructure.ErrInvalidFullName
	}

	return nil
}

func (v *Validator) ValidateUserData(data RegisterUserData) error {
	if err := v.ValidateEmail(data.Email); err != nil {
		return err
	}

	if err := v.ValidatePassword(data.Password); err != nil {
		return err
	}

	return v.ValidateFullName(data.FullName)
}

func (v *Validator) NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (v *Validator) NormalizeFullName(fullName string) string {
	return strings.TrimSpace(fullName)
}
