package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

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
