package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/loyalty/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListTransactions(ctx context.Context, ownerID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	return s.repo.ListTransactions(ctx, ownerID, customerID, page, pageSize)
}

func (s *Service) AwardPoints(ctx context.Context, input domain.AwardInput) error {
	return s.repo.CreateTransaction(ctx, input, "award")
}

func (s *Service) RedeemPoints(ctx context.Context, input domain.RedeemInput) error {
	return s.repo.CreateRedeemTransaction(ctx, input)
}

func (s *Service) GetConfig(ctx context.Context, ownerID int64) (domain.LoyaltyConfig, error) {
	return s.repo.GetConfig(ctx, ownerID)
}

func (s *Service) UpdateConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	return s.repo.UpsertConfig(ctx, input)
}

func (s *Service) ListMembers(ctx context.Context, ownerID int64, limit, offset int) ([]domain.LoyaltyMember, error) {
	return s.repo.ListMembers(ctx, ownerID, limit, offset)
}

func (s *Service) ListRewards(ctx context.Context, ownerID int64, activeOnly bool) ([]domain.LoyaltyReward, error) {
	return s.repo.ListRewards(ctx, ownerID, activeOnly)
}

func (s *Service) UpsertReward(ctx context.Context, rw domain.LoyaltyReward) (int64, error) {
	return s.repo.UpsertReward(ctx, rw)
}
