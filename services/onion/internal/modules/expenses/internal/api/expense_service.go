package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/expenses/internal/application"
	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

// listExpensesPageSize bounds ListExpenses server-side. The voxie
// finance.expenses.v1.ListExpensesRequest contract has no pagination fields
// yet, so the endpoint returns at most one maximum-size page instead of every
// row.
const listExpensesPageSize int32 = 500

type ExpensesAPI struct {
	Service *application.Service
}

func NewExpensesAPI(service *application.Service) *ExpensesAPI {
	return &ExpensesAPI{Service: service}
}

func (a *ExpensesAPI) ListExpenses(ctx context.Context, _ *financeexpensesv1.ListExpensesRequest) (*financeexpensesv1.ListExpensesResponse, error) {
	businessID, err := expenseBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	expenses, listErr := a.Service.ListPaginated(ctx, businessID, listExpensesPageSize, 0)
	if listErr != nil {
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to list expenses")
	}
	payloads := make([]*financeexpensesv1.Expense, 0, len(expenses))
	for _, expense := range expenses {
		payloads = append(payloads, newExpensePayload(expense))
	}
	return &financeexpensesv1.ListExpensesResponse{Expenses: payloads}, nil
}

func (a *ExpensesAPI) CreateExpense(ctx context.Context, req *financeexpensesv1.CreateExpenseRequest) (*financeexpensesv1.CreateExpenseResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetAmount() <= 0 {
		return nil, expenseStatusError(http.StatusBadRequest, "Amount must be greater than zero")
	}
	expenseDate, err := parseExpenseDate(req.GetExpenseDate())
	if err != nil {
		return nil, err
	}
	if validationErr := domain.ValidatExpenseInput(req.GetAmount(), req.GetCurrency(), req.GetDescription(), expenseDate); validationErr != nil {
		return nil, expenseStatusError(http.StatusBadRequest, validationErr.Error())
	}
	id, createErr := a.Service.Create(ctx, domain.ExpenseInput{
		OwnerID:         device.User.UserID,
		BusinessID:      businessID,
		CategoryID:      req.CategoryId,
		Amount:          req.GetAmount(),
		Currency:        req.GetCurrency(),
		Description:     req.GetDescription(),
		Vendor:          req.GetVendor(),
		PaymentMethod:   req.GetPaymentMethod(),
		ReceiptURL:      req.GetReceiptUrl(),
		IsTaxDeductible: req.GetIsTaxDeductible(),
		TaxAmount:       req.GetTaxAmount(),
		Notes:           req.GetNotes(),
		ExpenseDate:     expenseDate,
	})
	if createErr != nil {
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to create expense")
	}
	return &financeexpensesv1.CreateExpenseResponse{ExpenseId: id, Message: "Expense created successfully"}, nil
}

func (a *ExpensesAPI) UpdateExpense(ctx context.Context, req *financeexpensesv1.UpdateExpenseRequest) (*financeexpensesv1.UpdateExpenseResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	body := req.GetBody()
	if body == nil {
		return nil, expenseStatusError(http.StatusBadRequest, "Expense body is required")
	}
	if body.GetAmount() <= 0 {
		return nil, expenseStatusError(http.StatusBadRequest, "Amount must be greater than zero")
	}
	expenseDate, err := parseExpenseDate(body.GetExpenseDate())
	if err != nil {
		return nil, err
	}
	if validationErr := domain.ValidatExpenseInput(body.GetAmount(), body.GetCurrency(), body.GetDescription(), expenseDate); validationErr != nil {
		return nil, expenseStatusError(http.StatusBadRequest, validationErr.Error())
	}
	if updateErr := a.Service.Update(ctx, domain.ExpenseUpdateInput{
		ID:              req.GetId(),
		OwnerID:         device.User.UserID,
		BusinessID:      businessID,
		CategoryID:      body.CategoryId,
		Amount:          body.GetAmount(),
		Currency:        body.GetCurrency(),
		Description:     body.GetDescription(),
		Vendor:          body.GetVendor(),
		PaymentMethod:   body.GetPaymentMethod(),
		ReceiptURL:      body.GetReceiptUrl(),
		IsTaxDeductible: body.GetIsTaxDeductible(),
		TaxAmount:       body.GetTaxAmount(),
		Notes:           body.GetNotes(),
		ExpenseDate:     expenseDate,
	}); updateErr != nil {
		switch {
		case errors.Is(updateErr, domain.ErrExpenseNotFound):
			return nil, expenseStatusError(http.StatusNotFound, "Expense not found")
		case errors.Is(updateErr, domain.ErrApprovedExpenseImmutable):
			return nil, expenseStatusError(http.StatusConflict, "Approved expenses cannot be edited")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to update expense")
	}
	return &financeexpensesv1.UpdateExpenseResponse{Message: "Expense updated successfully"}, nil
}

func (a *ExpensesAPI) DeleteExpense(ctx context.Context, req *financeexpensesv1.DeleteExpenseRequest) (*financeexpensesv1.DeleteExpenseResponse, error) {
	businessID, err := expenseBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.Delete(ctx, businessID, req.GetId()); deleteErr != nil {
		switch {
		case errors.Is(deleteErr, domain.ErrExpenseNotFound):
			return nil, expenseStatusError(http.StatusNotFound, "Expense not found")
		case errors.Is(deleteErr, domain.ErrApprovedExpenseImmutable):
			return nil, expenseStatusError(http.StatusConflict, "Approved expenses cannot be deleted")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to delete expense")
	}
	return &financeexpensesv1.DeleteExpenseResponse{Message: "Expense deleted successfully"}, nil
}

func (a *ExpensesAPI) GetExpense(ctx context.Context, req *financeexpensesv1.GetExpenseRequest) (*financeexpensesv1.GetExpenseResponse, error) {
	businessID, err := expenseBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	expense, getErr := a.Service.GetByID(ctx, businessID, req.GetId())
	if getErr != nil {
		if errors.Is(getErr, domain.ErrExpenseNotFound) {
			return nil, expenseStatusError(http.StatusNotFound, "Expense not found")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to get expense")
	}
	return &financeexpensesv1.GetExpenseResponse{Expense: newExpensePayload(expense)}, nil
}

func (a *ExpensesAPI) UpdateExpenseStatus(ctx context.Context, req *financeexpensesv1.UpdateExpenseStatusRequest) (*financeexpensesv1.UpdateExpenseStatusResponse, error) {
	businessID, err := expenseBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil || body.GetStatus() == "" {
		return nil, expenseStatusError(http.StatusBadRequest, "Status is required")
	}
	if _, valid := domain.ValidExpenseStatuses[body.GetStatus()]; !valid {
		return nil, expenseStatusError(http.StatusBadRequest, "Invalid expense status")
	}
	if updateErr := a.Service.UpdateStatus(ctx, businessID, req.GetId(), body.GetStatus()); updateErr != nil {
		switch {
		case errors.Is(updateErr, domain.ErrExpenseNotFound):
			return nil, expenseStatusError(http.StatusNotFound, "Expense not found")
		case errors.Is(updateErr, domain.ErrInvalidStatusTransition):
			return nil, expenseStatusError(http.StatusConflict, "Expense review status transitions are one-way")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to update expense status")
	}
	return &financeexpensesv1.UpdateExpenseStatusResponse{Message: "Expense status updated successfully"}, nil
}

func (a *ExpensesAPI) ListExpenseCategories(ctx context.Context, _ *financeexpensesv1.ListExpenseCategoriesRequest) (*financeexpensesv1.ListExpenseCategoriesResponse, error) {
	businessID, err := expenseBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	categories, listErr := a.Service.ListCategories(ctx, businessID)
	if listErr != nil {
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to list expense categories")
	}
	payloads := make([]*financeexpensesv1.ExpenseCategory, 0, len(categories))
	for _, category := range categories {
		payloads = append(payloads, &financeexpensesv1.ExpenseCategory{
			Id:        category.ID,
			OwnerId:   category.OwnerID,
			Name:      category.Name,
			Color:     category.Color,
			CreatedAt: category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return &financeexpensesv1.ListExpenseCategoriesResponse{Categories: payloads}, nil
}

func newExpensePayload(expense domain.Expense) *financeexpensesv1.Expense {
	return &financeexpensesv1.Expense{
		Id:              expense.ID,
		OwnerId:         expense.OwnerID,
		CategoryId:      expense.CategoryID,
		Amount:          expense.Amount,
		Currency:        expense.Currency,
		Description:     expense.Description,
		Vendor:          expense.Vendor,
		PaymentMethod:   expense.PaymentMethod,
		ReceiptUrl:      expense.ReceiptURL,
		IsTaxDeductible: expense.IsTaxDeductible,
		TaxAmount:       expense.TaxAmount,
		Notes:           expense.Notes,
		ExpenseDate:     expense.ExpenseDate.Format("2006-01-02"),
		Status:          expense.Status,
		CreatedAt:       expense.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       expense.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func parseExpenseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Now(), nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, expenseStatusError(http.StatusBadRequest, "expense_date must be in YYYY-MM-DD format")
	}
	return parsed, nil
}

func expenseBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func expenseStatusError(code int, message string) error {
	return onekiterrors.New(code, message)
}
