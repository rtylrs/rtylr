package handlers

import (
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
	TotalSales       float64            `json:"totalSales"`
	TotalTax         float64            `json:"totalTax"`
	TotalTips        float64            `json:"totalTips"`
	TotalRefunds     float64            `json:"totalRefunds"`
	TotalOrders      int                `json:"totalOrders"`
	PaymentBreakdown map[string]float64 `json:"paymentBreakdown"`
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
		TotalSales:       summary.TotalSales,
		TotalTax:         summary.TotalTax,
		TotalTips:        summary.TotalTips,
		TotalRefunds:     summary.TotalRefunds,
		TotalOrders:      summary.TotalOrders,
		PaymentBreakdown: summary.PaymentBreakdown,
	}, nil
}
