package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Locations ───────────────────────────────────────────────────────────

func ListLocationsRoutes() []string { return []string{"/pos/locations"} }

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

func newLocationPayload(loc domain.Location) LocationPayload {
	return LocationPayload{
		ID:        loc.ID,
		OwnerID:   loc.OwnerID,
		Name:      loc.Name,
		Address:   loc.Address,
		Phone:     loc.Phone,
		IsActive:  loc.IsActive,
		CreatedAt: loc.CreatedAt,
	}
}

func (h *ListLocationsHandler) Serve(ctx context.Context, _ *ListLocationsRequest) (ListLocationsResponse, *errorModels.ServerError) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return ListLocationsResponse{}, authErr
	}

	ownerID := resolvedOwnerID(ctx, user)

	locations, err := h.Service.ListLocations(ctx, ownerID)
	if err != nil {
		return ListLocationsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list locations",
		}
	}

	payloads := make([]LocationPayload, 0, len(locations))
	for _, loc := range locations {
		payloads = append(payloads, newLocationPayload(loc))
	}

	return ListLocationsResponse{Locations: payloads}, nil
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

func (h *CreateLocationHandler) Validate(_ context.Context, req *CreateLocationRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Location name is required",
		}
	}

	return nil
}

func (h *CreateLocationHandler) Serve(ctx context.Context, req *CreateLocationRequest) (CreateLocationResponse, *errorModels.ServerError) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return CreateLocationResponse{}, authErr
	}

	ownerID := resolvedOwnerID(ctx, user)

	id, err := h.Service.CreateLocation(ctx, ownerID, domain.LocationInput{
		Name:    req.Name,
		Address: req.Address,
		Phone:   req.Phone,
	})
	if err != nil {
		return CreateLocationResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create location",
		}
	}

	return CreateLocationResponse{ID: id, Message: "Location created"}, nil
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func DashboardRoutes() []string { return []string{"/pos/dashboard"} }

type DashboardHandler struct{ Service *application.Service }

type DashboardRequest struct{}

type DashboardResponse struct {
	TotalSales    float64 `json:"total_sales"`
	OrderCount    int     `json:"order_count"`
	AvgOrder      float64 `json:"avg_order"`
	TableTurnover float64 `json:"table_turnover"`
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return DashboardResponse{}, authErr
	}

	ownerID := resolvedOwnerID(ctx, user)

	stats, err := h.Service.GetDashboardStats(ctx, ownerID)
	if err != nil {
		return DashboardResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get dashboard stats",
		}
	}

	return DashboardResponse{
		TotalSales:    stats.TotalSales,
		OrderCount:    stats.OrderCount,
		AvgOrder:      stats.AvgOrder,
		TableTurnover: stats.TableTurnover,
	}, nil
}
