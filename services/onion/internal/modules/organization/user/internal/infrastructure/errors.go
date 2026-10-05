package infrastructure

import "github.com/voxire/onion/internal/modules/organization/user/internal/domain"

var (
	ErrInvalidEmail       = domain.ErrInvalidEmail
	ErrInvalidPassword    = domain.ErrInvalidPassword
	ErrInvalidFullName    = domain.ErrInvalidFullName
	ErrInvalidCredentials = domain.ErrInvalidCredentials
	ErrUserAlreadyExists  = domain.ErrUserAlreadyExists
	ErrWeakPassword       = domain.ErrWeakPassword
	ErrUserNotFound       = domain.ErrUserNotFound
)
