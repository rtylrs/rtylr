package domain

// ProductMarginRow is one product's sales volume and cost, read from
// erp_products joined with pos_order_items.
type ProductMarginRow struct {
	ID             int64
	Name           string
	UnitsSold      float64
	RevenueMinor   int64
	CostMinorUnits *int64 // nil when cost_price is zero/unset
}
