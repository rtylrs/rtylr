package domain

// TotalValuationValue sums TotalValue across all valuation items.
func TotalValuationValue(items []ValuationItem) float64 {
	var total float64
	for _, item := range items {
		total += item.TotalValue
	}
	return total
}

// TotalValuationCost sums TotalCost across all valuation items.
func TotalValuationCost(items []ValuationItem) float64 {
	var total float64
	for _, item := range items {
		total += item.TotalCost
	}
	return total
}

// TotalItemCount sums TotalItems across all valuation items.
func TotalItemCount(items []ValuationItem) int64 {
	var count int64
	for _, item := range items {
		count += item.TotalItems
	}
	return count
}
