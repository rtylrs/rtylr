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
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Expense, error)
	GetByID(ctx context.Context, ownerID, id int64) (domain.Expense, error)
	Create(ctx context.Context, input domain.ExpenseInput) (int64, error)
	Update(ctx context.Context, input domain.ExpenseUpdateInput) error
	UpdateStatus(ctx context.Context, ownerID, id int64, status string) error
	Delete(ctx context.Context, ownerID, id int64) error
	ListCategories(ctx context.Context, ownerID int64) ([]domain.ExpenseCategory, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateExpensesScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(expensesScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Expense, error) {
	return cacher.Of[[]domain.Expense]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(expensesScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Expense, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, expensesTTL),
	)
}

func (s *Service) GetByID(ctx context.Context, ownerID, id int64) (domain.Expense, error) {
	key := cacheKeys.Scoped(expensesScope, ownerID, fmt.Sprintf("id=%d", id))
	return cacher.Of[domain.Expense]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Expense, error) {
			return s.repo.GetByID(ctx, ownerID, id)
		},
		cacher.WithExpiry(ctx, expensesTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.ExpenseInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateExpensesScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.ExpenseUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateExpensesScope(ctx, input.OwnerID)
	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, ownerID, id int64, status string) error {
	if err := s.repo.UpdateStatus(ctx, ownerID, id, status); err != nil {
		return err
	}
	invalidateExpensesScope(ctx, ownerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Delete(ctx, ownerID, id); err != nil {
		return err
	}
	invalidateExpensesScope(ctx, ownerID)
	return nil
}

func (s *Service) ListCategories(ctx context.Context, ownerID int64) ([]domain.ExpenseCategory, error) {
	return cacher.Of[[]domain.ExpenseCategory]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(expensesScope, ownerID, "categories"),
		func(ctx context.Context) ([]domain.ExpenseCategory, error) {
			return s.repo.ListCategories(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, 10*time.Minute),
	)
}
