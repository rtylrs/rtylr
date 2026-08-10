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

	// ErrPermissionDenied is returned when the caller is authenticated but
	// lacks the role permission an HR administration write requires. Handlers
	// map it to 403 and must never surface it as a 500.
	ErrPermissionDenied = errors.New("permission denied")
)
