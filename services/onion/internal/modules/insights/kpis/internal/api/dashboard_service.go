package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/kpis/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubdashboardv1 "github.com/voxire/voxie/gen/go/hub/dashboard/v1"
)

type DashboardAPI struct {
	Service *application.Service
}

func NewDashboardAPI(service *application.Service) *DashboardAPI {
	return &DashboardAPI{Service: service}
}

func (a *DashboardAPI) GetKPIs(ctx context.Context, _ *hubdashboardv1.GetKPIsRequest) (*hubdashboardv1.GetKPIsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if err := requireDashboardPermission(ctx); err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	kpi, err := a.Service.GetKPIs(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get KPIs")
	}

	return &hubdashboardv1.GetKPIsResponse{
		Revenue:       kpi.Revenue,
		OrderCount:    kpi.OrderCount,
		StaffCount:    kpi.StaffCount,
		CustomerCount: kpi.CustomerCount,
		LowStockCount: kpi.LowStockCount,
		ActiveApps:    kpi.ActiveApps,
	}, nil
}
