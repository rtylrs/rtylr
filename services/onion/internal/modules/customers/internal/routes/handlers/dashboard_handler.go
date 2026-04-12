package handlers

import (
	"context"
	"net/http"

	crmv1 "github.com/voxire/voxie/gen/go/crm/v1"

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

type DashboardResponse struct {
	TotalCustomers       int                    `json:"total_customers"`
	NewThisMonth         int                    `json:"new_this_month"`
	ActiveLoyaltyMembers int                    `json:"active_loyalty_members"`
	TopCustomers         []*crmv1.TopCustomer   `json:"top_customers"`
}

func newTopCustomerPayload(tc domain.TopCustomer) *crmv1.TopCustomer {
	return &crmv1.TopCustomer{
		Id:            tc.ID,
		FirstName:     tc.FirstName,
		LastName:      tc.LastName,
		Email:         tc.Email,
		TotalSpent:    tc.TotalSpent,
		LoyaltyPoints: int64(tc.LoyaltyPoints),
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

	topCustomers := make([]*crmv1.TopCustomer, len(stats.TopCustomers))
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
