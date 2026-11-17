// Package domain contains pure data types used across the insights module.
//
// Money-valued fields are stored as int64 "minor units" (e.g. cents) so the
// JSON contract never loses precision across the wire. Conversions from the
// DECIMAL columns in finance/pos/erp happen at the infrastructure boundary.
package domain

// MoneyCard is the shared envelope for a single headline metric card.
// The frontend renders `{ value, delta_pct }` consistently across widgets.
type MoneyCard struct {
	Value    int64
	DeltaPct float64
}

// CountCard is the same envelope for count-based metrics.
type CountCard struct {
	Value    int64
	DeltaPct float64
}

// PctCard wraps a percentage metric (0..100).
type PctCard struct {
	Value    float64
	DeltaPct float64
}

// DashboardSnapshot bundles the four headline cards for /insights/dashboard.
type DashboardSnapshot struct {
	Revenue   MoneyCard
	Customers CountCard
	Orders    CountCard
	Margin    PctCard
}

// SalesSummary is the CRM deal cards on the overview: the open pipeline as it
// stands now and the value won this month to date. RecentWonDeals counts wins
// since the start of last month so the client can tell "no CRM use" apart
// from a quiet month. MRR compares recurring revenue now with the start of
// the month; ActiveContracts counts the signed recurring CRM contracts behind it.
type SalesSummary struct {
	OpenPipelineValue int64
	OpenDeals         int64
	Won               MoneyCard
	RecentWonDeals    int64
	MRR               MoneyCard
	ActiveContracts   int64
}

// FlowLoggedTime is the time logged in Work (Flow) over a window, in minutes:
// the business total and its split per project, as Flow reports it.
type FlowLoggedTime struct {
	TotalMinutes int64
	Projects     []FlowProjectTime
}

// FlowProjectTime is one project's share of FlowLoggedTime. Visibility and
// Status decide whether the overview may name it.
type FlowProjectTime struct {
	ProjectID  int64
	Name       string
	Visibility string
	Status     string
	Minutes    int64
}

// HoursSummary is the Work time card on the overview: minutes logged this
// month to date against the same span last month, and the public projects
// that time went to. Private projects count in the total but are not named.
type HoursSummary struct {
	Logged       CountCard
	PriorMinutes int64
	Projects     []ProjectHours
}

// ProjectHours is the time logged on one public project this month to date.
type ProjectHours struct {
	ProjectID int64
	Name      string
	Minutes   int64
}

// DashboardProject is a Work project a viewer may open a project dashboard
// for. Flow decides which: public projects, and private ones Flow lets the
// viewer see.
type DashboardProject struct {
	ID         int64
	Name       string
	Status     string
	Visibility string
}

// ProjectOverdueTask is an open task past its due date. AssigneeID is zero
// when the task is unassigned.
type ProjectOverdueTask struct {
	ID         int64
	Title      string
	DueDate    string
	AssigneeID int64
}

// FlowProjectProgress is one project's task completion as Flow reports it:
// top-level tasks for progress, every open task past its due date for
// overdue, and the oldest few of those.
type FlowProjectProgress struct {
	TotalTasks     int64
	CompletedTasks int64
	OverdueTasks   int64
	Overdue        []ProjectOverdueTask
}

// FlowProjectTimeTotals is the minutes logged on one project: all time, and
// in the two windows the caller asked for.
type FlowProjectTimeTotals struct {
	AllTimeMinutes int64
	ThisMinutes    int64
	PriorMinutes   int64
}

// ProjectInvoices totals the invoices tagged with one project, in minor units.
type ProjectInvoices struct {
	InvoiceCount int64
	Invoiced     int64
	Paid         int64
	Outstanding  int64
	Overdue      int64
	OverdueCount int64
}

// ProjectDashboard is one project's page: progress, overdue work, the time
// logged this month against the same span last month and in total, and,
// when the viewer may see money, its invoices. Invoices is nil otherwise.
type ProjectDashboard struct {
	Project        DashboardProject
	TotalTasks     int64
	CompletedTasks int64
	ProgressPct    float64
	OverdueTasks   int64
	Overdue        []ProjectOverdueTask
	Logged         CountCard
	PriorMinutes   int64
	AllTimeMinutes int64
	Invoices       *ProjectInvoices
}

// RevenuePoint is a single daily revenue sample in minor units.
type RevenuePoint struct {
	Date             string
	AmountMinorUnits int64
}
