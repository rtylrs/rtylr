package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/store_categories/internal/domain"
	"github.com/voxire/onion/internal/modules/store_categories/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo *infrastructure.Repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.StoreCategory, error) {
	cacheKey := "store_categories:" + strconv.FormatInt(businessID, 10)
	return cacher.Of[[]domain.StoreCategory]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.StoreCategory, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.StoreCategoryInput) (int64, error) {
	if err := s.verifyParent(ctx, businessID, 0, input.ParentID); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.StoreCategoryInput) error {
	if err := s.verifyParent(ctx, businessID, id, input.ParentID); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

// verifyParent ensures a non-zero parent_id references a category owned by the
// same tenant, and that a category is not set as its own parent.
func (s *Service) verifyParent(ctx context.Context, businessID, selfID, parentID int64) error {
	if parentID == 0 {
		return nil
	}
	if parentID == selfID {
		return domain.ErrParentNotFound
	}
	if _, err := s.repo.GetByID(ctx, businessID, parentID); err != nil {
		return err
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	prefix := "store_categories:" + strconv.FormatInt(businessID, 10)
	cacher.Of[[]domain.StoreCategory]().Invalidate(ctx, []string{prefix})
}
