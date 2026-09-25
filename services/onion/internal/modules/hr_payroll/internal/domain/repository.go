package domain

import (
	"context"

	"github.com/voxire/onion/internal/platform/financejournal"
)

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}

// PermissionProvider lets the payroll service resolve which hr_employees row
// belongs to the calling platform user and check role-based permissions,
// without depending on hr_core directly.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
	HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error)
	HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error)
}

// FinanceJournalPoster lets the payroll service hand a completed run off to the
// finance module without depending on it directly. The finance module is
// expected to translate the call into the appropriate journal entries
// (debit payroll expense, credit payable / cash) inside its own bounded
// context. Implementations must be idempotent on (ownerID, runID).
type FinanceJournalPoster = financejournal.PayrollJournalPoster

// PayrollJournalEntry is the payload handed to the finance module when a
// payroll run is exported. Money figures are passed in the currency stored
// on the run record; finance is responsible for any conversion.
type PayrollJournalEntry = financejournal.PayrollJournalEntry

// NotificationDispatcher lets the payroll service emit notifications without
// depending on the notification module directly. Implementations should be
// non-blocking from the caller's perspective (queue + return).
type NotificationDispatcher interface {
	Dispatch(ctx context.Context, event NotificationEvent) error
}

// NotificationEvent is the cross-module notification envelope. Type is a
// stable string key (e.g. "payroll.payslip", "payroll.period_open") that the
// notification module maps to a template + delivery channel based on company
// config.
type NotificationEvent struct {
	OwnerID int64
	Type    string
	Subject string
	Payload map[string]any
	UserIDs []int64 // optional explicit recipients; empty = use type-based routing
}
