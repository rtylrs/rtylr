package domain

type InvoiceItem struct {
	ID          int64
	InvoiceID   int64
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	TaxRate     float64
	DiscountPct float64
	Total       float64
	SortOrder   int
}
type InvoiceItemInput struct {
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	TaxRate     float64
	DiscountPct float64
	SortOrder   int
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
