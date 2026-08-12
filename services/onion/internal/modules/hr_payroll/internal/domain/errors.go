package domain

import "errors"

var (
	ErrPayrollNotConfigured          = errors.New("payroll not configured for this company")
	ErrPayrollRunNotFound            = errors.New("payroll run not found")
	ErrPayrollAlreadyExported        = errors.New("payroll run has already been exported")
	ErrPayrollExportInProgress       = errors.New("payroll run export is already in progress")
	ErrPayrollRunAlreadyExists       = errors.New("a payroll run already exists for this period")
	ErrTimesheetNotApproved          = errors.New("not all timesheets are approved for this period")
	ErrOvertimeConfigMissing         = errors.New("overtime is enabled but overtime rates are not configured")
	ErrApprovalChainNotConfigured    = errors.New("approval chain not configured for payroll")
	ErrPayrollSelfApproval           = errors.New("payroll run creator cannot approve the same run")
	ErrPayrollApproverNotAllowed     = errors.New("approver is not allowed by the payroll approval chain")
	ErrInvalidDateRange              = errors.New("period start must be before period end")
	ErrPayrollRunNotApproved         = errors.New("payroll run must be approved before export")
	ErrFinanceIntegrationUnavailable = errors.New("finance integration is unavailable")
)
