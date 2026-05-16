package application

import (
	"context"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/price_lists/internal/domain"
	"github.com/voxire/onion/internal/modules/price_lists/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const defaultCurrency = "USD"

const (
	priceListsScope = "price_lists"
	priceListsTTL   = 10 * time.Minute
)

type repository interface {
	ListPriceLists(ctx context.Context, ownerID int64) ([]domain.PriceList, error)
	CreatePriceList(ctx context.Context, ownerID int64, input domain.PriceListInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidatePriceListsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(priceListsScope, ownerID))
	}
}

func (s *Service) ListPriceLists(ctx context.Context, ownerID int64) ([]domain.PriceList, error) {
	return cacher.Of[[]domain.PriceList]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(priceListsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.PriceList, error) {
			return s.repo.ListPriceLists(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, priceListsTTL),
	)
}

func (s *Service) CreatePriceList(ctx context.Context, ownerID int64, input domain.PriceListInput) (int64, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Currency = strings.TrimSpace(strings.ToUpper(input.Currency))
	if input.Currency == "" {
		input.Currency = defaultCurrency
	}
	id, err := s.repo.CreatePriceList(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidatePriceListsScope(ctx, ownerID)
	return id, nil
}
