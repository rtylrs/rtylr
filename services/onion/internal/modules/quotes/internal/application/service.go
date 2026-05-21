package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/quotes/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	quotesScope = "quotes"
	quotesTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Quote, error)
	Create(ctx context.Context, ownerID int64, input domain.QuoteInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.QuoteInput) error
	Delete(ctx context.Context, ownerID, id int64) error
	ConvertToInvoice(ctx context.Context, ownerID, quoteID int64) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateQuotesScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(quotesScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Quote, error) {
	return cacher.Of[[]domain.Quote]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(quotesScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Quote, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, quotesTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.QuoteInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateQuotesScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.QuoteInput) error {
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	invalidateQuotesScope(ctx, ownerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Delete(ctx, ownerID, id); err != nil {
		return err
	}
	invalidateQuotesScope(ctx, ownerID)
	return nil
}

func (s *Service) ConvertToInvoice(ctx context.Context, ownerID, quoteID int64) (int64, error) {
	id, err := s.repo.ConvertToInvoice(ctx, ownerID, quoteID)
	if err != nil {
		return 0, err
	}
	invalidateQuotesScope(ctx, ownerID)
	return id, nil
}
