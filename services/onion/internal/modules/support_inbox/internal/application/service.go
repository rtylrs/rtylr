package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/domain"
	"github.com/voxire/onion/internal/modules/support_inbox/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	supportInboxScope = "support_inbox"
	supportInboxTTL   = 30 * time.Second
)

type repository interface {
	ListInbox(ctx context.Context, businessID int64, page, pageSize int) ([]domain.InboxItem, int, error)
	UnreadCount(ctx context.Context, businessID int64) (int, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

type inboxListResult struct {
	Items []domain.InboxItem
	Total int
}

func (s *Service) ListInbox(ctx context.Context, businessID int64, page, pageSize int) ([]domain.InboxItem, int, error) {
	key := cacheKeys.Scoped(supportInboxScope, businessID, fmt.Sprintf("inbox|p=%d|ps=%d", page, pageSize))
	res, err := cacher.Of[inboxListResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (inboxListResult, error) {
			items, total, err := s.repo.ListInbox(ctx, businessID, page, pageSize)
			return inboxListResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, supportInboxTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) UnreadCount(ctx context.Context, businessID int64) (int, error) {
	return cacher.Of[int]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(supportInboxScope, businessID, "unread_count"),
		func(ctx context.Context) (int, error) {
			return s.repo.UnreadCount(ctx, businessID)
		},
		cacher.WithExpiry(ctx, supportInboxTTL),
	)
}
