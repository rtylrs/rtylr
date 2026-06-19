package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CreateExpenseRoutes() []string {
	return []string{"/finance/expenses"}
}

type CreateExpenseHandler struct {
	FinanceService *application.Service
}

type CreateExpenseRequest struct {
	CategoryID      int64   `json:"category_id"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	ExpenseDate     string  `json:"expense_date"`
	Vendor          string  `json:"vendor"`
	PaymentMethod   string  `json:"payment_method"`
	ReceiptURL      string  `json:"receipt_url"`
	IsTaxDeductible bool    `json:"is_tax_deductible"`
	TaxAmount       float64 `json:"tax_amount"`
	Notes           string  `json:"notes"`
}

type CreateExpenseResponse struct {
	Expense ExpensePayload `json:"expense"`
}

func (h *CreateExpenseHandler) Validate(_ context.Context, req *CreateExpenseRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Description) == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Description is required",
		}
	}

	if req.Amount <= 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Amount must be greater than zero",
		}
	}

	if strings.TrimSpace(req.ExpenseDate) == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Expense date is required",
		}
	}

	if _, err := time.Parse("2006-01-02", req.ExpenseDate); err != nil {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Expense date must be in YYYY-MM-DD format",
		}
	}

	return nil
}

func (h *CreateExpenseHandler) Serve(ctx context.Context, req *CreateExpenseRequest) (CreateExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return CreateExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.ExpenseInput{
		CategoryID:      req.CategoryID,
		Description:     req.Description,
		Amount:          req.Amount,
		Currency:        req.Currency,
		ExpenseDate:     req.ExpenseDate,
		Vendor:          req.Vendor,
		PaymentMethod:   req.PaymentMethod,
		ReceiptURL:      req.ReceiptURL,
		IsTaxDeductible: req.IsTaxDeductible,
		TaxAmount:       req.TaxAmount,
		Notes:           req.Notes,
	}

	e, err := h.FinanceService.CreateExpense(ctx, businessID, input)
	if err != nil {
		return CreateExpenseResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create expense",
		}
	}

	return CreateExpenseResponse{Expense: newExpensePayload(*e)}, nil
}
