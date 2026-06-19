package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/expenses/internal/application"
	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

func newExpensePayload(e domain.Expense) *financeexpensesv1.Expense {
	return &financeexpensesv1.Expense{
		Id:          e.ID,
		OwnerId:     e.OwnerID,
		CategoryId:  e.CategoryID,
		Amount:      e.Amount,
		Currency:    e.Currency,
		Description: e.Description,
		Vendor:      e.Vendor,
		ReceiptUrl:  e.ReceiptURL,
		ExpenseDate: e.ExpenseDate.Format("2006-01-02"),
		Status:      e.Status,
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   e.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func ExpenseListRoutes() []string { return []string{"/finance/expenses"} }

type (
	ListExpensesHandler  struct{ Service *application.Service }
	ListExpensesRequest  struct{}
	ListExpensesResponse struct {
		Expenses []*financeexpensesv1.Expense `json:"expenses"`
	}
)

func (h *ListExpensesHandler) Serve(ctx context.Context, _ *ListExpensesRequest) (ListExpensesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListExpensesResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListExpensesResponse{}, errorModels.PermissionDeniedServerError()
	}
	expenses, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListExpensesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list expenses"}
	}
	payloads := make([]*financeexpensesv1.Expense, len(expenses))
	for i, e := range expenses {
		payloads[i] = newExpensePayload(e)
	}
	return ListExpensesResponse{Expenses: payloads}, nil
}

// ─── Get by ID ────────────────────────────────────────────────────────────────

func ExpenseGetRoutes() []string { return []string{"/finance/expenses/{id}"} }

type (
	GetExpenseHandler  struct{ Service *application.Service }
	GetExpenseRequest  struct{}
	GetExpenseResponse struct {
		Expense *financeexpensesv1.Expense `json:"expense"`
	}
)

func (h *GetExpenseHandler) Serve(ctx context.Context, _ *GetExpenseRequest) (GetExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return GetExpenseResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
	}
	expense, getErr := h.Service.GetByID(ctx, businessID, id)
	if getErr != nil {
		if errors.Is(getErr, domain.ErrExpenseNotFound) {
			return GetExpenseResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Expense not found"}
		}
		return GetExpenseResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get expense"}
	}
	return GetExpenseResponse{Expense: newExpensePayload(expense)}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func ExpenseCreateRoutes() []string { return []string{"/finance/expenses"} }

type CreateExpenseHandler struct{ Service *application.Service }

type CreateExpenseRequest struct {
	CategoryID  *int64  `json:"category_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Vendor      string  `json:"vendor"`
	ReceiptURL  string  `json:"receipt_url"`
	ExpenseDate string  `json:"expense_date"`
}

type CreateExpenseResponse struct {
	ExpenseID int64  `json:"expense_id"`
	Message   string `json:"message"`
}

func (h *CreateExpenseHandler) Validate(_ context.Context, req *CreateExpenseRequest) *errorModels.ServerError {
	if req.Amount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Amount must be greater than zero"}
	}
	if req.ExpenseDate != "" {
		if _, err := time.Parse("2006-01-02", req.ExpenseDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "expense_date must be in YYYY-MM-DD format"}
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
	expDate := time.Now()
	if req.ExpenseDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpenseDate); err == nil {
			expDate = t
		}
	}
	id, err := h.Service.Create(ctx, domain.ExpenseInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
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
	CategoryID  *int64  `json:"category_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Vendor      string  `json:"vendor"`
	ReceiptURL  string  `json:"receipt_url"`
	ExpenseDate string  `json:"expense_date"`
}

type UpdateExpenseResponse struct {
	Message string `json:"message"`
}

func (h *UpdateExpenseHandler) Validate(_ context.Context, req *UpdateExpenseRequest) *errorModels.ServerError {
	if req.Amount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Amount must be greater than zero"}
	}
	if req.ExpenseDate != "" {
		if _, err := time.Parse("2006-01-02", req.ExpenseDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "expense_date must be in YYYY-MM-DD format"}
		}
	}
	return nil
}

func (h *UpdateExpenseHandler) Serve(ctx context.Context, req *UpdateExpenseRequest) (UpdateExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
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
		BusinessID:  businessID,
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

type (
	DeleteExpenseHandler  struct{ Service *application.Service }
	DeleteExpenseRequest  struct{}
	DeleteExpenseResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteExpenseHandler) Serve(ctx context.Context, _ *DeleteExpenseRequest) (DeleteExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteExpenseResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
	}
	if deleteErr := h.Service.Delete(ctx, businessID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrExpenseNotFound) {
			return DeleteExpenseResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Expense not found"}
		}
		return DeleteExpenseResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete expense"}
	}
	return DeleteExpenseResponse{Message: "Expense deleted successfully"}, nil
}

// ─── Approve / Reject ─────────────────────────────────────────────────────────

func ExpenseApproveRoutes() []string { return []string{"/finance/expenses/{id}/approve"} }
func ExpenseRejectRoutes() []string  { return []string{"/finance/expenses/{id}/reject"} }

type (
	ApproveExpenseHandler  struct{ Service *application.Service }
	ApproveExpenseRequest  struct{}
	ApproveExpenseResponse struct {
		Message string `json:"message"`
	}

	RejectExpenseHandler  struct{ Service *application.Service }
	RejectExpenseRequest  struct{}
	RejectExpenseResponse struct {
		Message string `json:"message"`
	}
)

func (h *ApproveExpenseHandler) Serve(ctx context.Context, _ *ApproveExpenseRequest) (ApproveExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ApproveExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ApproveExpenseResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
	}
	if updateErr := h.Service.UpdateStatus(ctx, businessID, id, "approved"); updateErr != nil {
		if errors.Is(updateErr, domain.ErrExpenseNotFound) {
			return ApproveExpenseResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Expense not found"}
		}
		return ApproveExpenseResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to approve expense"}
	}
	return ApproveExpenseResponse{Message: "Expense approved"}, nil
}

func (h *RejectExpenseHandler) Serve(ctx context.Context, _ *RejectExpenseRequest) (RejectExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RejectExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RejectExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return RejectExpenseResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
	}
	if updateErr := h.Service.UpdateStatus(ctx, businessID, id, "rejected"); updateErr != nil {
		if errors.Is(updateErr, domain.ErrExpenseNotFound) {
			return RejectExpenseResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Expense not found"}
		}
		return RejectExpenseResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to reject expense"}
	}
	return RejectExpenseResponse{Message: "Expense rejected"}, nil
}

// ─── Categories ───────────────────────────────────────────────────────────────

func ExpenseCategoriesRoutes() []string { return []string{"/finance/expenses/categories"} }

type (
	ListCategoriesHandler  struct{ Service *application.Service }
	ListCategoriesRequest  struct{}
	ListCategoriesResponse struct {
		Categories []*financeexpensesv1.ExpenseCategory `json:"categories"`
	}
)

func (h *ListCategoriesHandler) Serve(ctx context.Context, _ *ListCategoriesRequest) (ListCategoriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCategoriesResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListCategoriesResponse{}, errorModels.PermissionDeniedServerError()
	}
	cats, err := h.Service.ListCategories(ctx, businessID)
	if err != nil {
		return ListCategoriesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list expense categories"}
	}
	payloads := make([]*financeexpensesv1.ExpenseCategory, len(cats))
	for i, c := range cats {
		payloads[i] = &financeexpensesv1.ExpenseCategory{
			Id:        c.ID,
			OwnerId:   c.OwnerID,
			Name:      c.Name,
			Color:     c.Color,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return ListCategoriesResponse{Categories: payloads}, nil
}
