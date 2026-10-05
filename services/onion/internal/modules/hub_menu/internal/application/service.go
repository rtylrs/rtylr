package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

var ErrInvalidMenuItems = errors.New("invalid hub menu items")

const (
	maxBannerTitleLength = 255
	maxBannerBodyLength  = 2000
	maxBannerURLLength   = 2048
)

type repository interface {
	ListBanners(ctx context.Context, businessID int64) ([]domain.Banner, error)
	CreateBanner(ctx context.Context, ownerID, businessID int64, input domain.BannerInput) (int64, error)
	UpdateBanner(ctx context.Context, businessID, id int64, input domain.BannerInput) error
	DeleteBanner(ctx context.Context, businessID, id int64) error
	GetBanner(ctx context.Context, businessID, id int64) (domain.Banner, error)
	BusinessOwnerID(ctx context.Context, businessID int64) (int64, error)
	List(ctx context.Context, businessID int64) ([]*posmenuitemv1.MenuItem, error)
	Update(ctx context.Context, ownerID, businessID int64, items []*posmenuitemv1.MenuItem) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListBanners(ctx context.Context, businessID int64) ([]domain.Banner, error) {
	return s.repo.ListBanners(ctx, businessID)
}

func (s *Service) List(ctx context.Context, businessID int64) ([]*posmenuitemv1.MenuItem, error) {
	return s.repo.List(ctx, businessID)
}

func (s *Service) Update(ctx context.Context, businessID int64, items []*posmenuitemv1.MenuItem) error {
	if len(items) == 0 {
		return fmt.Errorf("%w: at least one menu item is required", ErrInvalidMenuItems)
	}
	for i, item := range items {
		if item == nil {
			return fmt.Errorf("%w: item[%d] is required", ErrInvalidMenuItems, i)
		}
	}
	ownerID, err := s.repo.BusinessOwnerID(ctx, businessID)
	if err != nil {
		return err
	}
	return s.repo.Update(ctx, ownerID, businessID, items)
}

func (s *Service) CreateBanner(ctx context.Context, businessID int64, input domain.BannerInput) (domain.Banner, error) {
	normalized, err := normalizeBannerInput(input)
	if err != nil {
		return domain.Banner{}, err
	}
	ownerID, err := s.repo.BusinessOwnerID(ctx, businessID)
	if err != nil {
		return domain.Banner{}, err
	}
	id, err := s.repo.CreateBanner(ctx, ownerID, businessID, normalized)
	if err != nil {
		return domain.Banner{}, err
	}
	return s.repo.GetBanner(ctx, businessID, id)
}

func (s *Service) UpdateBanner(ctx context.Context, businessID, id int64, input domain.BannerInput) (domain.Banner, error) {
	if id <= 0 {
		return domain.Banner{}, domain.ErrBannerNotFound
	}
	normalized, err := normalizeBannerInput(input)
	if err != nil {
		return domain.Banner{}, err
	}
	if err := s.repo.UpdateBanner(ctx, businessID, id, normalized); err != nil {
		return domain.Banner{}, err
	}
	return s.repo.GetBanner(ctx, businessID, id)
}

func (s *Service) DeleteBanner(ctx context.Context, businessID, id int64) error {
	if id <= 0 {
		return domain.ErrBannerNotFound
	}
	return s.repo.DeleteBanner(ctx, businessID, id)
}

func normalizeBannerInput(input domain.BannerInput) (domain.BannerInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.LinkURL = strings.TrimSpace(input.LinkURL)

	if input.Title == "" {
		return domain.BannerInput{}, fmt.Errorf("%w: title is required", domain.ErrBannerInvalidInput)
	}
	if len(input.Title) > maxBannerTitleLength {
		return domain.BannerInput{}, fmt.Errorf("%w: title must be %d characters or fewer", domain.ErrBannerInvalidInput, maxBannerTitleLength)
	}
	if len(input.Body) > maxBannerBodyLength {
		return domain.BannerInput{}, fmt.Errorf("%w: body must be %d characters or fewer", domain.ErrBannerInvalidInput, maxBannerBodyLength)
	}
	if len(input.ImageURL) > maxBannerURLLength {
		return domain.BannerInput{}, fmt.Errorf("%w: image_url must be %d characters or fewer", domain.ErrBannerInvalidInput, maxBannerURLLength)
	}
	if len(input.LinkURL) > maxBannerURLLength {
		return domain.BannerInput{}, fmt.Errorf("%w: link_url must be %d characters or fewer", domain.ErrBannerInvalidInput, maxBannerURLLength)
	}
	if !input.StartsAt.IsZero() && !input.EndsAt.IsZero() && input.EndsAt.Before(input.StartsAt) {
		return domain.BannerInput{}, fmt.Errorf("%w: ends_at must be after starts_at", domain.ErrBannerInvalidInput)
	}
	return input, nil
}
