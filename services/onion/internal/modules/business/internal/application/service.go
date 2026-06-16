package application

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
	"github.com/voxire/onion/internal/modules/business/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const businessProfileCacheTTL = 10 * time.Minute

type repository interface {
	GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error)
	GetPublicBusiness(ctx context.Context, businessID int64) (domain.PublicBusiness, error)
	LookupBySlug(ctx context.Context, slug string) (domain.SlugLookupResult, error)
	UpdateSlug(ctx context.Context, businessID int64, slug string) error
	ListUserBusinesses(ctx context.Context, userID int64) ([]domain.OwnedBusiness, error)
	CreateSubBusiness(ctx context.Context, name string, ownerUserID, parentBusinessID int64) (int64, error)
	CheckSubBusinessOwnership(ctx context.Context, ownerUserID, parentBusinessID int64) (bool, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error) {
	return cacher.Of[domain.BusinessProfile]().GetOrComputeSingle(
		ctx,
		cacheKeys.BusinessConfig(ownerID),
		func(ctx context.Context) (domain.BusinessProfile, error) {
			return s.repo.GetProfile(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, businessProfileCacheTTL),
	)
}

func (s *Service) GetPublicBusiness(ctx context.Context, businessID int64) (domain.PublicBusiness, error) {
	return s.repo.GetPublicBusiness(ctx, businessID)
}

func (s *Service) LookupBySlug(ctx context.Context, slug string) (domain.SlugLookupResult, error) {
	return s.repo.LookupBySlug(ctx, slug)
}

func (s *Service) UpdateSlug(ctx context.Context, businessID int64, slug string) error {
	if err := s.repo.UpdateSlug(ctx, businessID, slug); err != nil {
		return err
	}
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.Delete(ctx, []string{cacheKeys.BusinessConfig(businessID)})
	}
	return nil
}

func (s *Service) ListUserBusinesses(ctx context.Context, userID int64) ([]domain.OwnedBusiness, error) {
	return s.repo.ListUserBusinesses(ctx, userID)
}

var ErrSubBusinessNotOwned = errors.New("parent business not found or not owned by user")
var ErrSubBusinessNameEmpty = errors.New("business name is required")

func (s *Service) CreateSubBusiness(ctx context.Context, name string, ownerUserID, parentBusinessID int64) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, ErrSubBusinessNameEmpty
	}
	ok, err := s.repo.CheckSubBusinessOwnership(ctx, ownerUserID, parentBusinessID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrSubBusinessNotOwned
	}
	return s.repo.CreateSubBusiness(ctx, name, ownerUserID, parentBusinessID)
}
