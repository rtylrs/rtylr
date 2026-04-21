package application

import "context"

// ReportTemplatePayload is a single available report template. At P3 this
// list is static — real templates are stored in code rather than a table
// because the set is defined by the product, not the tenant.
//
// When this graduates to a DB-backed surface, move the list into a
// module-level config table so tenants can disable or rename templates
// (see CLAUDE.md §8 on business-rule configurability).
type ReportTemplatePayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListReportTemplates returns the static P3 catalog.
func (s *Service) ListReportTemplates(_ context.Context) []ReportTemplatePayload {
	return []ReportTemplatePayload{
		{ID: "revenue-monthly", Name: "Monthly Revenue", Description: "Month-by-month paid revenue over the last 12 months."},
		{ID: "revenue-daily", Name: "Daily Revenue", Description: "Paid revenue per day for a custom window."},
		{ID: "top-customers", Name: "Top Customers", Description: "Leaderboard of customers by lifetime spend."},
		{ID: "invoice-aging", Name: "Invoice Aging", Description: "A/R aging buckets with outstanding balances."},
		{ID: "expenses-category", Name: "Expenses by Category", Description: "Expense spend grouped by category."},
		{ID: "best-sellers", Name: "Best Sellers", Description: "Top-selling SKUs over the last 30 days."},
		{ID: "slow-movers", Name: "Slow Movers", Description: "SKUs with no sales in a trailing window."},
		{ID: "shift-coverage", Name: "Shift Coverage", Description: "Scheduled vs completed shifts by week."},
		{ID: "headcount", Name: "Headcount", Description: "Active staff count and MoM change."},
	}
}
