package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *DeleteInvoiceHandler) Validate(_ context.Context, req *DeleteInvoiceRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invoice ID is required")
	}

	return nil
}

func (h *DeleteInvoiceHandler) Serve(ctx context.Context, req *DeleteInvoiceRequest) (DeleteInvoiceResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return DeleteInvoiceResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteInvoiceResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.FinanceService.DeleteInvoice(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return DeleteInvoiceResponse{}, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		if errors.Is(err, domain.ErrInvoiceNotDeletable) {
			return DeleteInvoiceResponse{}, onekiterrors.New(http.StatusConflict, "Only draft invoices can be deleted")
		}
		return DeleteInvoiceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete invoice")
	}

	return DeleteInvoiceResponse{Message: "Invoice deleted successfully"}, nil
}
