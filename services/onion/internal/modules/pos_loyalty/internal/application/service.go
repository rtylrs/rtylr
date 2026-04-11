package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	return s.repo.GetConfig(ctx, businessID)
}

func (s *Service) UpdateConfig(ctx context.Context, input domain.UpdateLoyaltyConfigInput) error {
	return s.repo.UpsertConfig(ctx, input)
}

func (s *Service) ListTransactions(ctx context.Context, businessID, customerID int64) ([]domain.LoyaltyTransaction, error) {
	return s.repo.ListTransactions(ctx, businessID, customerID)
}

func (s *Service) GetBalance(ctx context.Context, businessID, customerID int64) (int32, error) {
	return s.repo.GetBalance(ctx, businessID, customerID)
}

func (s *Service) EarnPoints(ctx context.Context, input domain.EarnPointsInput) (int64, error) {
	currentBalance, err := s.repo.GetBalance(ctx, input.BusinessID, input.CustomerID)
	if err != nil {
		return 0, err
	}
	newBalance := currentBalance + input.Points
	note := "Points earned"
	return s.repo.CreateTransaction(ctx, input.BusinessID, input.CustomerID, input.OrderID, "earn", input.Points, newBalance, &note)
}

func (s *Service) RedeemPoints(ctx context.Context, input domain.RedeemPointsInput) (int64, error) {
	currentBalance, err := s.repo.GetBalance(ctx, input.BusinessID, input.CustomerID)
	if err != nil {
		return 0, err
	}
	if currentBalance < input.Points {
		return 0, domain.ErrInsufficientPoints
	}
	newBalance := currentBalance - input.Points
	note := "Points redeemed"
	return s.repo.CreateTransaction(ctx, input.BusinessID, input.CustomerID, nil, "redeem", -input.Points, newBalance, &note)
}
