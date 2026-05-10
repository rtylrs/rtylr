package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/billing/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"
)

func SummaryRoutes() []string  { return []string{"/billing/summary"} }
func InvoicesRoutes() []string { return []string{"/billing/invoices"} }

// -- Billing Summary --

type (
	SummaryHandler struct{ Service *application.Service }
	SummaryRequest struct{}
)

type SummaryResponse struct {
	Summary *hubv1.BillingSummary `json:"summary"`
}

func (h *SummaryHandler) Serve(ctx context.Context, _ *SummaryRequest) (SummaryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SummaryResponse{}, errorModels.PermissionDeniedServerError()
	}

	summary, err := h.Service.GetSummary(ctx, device.User.UserID)
	if err != nil {
		return SummaryResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get billing summary",
		}
	}

	bs := &hubv1.BillingSummary{
		Plan:         summary.Plan,
		Status:       summary.Status,
		CurrentUsage: summary.CurrentUsage,
		UsageLimit:   summary.UsageLimit,
	}
	if summary.NextBillingDate != nil {
		bs.NextBillingDate = summary.NextBillingDate.Format("2006-01-02T15:04:05Z07:00")
	}

	return SummaryResponse{Summary: bs}, nil
}

// -- List Invoices --

type (
	InvoicesHandler struct{ Service *application.Service }
	InvoicesRequest struct{}
)

type InvoicesResponse struct {
	Invoices []*hubv1.Invoice `json:"invoices"`
}

func (h *InvoicesHandler) Serve(ctx context.Context, _ *InvoicesRequest) (InvoicesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return InvoicesResponse{}, errorModels.PermissionDeniedServerError()
	}

	invoices, err := h.Service.ListInvoices(ctx, device.User.UserID)
	if err != nil {
		return InvoicesResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list invoices",
		}
	}

	payloads := make([]*hubv1.Invoice, 0, len(invoices))
	for _, inv := range invoices {
		iv := &hubv1.Invoice{
			Id:       inv.ID,
			Amount:   inv.Amount,
			Currency: inv.Currency,
			Status:   inv.Status,
			IssuedAt: inv.IssuedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if inv.PaidAt != nil {
			iv.PaidAt = inv.PaidAt.Format("2006-01-02T15:04:05Z07:00")
		}
		payloads = append(payloads, iv)
	}

	return InvoicesResponse{Invoices: payloads}, nil
}
