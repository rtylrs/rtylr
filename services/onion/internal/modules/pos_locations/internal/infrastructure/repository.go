package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Location, error) {
	rows, err := r.queries.ListLocationsByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list locations", logger.Error(err))
		return nil, err
	}

	locations := make([]domain.Location, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, domain.Location{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Address:    row.Address.String,
			Phone:      row.Phone.String,
			IsActive:   row.IsActive.Bool,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.CreatedAt.Time,
		})
	}
	return locations, nil
}

func (r *Repository) Create(ctx context.Context, input domain.LocationInput) (int64, error) {
	result, err := r.queries.CreateLocation(ctx, sqlc.CreateLocationParams{
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		Name:       input.Name,
		Address:    sql.NullString{String: input.Address, Valid: input.Address != ""},
		Phone:      sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create location", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.LocationUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	isActive := sql.NullBool{}
	if input.IsActive != nil {
		isActive = sql.NullBool{Bool: *input.IsActive, Valid: true}
	}
	result, err := r.queries.UpdateLocation(ctx, sqlc.UpdateLocationParams{
		Name:       input.Name,
		Address:    sql.NullString{String: input.Address, Valid: input.Address != ""},
		Phone:      sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		IsActive:   isActive,
		ID:         input.ID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update location", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrLocationNotFound
	}
	return nil
}
