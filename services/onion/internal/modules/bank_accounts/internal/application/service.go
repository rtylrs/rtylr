package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	bankAccountsScope = "bank_accounts"
	bankAccountsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.BankAccount, error)
	Create(ctx context.Context, ownerID int64, input domain.BankAccountInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateBankAccountsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(bankAccountsScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.BankAccount, error) {
	return cacher.Of[[]domain.BankAccount]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(bankAccountsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.BankAccount, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, bankAccountsTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.BankAccountInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateBankAccountsScope(ctx, ownerID)
	return id, nil
}
