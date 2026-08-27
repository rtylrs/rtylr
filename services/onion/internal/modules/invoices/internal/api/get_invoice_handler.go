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
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
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

type GetInvoiceResponse struct {
	Invoice  *financeinvoicesv1.InvoiceDetail `json:"invoice"`
	Payments []*financeinvoicesv1.Payment     `json:"payments"`
}

func (h *GetInvoiceHandler) Validate(_ context.Context, req *GetInvoiceRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invoice ID is required")
	}

	return nil
}

func (h *GetInvoiceHandler) Serve(ctx context.Context, req *GetInvoiceRequest) (GetInvoiceResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return GetInvoiceResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:invoices:view"); err != nil {
		return GetInvoiceResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetInvoiceResponse{}, onekiterrors.PermissionDenied()
	}

	inv, err := h.FinanceService.GetInvoice(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return GetInvoiceResponse{}, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}

		return GetInvoiceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get invoice")
	}

	payments := make([]*financeinvoicesv1.Payment, 0, len(inv.Payments))
	for _, p := range inv.Payments {
		payments = append(payments, &financeinvoicesv1.Payment{
			Id:          p.ID,
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
