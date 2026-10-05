package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

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
