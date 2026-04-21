package application

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/price_lists/internal/domain"
	"github.com/voxire/onion/internal/modules/price_lists/internal/infrastructure"
)

const defaultCurrency = "USD"

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListPriceLists(ctx context.Context, ownerID int64) ([]domain.PriceList, error) {
	return s.repo.ListPriceLists(ctx, ownerID)
}

func (s *Service) CreatePriceList(ctx context.Context, ownerID int64, input domain.PriceListInput) (int64, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Currency = strings.TrimSpace(strings.ToUpper(input.Currency))

	if input.Currency == "" {
		input.Currency = defaultCurrency
	}

	return s.repo.CreatePriceList(ctx, ownerID, input)
}
