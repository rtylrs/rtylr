package infrastructure

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidFullName    = errors.New("invalid full name")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrWeakPassword       = errors.New("password does not meet security requirements")
	ErrUserNotFound       = errors.New("user not found")
)
