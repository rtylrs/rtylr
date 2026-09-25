package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmdashboardv1 "github.com/voxire/voxie/gen/go/crm/dashboard/v1"
)

func DashboardRoutes() []string {
	return []string{"/crm/dashboard"}
}

type DashboardHandler struct {
	CRMService *application.Service
}

type DashboardRequest struct{}

type DashboardResponse struct {
	TotalCustomers       int                           `json:"total_customers"`
	NewThisMonth         int                           `json:"new_this_month"`
	ActiveLoyaltyMembers int                           `json:"active_loyalty_members"`
	TopCustomers         []*crmdashboardv1.TopCustomer `json:"top_customers"`
}

func newTopCustomerPayload(tc domain.TopCustomer) *crmdashboardv1.TopCustomer {
	return &crmdashboardv1.TopCustomer{
		Id:            tc.ID,
		FirstName:     tc.FirstName,
		LastName:      tc.LastName,
		Email:         tc.Email,
		TotalSpent:    tc.TotalSpent,
		LoyaltyPoints: int64(tc.LoyaltyPoints),
		LoyaltyTier:   tc.LoyaltyTier,
	}
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DashboardResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return DashboardResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DashboardResponse{}, onekiterrors.PermissionDenied()
	}

	stats, err := h.CRMService.GetDashboardStats(ctx, businessID)
	if err != nil {
		return DashboardResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get dashboard stats")
	}

	topCustomers := make([]*crmdashboardv1.TopCustomer, len(stats.TopCustomers))
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
