package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Location, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, address, phone, is_active, created_at, updated_at
		 FROM pos_locations WHERE owner_id = ? ORDER BY name`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list locations", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	locations := make([]domain.Location, 0)
	for rows.Next() {
		l, scanErr := scanLocation(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.LocationInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO pos_locations (owner_id, name, address, phone, is_active) VALUES (?, ?, ?, ?, ?)`,
		input.OwnerID, input.Name, nullableString(input.Address), nullableString(input.Phone), input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create location", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.LocationUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE pos_locations SET name = ?, address = ?, phone = ?, is_active = ? WHERE id = ? AND owner_id = ?`,
		input.Name, nullableString(input.Address), nullableString(input.Phone), input.IsActive, input.ID, input.OwnerID,
	)
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

func scanLocation(rows *sql.Rows) (domain.Location, error) {
	var l domain.Location
	var address, phone sql.NullString
	if err := rows.Scan(&l.ID, &l.OwnerID, &l.Name, &address, &phone, &l.IsActive, &l.CreatedAt, &l.UpdatedAt); err != nil {
		return domain.Location{}, err
	}
	l.Address = address.String
	l.Phone = phone.String
	return l, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
