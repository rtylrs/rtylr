package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/dashboard
//
// Returns the four headline KPI cards: MTD revenue, customers, orders,
// and margin %, each paired with a prior-MTD delta_pct.

func GetDashboardRoutes() []string { return []string{"/insights/dashboard"} }

type GetDashboardHandler struct{ Service *application.Service }

type GetDashboardRequest struct{}

type GetDashboardResponse struct {
	Revenue   domain.MoneyCard `json:"revenue"`
	Customers domain.CountCard `json:"customers"`
	Orders    domain.CountCard `json:"orders"`
	Margin    domain.PctCard   `json:"margin"`
}

func (h *GetDashboardHandler) Serve(ctx context.Context, _ *GetDashboardRequest) (GetDashboardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetDashboardResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetDashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	snap, err := h.Service.GetDashboard(ctx, businessID)
	if err != nil {
		return GetDashboardResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load insights dashboard",
		}
	}

	return GetDashboardResponse{
		Revenue:   snap.Revenue,
		Customers: snap.Customers,
		Orders:    snap.Orders,
		Margin:    snap.Margin,
	}, nil
}
