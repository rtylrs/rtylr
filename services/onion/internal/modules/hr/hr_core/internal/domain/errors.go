package domain

import "errors"

var (
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrDepartmentNotFound   = errors.New("department not found")
	ErrLocationNotFound     = errors.New("location not found")
	ErrRoleNotFound         = errors.New("role not found")
	ErrWorkScheduleNotFound = errors.New("work schedule not found")
	ErrInviteTokenInvalid   = errors.New("invite token invalid or expired")
	ErrInviteTokenExpired   = errors.New("invite token expired")
	// ErrInviteRecipientMismatch is returned when the account trying to accept
	// an invite is not the one the invite was addressed to. It deliberately
	// covers an account that does not exist, so acceptance cannot be used to
	// probe which user ids have accounts.
	ErrInviteRecipientMismatch = errors.New("invite was not addressed to this account")
	ErrSetupNotComplete        = errors.New("hr setup not complete")
	ErrSetupStepNotFound       = errors.New("hr setup step not found")
	ErrUserAlreadyLinked       = errors.New("user is already linked to another employee record")

	// ErrPermissionDenied is returned when the caller is authenticated but
	// lacks the role permission an HR administration write requires. Handlers
	// map it to 403 and must never surface it as a 500.
	ErrPermissionDenied = errors.New("permission denied")
)
