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
)

func UpdateExpenseRoutes() []string {
	return []string{"/finance/expenses"}
}

type UpdateExpenseHandler struct {
	FinanceService *application.Service
}

type UpdateExpenseRequest struct {
	ID              int64   `json:"id"`
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

type UpdateExpenseResponse struct {
	Expense ExpensePayload `json:"expense"`
}

func (h *UpdateExpenseHandler) Validate(_ context.Context, req *UpdateExpenseRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Expense ID is required")
	}

	if strings.TrimSpace(req.Description) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Description is required")
	}

	if req.Amount <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Amount must be greater than zero")
	}

	if strings.TrimSpace(req.ExpenseDate) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Expense date is required")
	}

	if _, err := time.Parse("2006-01-02", req.ExpenseDate); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "Expense date must be in YYYY-MM-DD format")
	}

	return nil
}

func (h *UpdateExpenseHandler) Serve(ctx context.Context, req *UpdateExpenseRequest) (UpdateExpenseResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return UpdateExpenseResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:expenses:create"); err != nil {
		return UpdateExpenseResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateExpenseResponse{}, onekiterrors.PermissionDenied()
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

	e, err := h.FinanceService.UpdateExpense(ctx, businessID, req.ID, input)
	if err != nil {
		if errors.Is(err, domain.ErrExpenseNotFound) {
			return UpdateExpenseResponse{}, onekiterrors.New(http.StatusNotFound, "Expense not found")
		}

		return UpdateExpenseResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update expense")
	}

	return UpdateExpenseResponse{Expense: newExpensePayload(*e)}, nil
}
