package domain

// Payroll run status values.
const (
	PayrollStatusDraft     = "draft"
	PayrollStatusInReview  = "in_review"
	PayrollStatusApproved  = "approved"
	PayrollStatusExporting = "exporting"
	PayrollStatusExported  = "exported"
)

// Payroll period type values.
const (
	PayrollPeriodMonthly     = "monthly"
	PayrollPeriodBiWeekly    = "bi_weekly"
	PayrollPeriodWeekly      = "weekly"
	PayrollPeriodSemiMonthly = "semi_monthly"
)
