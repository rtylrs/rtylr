package domain

import "time"

// DashboardStats is the top-of-page summary used by the ERP dashboard.
// Fields are additive — new counters can be appended without breaking
// the frontend because missing fields deserialize to zero values.
type DashboardStats struct {
	TotalProducts  int            `json:"total_products"`
	LowStockCount  int            `json:"low_stock_count"`
	PendingOrders  int            `json:"pending_orders"`
	TotalSuppliers int            `json:"total_suppliers"`
	InventoryValue float64        `json:"inventory_value"`
	LowStockItems  []LowStockItem `json:"low_stock_items"`
	RecentOrders   []RecentPO     `json:"recent_orders"`
	TopCategories  []TopCategory  `json:"top_categories"`
}

// LowStockItem surfaces on the dashboard sidebar. Canonical source is the
// inventory module's /erp/inventory/low-stock endpoint — this is a
// trimmed projection for the dashboard's summary card.
type LowStockItem struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	CurrentStock  float64 `json:"current_stock"`
	ReorderPoint  float64 `json:"reorder_point"`
	WarehouseName string  `json:"warehouse_name"`
}

// RecentPO is a compact view of a purchase order used on the dashboard.
type RecentPO struct {
	ID           int64     `json:"id"`
	PONumber     string    `json:"po_number"`
	SupplierName string    `json:"supplier_name"`
	Status       string    `json:"status"`
	Total        float64   `json:"total"`
	CreatedAt    time.Time `json:"created_at"`
}

// TopCategory = category name + product count, used to render the
// "top categories" mini-chart on the ERP dashboard.
type TopCategory struct {
	Name         string `json:"name"`
	ProductCount int    `json:"product_count"`
}

// ReorderSuggestion is a row on the Reorder screen. The backend picks
// products whose stock is at or below their reorder threshold, and
// estimates suggested quantity based on the recorded reorder_quantity.
type ReorderSuggestion struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	CurrentStock  float64 `json:"current_stock"`
	ReorderPoint  float64 `json:"reorder_point"`
	SuggestedQty  float64 `json:"suggested_quantity"`
	AvgDailySales float64 `json:"avg_daily_sales"`
	LeadTimeDays  int     `json:"lead_time_days"`
	Confidence    string  `json:"confidence"` // "high" | "medium" | "low"
}

type ReorderSuggestionsResult struct {
	Suggestions []ReorderSuggestion
	Truncated   bool
}

// ExpiryItem is a row on the Expiry screen. Until batch tracking is
// modelled, this endpoint returns an empty list — stubbed so the frontend
// has a real endpoint to call instead of 404'ing.
type ExpiryItem struct {
	ProductID     int64     `json:"product_id"`
	Name          string    `json:"name"`
	SKU           string    `json:"sku"`
	BatchID       string    `json:"batch_id"`
	ExpiryDate    time.Time `json:"expiry_date"`
	Quantity      float64   `json:"quantity"`
	WarehouseName string    `json:"warehouse_name"`
}
