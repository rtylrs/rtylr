package domain

// Payroll run status values.
const (
	PayrollStatusDraft           = "draft"
	PayrollStatusPendingApproval = "pending_approval"
	PayrollStatusApproved        = "approved"
	PayrollStatusExporting       = "exporting"
	PayrollStatusExported        = "exported"
)

// Payroll period type values.
const (
	PayrollPeriodMonthly     = "monthly"
	PayrollPeriodBiWeekly    = "bi_weekly"
	PayrollPeriodWeekly      = "weekly"
	PayrollPeriodSemiMonthly = "semi_monthly"
)
