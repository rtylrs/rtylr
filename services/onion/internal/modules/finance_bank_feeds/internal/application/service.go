package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	bankFeedsScope    = "finance_bank_feeds"
	bankFeedsTTL      = 2 * time.Minute
	bankAccountsScope = "bank_accounts"
)

type repository interface {
	List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error)
	Sync(ctx context.Context, businessID int64) (int, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

type bankFeedsListResult struct {
	Items []domain.BankFeedTransaction
	Total int64
}

func (s *Service) List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	key := cacheKeys.Scoped(bankFeedsScope, businessID, fmt.Sprintf("list|l=%d|o=%d", limit, offset))
	res, err := cacher.Of[bankFeedsListResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (bankFeedsListResult, error) {
			items, total, err := s.repo.List(ctx, businessID, limit, offset)
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
