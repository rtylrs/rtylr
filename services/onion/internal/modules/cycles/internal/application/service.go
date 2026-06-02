package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/cycles/internal/domain"
	"github.com/voxire/onion/internal/modules/cycles/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	cyclesScope = "cycles"
	cyclesTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64, workspaceID *int64) ([]domain.Cycle, error)
	WorkspaceBelongsToOwner(ctx context.Context, ownerID, workspaceID int64) (bool, error)
	Create(ctx context.Context, input domain.CycleInput) (int64, error)
	Update(ctx context.Context, input domain.CycleUpdateInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateCyclesScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(cyclesScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64, workspaceID *int64) ([]domain.Cycle, error) {
	ws := int64(0)
	if workspaceID != nil {
		ws = *workspaceID
	}
	key := cacheKeys.Scoped(cyclesScope, ownerID, fmt.Sprintf("list|w=%d", ws))
	return cacher.Of[[]domain.Cycle]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Cycle, error) {
			return s.repo.List(ctx, ownerID, workspaceID)
		},
		cacher.WithExpiry(ctx, cyclesTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.CycleInput) (int64, error) {
	// A cycle must live in a workspace the caller owns. Without this check a user
	// could attach a cycle to another tenant's workspace (cross-tenant write).
	owns, err := s.repo.WorkspaceBelongsToOwner(ctx, input.OwnerID, input.WorkspaceID)
	if err != nil {
		return 0, err
	}
	if !owns {
		return 0, domain.ErrWorkspaceNotFound
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateCyclesScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.CycleUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateCyclesScope(ctx, input.OwnerID)
	return nil
}
