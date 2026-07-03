package handlers

import (
	"context"
	"net/http"

	possummaryv1 "github.com/voxire/voxie/gen/go/pos/day_summary/v1"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type DaySummaryAPI struct {
	Service *application.Service
}

func NewDaySummaryAPI(service *application.Service) *DaySummaryAPI {
	return &DaySummaryAPI{Service: service}
}

func (a *DaySummaryAPI) DaySummary(ctx context.Context, req *possummaryv1.DaySummaryRequest) (*possummaryv1.DaySummaryResponse, error) {
	if !reDate.MatchString(req.GetDate()) {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "date must be in YYYY-MM-DD format"}
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, errorModels.PermissionDeniedServerError()
	}

	summary, err := a.Service.GetDaySummary(ctx, businessID, req.GetDate())
	if err != nil {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get day summary"}
	}

	return &possummaryv1.DaySummaryResponse{
		TotalSales:       summary.TotalSales,
		TotalTax:         summary.TotalTax,
		TotalTips:        summary.TotalTips,
		TotalRefunds:     summary.TotalRefunds,
		TotalOrders:      int32(summary.TotalOrders),
		PaymentBreakdown: summary.PaymentBreakdown,
	}, nil
}
