package handlers

import (
	"context"
	"net/http"

	bookv1 "github.com/voxire/voxie/gen/go/book/v1"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func DashboardRoutes() []string { return []string{"/book/dashboard"} }

type DashboardHandler struct{ Service *application.Service }
type DashboardRequest struct{}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (*bookv1.DashboardResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return nil, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	data, err := h.Service.GetDashboard(ctx, authUser.UserID)
	if err != nil { return nil, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to get dashboard"} }
	return &bookv1.DashboardResponse{TodayCount: int32(data.TodayCount), UpcomingCount: int32(data.UpcomingCount), CancellationRate: data.CancellationRate, Revenue: data.Revenue}, nil
}
