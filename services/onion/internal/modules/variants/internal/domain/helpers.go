package domain

// Margin returns the gross margin percentage for the variant.
// Returns 0 when sale price is zero to avoid a divide-by-zero.
func (v Variant) Margin() float64 {
	if v.SalePrice == 0 {
		return 0
	}
	return (v.SalePrice - v.CostPrice) / v.SalePrice * 100
}

// GrossProfit returns the gross profit per unit for the variant.
func (v Variant) GrossProfit() float64 {
	return v.SalePrice - v.CostPrice
}

// StockValue returns the total stock value at cost price.
func (v Variant) StockValue() float64 {
	return v.StockQuantity * v.CostPrice
}
