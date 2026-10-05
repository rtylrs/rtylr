package application

import (
	"context"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/price_lists/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const defaultCurrency = "USD"

const (
	priceListsScope = "price_lists"
	priceListsTTL   = 10 * time.Minute
)

type repository interface {
	ListPriceLists(ctx context.Context, businessID int64) ([]domain.PriceList, error)
	CreatePriceList(ctx context.Context, businessID int64, input domain.PriceListInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidatePriceListsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(priceListsScope, businessID))
}

func (s *Service) ListPriceLists(ctx context.Context, businessID int64) ([]domain.PriceList, error) {
	return cacher.Of[[]domain.PriceList]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(priceListsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.PriceList, error) {
			return s.repo.ListPriceLists(ctx, businessID)
		},
		cacher.WithExpiry(ctx, priceListsTTL),
	)
}

func (s *Service) CreatePriceList(ctx context.Context, businessID int64, input domain.PriceListInput) (int64, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Currency = strings.TrimSpace(strings.ToUpper(input.Currency))
	if input.Currency == "" {
		input.Currency = defaultCurrency
	}
	id, err := s.repo.CreatePriceList(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidatePriceListsScope(ctx, businessID)
	return id, nil
}
