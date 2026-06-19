package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/revenue?from=&to=
//
// Returns a contiguous daily paid-revenue series (int64 minor units) over
// the requested window. Defaults to the last 30 days when unspecified.

func GetRevenueSeriesRoutes() []string { return []string{"/insights/revenue"} }

type GetRevenueSeriesHandler struct{ Service *application.Service }

type GetRevenueSeriesRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GetRevenueSeriesResponse struct {
	Series []domain.RevenuePoint `json:"series"`
}

func (h *GetRevenueSeriesHandler) Validate(_ context.Context, req *GetRevenueSeriesRequest) *errorModels.ServerError {
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	return nil
}

func (h *GetRevenueSeriesHandler) Serve(ctx context.Context, req *GetRevenueSeriesRequest) (GetRevenueSeriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetRevenueSeriesResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetRevenueSeriesResponse{}, errorModels.PermissionDeniedServerError()
	}

	from, to, err := application.ResolveDateRange(req.From, req.To)
	if err != nil {
		return GetRevenueSeriesResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid date range",
		}
	}

	series, err := h.Service.GetRevenueSeries(ctx, businessID, from, to)
	if err != nil {
		return GetRevenueSeriesResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load revenue series",
		}
	}
	return GetRevenueSeriesResponse{Series: series}, nil
}
