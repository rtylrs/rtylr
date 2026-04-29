package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/infrastructure"
)

type repository interface {
	List(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, ownerID, limit, offset)
}

// Sync is a stub until the real bank connector (Plaid, MX, etc.) lands
// post-launch. It currently imports nothing and always returns a zero
// count so the frontend can wire up the action on the final contract.
func (s *Service) Sync(_ context.Context, _ int64) (int, error) {
	return 0, nil
}
