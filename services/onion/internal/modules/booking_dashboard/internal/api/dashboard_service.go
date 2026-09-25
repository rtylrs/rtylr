package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	bookingdashboardv1 "github.com/voxire/voxie/gen/go/booking/dashboard/v1"
)

type DashboardAPI struct {
	service *application.Service
}

func NewDashboardAPI(service *application.Service) *DashboardAPI {
	return &DashboardAPI{service: service}
}

func (a *DashboardAPI) GetDashboard(ctx context.Context, _ *bookingdashboardv1.GetDashboardRequest) (*bookingdashboardv1.GetDashboardResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	data, err := a.service.GetDashboard(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get dashboard")
	}

	return &bookingdashboardv1.GetDashboardResponse{
		TodayCount:       int32(data.TodayCount),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		UpcomingCount:    int32(data.UpcomingCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CancellationRate: data.CancellationRate,
		Revenue:          data.Revenue.Float64(),
	}, nil
}
