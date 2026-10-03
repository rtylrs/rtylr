package application

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const ConfigCacheTTL = 500 * time.Minute // 500 minutes = 8 hours

type repository interface {
	GetConfig(ctx context.Context, ownerID string) (domain.Config, error)
	CreateConfig(ctx context.Context, input domain.ConfigInput) error
	UpdateConfig(ctx context.Context, input domain.ConfigInput) error
	GetThemeByOwnerID(ctx context.Context, id string) (domain.Theme, error)
}

type Service struct {
	repo repository
	db   *sql.DB
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
		db:   db,
	}
}

func (s *Service) DB() *sql.DB { return s.db }
func (s *Service) GetConfig(ctx context.Context, ownerID string) (domain.Config, error) {
	idN, err := strconv.ParseInt(ownerID, 10, 64)
	if err != nil {
		return domain.Config{}, domain.ErrInvalidOwnerID
	}
	cacheKey := cacheKeys.BusinessConfig(idN)

	return cacher.Of[domain.Config]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) (domain.Config, error) {
		cfg, err := s.repo.GetConfig(ctx, ownerID)
		if err == nil {
			return cfg, nil
		}
		if !errors.Is(err, domain.ErrConfigNotFound) {
			return domain.Config{}, err
		}
		if seedErr := s.repo.CreateConfig(ctx, domain.ConfigInput{OwnerId: ownerID}); seedErr != nil {
			return domain.Config{}, seedErr
		}
		return s.repo.GetConfig(ctx, ownerID)
	}, cacher.WithExpiry(ctx, ConfigCacheTTL))
}

func (s *Service) CreateConfig(ctx context.Context, input domain.ConfigInput) error {
	return s.repo.CreateConfig(ctx, input)
}

func (s *Service) UpdateConfig(ctx context.Context, ownerID string, input domain.ConfigInput) error {
	idN, parseErr := strconv.ParseInt(ownerID, 10, 64)
	if parseErr != nil {
		return domain.ErrInvalidOwnerID
	}
	input.OwnerId = ownerID
	err := s.repo.UpdateConfig(ctx, input)
	if err != nil {
		return err
	}
	cacher.InvalidateScope(ctx, cacheKeys.BusinessPrefix(idN))
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix("hr_core", idN))
	return nil
}

func (s *Service) GetThemeByOwnerID(ctx context.Context, id string) (domain.Theme, error) {
	idN, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return domain.Theme{}, domain.ErrInvalidOwnerID
	}
	return cacher.Of[domain.Theme]().GetOrComputeSingle(ctx, cacheKeys.BusinessTheme(idN), func(ctx context.Context) (domain.Theme, error) {
		return s.repo.GetThemeByOwnerID(ctx, id)
	}, cacher.WithExpiry(ctx, ConfigCacheTTL))
}
