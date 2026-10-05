package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/erp/price_lists/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/price_lists/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListPriceLists(ctx context.Context, businessID int64) ([]domain.PriceList, error) {
	rows, err := r.queries.ListPriceLists(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list price lists",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)

		return nil, err
	}

	result := make([]domain.PriceList, 0, len(rows))
	for _, row := range rows {
		pl := domain.PriceList{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Currency:   row.Currency,
			IsDefault:  row.IsDefault,
		}

		if row.CreatedAt.Valid {
			pl.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			pl.UpdatedAt = row.UpdatedAt.Time
		}

		result = append(result, pl)
	}

	return result, nil
}

func (r *Repository) CreatePriceList(ctx context.Context, businessID int64, input domain.PriceListInput) (int64, error) {
	res, err := r.queries.CreatePriceList(ctx, sqlc.CreatePriceListParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       input.Name,
		Currency:   input.Currency,
		IsDefault:  input.IsDefault,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create price list",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)

		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read inserted price list id",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)

		return 0, err
	}

	return id, nil
}
