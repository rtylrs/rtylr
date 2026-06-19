package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
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

	if device.User == nil {
		return DeleteInvoiceResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteInvoiceResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FinanceService.DeleteInvoice(ctx, businessID, req.ID); err != nil {
		return DeleteInvoiceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete invoice",
		}
	}

	return DeleteInvoiceResponse{Message: "Invoice deleted successfully"}, nil
}
