package handlers

import (
	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/application"
)

func DaySummaryRoutes() []string { return []string{"/pos/day-summary"} }

type DaySummaryHandler struct{ Service *application.Service }

type DaySummaryRequest struct {
	Date string `json:"date"`
}

type DaySummaryResponse struct {
	Summary *posv1.DaySummary `json:"summary"`
}

func (h *DaySummaryHandler) Serve(ctx context.Context, req *DaySummaryRequest) (DaySummaryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DaySummaryResponse{}, errorModels.PermissionDeniedServerError()
	}

	summary, err := h.Service.GetDaySummary(ctx, device.User.UserID, req.Date)
	if err != nil {
		return DaySummaryResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get day summary"}
	}

	return DaySummaryResponse{
		Summary: &posv1.DaySummary{
			Date:         req.Date,
			TotalSales:   summary.TotalSales,
			TotalTax:     summary.TotalTax,
			TotalTips:    summary.TotalTips,
			TotalRefunds: summary.TotalRefunds,
			TotalOrders:  int32(summary.TotalOrders),
		},
	}, nil
}
