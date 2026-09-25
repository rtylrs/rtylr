package domain

type InvoiceItem struct {
	ID          int64   `json:"id"`
	InvoiceID   int64   `json:"invoice_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	DiscountPct float64 `json:"discount_pct"`
	Total       float64 `json:"total"`
	SortOrder   int     `json:"sort_order"`
}
type InvoiceItemInput struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	DiscountPct float64 `json:"discount_pct"`
	SortOrder   int     `json:"sort_order"`
}

// LineComponents splits a line into its pre-tax base and its own tax so
// header-level discount and tax can be applied without compounding item tax.
type LineComponents struct {
	Base float64
	Tax  float64
}

// CalculateComponents returns the line's pre-tax base (after per-item
// discount) and its per-item tax amount.
func (i InvoiceItemInput) CalculateComponents() LineComponents {
	lineBase := i.Quantity * i.UnitPrice
	discountAmt := lineBase * (i.DiscountPct / 100)
	base := lineBase - discountAmt

	return LineComponents{Base: base, Tax: base * (i.TaxRate / 100)}
}

// CalculateTotal computes the line total after per-item discount and tax.
func (i InvoiceItemInput) CalculateTotal() float64 {
	components := i.CalculateComponents()

	return components.Base + components.Tax
}
