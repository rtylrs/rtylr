package domain

// TotalStockValue sums StockValue across all items in a report.
func TotalStockValue(items []InventoryReportItem) float64 {
	var total float64
	for _, item := range items {
		total += item.StockValue
	}
	return total
}

// LowStockCount returns the number of low-stock items in a report.
func LowStockCount(items []InventoryReportItem) int {
	count := 0
	for _, item := range items {
		if item.IsLowStock {
			count++
		}
	}
	return count
}

// FilterLowStock returns only the low-stock items from a report.
func FilterLowStock(items []InventoryReportItem) []InventoryReportItem {
	var out []InventoryReportItem
	for _, item := range items {
		if item.IsLowStock {
			out = append(out, item)
		}
	}
	return out
}
