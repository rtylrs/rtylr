package domain

import "errors"

var (
	ErrLeaveTypeNotFound          = errors.New("leave type not found")
	ErrInsufficientLeaveBalance   = errors.New("insufficient leave balance")
	ErrLeaveNotConfigured         = errors.New("leave types not configured for this company")
	ErrLeaveRequestNotFound       = errors.New("leave request not found")
	ErrLeaveRequestNotPending     = errors.New("leave request is no longer pending")
	ErrLeaveBalanceNotFound       = errors.New("leave balance not found")
	ErrLeaveConflict              = errors.New("leave request overlaps with existing approved or pending leave")
	ErrApprovalChainNotConfigured = errors.New("approval chain not configured for leave requests")
	ErrLeaveAccessDenied          = errors.New("leave access denied")
	ErrLeaveApproverNotAllowed    = errors.New("approver is not allowed by the leave approval chain")
	ErrLeaveSelfApproval          = errors.New("employees cannot approve their own leave requests")
	ErrInvalidDateRange           = errors.New("start date must be before or equal to end date")
	ErrInsufficientNotice         = errors.New("insufficient notice period for this leave type")
)
