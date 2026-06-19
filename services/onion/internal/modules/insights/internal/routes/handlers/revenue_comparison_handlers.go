package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/revenue/comparison?from=&to=
//
// Period-over-period paid revenue totals plus a computed delta percentage.

func GetRevenueComparisonRoutes() []string { return []string{"/insights/revenue/comparison"} }

type GetRevenueComparisonHandler struct{ Service *application.Service }

type GetRevenueComparisonRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GetRevenueComparisonResponse struct {
	ThisPeriod int64   `json:"this_period"`
	LastPeriod int64   `json:"last_period"`
	DeltaPct   float64 `json:"delta_pct"`
}

func (h *GetRevenueComparisonHandler) Validate(_ context.Context, req *GetRevenueComparisonRequest) *errorModels.ServerError {
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	return nil
}

func (h *GetRevenueComparisonHandler) Serve(ctx context.Context, req *GetRevenueComparisonRequest) (GetRevenueComparisonResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetRevenueComparisonResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetRevenueComparisonResponse{}, errorModels.PermissionDeniedServerError()
	}
	from, to, err := application.ResolveDateRange(req.From, req.To)
	if err != nil {
		return GetRevenueComparisonResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid date range",
		}
	}
	payload, err := h.Service.GetRevenueComparison(ctx, businessID, from, to)
	if err != nil {
		return GetRevenueComparisonResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load revenue comparison",
		}
	}
	return GetRevenueComparisonResponse{
		ThisPeriod: payload.ThisPeriodMinorUnits,
		LastPeriod: payload.LastPeriodMinorUnits,
		DeltaPct:   payload.DeltaPct,
	}, nil
}
