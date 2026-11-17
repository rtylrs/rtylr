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

// RevenuePoint is a single daily revenue sample in minor units.
type RevenuePoint struct {
	Date             string
	AmountMinorUnits int64
}
