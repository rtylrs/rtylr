package domain

import "time"

// DashboardStats is the top-of-page summary used by the ERP dashboard.
// Fields are additive — new counters can be appended without breaking
// the frontend because missing fields deserialize to zero values.
type DashboardStats struct {
	TotalProducts  int
	LowStockCount  int
	PendingOrders  int
	TotalSuppliers int
	InventoryValue float64
	LowStockItems  []LowStockItem
	RecentOrders   []RecentPO
	TopCategories  []TopCategory
}

// LowStockItem surfaces on the dashboard sidebar. Canonical source is the
// inventory module's /erp/inventory/low-stock endpoint — this is a
// trimmed projection for the dashboard's summary card.
type LowStockItem struct {
	ProductID     int64
	Name          string
	SKU           string
	CurrentStock  float64
	ReorderPoint  float64
	WarehouseName string
}

// RecentPO is a compact view of a purchase order used on the dashboard.
type RecentPO struct {
	ID           int64
	PONumber     string
	SupplierName string
	Status       string
	Total        float64
	CreatedAt    time.Time
}

// TopCategory = category name + product count, used to render the
// "top categories" mini-chart on the ERP dashboard.
type TopCategory struct {
	Name         string
	ProductCount int
}

// ReorderSuggestion is a row on the Reorder screen. The backend picks
// products whose stock is at or below their reorder threshold, and
// estimates suggested quantity based on the recorded reorder_quantity.
type ReorderSuggestion struct {
	ProductID     int64
	Name          string
	SKU           string
	CurrentStock  float64
	ReorderPoint  float64
	SuggestedQty  float64
	AvgDailySales float64
	LeadTimeDays  int
	Confidence    string // "high" | "medium" | "low"
}

type ReorderSuggestionsResult struct {
	Suggestions []ReorderSuggestion
	Truncated   bool
}

// ExpiryItem is a row on the Expiry screen. Until batch tracking is
// modelled, this endpoint returns an empty list — stubbed so the frontend
// has a real endpoint to call instead of 404'ing.
type ExpiryItem struct {
	ProductID     int64
	Name          string
	SKU           string
	BatchID       string
	ExpiryDate    time.Time
	Quantity      float64
	WarehouseName string
}
