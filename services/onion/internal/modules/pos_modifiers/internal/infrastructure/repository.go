package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Modifier, error) {
	rows, err := r.queries.ListModifiersByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list modifiers", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Modifier, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Modifier{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			PriceDelta: decimalToFloat(row.PriceDelta),
			GroupName:  row.GroupName.String,
			IsRequired: row.IsRequired.Bool,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		})
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.ModifierInput) (int64, error) {
	result, err := r.queries.CreateModifier(ctx, sqlc.CreateModifierParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Name:       input.Name,
		PriceDelta: floatToDecimal(input.PriceDelta),
		GroupName:  sql.NullString{String: input.GroupName, Valid: input.GroupName != ""},
		IsRequired: sql.NullBool{Bool: input.IsRequired, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create modifier", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.ModifierInput) error {
	result, err := r.queries.UpdateModifier(ctx, sqlc.UpdateModifierParams{
		Name:       input.Name,
		PriceDelta: floatToDecimal(input.PriceDelta),
		GroupName:  sql.NullString{String: input.GroupName, Valid: input.GroupName != ""},
		IsRequired: sql.NullBool{Bool: input.IsRequired, Valid: true},
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update modifier", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrModifierNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	result, err := r.queries.DeleteModifier(ctx, sqlc.DeleteModifierParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete modifier", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrModifierNotFound
	}
	return nil
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse modifier decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

var floatToDecimal = sqlutil.FloatToDecimal
