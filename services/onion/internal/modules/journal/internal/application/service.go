package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	journalScope = "journal"
	journalTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.JournalEntry, error)
	Create(ctx context.Context, ownerID int64, input domain.JournalEntryInput) (int64, error)
	Post(ctx context.Context, ownerID, id int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateJournalScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(journalScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID, page, pageSize int64) ([]domain.JournalEntry, bool, error) {
	limit, offset := paginate(page, pageSize)
	key := cacheKeys.Scoped(journalScope, ownerID, fmt.Sprintf("list|p=%d|ps=%d", page, pageSize))
	rows, err := cacher.Of[[]domain.JournalEntry]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.JournalEntry, error) {
			return s.repo.List(ctx, ownerID, int32(limit+1), int32(offset))
		},
		cacher.WithExpiry(ctx, journalTTL),
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

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.JournalEntryInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateJournalScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Post(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Post(ctx, ownerID, id); err != nil {
		return err
	}
	invalidateJournalScope(ctx, ownerID)
	return nil
}
