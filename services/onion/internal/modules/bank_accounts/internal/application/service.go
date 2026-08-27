package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/money"
)

const (
	bankAccountsScope = "bank_accounts"
	bankAccountsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.BankAccount, error)
	Create(ctx context.Context, businessID int64, input domain.BankAccountInput) (int64, error)
	TotalBalance(ctx context.Context, businessID int64) (money.Amount, error)
	CountConnected(ctx context.Context, businessID int64) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateBankAccountsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(bankAccountsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.BankAccount, error) {
	return cacher.Of[[]domain.BankAccount]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(bankAccountsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.BankAccount, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, bankAccountsTTL),
	)
}

func (s *Service) GetByID(ctx context.Context, businessID, bankAccountID int64) (domain.BankAccount, error) {
	accounts, err := s.List(ctx, businessID)
	if err != nil {
		return domain.BankAccount{}, err
	}
	for _, account := range accounts {
		if account.ID == bankAccountID {
			return account, nil
		}
	}
	return domain.BankAccount{}, domain.ErrBankAccountNotFound
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.BankAccountInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateBankAccountsScope(ctx, businessID)
	return id, nil
}

// TotalBalance returns the aggregate balance across all bank accounts for the
// owner. It queries the database directly (no cache) because balance data is
// time-sensitive and should never be stale.
func (s *Service) TotalBalance(ctx context.Context, businessID int64) (money.Amount, error) {
	return s.repo.TotalBalance(ctx, businessID)
}

// ConnectedCount returns how many accounts have is_connected=TRUE for the
// owner. Useful for showing integration status in the dashboard.
func (s *Service) ConnectedCount(ctx context.Context, businessID int64) (int64, error) {
	return s.repo.CountConnected(ctx, businessID)
}
