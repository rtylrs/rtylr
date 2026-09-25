// Package financeaccounts defines the finance account configuration boundary.
package financeaccounts

import (
	"context"
	"errors"
)

var ErrPayrollAccountsNotConfigured = errors.New("payroll accounts are not configured")

// PayrollAccountMapping identifies the tenant-selected chart-of-accounts
// entries used by payroll posting. Zero deductions means deductions are not
// configured because the payroll run had no deductions.
type PayrollAccountMapping struct {
	BusinessID                 int64
	ExpenseAccountID           int64
	PayableAccountID           int64
	DeductionsPayableAccountID int64
}

// PayrollAccountProvider is implemented by the accounts module and consumed
// by journal. Account selection remains tenant configuration, not a journal
// business rule.
type PayrollAccountProvider interface {
	GetPayrollAccountMapping(ctx context.Context, businessID int64) (PayrollAccountMapping, error)
}
