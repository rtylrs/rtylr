package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func CreateExpenseRoutes() []string {
	return []string{"/finance/expenses"}
}

type CreateExpenseHandler struct {
	FinanceService *application.Service
}

type CreateExpenseRequest struct {
	CategoryID      int64   `json:"categoryId"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	ExpenseDate     string  `json:"expenseDate"`
	Vendor          string  `json:"vendor"`
	PaymentMethod   string  `json:"paymentMethod"`
	ReceiptURL      string  `json:"receiptUrl"`
	IsTaxDeductible bool    `json:"isTaxDeductible"`
	TaxAmount       float64 `json:"taxAmount"`
	Notes           string  `json:"notes"`
}

type CreateExpenseResponse struct {
	Expense ExpensePayload `json:"expense"`
}

func (h *CreateExpenseHandler) Validate(_ context.Context, req *CreateExpenseRequest) *errorModels.ServerError {
	if req.Description == "" {
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

	if req.ExpenseDate == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Expense date is required",
		}
	}

	return nil
}

func (h *CreateExpenseHandler) Serve(ctx context.Context, req *CreateExpenseRequest) (CreateExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

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

	e, err := h.FinanceService.CreateExpense(ctx, device.User.UserID, input)
	if err != nil {
		return CreateExpenseResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create expense",
		}
	}

	return CreateExpenseResponse{Expense: newExpensePayload(*e)}, nil
}
