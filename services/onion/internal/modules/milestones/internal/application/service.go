package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/milestones/internal/domain"
	"github.com/voxire/onion/internal/modules/milestones/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	milestonesScope = "milestones"
	milestonesTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64, projectID *int64) ([]domain.Milestone, error)
	GetProject(ctx context.Context, projectID int64, ownerID int64) (int64, error)
	Create(ctx context.Context, input domain.MilestoneInput) (int64, error)
	Update(ctx context.Context, input domain.MilestoneUpdateInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateMilestonesScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(milestonesScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64, projectID *int64) ([]domain.Milestone, error) {
	pid := int64(0)
	if projectID != nil {
		pid = *projectID
	}
	key := cacheKeys.Scoped(milestonesScope, ownerID, fmt.Sprintf("list|p=%d", pid))
	return cacher.Of[[]domain.Milestone]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Milestone, error) {
			return s.repo.List(ctx, ownerID, projectID)
		},
		cacher.WithExpiry(ctx, milestonesTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	// Verify the target project belongs to the authenticated owner before
	// allowing a milestone to be attached to it. Prevents creating milestones
	// against another tenant's projects.
	if _, err := s.repo.GetProject(ctx, input.ProjectID, input.OwnerID); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateMilestonesScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, input domain.MilestoneUpdateInput) error {
	if err := s.repo.Update(ctx, input); err != nil {
		return err
	}
	invalidateMilestonesScope(ctx, input.OwnerID)
	return nil
}
