package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
)

// ─── List Locations ───────────────────────────────────────────────────────────

type ListLocationsHandler struct{ Service *application.Service }

type ListLocationsRequest struct{}

type LocationPayload struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type ListLocationsResponse struct {
	Locations []LocationPayload `json:"locations"`
}

// ─── Create Location ──────────────────────────────────────────────────────────

type CreateLocationHandler struct{ Service *application.Service }

type CreateLocationRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type CreateLocationResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

type DashboardHandler struct{ Service *application.Service }

type DashboardRequest struct{}

type DashboardResponse struct {
	TotalSales    float64 `json:"total_sales"`
	OrderCount    int     `json:"order_count"`
	AvgOrder      float64 `json:"avg_order"`
	TableTurnover float64 `json:"table_turnover"`
}
