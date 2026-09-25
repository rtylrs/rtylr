package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/booking_resources/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_resources/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Resource, error) {
	rows, err := r.queries.ListResources(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list resources", logger.Error(err))
		return nil, err
	}
	items := make([]domain.Resource, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.ResourceInput) (int64, error) {
	result, err := r.queries.CreateResource(ctx, sqlc.CreateResourceParams{
		OwnerID: businessID, BusinessID: businessID, Name: input.Name,
		Type:     sqlc.NullBookingResourcesType{BookingResourcesType: sqlc.BookingResourcesType(input.Type), Valid: input.Type != ""},
		Capacity: sql.NullInt32{Int32: int32(input.Capacity), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IsActive: sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create resource", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to retrieve last insert ID for resource", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.ResourceInput) error {
	err := r.queries.UpdateResource(ctx, sqlc.UpdateResourceParams{
		Name:       input.Name,
		Type:       sqlc.NullBookingResourcesType{BookingResourcesType: sqlc.BookingResourcesType(input.Type), Valid: input.Type != ""},
		Capacity:   sql.NullInt32{Int32: int32(input.Capacity), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID: businessID, ID: id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update resource", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteResource(ctx, sqlc.DeleteResourceParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete resource", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.BookingResource) domain.Resource {
	res := domain.Resource{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name}
	if row.Type.Valid {
		res.Type = string(row.Type.BookingResourcesType)
	}
	if row.Capacity.Valid {
		res.Capacity = int(row.Capacity.Int32)
	}
	if row.IsActive.Valid {
		res.IsActive = row.IsActive.Bool
	}
	if row.CreatedAt.Valid {
		res.CreatedAt = row.CreatedAt.Time
	}
	return res
}
