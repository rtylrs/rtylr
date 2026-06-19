package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func SendInvoiceRoutes() []string {
	return []string{"/finance/invoices/{id}/send"}
}

type SendInvoiceHandler struct {
	FinanceService *application.Service
}

type SendInvoiceRequest struct {
	ID int64 `json:"id"`
}

type SendInvoiceResponse struct {
	Message string `json:"message"`
}

func (h *SendInvoiceHandler) Validate(_ context.Context, req *SendInvoiceRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		}
	}

	return nil
}

func (h *SendInvoiceHandler) Serve(ctx context.Context, req *SendInvoiceRequest) (SendInvoiceResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return SendInvoiceResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SendInvoiceResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FinanceService.SendInvoice(ctx, businessID, req.ID); err != nil {
		return SendInvoiceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send invoice",
		}
	}

	return SendInvoiceResponse{Message: "Invoice sent successfully"}, nil
}
