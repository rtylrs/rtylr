package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	bankFeedsScope    = "finance_bank_feeds"
	bankFeedsTTL      = 2 * time.Minute
	bankAccountsScope = "bank_accounts"
)

var (
	ErrBankAccountNotFound = errors.New("bank account not found")
	ErrInvalidBankAccount  = errors.New("invalid bank account")
)

type repository interface {
	List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error)
	Sync(ctx context.Context, businessID int64) (int, error)
}

type accountRepository interface {
	ListByAccount(ctx context.Context, businessID, bankAccountID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error)
	BankAccountExists(ctx context.Context, businessID, bankAccountID int64) (bool, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

type bankFeedsListResult struct {
	Items []domain.BankFeedTransaction
	Total int64
}

func (s *Service) List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	return s.list(ctx, businessID, 0, limit, offset)
}

func (s *Service) ListByAccount(ctx context.Context, businessID, bankAccountID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	if bankAccountID < 0 {
		return nil, 0, ErrInvalidBankAccount
	}
	if bankAccountID == 0 {
		return s.list(ctx, businessID, 0, limit, offset)
	}

	repo, ok := s.repo.(accountRepository)
	if !ok {
		return nil, 0, errors.New("bank-feed repository does not support account filtering")
	}
	exists, err := repo.BankAccountExists(ctx, businessID, bankAccountID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, ErrBankAccountNotFound
	}
	return s.list(ctx, businessID, bankAccountID, limit, offset)
}

func (s *Service) list(ctx context.Context, businessID, bankAccountID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	key := cacheKeys.Scoped(bankFeedsScope, businessID, fmt.Sprintf("list|a=%d|l=%d|o=%d", bankAccountID, limit, offset))
	res, err := cacher.Of[bankFeedsListResult]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (bankFeedsListResult, error) {
			var items []domain.BankFeedTransaction
			var total int64
			var err error
			if bankAccountID > 0 {
				items, total, err = s.repo.(accountRepository).ListByAccount(ctx, businessID, bankAccountID, limit, offset)
			} else {
				items, total, err = s.repo.List(ctx, businessID, limit, offset)
			}
			return bankFeedsListResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, bankFeedsTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) Sync(ctx context.Context, businessID int64) (int, error) {
	imported, err := s.repo.Sync(ctx, businessID)
	if err != nil {
		return 0, err
	}

	cacher.InvalidateScope(ctx,
		cacheKeys.ScopedPrefix(bankFeedsScope, businessID),
		cacheKeys.ScopedPrefix(bankAccountsScope, businessID),
	)
	return imported, nil
}
