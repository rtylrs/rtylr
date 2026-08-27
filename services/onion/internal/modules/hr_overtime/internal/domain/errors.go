package domain

import "errors"

var (
	ErrOvertimeConfigNotFound       = errors.New("overtime config not found")
	ErrOvertimePayrollConfigMissing = errors.New("overtime payroll config is required for payroll processing")
	ErrOvertimeRequestNotFound      = errors.New("overtime request not found")
	ErrOvertimeRequestNotPending    = errors.New("overtime request is no longer pending")
	ErrOvertimeAccessDenied         = errors.New("overtime access denied")
	ErrOvertimeApproverNotAllowed   = errors.New("approver is not allowed by the overtime approval chain")
	ErrOvertimeSelfApproval         = errors.New("employees cannot approve their own overtime requests")
	ErrApprovalChainNotConfigured   = errors.New("approval chain not configured for overtime requests")
)
