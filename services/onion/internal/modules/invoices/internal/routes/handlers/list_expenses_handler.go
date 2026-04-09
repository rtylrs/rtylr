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

func ListExpensesRoutes() []string {
	return []string{"/finance/expenses"}
}

type ListExpensesHandler struct {
	FinanceService *application.Service
}

type ListExpensesRequest struct {
	CategoryID int64  `json:"categoryId"`
	Search     string `json:"search"`
	DateFrom   string `json:"dateFrom"`
	DateTo     string `json:"dateTo"`
}

type ExpensePayload struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"ownerId"`
	CategoryID      int64     `json:"categoryId"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ExpenseDate     time.Time `json:"expenseDate"`
	Vendor          string    `json:"vendor"`
	PaymentMethod   string    `json:"paymentMethod"`
	ReceiptURL      string    `json:"receiptUrl"`
	IsTaxDeductible bool      `json:"isTaxDeductible"`
	TaxAmount       float64   `json:"taxAmount"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ListExpensesResponse struct {
	Expenses []ExpensePayload `json:"expenses"`
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
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func (h *ListExpensesHandler) Serve(ctx context.Context, req *ListExpensesRequest) (ListExpensesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	expenses, err := h.FinanceService.ListExpenses(ctx, device.User.UserID, req.CategoryID, req.Search, req.DateFrom, req.DateTo)
	if err != nil {
		return ListExpensesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list expenses",
		}
	}

	payloads := make([]ExpensePayload, 0, len(expenses))
	for _, e := range expenses {
		payloads = append(payloads, newExpensePayload(e))
	}

	return ListExpensesResponse{Expenses: payloads}, nil
}
