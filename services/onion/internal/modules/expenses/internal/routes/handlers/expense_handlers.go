package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/expenses/internal/application"
	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
)

type ExpensePayload struct {
	ID          int64   `json:"id"`
	OwnerID     int64   `json:"ownerId"`
	CategoryID  *int64  `json:"categoryId"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Vendor      string  `json:"vendor"`
	ReceiptURL  string  `json:"receiptUrl"`
	ExpenseDate string  `json:"expenseDate"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

func newExpensePayload(e domain.Expense) ExpensePayload {
	return ExpensePayload{
		ID:          e.ID,
		OwnerID:     e.OwnerID,
		CategoryID:  e.CategoryID,
		Amount:      e.Amount,
		Currency:    e.Currency,
		Description: e.Description,
		Vendor:      e.Vendor,
		ReceiptURL:  e.ReceiptURL,
		ExpenseDate: e.ExpenseDate.Format("2006-01-02"),
		Status:      e.Status,
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   e.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func ExpenseListRoutes() []string { return []string{"/finance/expenses"} }

type ListExpensesHandler struct{ Service *application.Service }
type ListExpensesRequest struct{}
type ListExpensesResponse struct {
	Expenses []ExpensePayload `json:"expenses"`
}

func (h *ListExpensesHandler) Serve(ctx context.Context, _ *ListExpensesRequest) (ListExpensesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListExpensesResponse{}, errorModels.PermissionDeniedServerError()
	}
	expenses, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListExpensesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list expenses"}
	}
	payloads := make([]ExpensePayload, len(expenses))
	for i, e := range expenses {
		payloads[i] = newExpensePayload(e)
	}
	return ListExpensesResponse{Expenses: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func ExpenseCreateRoutes() []string { return []string{"/finance/expenses"} }

type CreateExpenseHandler struct{ Service *application.Service }

type CreateExpenseRequest struct {
	CategoryID  *int64  `json:"categoryId"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Vendor      string  `json:"vendor"`
	ReceiptURL  string  `json:"receiptUrl"`
	ExpenseDate string  `json:"expenseDate"`
}

type CreateExpenseResponse struct {
	ExpenseID int64  `json:"expenseId"`
	Message   string `json:"message"`
}

func (h *CreateExpenseHandler) Validate(_ context.Context, req *CreateExpenseRequest) *errorModels.ServerError {
	if req.Amount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Amount must be greater than zero"}
	}
	return nil
}

func (h *CreateExpenseHandler) Serve(ctx context.Context, req *CreateExpenseRequest) (CreateExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	expDate := time.Now()
	if req.ExpenseDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpenseDate); err == nil {
			expDate = t
		}
	}
	id, err := h.Service.Create(ctx, domain.ExpenseInput{
		OwnerID:     device.User.UserID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Vendor:      req.Vendor,
		ReceiptURL:  req.ReceiptURL,
		ExpenseDate: expDate,
	})
	if err != nil {
		return CreateExpenseResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create expense"}
	}
	return CreateExpenseResponse{ExpenseID: id, Message: "Expense created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func ExpenseUpdateRoutes() []string { return []string{"/finance/expenses/{id}"} }

type UpdateExpenseHandler struct{ Service *application.Service }

type UpdateExpenseRequest struct {
	CategoryID  *int64  `json:"categoryId"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Vendor      string  `json:"vendor"`
	ReceiptURL  string  `json:"receiptUrl"`
	ExpenseDate string  `json:"expenseDate"`
}

type UpdateExpenseResponse struct {
	Message string `json:"message"`
}

func (h *UpdateExpenseHandler) Validate(_ context.Context, req *UpdateExpenseRequest) *errorModels.ServerError {
	if req.Amount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Amount must be greater than zero"}
	}
	return nil
}

func (h *UpdateExpenseHandler) Serve(ctx context.Context, req *UpdateExpenseRequest) (UpdateExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateExpenseResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
	}
	expDate := time.Now()
	if req.ExpenseDate != "" {
		if t, parseErr := time.Parse("2006-01-02", req.ExpenseDate); parseErr == nil {
			expDate = t
		}
	}
	if updateErr := h.Service.Update(ctx, domain.ExpenseUpdateInput{
		ID:          id,
		OwnerID:     device.User.UserID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Vendor:      req.Vendor,
		ReceiptURL:  req.ReceiptURL,
		ExpenseDate: expDate,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrExpenseNotFound) {
			return UpdateExpenseResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Expense not found"}
		}
		return UpdateExpenseResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update expense"}
	}
	return UpdateExpenseResponse{Message: "Expense updated successfully"}, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func ExpenseDeleteRoutes() []string { return []string{"/finance/expenses/{id}"} }

type DeleteExpenseHandler struct{ Service *application.Service }
type DeleteExpenseRequest struct{}
type DeleteExpenseResponse struct {
	Message string `json:"message"`
}

func (h *DeleteExpenseHandler) Serve(ctx context.Context, _ *DeleteExpenseRequest) (DeleteExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteExpenseResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
	}
	if deleteErr := h.Service.Delete(ctx, device.User.UserID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrExpenseNotFound) {
			return DeleteExpenseResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Expense not found"}
		}
		return DeleteExpenseResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete expense"}
	}
	return DeleteExpenseResponse{Message: "Expense deleted successfully"}, nil
}

// ─── Categories ───────────────────────────────────────────────────────────────

func ExpenseCategoriesRoutes() []string { return []string{"/finance/expenses/categories"} }

type ListCategoriesHandler struct{ Service *application.Service }
type ListCategoriesRequest struct{}
type ListCategoriesResponse struct {
	Categories []CategoryPayload `json:"categories"`
}

type CategoryPayload struct {
	ID        int64  `json:"id"`
	OwnerID   int64  `json:"ownerId"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"createdAt"`
}

func (h *ListCategoriesHandler) Serve(ctx context.Context, _ *ListCategoriesRequest) (ListCategoriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCategoriesResponse{}, errorModels.PermissionDeniedServerError()
	}
	cats, err := h.Service.ListCategories(ctx, device.User.UserID)
	if err != nil {
		return ListCategoriesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list expense categories"}
	}
	payloads := make([]CategoryPayload, len(cats))
	for i, c := range cats {
		payloads[i] = CategoryPayload{
			ID:        c.ID,
			OwnerID:   c.OwnerID,
			Name:      c.Name,
			Color:     c.Color,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return ListCategoriesResponse{Categories: payloads}, nil
}
