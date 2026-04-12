package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
)

func DashboardRoutes() []string {
	return []string{"/crm/dashboard"}
}

type DashboardHandler struct {
	CRMService *application.Service
}

type DashboardRequest struct{}

type TopCustomerPayload struct {
	ID            int64   `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	TotalSpent    float64 `json:"total_spent"`
	LoyaltyPoints int     `json:"loyalty_points"`
	LoyaltyTier   string  `json:"loyalty_tier"`
}

type DashboardResponse struct {
	TotalCustomers       int                  `json:"total_customers"`
	NewThisMonth         int                  `json:"new_this_month"`
	ActiveLoyaltyMembers int                  `json:"active_loyalty_members"`
	TopCustomers         []TopCustomerPayload `json:"top_customers"`
}

func newTopCustomerPayload(tc domain.TopCustomer) TopCustomerPayload {
	return TopCustomerPayload{
		ID:            tc.ID,
		FirstName:     tc.FirstName,
		LastName:      tc.LastName,
		Email:         tc.Email,
		TotalSpent:    tc.TotalSpent,
		LoyaltyPoints: tc.LoyaltyPoints,
		LoyaltyTier:   tc.LoyaltyTier,
	}
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	stats, err := h.CRMService.GetDashboardStats(ctx, device.User.UserID)
	if err != nil {
		return DashboardResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get dashboard stats",
		}
	}

	topCustomers := make([]TopCustomerPayload, len(stats.TopCustomers))
	for i, tc := range stats.TopCustomers {
		topCustomers[i] = newTopCustomerPayload(tc)
	}

	return DashboardResponse{
		TotalCustomers:       stats.TotalCustomers,
		NewThisMonth:         stats.NewThisMonth,
		ActiveLoyaltyMembers: stats.ActiveLoyaltyMembers,
		TopCustomers:         topCustomers,
	}, nil
}
