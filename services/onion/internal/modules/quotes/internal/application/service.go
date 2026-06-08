package application

import (
	"context"
	"database/sql"
	"fmt"
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
	List(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.Quote, error)
	Create(ctx context.Context, ownerID int64, input domain.QuoteInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.QuoteInput) error
	UpdateStatus(ctx context.Context, ownerID, id int64, status string) error
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

func (s *Service) List(ctx context.Context, ownerID, page, pageSize int64) ([]domain.Quote, bool, error) {
	limit, offset := paginate(page, pageSize)
	key := cacheKeys.Scoped(quotesScope, ownerID, fmt.Sprintf("list|p=%d|ps=%d", page, pageSize))
	rows, err := cacher.Of[[]domain.Quote]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Quote, error) {
			return s.repo.List(ctx, ownerID, int32(limit+1), int32(offset))
		},
		cacher.WithExpiry(ctx, quotesTTL),
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func paginate(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return int(pageSize), int((page - 1) * pageSize)
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

func (s *Service) UpdateStatus(ctx context.Context, ownerID, id int64, status string) error {
	if err := s.repo.UpdateStatus(ctx, ownerID, id, status); err != nil {
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
