package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListBySession(ctx context.Context, businessID, sessionID int64) ([]domain.CashMovement, error) {
	return s.repo.ListBySession(ctx, businessID, sessionID)
}

func (s *Service) CashDrop(ctx context.Context, input domain.CreateCashMovementInput) (int64, error) {
	input.Type = "drop"
	return s.repo.Create(ctx, input)
}

func (s *Service) CashPickup(ctx context.Context, input domain.CreateCashMovementInput) (int64, error) {
	input.Type = "pickup"
	return s.repo.Create(ctx, input)
}

func (s *Service) CashAdjust(ctx context.Context, input domain.CreateCashMovementInput) (int64, error) {
	input.Type = "adjustment"
	return s.repo.Create(ctx, input)
}
