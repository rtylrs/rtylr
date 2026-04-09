package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func CreateInvoiceRoutes() []string {
	return []string{"/finance/invoices"}
}

type CreateInvoiceHandler struct {
	FinanceService *application.Service
}

type InvoiceItemPayload struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unitPrice"`
	TaxRate     float64 `json:"taxRate"`
	DiscountPct float64 `json:"discountPct"`
	SortOrder   int     `json:"sortOrder"`
	Total       float64 `json:"total"`
}

type CreateInvoiceRequest struct {
	ClientName    string               `json:"clientName"`
	ClientEmail   string               `json:"clientEmail"`
	ClientAddress string               `json:"clientAddress"`
	ClientTaxID   string               `json:"clientTaxId"`
	IssueDate     string               `json:"issueDate"`
	DueDate       string               `json:"dueDate"`
	Currency      string               `json:"currency"`
	DiscountType  string               `json:"discountType"`
	DiscountValue float64              `json:"discountValue"`
	TaxRate       float64              `json:"taxRate"`
	Notes         string               `json:"notes"`
	Terms         string               `json:"terms"`
	Items         []InvoiceItemPayload `json:"items"`
}

type InvoiceDetailPayload struct {
	ID             int64                `json:"id"`
	OwnerID        int64                `json:"ownerId"`
	InvoiceNumber  string               `json:"invoiceNumber"`
	ClientName     string               `json:"clientName"`
	ClientEmail    string               `json:"clientEmail"`
	ClientAddress  string               `json:"clientAddress"`
	ClientTaxID    string               `json:"clientTaxId"`
	Status         string               `json:"status"`
	IssueDate      time.Time            `json:"issueDate"`
	DueDate        time.Time            `json:"dueDate"`
	PaidDate       *time.Time           `json:"paidDate"`
	Currency       string               `json:"currency"`
	Subtotal       float64              `json:"subtotal"`
	DiscountType   string               `json:"discountType"`
	DiscountValue  float64              `json:"discountValue"`
	DiscountAmount float64              `json:"discountAmount"`
	TaxRate        float64              `json:"taxRate"`
	TaxAmount      float64              `json:"taxAmount"`
	Total          float64              `json:"total"`
	AmountPaid     float64              `json:"amountPaid"`
	AmountDue      float64              `json:"amountDue"`
	Notes          string               `json:"notes"`
	Terms          string               `json:"terms"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
	Items          []InvoiceItemPayload `json:"items"`
}

type CreateInvoiceResponse struct {
	Invoice InvoiceDetailPayload `json:"invoice"`
}

func newInvoiceDetailPayload(inv *domain.Invoice) InvoiceDetailPayload {
	items := make([]InvoiceItemPayload, 0, len(inv.Items))
	for _, item := range inv.Items {
		items = append(items, InvoiceItemPayload{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			TaxRate:     item.TaxRate,
			DiscountPct: item.DiscountPct,
			SortOrder:   item.SortOrder,
			Total:       item.Total,
		})
	}

	return InvoiceDetailPayload{
		ID:             inv.ID,
		OwnerID:        inv.OwnerID,
		InvoiceNumber:  inv.InvoiceNumber,
		ClientName:     inv.ClientName,
		ClientEmail:    inv.ClientEmail,
		ClientAddress:  inv.ClientAddress,
		ClientTaxID:    inv.ClientTaxID,
		Status:         inv.Status,
		IssueDate:      inv.IssueDate,
		DueDate:        inv.DueDate,
		PaidDate:       inv.PaidDate,
		Currency:       inv.Currency,
		Subtotal:       inv.Subtotal,
		DiscountType:   inv.DiscountType,
		DiscountValue:  inv.DiscountValue,
		DiscountAmount: inv.DiscountAmount,
		TaxRate:        inv.TaxRate,
		TaxAmount:      inv.TaxAmount,
		Total:          inv.Total,
		AmountPaid:     inv.AmountPaid,
		AmountDue:      inv.AmountDue,
		Notes:          inv.Notes,
		Terms:          inv.Terms,
		CreatedAt:      inv.CreatedAt,
		UpdatedAt:      inv.UpdatedAt,
		Items:          items,
	}
}

func (h *CreateInvoiceHandler) Validate(_ context.Context, req *CreateInvoiceRequest) *errorModels.ServerError {
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

	if len(req.Items) == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "At least one item is required",
		}
	}

	return nil
}

func (h *CreateInvoiceHandler) Serve(ctx context.Context, req *CreateInvoiceRequest) (CreateInvoiceResponse, *errorModels.ServerError) {
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

	input := domain.InvoiceInput{
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

	inv, err := h.FinanceService.CreateInvoice(ctx, device.User.UserID, input)
	if err != nil {
		return CreateInvoiceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create invoice",
		}
	}

	return CreateInvoiceResponse{Invoice: newInvoiceDetailPayload(inv)}, nil
}
