package domain

// Margin returns the gross margin percentage for the product.
// Returns 0 when the sale price is zero to avoid a divide-by-zero panic.
func (p Product) Margin() float64 {
	if p.SalePrice == 0 {
		return 0
	}
	return (p.SalePrice - p.CostPrice) / p.SalePrice * 100
}

// GrossProfit returns the gross profit per unit.
func (p Product) GrossProfit() float64 {
	return p.SalePrice - p.CostPrice
}

// StockValue returns the total stock value at cost price.
func (p Product) StockValue() float64 {
	return p.StockQuantity * p.CostPrice
}

// IsLowStock reports whether the product stock is below its threshold.
// Products with a zero threshold are never considered low-stock.
func (p Product) IsLowStock() bool {
	if p.LowStockThreshold == 0 {
		return false
	}
	return p.StockQuantity <= p.LowStockThreshold
}
