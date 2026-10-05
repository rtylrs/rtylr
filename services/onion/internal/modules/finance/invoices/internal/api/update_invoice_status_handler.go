package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type UpdateInvoiceStatusHandler struct {
	FinanceService *application.Service
}

type UpdateInvoiceStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type UpdateInvoiceStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateInvoiceStatusHandler) Validate(_ context.Context, req *UpdateInvoiceStatusRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invoice ID is required")
	}

	if req.Status == "" {
		return onekiterrors.New(http.StatusBadRequest, "Status is required")
	}

	// The valid-status set is the domain's. "paid" and "partial" are derived
	// from the recorded payments, so a status-only write cannot set them.
	if err := domain.ValidateSettableInvoiceStatus(req.Status); err != nil {
		return invoiceServiceError(err, "Failed to update invoice status")
	}

	return nil
}
