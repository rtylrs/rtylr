package domain

import "errors"

var (
	ErrInvalidEmail                     = errors.New("invalid email format")
	ErrInvalidPassword                  = errors.New("invalid password")
	ErrInvalidFullName                  = errors.New("invalid full name")
	ErrInvalidCredentials               = errors.New("invalid credentials")
	ErrUserAlreadyExists                = errors.New("user already exists")
	ErrWeakPassword                     = errors.New("password does not meet security requirements")
	ErrUserNotFound                     = errors.New("user not found")
	ErrInvalidEmailVerificationCode     = errors.New("invalid or expired email verification code")
	ErrSupersededEmailVerificationCode  = errors.New("email verification code was issued in an older format; request a new one")
	ErrTooManyEmailVerificationAttempts = errors.New("too many email verification attempts")
)

type EmailVerification struct {
	UserID int64
	Code   string
	Token  string
}

type UserAccessState struct {
	HasPermissions  bool
	HasActiveAccess bool
}
