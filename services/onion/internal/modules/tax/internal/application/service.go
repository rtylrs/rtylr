package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	"github.com/voxire/onion/internal/modules/tax/internal/infrastructure"
)

type repository interface {
	GetConfig(ctx context.Context, ownerID int64) (domain.TaxConfig, error)
	UpsertConfig(ctx context.Context, input domain.TaxConfigInput) error
	GetPLReport(ctx context.Context, ownerID int64) (domain.PLReport, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) GetConfig(ctx context.Context, ownerID int64) (domain.TaxConfig, error) {
	return s.repo.GetConfig(ctx, ownerID)
}

func (s *Service) UpdateConfig(ctx context.Context, input domain.TaxConfigInput) error {
	return s.repo.UpsertConfig(ctx, input)
}

func (s *Service) GetPLReport(ctx context.Context, ownerID int64) (domain.PLReport, error) {
	return s.repo.GetPLReport(ctx, ownerID)
}
