package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func DeleteInvoiceRoutes() []string {
	return []string{"/finance/invoices"}
}

type DeleteInvoiceHandler struct {
	FinanceService *application.Service
}

type DeleteInvoiceRequest struct {
	ID int64 `json:"id"`
}

type DeleteInvoiceResponse struct {
	Message string `json:"message"`
}

func (h *DeleteInvoiceHandler) Validate(_ context.Context, req *DeleteInvoiceRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		}
	}

	return nil
}

func (h *DeleteInvoiceHandler) Serve(ctx context.Context, req *DeleteInvoiceRequest) (DeleteInvoiceResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	if err := h.FinanceService.DeleteInvoice(ctx, device.User.UserID, req.ID); err != nil {
		return DeleteInvoiceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete invoice",
		}
	}

	return DeleteInvoiceResponse{Message: "Invoice deleted successfully"}, nil
}
