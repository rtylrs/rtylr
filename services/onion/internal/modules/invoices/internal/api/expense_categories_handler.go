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

// ---------------------------------------------------------------------------
// List Expense Categories
// ---------------------------------------------------------------------------

func ListExpenseCategoriesRoutes() []string {
	return []string{"/finance/expense-categories"}
}

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

func newExpenseCategoryPayload(c domain.ExpenseCategory) ExpenseCategoryPayload {
	return ExpenseCategoryPayload{
		ID:        c.ID,
		OwnerID:   c.OwnerID,
		Name:      c.Name,
		Color:     c.Color,
		Icon:      c.Icon,
		CreatedAt: c.CreatedAt,
	}
}

func (h *ListExpenseCategoriesHandler) Serve(ctx context.Context, _ *ListExpenseCategoriesRequest) (ListExpenseCategoriesResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return ListExpenseCategoriesResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:expenses:view"); err != nil {
		return ListExpenseCategoriesResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListExpenseCategoriesResponse{}, onekiterrors.PermissionDenied()
	}

	categories, err := h.FinanceService.ListExpenseCategories(ctx, businessID)
	if err != nil {
		return ListExpenseCategoriesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list expense categories")
	}

	payloads := make([]ExpenseCategoryPayload, 0, len(categories))
	for _, c := range categories {
		payloads = append(payloads, newExpenseCategoryPayload(c))
	}

	return ListExpenseCategoriesResponse{Categories: payloads}, nil
}

// ---------------------------------------------------------------------------
// Create Expense Category
// ---------------------------------------------------------------------------

func CreateExpenseCategoryRoutes() []string {
	return []string{"/finance/expense-categories"}
}

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

func (h *CreateExpenseCategoryHandler) Validate(_ context.Context, req *CreateExpenseCategoryRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Category name is required")
	}

	return nil
}

func (h *CreateExpenseCategoryHandler) Serve(ctx context.Context, req *CreateExpenseCategoryRequest) (CreateExpenseCategoryResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return CreateExpenseCategoryResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:expenses:create"); err != nil {
		return CreateExpenseCategoryResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateExpenseCategoryResponse{}, onekiterrors.PermissionDenied()
	}

	input := domain.ExpenseCategoryInput{
		Name:  req.Name,
		Color: req.Color,
		Icon:  req.Icon,
	}

	c, err := h.FinanceService.CreateExpenseCategory(ctx, businessID, input)
	if err != nil {
		return CreateExpenseCategoryResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create expense category")
	}

	return CreateExpenseCategoryResponse{Category: newExpenseCategoryPayload(*c)}, nil
}
