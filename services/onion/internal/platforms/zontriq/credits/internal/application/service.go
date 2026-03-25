package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) GetZontriqUserID(ctx context.Context, userID int64) (int32, error) {
	zontriqUserID, err := s.repo.GetZontriqUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return zontriqUserID, nil
}

func (s *Service) GetCredits(ctx context.Context, userID int32) (domain.Credits, error) {
	credits, err := s.repo.GetCreditsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrCreditsNotFound) {
			credits, err = s.repo.CreateCredits(ctx, userID)
			if err != nil {
				return domain.Credits{}, fmt.Errorf("failed to create credits: %w", err)
			}

			return credits, nil
		}

		return domain.Credits{}, fmt.Errorf("failed to get credits: %w", err)
	}

	return credits, nil
}

type PurchaseCreditsResult struct {
	Credits     domain.Credits
	Transaction domain.CreditTransaction
}

func (s *Service) PurchaseCredits(ctx context.Context, userID, amount int32, paymentMethod string) (PurchaseCreditsResult, error) {
	_, err := s.GetCredits(ctx, userID)
	if err != nil {
		return PurchaseCreditsResult{}, fmt.Errorf("failed to ensure credits exist: %w", err)
	}

	credits, err := s.repo.AddPurchaseCredits(ctx, userID, amount)
	if err != nil {
		return PurchaseCreditsResult{}, fmt.Errorf("failed to purchase credits: %w", err)
	}

	description := fmt.Sprintf("Purchased %d credits via %s", amount, paymentMethod)

	transaction, err := s.repo.CreateCreditTransaction(ctx, userID, amount, "purchase", description, "payment", 0)
	if err != nil {
		return PurchaseCreditsResult{}, fmt.Errorf("failed to create purchase transaction: %w", err)
	}

	return PurchaseCreditsResult{
		Credits:     credits,
		Transaction: transaction,
	}, nil
}

type SpendCreditsResult struct {
	Credits     domain.Credits
	Transaction domain.CreditTransaction
}

func (s *Service) SpendCredits(ctx context.Context, userID, amount int32, description, referenceType string, referenceID int32) (SpendCreditsResult, error) {
	credits, err := s.repo.SpendCredits(ctx, userID, amount)
	if err != nil {
		return SpendCreditsResult{}, fmt.Errorf("failed to spend credits: %w", err)
	}

	transaction, err := s.repo.CreateCreditTransaction(ctx, userID, amount, "spend", description, referenceType, referenceID)
	if err != nil {
		return SpendCreditsResult{}, fmt.Errorf("failed to create spend transaction: %w", err)
	}

	return SpendCreditsResult{
		Credits:     credits,
		Transaction: transaction,
	}, nil
}

type ListTransactionsParams struct {
	UserID   int32
	Page     int32
	PageSize int32
}

type ListTransactionsResult struct {
	Transactions []domain.CreditTransaction
	TotalCount   int64
}

func (s *Service) ListTransactions(ctx context.Context, params ListTransactionsParams) (ListTransactionsResult, error) {
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	if params.Page <= 0 {
		params.Page = 1
	}

	offset := (params.Page - 1) * params.PageSize

	transactions, err := s.repo.ListCreditTransactionsByUserID(ctx, params.UserID, params.PageSize, offset)
	if err != nil {
		return ListTransactionsResult{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	totalCount, err := s.repo.CountCreditTransactionsByUserID(ctx, params.UserID)
	if err != nil {
		return ListTransactionsResult{}, fmt.Errorf("failed to count transactions: %w", err)
	}

	return ListTransactionsResult{
		Transactions: transactions,
		TotalCount:   totalCount,
	}, nil
}
