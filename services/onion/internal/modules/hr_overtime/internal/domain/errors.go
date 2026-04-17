package domain

import "errors"

var (
	ErrOvertimeConfigNotFound       = errors.New("overtime config not found")
	ErrOvertimePayrollConfigMissing = errors.New("overtime payroll config is required for payroll processing")
	ErrOvertimeRequestNotFound      = errors.New("overtime request not found")
	ErrApprovalChainNotConfigured   = errors.New("approval chain not configured for overtime requests")
)
