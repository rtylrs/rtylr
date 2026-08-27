package api

import (
	"context"
	"net/http"
	"regexp"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	possummaryv1 "github.com/voxire/voxie/gen/go/pos/day_summary/v1"
)

var reDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type DaySummaryAPI struct {
	Service *application.Service
}

func NewDaySummaryAPI(service *application.Service) *DaySummaryAPI {
	return &DaySummaryAPI{Service: service}
}

func (a *DaySummaryAPI) DaySummary(ctx context.Context, req *possummaryv1.DaySummaryRequest) (*possummaryv1.DaySummaryResponse, error) {
	if !reDate.MatchString(req.GetDate()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "date must be in YYYY-MM-DD format")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppPOS, "pos:reports:view") || !authPkg.HasAnyAppScope(device.User, authorization.AppPOS) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	summary, err := a.Service.GetDaySummary(ctx, businessID, req.GetDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get day summary")
	}

	return &possummaryv1.DaySummaryResponse{
		TotalSales:       summary.TotalSales,
		TotalTax:         summary.TotalTax,
		TotalTips:        summary.TotalTips,
		TotalRefunds:     summary.TotalRefunds,
		TotalOrders:      int32(summary.TotalOrders), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PaymentBreakdown: summary.PaymentBreakdown,
	}, nil
}
