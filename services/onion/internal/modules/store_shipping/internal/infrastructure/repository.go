package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/store_shipping/internal/domain"
	"github.com/voxire/onion/internal/modules/store_shipping/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.ShippingMethod, error) {
	rows, err := r.queries.ListShippingMethods(ctx, ownerID)
	if err != nil { logger.Ctx(ctx).Error("Failed to list shipping methods", logger.Error(err)); return nil, err }
	items := make([]domain.ShippingMethod, 0, len(rows))
	for _, row := range rows { items = append(items, fromRow(row)) }
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.ShippingMethodInput) (int64, error) {
	result, err := r.queries.CreateShippingMethod(ctx, sqlc.CreateShippingMethodParams{
		OwnerID: ownerID, Name: input.Name,
		Price: strconv.FormatFloat(input.Price, 'f', 2, 64),
		EstimatedDays: sql.NullInt32{Int32: int32(input.EstimatedDays), Valid: true},
		IsActive: sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil { logger.Ctx(ctx).Error("Failed to create shipping method", logger.Error(err)); return 0, err }
	id, _ := result.LastInsertId()
	return id, nil
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.ShippingMethodInput) error {
	err := r.queries.UpdateShippingMethod(ctx, sqlc.UpdateShippingMethodParams{
		Name: input.Name, Price: strconv.FormatFloat(input.Price, 'f', 2, 64),
		EstimatedDays: sql.NullInt32{Int32: int32(input.EstimatedDays), Valid: true},
		IsActive: sql.NullBool{Bool: input.IsActive, Valid: true},
		OwnerID: ownerID, ID: id,
	})
	if err != nil { logger.Ctx(ctx).Error("Failed to update shipping method", logger.Error(err)); return err }
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteShippingMethod(ctx, sqlc.DeleteShippingMethodParams{OwnerID: ownerID, ID: id})
	if err != nil { logger.Ctx(ctx).Error("Failed to delete shipping method", logger.Error(err)); return err }
	return nil
}

func fromRow(row sqlc.StoreShippingMethod) domain.ShippingMethod {
	s := domain.ShippingMethod{ID: row.ID, OwnerID: row.OwnerID, Name: row.Name}
	s.Price, _ = strconv.ParseFloat(row.Price, 64)
	if row.EstimatedDays.Valid { s.EstimatedDays = int(row.EstimatedDays.Int32) }
	if row.IsActive.Valid { s.IsActive = row.IsActive.Bool }
	if row.CreatedAt.Valid { s.CreatedAt = row.CreatedAt.Time }
	return s
}
