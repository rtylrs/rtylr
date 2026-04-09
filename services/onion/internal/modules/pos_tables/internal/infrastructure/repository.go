package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Table, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, location_id, label, seats, status, created_at, updated_at
		 FROM pos_tables WHERE owner_id = ? ORDER BY label`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list tables", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	tables := make([]domain.Table, 0)
	for rows.Next() {
		t, scanErr := scanTable(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.TableInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO pos_tables (owner_id, location_id, label, seats, status) VALUES (?, ?, ?, ?, 'available')`,
		input.OwnerID, input.LocationID, input.Label, input.Seats,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create table", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TableUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE pos_tables SET location_id = ?, label = ?, seats = ?, status = ? WHERE id = ? AND owner_id = ?`,
		input.LocationID, input.Label, input.Seats, input.Status, input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update table", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTableNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM pos_tables WHERE id = ? AND owner_id = ?", id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete table", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTableNotFound
	}
	return nil
}

func scanTable(rows *sql.Rows) (domain.Table, error) {
	var t domain.Table
	var locID sql.NullInt64
	var status sql.NullString
	if err := rows.Scan(&t.ID, &t.OwnerID, &locID, &t.Label, &t.Seats, &status, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return domain.Table{}, err
	}
	if locID.Valid {
		t.LocationID = &locID.Int64
	}
	t.Status = status.String
	return t, nil
}
