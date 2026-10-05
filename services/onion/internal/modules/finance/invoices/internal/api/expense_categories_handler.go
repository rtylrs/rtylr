package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

// ---------------------------------------------------------------------------
// List Expense Categories
// ---------------------------------------------------------------------------

type ListExpenseCategoriesHandler struct {
	FinanceService *application.Service
}

type ListExpenseCategoriesRequest struct{}

type ExpenseCategoryPayload struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
}

type ListExpenseCategoriesResponse struct {
	Categories []ExpenseCategoryPayload `json:"categories"`
}

// ---------------------------------------------------------------------------
// Create Expense Category
// ---------------------------------------------------------------------------

type CreateExpenseCategoryHandler struct {
	FinanceService *application.Service
}

type CreateExpenseCategoryRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type CreateExpenseCategoryResponse struct {
	Category ExpenseCategoryPayload `json:"category"`
}
