package handlers

import (
	"context"
	"net/http"

	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"

	"github.com/voxire/onion/internal/modules/billing/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SummaryRoutes() []string  { return []string{"/billing/summary"} }
func InvoicesRoutes() []string { return []string{"/billing/invoices"} }

// -- Billing Summary --

type SummaryHandler struct{ Service *application.Service }
type SummaryRequest struct{}

type SummaryResponse struct {
	Plan            string     `json:"plan"`
	Status          string     `json:"status"`
	CurrentUsage    float64    `json:"current_usage"`
	UsageLimit      float64    `json:"usage_limit"`
	NextBillingDate *time.Time `json:"next_billing_date"`
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

	return SummaryResponse{
		Plan:            summary.Plan,
		Status:          summary.Status,
		CurrentUsage:    summary.CurrentUsage,
		UsageLimit:      summary.UsageLimit,
		NextBillingDate: summary.NextBillingDate,
	}, nil
}

// -- List Invoices --

type InvoicesHandler struct{ Service *application.Service }
type InvoicesRequest struct{}

type InvoicePayload struct {
	ID       int64      `json:"id"`
	Amount   float64    `json:"amount"`
	Currency string     `json:"currency"`
	Status   string     `json:"status"`
	IssuedAt time.Time  `json:"issued_at"`
	PaidAt   *time.Time `json:"paid_at"`
}

type InvoicesResponse struct {
	Invoices []InvoicePayload `json:"invoices"`
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

	payloads := make([]InvoicePayload, 0, len(invoices))
	for _, inv := range invoices {
		payloads = append(payloads, InvoicePayload{
			ID:       inv.ID,
			Amount:   inv.Amount,
			Currency: inv.Currency,
			Status:   inv.Status,
			IssuedAt: inv.IssuedAt,
			PaidAt:   inv.PaidAt,
		})
	}

	return InvoicesResponse{Invoices: payloads}, nil
}
