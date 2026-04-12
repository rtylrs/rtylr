package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func GetInvoiceRoutes() []string {
	return []string{"/finance/invoices/detail"}
}

type GetInvoiceHandler struct {
	FinanceService *application.Service
}

type GetInvoiceRequest struct {
	ID int64 `json:"id"`
}

type PaymentPayload struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	PaymentDate string  `json:"payment_date"`
	Method      string  `json:"method"`
	Reference   string  `json:"reference"`
	Notes       string  `json:"notes"`
}

type GetInvoiceResponse struct {
	Invoice  InvoiceDetailPayload `json:"invoice"`
	Payments []PaymentPayload     `json:"payments"`
}

func (h *GetInvoiceHandler) Validate(_ context.Context, req *GetInvoiceRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		}
	}

	return nil
}

func (h *GetInvoiceHandler) Serve(ctx context.Context, req *GetInvoiceRequest) (GetInvoiceResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	inv, err := h.FinanceService.GetInvoice(ctx, device.User.UserID, req.ID)
	if err != nil {
		if err == domain.ErrInvoiceNotFound {
			return GetInvoiceResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Invoice not found",
			}
		}

		return GetInvoiceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get invoice",
		}
	}

	payments := make([]PaymentPayload, 0, len(inv.Payments))
	for _, p := range inv.Payments {
		payments = append(payments, PaymentPayload{
			ID:          p.ID,
			Amount:      p.Amount,
			Currency:    p.Currency,
			PaymentDate: p.PaymentDate.Format("2006-01-02"),
			Method:      p.Method,
			Reference:   p.Reference,
			Notes:       p.Notes,
		})
	}

	return GetInvoiceResponse{
		Invoice:  newInvoiceDetailPayload(inv),
		Payments: payments,
	}, nil
}
