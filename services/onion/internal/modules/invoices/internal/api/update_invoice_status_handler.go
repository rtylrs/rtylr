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

func UpdateInvoiceStatusRoutes() []string {
	return []string{"/finance/invoices/status"}
}

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

var validInvoiceStatuses = map[string]struct{}{
	"draft":     {},
	"sent":      {},
	"viewed":    {},
	"partial":   {},
	"paid":      {},
	"overdue":   {},
	"cancelled": {},
	"void":      {},
}

func (h *UpdateInvoiceStatusHandler) Validate(_ context.Context, req *UpdateInvoiceStatusRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invoice ID is required")
	}

	if req.Status == "" {
		return onekiterrors.New(http.StatusBadRequest, "Status is required")
	}

	if _, ok := validInvoiceStatuses[req.Status]; !ok {
		return onekiterrors.New(http.StatusBadRequest, "Invalid invoice status")
	}

	return nil
}

func (h *UpdateInvoiceStatusHandler) Serve(ctx context.Context, req *UpdateInvoiceStatusRequest) (UpdateInvoiceStatusResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return UpdateInvoiceStatusResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return UpdateInvoiceStatusResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateInvoiceStatusResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.FinanceService.UpdateInvoiceStatus(ctx, businessID, req.ID, req.Status); err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return UpdateInvoiceStatusResponse{}, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		if errors.Is(err, domain.ErrInvoiceNotDraft) {
			return UpdateInvoiceStatusResponse{}, onekiterrors.New(http.StatusConflict, "Invoice lifecycle does not allow this status change")
		}
		return UpdateInvoiceStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update invoice status")
	}

	return UpdateInvoiceStatusResponse{Message: "Invoice status updated successfully"}, nil
}
