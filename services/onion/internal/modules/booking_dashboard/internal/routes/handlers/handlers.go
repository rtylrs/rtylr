package handlers

import (
	bookv1 "github.com/voxire/voxie/gen/go/book/v1"

	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func DashboardRoutes() []string { return []string{"/book/dashboard"} }

type DashboardHandler struct{ Service *application.Service }
type DashboardRequest struct{}
type DashboardResponse struct {
	Stats *bookv1.DashboardStats `json:"stats"`
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return DashboardResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	data, err := h.Service.GetDashboard(ctx, authUser.UserID)
	if err != nil { return DashboardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to get dashboard"} }
	return DashboardResponse{Stats: &bookv1.DashboardStats{
		TodayAppointments:    int32(data.TodayCount),
		UpcomingAppointments: int32(data.UpcomingCount),
		TotalRevenue:         data.Revenue,
	}}, nil
}
