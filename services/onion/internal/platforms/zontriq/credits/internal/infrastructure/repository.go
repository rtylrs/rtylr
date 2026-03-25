package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/domain"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/credits/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetCreditsByUserID(ctx context.Context, userID int32) (domain.Credits, error) {
	result, err := db.NewReadExecutor[domain.Credits, sqlc.ZontriqCredit](db.MainDatabase).Fetch(ctx, r.queries.GetCreditsByUserID, userID)
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return domain.Credits{}, domain.ErrCreditsNotFound
		}

		return domain.Credits{}, fmt.Errorf("failed to get credits: %w", err)
	}

	return result, nil
}

func (r *Repository) CreateCredits(ctx context.Context, userID int32) (domain.Credits, error) {
	result, err := db.NewWriteExecutor[domain.Credits, sqlc.ZontriqCredit](db.MainDatabase).Insert(ctx, r.queries.CreateCredits, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create credits", logger.Error(err))

		return domain.Credits{}, fmt.Errorf("failed to create credits: %w", err)
	}

	return result, nil
}

func (r *Repository) AddPurchaseCredits(ctx context.Context, userID, amount int32) (domain.Credits, error) {
	params := sqlc.AddPurchaseCreditsParams{
		ZontriqUserID: userID,
		Balance:       amount,
	}

	result, err := db.NewWriteExecutor[domain.Credits, sqlc.ZontriqCredit](db.MainDatabase).Insert(ctx, r.queries.AddPurchaseCredits, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to add purchase credits", logger.Error(err))

		return domain.Credits{}, fmt.Errorf("failed to add purchase credits: %w", err)
	}

	return result, nil
}

func (r *Repository) SpendCredits(ctx context.Context, userID, amount int32) (domain.Credits, error) {
	params := sqlc.SpendCreditsParams{
		ZontriqUserID: userID,
		Balance:       amount,
	}

	result, err := db.NewWriteExecutor[domain.Credits, sqlc.ZontriqCredit](db.MainDatabase).Insert(ctx, r.queries.SpendCredits, params)
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return domain.Credits{}, domain.ErrInsufficientCredits
		}

		logger.Ctx(ctx).Error("Failed to spend credits", logger.Error(err))

		return domain.Credits{}, fmt.Errorf("failed to spend credits: %w", err)
	}

	return result, nil
}

func (r *Repository) CreateCreditTransaction(ctx context.Context, userID, amount int32, txType, description, referenceType string, referenceID int32) (domain.CreditTransaction, error) {
	params := sqlc.CreateCreditTransactionParams{
		ZontriqUserID: userID,
		Amount:        amount,
		Type:          txType,
		Description:   pgtype.Text{String: description, Valid: description != ""},
		ReferenceType: pgtype.Text{String: referenceType, Valid: referenceType != ""},
		ReferenceID:   pgtype.Int4{Int32: referenceID, Valid: referenceID > 0},
	}

	result, err := db.NewWriteExecutor[domain.CreditTransaction, sqlc.ZontriqCreditTransaction](db.MainDatabase).Insert(ctx, r.queries.CreateCreditTransaction, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create credit transaction", logger.Error(err))

		return domain.CreditTransaction{}, fmt.Errorf("failed to create credit transaction: %w", err)
	}

	return result, nil
}

func (r *Repository) ListCreditTransactionsByUserID(ctx context.Context, userID, limit, offset int32) ([]domain.CreditTransaction, error) {
	params := sqlc.ListCreditTransactionsByUserIDParams{
		ZontriqUserID: userID,
		Limit:         limit,
		Offset:        offset,
	}

	results, err := db.NewReadExecutor[domain.CreditTransaction, sqlc.ZontriqCreditTransaction](db.MainDatabase).FetchOrdered(ctx, r.queries.ListCreditTransactionsByUserID, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list credit transactions", logger.Error(err))

		return nil, fmt.Errorf("failed to list credit transactions: %w", err)
	}

	return results, nil
}

func (r *Repository) CountCreditTransactionsByUserID(ctx context.Context, userID int32) (int64, error) {
	count, err := r.queries.CountCreditTransactionsByUserID(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count credit transactions", logger.Error(err))

		return 0, fmt.Errorf("failed to count credit transactions: %w", err)
	}

	return count, nil
}

func (r *Repository) GetZontriqUserID(ctx context.Context, userID int64) (int32, error) {
	zontriqUserID, err := r.queries.GetZontriqUserIDByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return zontriqUserID, nil
}
