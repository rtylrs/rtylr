package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/billing/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubbillingv1 "github.com/voxire/voxie/gen/go/hub/billing/v1"
)

type BillingAPI struct {
	Service *application.Service
}

func NewBillingAPI(service *application.Service) *BillingAPI {
	return &BillingAPI{Service: service}
}

func (a *BillingAPI) GetBillingSummary(ctx context.Context, _ *hubbillingv1.GetBillingSummaryRequest) (*hubbillingv1.GetBillingSummaryResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewBilling(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	summary, err := a.Service.GetSummary(ctx, device.User.BusinessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get billing summary")
	}

	bs := &hubbillingv1.BillingSummary{
		Plan:         summary.Plan,
		Status:       summary.Status,
		CurrentUsage: summary.CurrentUsage,
		UsageLimit:   summary.UsageLimit,
	}
	if summary.NextBillingDate != nil {
		bs.NextBillingDate = summary.NextBillingDate.Format("2006-01-02T15:04:05Z07:00")
	}

	return &hubbillingv1.GetBillingSummaryResponse{Summary: bs}, nil
}

func (a *BillingAPI) ListInvoices(ctx context.Context, _ *hubbillingv1.ListInvoicesRequest) (*hubbillingv1.ListInvoicesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewBilling(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	invoices, err := a.Service.ListInvoices(ctx, device.User.BusinessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list invoices")
	}

	payloads := make([]*hubbillingv1.Invoice, 0, len(invoices))
	for _, inv := range invoices {
		iv := &hubbillingv1.Invoice{
			Id:       inv.ID,
			Amount:   inv.Amount.Float64(),
			Currency: inv.Currency,
			Status:   inv.Status,
			IssuedAt: inv.IssuedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if inv.PaidAt != nil {
			iv.PaidAt = inv.PaidAt.Format("2006-01-02T15:04:05Z07:00")
		}
		payloads = append(payloads, iv)
	}

	return &hubbillingv1.ListInvoicesResponse{Invoices: payloads}, nil
}

// canViewBilling reports whether the user may read the business's billing,
// which belongs to the business, not to each member.
func canViewBilling(user *authPkg.AuthUser) bool {
	return authPkg.IsBusinessAdmin(user) ||
		authPkg.HasDashboardPermission(user, "hub:billing:view") ||
		authPkg.HasDashboardPermission(user, "hub:billing:manage")
}
