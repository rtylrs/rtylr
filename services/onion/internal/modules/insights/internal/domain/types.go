// Package domain contains pure data types used across the insights module.
//
// Money-valued fields are stored as int64 "minor units" (e.g. cents) so the
// JSON contract never loses precision across the wire. Conversions from the
// DECIMAL columns in finance/pos/erp happen at the infrastructure boundary.
package domain

// MoneyCard is the shared envelope for a single headline metric card.
// The frontend renders `{ value, delta_pct }` consistently across widgets.
type MoneyCard struct {
	Value    int64   `json:"value"`
	DeltaPct float64 `json:"delta_pct"`
}

// CountCard is the same envelope for count-based metrics.
type CountCard struct {
	Value    int64   `json:"value"`
	DeltaPct float64 `json:"delta_pct"`
}

// PctCard wraps a percentage metric (0..100).
type PctCard struct {
	Value    float64 `json:"value"`
	DeltaPct float64 `json:"delta_pct"`
}

// DashboardSnapshot bundles the four headline cards for /insights/dashboard.
type DashboardSnapshot struct {
	Revenue   MoneyCard `json:"revenue"`
	Customers CountCard `json:"customers"`
	Orders    CountCard `json:"orders"`
	Margin    PctCard   `json:"margin"`
}

// RevenuePoint is a single daily revenue sample in minor units.
type RevenuePoint struct {
	Date             string `json:"date"`
	AmountMinorUnits int64  `json:"amount_minor_units"`
}
