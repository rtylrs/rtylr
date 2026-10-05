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

// RevenuePoint is a single daily revenue sample in minor units.
type RevenuePoint struct {
	Date             string
	AmountMinorUnits int64
}
