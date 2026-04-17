package domain

import "errors"

var (
	ErrEmployeeNotFound   = errors.New("employee not found")
	ErrDepartmentNotFound = errors.New("department not found")
	ErrLocationNotFound   = errors.New("location not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrInviteTokenInvalid = errors.New("invite token invalid or expired")
	ErrInviteTokenExpired = errors.New("invite token expired")
	ErrSetupNotComplete   = errors.New("hr setup not complete")
)
