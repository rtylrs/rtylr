package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	expensesScope = "expenses"
	expensesTTL   = 5 * time.Minute

	defaultListLimit int32 = 50
	maxListLimit     int32 = 500
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Expense, error)
	ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Expense, error)
	GetByID(ctx context.Context, businessID, id int64) (domain.Expense, error)
	Create(ctx context.Context, input domain.ExpenseInput) (int64, error)
	Update(ctx context.Context, input domain.ExpenseUpdateInput) error
	UpdateStatus(ctx context.Context, businessID, id int64, status string) error
	Delete(ctx context.Context, businessID, id int64) error
	ListCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error)
	GetCategoryByID(ctx context.Context, businessID, id int64) (domain.ExpenseCategory, error)
	FindCategoryByName(ctx context.Context, businessID int64, name string) (domain.ExpenseCategory, error)
	CreateCategory(ctx context.Context, input domain.ExpenseCategoryInput) (int64, error)
	UpdateCategory(ctx context.Context, businessID, id int64, input domain.ExpenseCategoryInput) error
	CountExpensesInCategory(ctx context.Context, businessID, id int64) (int64, error)
	DeleteCategory(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

// invoicesScope is where the invoices module caches the finance dashboard,
// whose expense totals and pending count read this module's table.
const invoicesScope = "invoices"

func invalidateExpensesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(expensesScope, businessID))
	// Approving an expense moves it into the dashboard's totals; without this
	// the dashboard kept the old figures for up to a minute.
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(invoicesScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Expense, error) {
	return cacher.Of[[]domain.Expense]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(expensesScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Expense, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, expensesTTL),
	)
}

// ListPaginated returns a single page of expenses. A non-positive limit
// defaults to 50 and limit is capped at 500; a negative offset is clamped to
// zero.
func (s *Service) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Expense, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPaginated(ctx, businessID, limit, offset)
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.Expense, error) {
	key := cacheKeys.Scoped(expensesScope, businessID, fmt.Sprintf("id=%d", id))
	return cacher.Of[domain.Expense]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (domain.Expense, error) {
			return s.repo.GetByID(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, expensesTTL),
	)
}

// requireCategory rejects a category id that does not belong to the business.
// A nil or zero id means "uncategorised" and is always allowed.
func (s *Service) requireCategory(ctx context.Context, businessID int64, categoryID *int64) error {
	if categoryID == nil || *categoryID == 0 {
		return nil
	}
	_, err := s.repo.GetCategoryByID(ctx, businessID, *categoryID)
	return err
}

func (s *Service) Create(ctx context.Context, input domain.ExpenseInput) (int64, error) {
	if err := s.requireCategory(ctx, input.BusinessID, input.CategoryID); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateExpensesScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.ExpenseUpdateInput) error {
	if err := s.requireCategory(ctx, input.BusinessID, input.CategoryID); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateExpensesScope(ctx, input.BusinessID)
	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, businessID, id int64, status string) error {
	if err := s.repo.UpdateStatus(ctx, businessID, id, status); err != nil {
		return err
	}
	invalidateExpensesScope(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateExpensesScope(ctx, businessID)
	return nil
}

// ListCategories reads the table directly. It used to be cached under the
// expenses scope, but the invoices module also writes categories and never
// cleared that scope, so a new category stayed invisible for ten minutes.
func (s *Service) ListCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error) {
	return s.repo.ListCategories(ctx, businessID)
}

// CreateExpenseCategory stores a new category, or returns the existing one when the
// business already has a category with that name (ignoring case). created
// reports which happened, so a double-submitted form never leaves duplicates.
func (s *Service) CreateExpenseCategory(ctx context.Context, input domain.ExpenseCategoryInput) (domain.ExpenseCategory, bool, error) {
	input.Name = strings.TrimSpace(input.Name)
	existing, err := s.repo.FindCategoryByName(ctx, input.BusinessID, input.Name)
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, domain.ErrCategoryNotFound) {
		return domain.ExpenseCategory{}, false, err
	}
	id, err := s.repo.CreateCategory(ctx, input)
	if errors.Is(err, domain.ErrCategoryNameTaken) {
		// A concurrent request won the insert; hand back its category.
		existing, findErr := s.repo.FindCategoryByName(ctx, input.BusinessID, input.Name)
		return existing, false, findErr
	}
	if err != nil {
		return domain.ExpenseCategory{}, false, err
	}
	invalidateExpensesScope(ctx, input.BusinessID)
	category, err := s.repo.GetCategoryByID(ctx, input.BusinessID, id)
	if err != nil {
		return domain.ExpenseCategory{}, false, err
	}
	return category, true, nil
}

// UpdateExpenseCategory renames or recolours a category. Taking a name another
// category already uses fails with domain.ErrCategoryNameTaken.
func (s *Service) UpdateExpenseCategory(ctx context.Context, businessID, id int64, input domain.ExpenseCategoryInput) (domain.ExpenseCategory, error) {
	input.Name = strings.TrimSpace(input.Name)
	if _, err := s.repo.GetCategoryByID(ctx, businessID, id); err != nil {
		return domain.ExpenseCategory{}, err
	}
	clash, err := s.repo.FindCategoryByName(ctx, businessID, input.Name)
	switch {
	case err == nil && clash.ID != id:
		return domain.ExpenseCategory{}, domain.ErrCategoryNameTaken
	case err != nil && !errors.Is(err, domain.ErrCategoryNotFound):
		return domain.ExpenseCategory{}, err
	}
	if err := s.repo.UpdateCategory(ctx, businessID, id, input); err != nil {
		return domain.ExpenseCategory{}, err
	}
	invalidateExpensesScope(ctx, businessID)
	return s.repo.GetCategoryByID(ctx, businessID, id)
}

// DeleteExpenseCategory removes a category no expense is filed under. A category
// that still has expenses fails with domain.ErrCategoryInUse: approved
// expenses can never be re-filed, so deleting it would orphan their reporting.
func (s *Service) DeleteExpenseCategory(ctx context.Context, businessID, id int64) error {
	if _, err := s.repo.GetCategoryByID(ctx, businessID, id); err != nil {
		return err
	}
	count, err := s.repo.CountExpensesInCategory(ctx, businessID, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrCategoryInUse
	}
	if err := s.repo.DeleteCategory(ctx, businessID, id); err != nil {
		return err
	}
	invalidateExpensesScope(ctx, businessID)
	return nil
}
