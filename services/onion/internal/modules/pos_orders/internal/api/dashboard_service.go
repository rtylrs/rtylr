package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	posdashboardv1 "github.com/voxire/voxie/gen/go/pos/dashboard/v1"
)

type DashboardAPI struct {
	Service *application.Service
}

func NewDashboardAPI(service *application.Service) *DashboardAPI {
	return &DashboardAPI{Service: service}
}

func (a *DashboardAPI) Dashboard(ctx context.Context, _ *posdashboardv1.DashboardRequest) (*posdashboardv1.DashboardResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:reports:view")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	stats, err := a.Service.GetDashboardStats(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get dashboard stats")
	}

	return &posdashboardv1.DashboardResponse{
		TotalSales:    stats.TotalSales,
		OrderCount:    int32(stats.OrderCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		AvgOrder:      stats.AvgOrder,
		TableTurnover: stats.TableTurnover,
	}, nil
}
