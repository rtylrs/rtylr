package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

func UpdateInvoiceRoutes() []string {
	return []string{"/finance/invoices"}
}

type UpdateInvoiceHandler struct {
	FinanceService *application.Service
}

type UpdateInvoiceRequest struct {
	ID            int64                `json:"id"`
	ClientName    string               `json:"client_name"`
	ClientEmail   string               `json:"client_email"`
	ClientAddress string               `json:"client_address"`
	ClientTaxID   string               `json:"client_tax_id"`
	IssueDate     string               `json:"issue_date"`
	DueDate       string               `json:"due_date"`
	Currency      string               `json:"currency"`
	DiscountType  string               `json:"discount_type"`
	DiscountValue float64              `json:"discount_value"`
	TaxRate       float64              `json:"tax_rate"`
	Notes         string               `json:"notes"`
	Terms         string               `json:"terms"`
	Items         []InvoiceItemPayload `json:"items"`
}

type UpdateInvoiceResponse struct {
	Invoice *financeinvoicesv1.InvoiceDetail `json:"invoice"`
}

func (h *UpdateInvoiceHandler) Validate(_ context.Context, req *UpdateInvoiceRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invoice ID is required")
	}

	if strings.TrimSpace(req.ClientName) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Client name is required")
	}

	if strings.TrimSpace(req.IssueDate) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Issue date is required")
	}

	if _, err := time.Parse("2006-01-02", req.IssueDate); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "Issue date must be in YYYY-MM-DD format")
	}

	if strings.TrimSpace(req.DueDate) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Due date is required")
	}

	if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "Due date must be in YYYY-MM-DD format")
	}

	return nil
}

func (h *UpdateInvoiceHandler) Serve(ctx context.Context, req *UpdateInvoiceRequest) (UpdateInvoiceResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return UpdateInvoiceResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:invoices:create"); err != nil {
		return UpdateInvoiceResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateInvoiceResponse{}, onekiterrors.PermissionDenied()
	}

	items := make([]domain.InvoiceItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.InvoiceItemInput{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			TaxRate:     item.TaxRate,
			DiscountPct: item.DiscountPct,
			SortOrder:   item.SortOrder,
		})
	}

	input := domain.InvoiceUpdateInput{
		ClientName:    req.ClientName,
		ClientEmail:   req.ClientEmail,
		ClientAddress: req.ClientAddress,
		ClientTaxID:   req.ClientTaxID,
		IssueDate:     req.IssueDate,
		DueDate:       req.DueDate,
		Currency:      req.Currency,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		TaxRate:       req.TaxRate,
		Notes:         req.Notes,
		Terms:         req.Terms,
		Items:         items,
	}

	inv, err := h.FinanceService.UpdateInvoice(ctx, businessID, req.ID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return UpdateInvoiceResponse{}, onekiterrors.New(http.StatusNotFound, "Invoice not found")
		}
		if errors.Is(err, domain.ErrInvoiceNotDraft) {
			return UpdateInvoiceResponse{}, onekiterrors.New(http.StatusConflict, "Only draft invoices can be edited")
		}

		return UpdateInvoiceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update invoice")
	}

	return UpdateInvoiceResponse{Invoice: newInvoiceDetailPayload(inv)}, nil
}
