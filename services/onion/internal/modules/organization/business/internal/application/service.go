package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/organization/business/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	businessProfileCacheTTL = 10 * time.Minute
	userBusinessesCacheTTL  = time.Minute
)

type repository interface {
	GetBusinessOwnerID(ctx context.Context, businessID int64) (int64, error)
	GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error)
	GetPublicBusiness(ctx context.Context, businessID int64) (domain.PublicBusiness, error)
	LookupBySlug(ctx context.Context, slug string) (domain.SlugLookupResult, error)
	UpdateSlug(ctx context.Context, businessID, ownerUserID int64, slug string) (bool, error)
	ListUserBusinesses(ctx context.Context, userID int64) ([]domain.OwnedBusiness, error)
	CreateSubBusiness(ctx context.Context, name string, ownerUserID, parentBusinessID int64) (int64, error)
	CheckSubBusinessOwnership(ctx context.Context, ownerUserID, parentBusinessID int64) (bool, error)
	RenameBusiness(ctx context.Context, businessID, ownerUserID int64, name string) (bool, error)
	SetBusinessLogo(ctx context.Context, businessID, ownerUserID int64, logoURL string) (bool, error)
	DeactivateBusiness(ctx context.Context, businessID, ownerUserID int64) (bool, error)
	CountActiveOwnedBusinesses(ctx context.Context, ownerUserID int64) (int64, error)
}

type Service struct {
	repo            repository
	publicLocations PublicLocationReader
}

type PublicLocationReader interface {
	ListPublicLocations(ctx context.Context, businessID int64) ([]domain.PublicLocation, error)
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) WithPublicLocationReader(reader PublicLocationReader) *Service {
	s.publicLocations = reader
	return s
}

func (s *Service) GetProfile(ctx context.Context, ownerID int64) (domain.BusinessProfile, error) {
	return cacher.Of[domain.BusinessProfile]().GetOrComputeSingle(
		ctx,
		cacheKeys.BusinessProfile(ownerID),
		func(ctx context.Context) (domain.BusinessProfile, error) {
			return s.repo.GetProfile(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, businessProfileCacheTTL),
	)
}

func (s *Service) GetPublicBusiness(ctx context.Context, businessID int64) (domain.PublicBusiness, error) {
	return s.repo.GetPublicBusiness(ctx, businessID)
}

func (s *Service) IsOwner(ctx context.Context, userID int64) (bool, error) {
	if userID <= 0 {
		return false, nil
	}
	count, err := s.repo.CountActiveOwnedBusinesses(ctx, userID)
	return count > 0, err
}

func (s *Service) GetBusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	return s.repo.GetBusinessOwnerID(ctx, businessID)
}

func (s *Service) LookupBySlug(ctx context.Context, slug string) (domain.SlugLookupResult, error) {
	result, err := s.repo.LookupBySlug(ctx, slug)
	if err != nil {
		return domain.SlugLookupResult{}, err
	}
	if s.publicLocations == nil {
		return result, nil
	}
	result.Locations, err = s.publicLocations.ListPublicLocations(ctx, result.Business.ID)
	if err != nil {
		return domain.SlugLookupResult{}, err
	}
	return result, nil
}

// UpdateSlug renames the storefront slug of businessID on behalf of ownerUserID.
// The owner comes from the authenticated caller, so a business the caller does
// not own is reported as not owned instead of being renamed.
func (s *Service) UpdateSlug(ctx context.Context, businessID, ownerUserID int64, slug string) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if domain.ReservedBusinessSlug(slug) {
		return ErrSlugReserved
	}
	updated, err := s.repo.UpdateSlug(ctx, businessID, ownerUserID, slug)
	if err != nil {
		return err
	}
	if !updated {
		return ErrBusinessNotOwned
	}
	cacher.Of[domain.BusinessProfile]().Delete(ctx, []string{cacheKeys.BusinessProfile(businessID)})
	return nil
}

func (s *Service) ListUserBusinesses(ctx context.Context, userID int64) ([]domain.OwnedBusiness, error) {
	return cacher.Of[[]domain.OwnedBusiness]().GetOrComputeSingle(
		ctx,
		cacheKeys.UserBusinesses(userID),
		func(ctx context.Context) ([]domain.OwnedBusiness, error) {
			return s.repo.ListUserBusinesses(ctx, userID)
		},
		cacher.WithExpiry(ctx, userBusinessesCacheTTL),
	)
}

var (
	ErrSlugReserved             = domain.ErrSlugReserved
	ErrSubBusinessNotOwned      = errors.New("parent business not found or not owned by user")
	ErrSubBusinessNameEmpty     = errors.New("business name is required")
	ErrBusinessNotOwned         = errors.New("business not found or not owned by user")
	ErrBusinessNameEmpty        = errors.New("business name is required")
	ErrLogoURLEmpty             = errors.New("logo url is required")
	ErrCannotDeleteOnlyBusiness = errors.New("cannot delete the only remaining business")
)

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
	id, err := s.repo.CreateSubBusiness(ctx, name, ownerUserID, parentBusinessID)
	if err != nil {
		return 0, err
	}
	cacher.Of[[]domain.OwnedBusiness]().Delete(ctx, []string{cacheKeys.UserBusinesses(ownerUserID)})
	return id, nil
}

func (s *Service) RenameBusiness(ctx context.Context, businessID, ownerUserID int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrBusinessNameEmpty
	}

	updated, err := s.repo.RenameBusiness(ctx, businessID, ownerUserID, name)
	if err != nil {
		return err
	}
	if !updated {
		return ErrBusinessNotOwned
	}

	cacher.Of[domain.BusinessProfile]().Delete(ctx, []string{cacheKeys.BusinessProfile(businessID)})
	cacher.Of[[]domain.OwnedBusiness]().Delete(ctx, []string{cacheKeys.UserBusinesses(ownerUserID)})
	return nil
}

func (s *Service) SetLogo(ctx context.Context, businessID, ownerUserID int64, logoURL string) error {
	logoURL = strings.TrimSpace(logoURL)
	if logoURL == "" {
		return ErrLogoURLEmpty
	}

	updated, err := s.repo.SetBusinessLogo(ctx, businessID, ownerUserID, logoURL)
	if err != nil {
		return err
	}
	if !updated {
		return ErrBusinessNotOwned
	}

	cacher.Of[domain.BusinessProfile]().Delete(ctx, []string{cacheKeys.BusinessProfile(businessID)})
	cacher.Of[[]domain.OwnedBusiness]().Delete(ctx, []string{cacheKeys.UserBusinesses(ownerUserID)})
	return nil
}

func (s *Service) DeleteBusiness(ctx context.Context, businessID, ownerUserID int64) error {
	count, err := s.repo.CountActiveOwnedBusinesses(ctx, ownerUserID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrCannotDeleteOnlyBusiness
	}

	deleted, err := s.repo.DeactivateBusiness(ctx, businessID, ownerUserID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrBusinessNotOwned
	}

	cacher.Of[domain.BusinessProfile]().Delete(ctx, []string{cacheKeys.BusinessProfile(businessID)})
	cacher.Of[[]domain.OwnedBusiness]().Delete(ctx, []string{cacheKeys.UserBusinesses(ownerUserID)})
	return nil
}
