package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_combos/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Combo, error) {
	rows, err := r.queries.ListCombosByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list combos", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Combo, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Combo{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			BusinessID:  row.BusinessID,
			Name:        row.Name,
			Description: row.Description.String,
			Price:       decimalToFloat(row.Price),
			Items:       string(row.Items),
			IsActive:    row.IsActive.Bool,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.ComboInput) (int64, error) {
	result, err := r.queries.CreateCombo(ctx, sqlc.CreateComboParams{
		OwnerID:     ownerID,
		BusinessID:  businessID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Price:       floatToDecimal(input.Price),
		Items:       itemsToJSON(input.Items),
		IsActive:    sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create combo", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.ComboInput) error {
	result, err := r.queries.UpdateCombo(ctx, sqlc.UpdateComboParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Price:       floatToDecimal(input.Price),
		Items:       itemsToJSON(input.Items),
		IsActive:    sql.NullBool{Bool: input.IsActive, Valid: true},
		ID:          id,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update combo", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrComboNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	result, err := r.queries.DeleteCombo(ctx, sqlc.DeleteComboParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete combo", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrComboNotFound
	}
	return nil
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse combo decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

var floatToDecimal = sqlutil.FloatToDecimal

func itemsToJSON(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("[]")
	}
	return json.RawMessage(s)
}
