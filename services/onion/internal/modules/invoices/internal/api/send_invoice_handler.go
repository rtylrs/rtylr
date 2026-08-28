package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *SendInvoiceHandler) Validate(_ context.Context, req *SendInvoiceRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invoice ID is required")
	}

	return nil
}

func (h *SendInvoiceHandler) Serve(ctx context.Context, req *SendInvoiceRequest) (SendInvoiceResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return SendInvoiceResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:invoices:send"); err != nil {
		return SendInvoiceResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SendInvoiceResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.FinanceService.SendInvoice(ctx, businessID, req.ID); err != nil {
		return SendInvoiceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to send invoice")
	}

	return SendInvoiceResponse{Message: "Invoice sent successfully"}, nil
}
