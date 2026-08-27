package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/store_shipping/internal/domain"
	"github.com/voxire/onion/internal/modules/store_shipping/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.ShippingMethod, error) {
	rows, err := r.queries.ListShippingMethods(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list shipping methods", logger.Error(err))
		return nil, err
	}
	items := make([]domain.ShippingMethod, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(ctx, row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.ShippingMethodInput) (int64, error) {
	result, err := r.queries.CreateShippingMethod(ctx, sqlc.CreateShippingMethodParams{
		OwnerID: ownerID, BusinessID: businessID, Name: input.Name,
		Price:         strconv.FormatFloat(input.Price, 'f', 2, 64),
		EstimatedDays: sql.NullInt32{Int32: int32(input.EstimatedDays), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create shipping method", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.ShippingMethodInput) error {
	err := r.queries.UpdateShippingMethod(ctx, sqlc.UpdateShippingMethodParams{
		Name: input.Name, Price: strconv.FormatFloat(input.Price, 'f', 2, 64),
		EstimatedDays: sql.NullInt32{Int32: int32(input.EstimatedDays), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID:    businessID, ID: id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update shipping method", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteShippingMethod(ctx, sqlc.DeleteShippingMethodParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete shipping method", logger.Error(err))
		return err
	}
	return nil
}

func parseMonetary(ctx context.Context, s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to parse monetary value", logger.String("value", s), logger.Error(err))
		return 0
	}
	return v
}

func fromRow(ctx context.Context, row sqlc.StoreShippingMethod) domain.ShippingMethod {
	s := domain.ShippingMethod{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name}
	s.Price = parseMonetary(ctx, row.Price)
	if row.EstimatedDays.Valid {
		s.EstimatedDays = int(row.EstimatedDays.Int32)
	}
	if row.IsActive.Valid {
		s.IsActive = row.IsActive.Bool
	}
	if row.CreatedAt.Valid {
		s.CreatedAt = row.CreatedAt.Time
	}
	return s
}
