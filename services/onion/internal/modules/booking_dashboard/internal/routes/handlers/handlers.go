package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DashboardRoutes() []string { return []string{"/book/dashboard"} }

type (
	DashboardHandler  struct{ Service *application.Service }
	DashboardRequest  struct{}
	DashboardResponse struct {
		TodayCount       int     `json:"today_count"`
		UpcomingCount    int     `json:"upcoming_count"`
		CancellationRate float64 `json:"cancellation_rate"`
		Revenue          float64 `json:"revenue"`
	}
)

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DashboardResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DashboardResponse{}, errorModels.PermissionDeniedServerError()
	}
	data, err := h.Service.GetDashboard(ctx, businessID)
	if err != nil {
		return DashboardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to get dashboard"}
	}
	return DashboardResponse{TodayCount: data.TodayCount, UpcomingCount: data.UpcomingCount, CancellationRate: data.CancellationRate, Revenue: data.Revenue}, nil
}
