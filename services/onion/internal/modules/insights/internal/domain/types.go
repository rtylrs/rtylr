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

// RevenuePoint is a single daily revenue sample in minor units.
type RevenuePoint struct {
	Date             string
	AmountMinorUnits int64
}
