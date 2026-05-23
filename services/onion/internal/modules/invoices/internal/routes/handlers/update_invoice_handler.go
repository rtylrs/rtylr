package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	finv1 "github.com/voxire/voxie/gen/go/finance/v1"
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
	Invoice *finv1.InvoiceDetail `json:"invoice"`
}

func (h *UpdateInvoiceHandler) Validate(_ context.Context, req *UpdateInvoiceRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invoice ID is required",
		}
	}

	if req.ClientName == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Client name is required",
		}
	}

	if req.IssueDate == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Issue date is required",
		}
	}

	if req.DueDate == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Due date is required",
		}
	}

	return nil
}

func (h *UpdateInvoiceHandler) Serve(ctx context.Context, req *UpdateInvoiceRequest) (UpdateInvoiceResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

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

	inv, err := h.FinanceService.UpdateInvoice(ctx, device.User.UserID, req.ID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return UpdateInvoiceResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Invoice not found",
			}
		}

		return UpdateInvoiceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update invoice",
		}
	}

	return UpdateInvoiceResponse{Invoice: newInvoiceDetailPayload(inv)}, nil
}
