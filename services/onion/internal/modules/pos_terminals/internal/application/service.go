package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_terminals/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posTerminalsScope = "pos_terminals"
	posTerminalsTTL   = 1 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Terminal, error)
	Create(ctx context.Context, input domain.TerminalInput) (int64, error)
	Update(ctx context.Context, input domain.TerminalUpdateInput) error
	Heartbeat(ctx context.Context, terminalID, businessID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidatePosTerminalsScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posTerminalsScope, businessID))
	}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Terminal, error) {
	return cacher.Of[[]domain.Terminal]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posTerminalsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Terminal, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posTerminalsTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.TerminalInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidatePosTerminalsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.TerminalUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidatePosTerminalsScope(ctx, input.BusinessID)
	return nil
}

// Heartbeat is a hot-path liveness ping that runs every few seconds; we
// intentionally do NOT invalidate the list cache on heartbeat (the TTL ages it
// out fast enough and the field that changes is last_seen, not membership).
func (s *Service) Heartbeat(ctx context.Context, terminalID, businessID int64) error {
	return s.repo.Heartbeat(ctx, terminalID, businessID)
}
