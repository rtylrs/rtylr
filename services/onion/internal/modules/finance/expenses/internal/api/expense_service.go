package api

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/application"
	"github.com/voxire/onion/internal/modules/finance/expenses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

// listExpensesDefaultPageSize and listExpensesMaxPageSize bound the
// finance.expenses.v1.ListExpenses page: an omitted or non-positive page_size
// asks for 50 rows and no client can request more than 500 in one page.
const (
	listExpensesDefaultPageSize int64 = 50
	listExpensesMaxPageSize     int64 = 500
)

type ExpensesAPI struct {
	Service *application.Service
}

// Registration only checks the interface at runtime, so a missing handler
// would otherwise surface as a startup panic.
var _ financeexpensesv1.ExpensesServiceServer = (*ExpensesAPI)(nil)

// maxCategoryNameLength matches finance_expense_categories.name varchar(255).
const maxCategoryNameLength = 255

func NewExpensesAPI(service *application.Service) *ExpensesAPI {
	return &ExpensesAPI{Service: service}
}

func (a *ExpensesAPI) ListExpenses(ctx context.Context, req *financeexpensesv1.ListExpensesRequest) (*financeexpensesv1.ListExpensesResponse, error) {
	if err := requireExpensePermission(ctx, "finance:expenses:view"); err != nil {
		return nil, err
	}
	businessID, err := expenseBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize := req.GetPage(), req.GetPageSize()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = listExpensesDefaultPageSize
	}
	if pageSize > listExpensesMaxPageSize {
		pageSize = listExpensesMaxPageSize
	}
	// Clamp the offset before narrowing to int32 so an absurd page number can
	// never wrap into a mid-table window; any offset past the table is empty.
	offset := int64(0)
	if page-1 > math.MaxInt32/pageSize {
		offset = math.MaxInt32
	} else {
		offset = (page - 1) * pageSize
	}
	expenses, listErr := a.Service.ListPaginated(ctx, businessID, int32(pageSize), int32(offset))
	if listErr != nil {
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to list expenses")
	}
	payloads := make([]*financeexpensesv1.Expense, 0, len(expenses))
	for _, expense := range expenses {
		payloads = append(payloads, newExpensePayload(expense))
	}
	return &financeexpensesv1.ListExpensesResponse{
		Expenses: payloads,
		Page:     page,
		PageSize: pageSize,
		// A full page may be followed by more rows; the client stops when a
		// short page comes back.
		HasMore: len(expenses) == int(pageSize),
	}, nil
}

func (a *ExpensesAPI) CreateExpense(ctx context.Context, req *financeexpensesv1.CreateExpenseRequest) (*financeexpensesv1.CreateExpenseResponse, error) {
	user, permissionErr := requireExpenseUserPermission(ctx, "finance:expenses:create")
	if permissionErr != nil {
		return nil, permissionErr
	}
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
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
	if validationErr := domain.ValidatePaymentMethod(req.GetPaymentMethod()); validationErr != nil {
		return nil, expenseStatusError(http.StatusBadRequest, validationErr.Error())
	}
	id, createErr := a.Service.Create(ctx, domain.ExpenseInput{
		OwnerID:         user.UserID,
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
		if errors.Is(createErr, domain.ErrCategoryNotFound) {
			return nil, expenseStatusError(http.StatusBadRequest, "Expense category not found")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to create expense")
	}
	return &financeexpensesv1.CreateExpenseResponse{ExpenseId: id, Message: "Expense created successfully"}, nil
}

func (a *ExpensesAPI) UpdateExpense(ctx context.Context, req *financeexpensesv1.UpdateExpenseRequest) (*financeexpensesv1.UpdateExpenseResponse, error) {
	user, permissionErr := requireExpenseUserPermission(ctx, "finance:expenses:create")
	if permissionErr != nil {
		return nil, permissionErr
	}
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
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
	if validationErr := domain.ValidatePaymentMethod(body.GetPaymentMethod()); validationErr != nil {
		return nil, expenseStatusError(http.StatusBadRequest, validationErr.Error())
	}
	if updateErr := a.Service.Update(ctx, domain.ExpenseUpdateInput{
		ID:              req.GetId(),
		OwnerID:         user.UserID,
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
		case errors.Is(updateErr, domain.ErrCategoryNotFound):
			return nil, expenseStatusError(http.StatusBadRequest, "Expense category not found")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to update expense")
	}
	return &financeexpensesv1.UpdateExpenseResponse{Message: "Expense updated successfully"}, nil
}

func (a *ExpensesAPI) DeleteExpense(ctx context.Context, req *financeexpensesv1.DeleteExpenseRequest) (*financeexpensesv1.DeleteExpenseResponse, error) {
	if err := requireExpensePermission(ctx, "finance:expenses:create"); err != nil {
		return nil, err
	}
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
	if err := requireExpensePermission(ctx, "finance:expenses:view"); err != nil {
		return nil, err
	}
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
	if err := requireExpensePermission(ctx, "finance:expenses:approve"); err != nil {
		return nil, err
	}
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
	if err := requireExpensePermission(ctx, "finance:expenses:view"); err != nil {
		return nil, err
	}
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
		payloads = append(payloads, newCategoryPayload(category))
	}
	return &financeexpensesv1.ListExpenseCategoriesResponse{Categories: payloads}, nil
}

func (a *ExpensesAPI) CreateExpenseCategory(ctx context.Context, req *financeexpensesv1.CreateExpenseCategoryRequest) (*financeexpensesv1.CreateExpenseCategoryResponse, error) {
	user, permissionErr := requireExpenseUserPermission(ctx, "finance:expenses:create")
	if permissionErr != nil {
		return nil, permissionErr
	}
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	name, nameErr := categoryName(req.GetName())
	if nameErr != nil {
		return nil, nameErr
	}
	category, created, createErr := a.Service.CreateExpenseCategory(ctx, domain.ExpenseCategoryInput{
		OwnerID:    user.UserID,
		BusinessID: businessID,
		Name:       name,
		Color:      strings.TrimSpace(req.GetColor()),
	})
	if createErr != nil {
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to create expense category")
	}
	return &financeexpensesv1.CreateExpenseCategoryResponse{Category: newCategoryPayload(category), Created: created}, nil
}

func (a *ExpensesAPI) UpdateExpenseCategory(ctx context.Context, req *financeexpensesv1.UpdateExpenseCategoryRequest) (*financeexpensesv1.UpdateExpenseCategoryResponse, error) {
	user, permissionErr := requireExpenseUserPermission(ctx, "finance:expenses:create")
	if permissionErr != nil {
		return nil, permissionErr
	}
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	body := req.GetBody()
	if body == nil {
		return nil, expenseStatusError(http.StatusBadRequest, "Category body is required")
	}
	name, nameErr := categoryName(body.GetName())
	if nameErr != nil {
		return nil, nameErr
	}
	category, updateErr := a.Service.UpdateExpenseCategory(ctx, businessID, req.GetId(), domain.ExpenseCategoryInput{
		BusinessID: businessID,
		Name:       name,
		Color:      strings.TrimSpace(body.GetColor()),
	})
	if updateErr != nil {
		switch {
		case errors.Is(updateErr, domain.ErrCategoryNotFound):
			return nil, expenseStatusError(http.StatusNotFound, "Expense category not found")
		case errors.Is(updateErr, domain.ErrCategoryNameTaken):
			return nil, expenseStatusError(http.StatusConflict, "Another category already uses this name")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to update expense category")
	}
	return &financeexpensesv1.UpdateExpenseCategoryResponse{Category: newCategoryPayload(category)}, nil
}

func (a *ExpensesAPI) DeleteExpenseCategory(ctx context.Context, req *financeexpensesv1.DeleteExpenseCategoryRequest) (*financeexpensesv1.DeleteExpenseCategoryResponse, error) {
	if err := requireExpensePermission(ctx, "finance:expenses:create"); err != nil {
		return nil, err
	}
	businessID, err := expenseBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteExpenseCategory(ctx, businessID, req.GetId()); deleteErr != nil {
		switch {
		case errors.Is(deleteErr, domain.ErrCategoryNotFound):
			return nil, expenseStatusError(http.StatusNotFound, "Expense category not found")
		case errors.Is(deleteErr, domain.ErrCategoryInUse):
			return nil, expenseStatusError(http.StatusConflict, "This category still has expenses and cannot be deleted")
		}
		return nil, expenseStatusError(http.StatusInternalServerError, "Failed to delete expense category")
	}
	return &financeexpensesv1.DeleteExpenseCategoryResponse{Message: "Expense category deleted successfully"}, nil
}

func categoryName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", expenseStatusError(http.StatusBadRequest, "Category name is required")
	}
	if utf8.RuneCountInString(name) > maxCategoryNameLength {
		return "", expenseStatusError(http.StatusBadRequest, "Category name is too long")
	}
	return name, nil
}

func newCategoryPayload(category domain.ExpenseCategory) *financeexpensesv1.ExpenseCategory {
	return &financeexpensesv1.ExpenseCategory{
		Id:        category.ID,
		OwnerId:   category.OwnerID,
		Name:      category.Name,
		Color:     category.Color,
		CreatedAt: category.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
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

func requireExpensePermission(ctx context.Context, permission string) error {
	_, err := requireExpenseUserPermission(ctx, permission)
	return err
}

func requireExpenseUserPermission(ctx context.Context, permission string) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppFinance, permission) {
		return nil, onekiterrors.PermissionDenied()
	}
	return device.User, nil
}

func expenseStatusError(code int, message string) error {
	return onekiterrors.New(code, message)
}
