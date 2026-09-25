package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListExpensesRoutes() []string {
	return []string{"/finance/expenses"}
}

type ListExpensesHandler struct {
	FinanceService *application.Service
}

type ListExpensesRequest struct {
	CategoryID int64  `json:"category_id"`
	Search     string `json:"search"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
	Page       int64  `json:"page"`
	PageSize   int64  `json:"page_size"`
}

type ExpensePayload struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	CategoryID      int64     `json:"category_id"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ExpenseDate     time.Time `json:"expense_date"`
	Vendor          string    `json:"vendor"`
	PaymentMethod   string    `json:"payment_method"`
	ReceiptURL      string    `json:"receipt_url"`
	IsTaxDeductible bool      `json:"is_tax_deductible"`
	TaxAmount       float64   `json:"tax_amount"`
	Notes           string    `json:"notes"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ListExpensesResponse struct {
	Expenses []ExpensePayload `json:"expenses"`
	Page     int64            `json:"page"`
	PageSize int64            `json:"page_size"`
	HasMore  bool             `json:"has_more"`
}

func newExpensePayload(e domain.Expense) ExpensePayload {
	return ExpensePayload{
		ID:              e.ID,
		OwnerID:         e.OwnerID,
		CategoryID:      e.CategoryID,
		Description:     e.Description,
		Amount:          e.Amount,
		Currency:        e.Currency,
		ExpenseDate:     e.ExpenseDate,
		Vendor:          e.Vendor,
		PaymentMethod:   e.PaymentMethod,
		ReceiptURL:      e.ReceiptURL,
		IsTaxDeductible: e.IsTaxDeductible,
		TaxAmount:       e.TaxAmount,
		Notes:           e.Notes,
		Status:          e.Status,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func (h *ListExpensesHandler) Serve(ctx context.Context, req *ListExpensesRequest) (ListExpensesResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return ListExpensesResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:expenses:view"); err != nil {
		return ListExpensesResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListExpensesResponse{}, onekiterrors.PermissionDenied()
	}

	expenses, hasMore, err := h.FinanceService.ListExpenses(ctx, businessID, req.CategoryID, req.Search, req.DateFrom, req.DateTo, req.Page, req.PageSize)
	if err != nil {
		return ListExpensesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list expenses")
	}

	payloads := make([]ExpensePayload, 0, len(expenses))
	for _, e := range expenses {
		payloads = append(payloads, newExpensePayload(e))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListExpensesResponse{Expenses: payloads, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}
