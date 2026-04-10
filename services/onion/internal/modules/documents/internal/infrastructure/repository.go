package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Document, error) {
	query := `
		SELECT id, owner_id, staff_id, name, type, file_url, expiry_date, created_at
		FROM hr_documents
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list documents", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.Document
	for rows.Next() {
		var item domain.Document
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.DocumentInput) (int64, error) {
	query := `
		INSERT INTO hr_documents (owner_id, staff_id, name, type, file_url, expiry_date)
		VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		sql.NullInt64{Int64: input.StaffID, Valid: input.StaffID != 0},
		input.Name,
		sql.NullString{String: input.Type, Valid: input.Type != ""},
		sql.NullString{String: input.FileURL, Valid: input.FileURL != ""},
		sql.NullString{String: input.ExpiryDate, Valid: input.ExpiryDate != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create document", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM hr_documents WHERE id = ? AND owner_id = ?`, id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete document", logger.Error(err))
		return err
	}
	return nil
}
