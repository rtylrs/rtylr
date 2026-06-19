package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─── GET /erp/dashboard ───────────────────────────────────────────────────────

func DashboardRoutes() []string { return []string{"/erp/dashboard"} }

type DashboardHandler struct{ Service *application.Service }

type DashboardRequest struct{}

type LowStockPayload struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	CurrentStock  float64 `json:"current_stock"`
	ReorderPoint  float64 `json:"reorder_point"`
	WarehouseName string  `json:"warehouse_name"`
}

type RecentPOPayload struct {
	ID           int64   `json:"id"`
	PONumber     string  `json:"po_number"`
	SupplierName string  `json:"supplier_name"`
	Status       string  `json:"status"`
	Total        float64 `json:"total"`
	CreatedAt    string  `json:"created_at"`
}

type TopCategoryPayload struct {
	Name         string `json:"name"`
	ProductCount int    `json:"product_count"`
}

type DashboardResponse struct {
	TotalProducts  int                  `json:"total_products"`
	LowStockCount  int                  `json:"low_stock_count"`
	PendingOrders  int                  `json:"pending_orders"`
	TotalSuppliers int                  `json:"total_suppliers"`
	InventoryValue float64              `json:"inventory_value"`
	LowStockItems  []LowStockPayload    `json:"low_stock_items"`
	RecentOrders   []RecentPOPayload    `json:"recent_orders"`
	TopCategories  []TopCategoryPayload `json:"top_categories"`
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DashboardResponse{}, errorModels.PermissionDeniedServerError()
	}
	stats, err := h.Service.GetDashboard(ctx, ownerID)
	if err != nil {
		return DashboardResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load dashboard",
		}
	}

	return DashboardResponse{
		TotalProducts:  stats.TotalProducts,
		LowStockCount:  stats.LowStockCount,
		PendingOrders:  stats.PendingOrders,
		TotalSuppliers: stats.TotalSuppliers,
		InventoryValue: stats.InventoryValue,
		LowStockItems:  toLowStockPayloads(stats.LowStockItems),
		RecentOrders:   toRecentPOPayloads(stats.RecentOrders),
		TopCategories:  toTopCategoryPayloads(stats.TopCategories),
	}, nil
}

func toLowStockPayloads(items []domain.LowStockItem) []LowStockPayload {
	out := make([]LowStockPayload, 0, len(items))
	for _, it := range items {
		out = append(out, LowStockPayload{
			ProductID:     it.ProductID,
			Name:          it.Name,
			SKU:           it.SKU,
			CurrentStock:  it.CurrentStock,
			ReorderPoint:  it.ReorderPoint,
			WarehouseName: it.WarehouseName,
		})
	}
	return out
}

func toRecentPOPayloads(pos []domain.RecentPO) []RecentPOPayload {
	out := make([]RecentPOPayload, 0, len(pos))
	for _, p := range pos {
		out = append(out, RecentPOPayload{
			ID:           p.ID,
			PONumber:     p.PONumber,
			SupplierName: p.SupplierName,
			Status:       p.Status,
			Total:        p.Total,
			CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func toTopCategoryPayloads(cats []domain.TopCategory) []TopCategoryPayload {
	out := make([]TopCategoryPayload, 0, len(cats))
	for _, c := range cats {
		out = append(out, TopCategoryPayload{Name: c.Name, ProductCount: c.ProductCount})
	}
	return out
}

// ─── GET /erp/reorder-suggestions ─────────────────────────────────────────────

func ReorderSuggestionsRoutes() []string { return []string{"/erp/reorder-suggestions"} }

type ReorderSuggestionsHandler struct{ Service *application.Service }

type ReorderSuggestionsRequest struct{}

type ReorderSuggestionPayload struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	CurrentStock  float64 `json:"current_stock"`
	ReorderPoint  float64 `json:"reorder_point"`
	SuggestedQty  float64 `json:"suggested_quantity"`
	AvgDailySales float64 `json:"avg_daily_sales"`
	LeadTimeDays  int     `json:"lead_time_days"`
	Confidence    string  `json:"confidence"`
}

type ReorderSuggestionsResponse struct {
	Suggestions []ReorderSuggestionPayload `json:"suggestions"`
}

func (h *ReorderSuggestionsHandler) Serve(ctx context.Context, _ *ReorderSuggestionsRequest) (ReorderSuggestionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ReorderSuggestionsResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ReorderSuggestionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	list, err := h.Service.ListReorderSuggestions(ctx, ownerID)
	if err != nil {
		return ReorderSuggestionsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load reorder suggestions",
		}
	}

	payloads := make([]ReorderSuggestionPayload, 0, len(list))
	for _, s := range list {
		payloads = append(payloads, ReorderSuggestionPayload{
			ProductID:     s.ProductID,
			Name:          s.Name,
			SKU:           s.SKU,
			CurrentStock:  s.CurrentStock,
			ReorderPoint:  s.ReorderPoint,
			SuggestedQty:  s.SuggestedQty,
			AvgDailySales: s.AvgDailySales,
			LeadTimeDays:  s.LeadTimeDays,
			Confidence:    s.Confidence,
		})
	}
	return ReorderSuggestionsResponse{Suggestions: payloads}, nil
}

// ─── GET /erp/expiry ──────────────────────────────────────────────────────────

func ExpiryRoutes() []string { return []string{"/erp/expiry"} }

type ExpiryHandler struct{ Service *application.Service }

type ExpiryRequest struct {
	DaysAhead int `json:"days_ahead"`
}

type ExpiryItemPayload struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	BatchID       string  `json:"batch_id"`
	ExpiryDate    string  `json:"expiry_date"`
	Quantity      float64 `json:"quantity"`
	WarehouseName string  `json:"warehouse_name"`
}

type ExpiryResponse struct {
	Items []ExpiryItemPayload `json:"items"`
}

func (h *ExpiryHandler) Serve(ctx context.Context, req *ExpiryRequest) (ExpiryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ExpiryResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ExpiryResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.ListExpiringItems(ctx, ownerID, req.DaysAhead)
	if err != nil {
		return ExpiryResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load expiry report",
		}
	}

	payloads := make([]ExpiryItemPayload, 0, len(items))
	for _, it := range items {
		payloads = append(payloads, ExpiryItemPayload{
			ProductID:     it.ProductID,
			Name:          it.Name,
			SKU:           it.SKU,
			BatchID:       it.BatchID,
			ExpiryDate:    it.ExpiryDate.Format("2006-01-02"),
			Quantity:      it.Quantity,
			WarehouseName: it.WarehouseName,
		})
	}
	return ExpiryResponse{Items: payloads}, nil
}
