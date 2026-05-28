package domain

// Leave request status values.
const (
	LeaveStatusPending  = "pending"
	LeaveStatusApproved = "approved"
	LeaveStatusRejected = "rejected"
	LeaveStatusCanceled = "canceled"
)

// Accrual rule types for leave types.
const (
	AccrualRuleNone    = "none"
	AccrualRuleMonthly = "monthly"
	AccrualRuleAnnual  = "annual"
)
