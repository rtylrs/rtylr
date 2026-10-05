package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

type DeleteExpenseHandler struct {
	FinanceService *application.Service
}

type DeleteExpenseRequest struct {
	ID int64 `json:"id"`
}

type DeleteExpenseResponse struct {
	Message string `json:"message"`
}
