package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posCashMovementsScope = "pos_cash_movements"
	posCashMovementsTTL   = 30 * time.Second
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosCashMovementsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posCashMovementsScope, businessID))
}

func (s *Service) ListBySession(ctx context.Context, businessID, sessionID int64) ([]domain.CashMovement, error) {
	key := cacheKeys.Scoped(posCashMovementsScope, businessID, fmt.Sprintf("sess=%d", sessionID))
	return cacher.Of[[]domain.CashMovement]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.CashMovement, error) {
			return s.repo.ListBySession(ctx, businessID, sessionID)
		},
		cacher.WithExpiry(ctx, posCashMovementsTTL),
	)
}

func (s *Service) create(ctx context.Context, input domain.CreateCashMovementInput, kind string) (int64, error) {
	input.Type = kind
	if err := domain.ValidateCreateCashMovementInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidatePosCashMovementsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) CashDrop(ctx context.Context, input domain.CreateCashMovementInput) (int64, error) {
	return s.create(ctx, input, "drop")
}

func (s *Service) CashPickup(ctx context.Context, input domain.CreateCashMovementInput) (int64, error) {
	return s.create(ctx, input, "pickup")
}

func (s *Service) CashAdjust(ctx context.Context, input domain.CreateCashMovementInput) (int64, error) {
	return s.create(ctx, input, "adjustment")
}
