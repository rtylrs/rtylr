package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
	"github.com/voxire/onion/internal/modules/expenses/internal/infrastructure"
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
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateExpensesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(expensesScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Expense, error) {
	return cacher.Of[[]domain.Expense]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(expensesScope, businessID, "list"),
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
	return cacher.Of[domain.Expense]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Expense, error) {
			return s.repo.GetByID(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, expensesTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.ExpenseInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateExpensesScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.ExpenseUpdateInput) error {
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

func (s *Service) ListCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error) {
	return cacher.Of[[]domain.ExpenseCategory]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(expensesScope, businessID, "categories"),
		func(ctx context.Context) ([]domain.ExpenseCategory, error) {
			return s.repo.ListCategories(ctx, businessID)
		},
		cacher.WithExpiry(ctx, 10*time.Minute),
	)
}
