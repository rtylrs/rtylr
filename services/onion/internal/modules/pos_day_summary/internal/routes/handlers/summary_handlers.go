package handlers

import (
	"context"
	"net/http"
	"regexp"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

var reDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func DaySummaryRoutes() []string { return []string{"/pos/day-summary"} }

type DaySummaryHandler struct{ Service *application.Service }

type DaySummaryRequest struct {
	Date string `json:"date"`
}

type DaySummaryResponse struct {
	TotalSales       float64            `json:"total_sales"`
	TotalTax         float64            `json:"total_tax"`
	TotalTips        float64            `json:"total_tips"`
	TotalRefunds     float64            `json:"total_refunds"`
	TotalOrders      int                `json:"total_orders"`
	PaymentBreakdown map[string]float64 `json:"payment_breakdown"`
}

func (h *DaySummaryHandler) Validate(_ context.Context, req *DaySummaryRequest) *errorModels.ServerError {
	if req.Date == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "date is required"}
	}
	if !reDate.MatchString(req.Date) {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "date must be in YYYY-MM-DD format"}
	}
	return nil
}

func (h *DaySummaryHandler) Serve(ctx context.Context, req *DaySummaryRequest) (DaySummaryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DaySummaryResponse{}, errorModels.PermissionDeniedServerError()
	}
	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)

	summary, err := h.Service.GetDaySummary(ctx, ownerID, req.Date)
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
