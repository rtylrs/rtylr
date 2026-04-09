package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/segments/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Segment, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, description, filters, customer_count, created_at, updated_at
		 FROM crm_segments WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list segments", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	segments := make([]domain.Segment, 0)
	for rows.Next() {
		var seg domain.Segment
		var desc sql.NullString
		var filters []byte
		if scanErr := rows.Scan(&seg.ID, &seg.OwnerID, &seg.Name, &desc, &filters, &seg.CustomerCount, &seg.CreatedAt, &seg.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		seg.Description = desc.String
		seg.Filters = filters
		segments = append(segments, seg)
	}
	return segments, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, ownerID, id int64) (domain.Segment, error) {
	var seg domain.Segment
	var desc sql.NullString
	var filters []byte
	err := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, name, description, filters, customer_count, created_at, updated_at
		 FROM crm_segments WHERE id = ? AND owner_id = ?`, id, ownerID,
	).Scan(&seg.ID, &seg.OwnerID, &seg.Name, &desc, &filters, &seg.CustomerCount, &seg.CreatedAt, &seg.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}
		return domain.Segment{}, err
	}
	seg.Description = desc.String
	seg.Filters = filters
	return seg, nil
}

func (r *Repository) Create(ctx context.Context, input domain.SegmentInput) (int64, error) {
	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_segments (owner_id, name, description, filters) VALUES (?, ?, ?, ?)`,
		input.OwnerID, input.Name, nullableString(input.Description), filters,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create segment", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.SegmentUpdateInput) error {
	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}
	result, err := r.pool.ExecContext(ctx,
		`UPDATE crm_segments SET name = ?, description = ?, filters = ? WHERE id = ? AND owner_id = ?`,
		input.Name, nullableString(input.Description), filters, input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update segment", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrSegmentNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM crm_segments WHERE id = ? AND owner_id = ?", id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete segment", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrSegmentNotFound
	}
	return nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
