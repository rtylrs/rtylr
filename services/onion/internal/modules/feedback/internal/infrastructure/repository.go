package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64, feedbackType string, page, pageSize int) ([]domain.Feedback, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	countQuery := "SELECT COUNT(*) FROM crm_feedback WHERE owner_id = ?"
	args := []interface{}{ownerID}

	if feedbackType != "" {
		countQuery += " AND type = ?"
		args = append(args, feedbackType)
	}

	var total int

	if err := r.pool.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count feedback", logger.Error(err))

		return nil, 0, err
	}

	listQuery := `SELECT id, owner_id, customer_id, type, score, comment, source, created_at
		 FROM crm_feedback WHERE owner_id = ?`
	listArgs := []interface{}{ownerID}

	if feedbackType != "" {
		listQuery += " AND type = ?"
		listArgs = append(listArgs, feedbackType)
	}

	listQuery += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	listArgs = append(listArgs, pageSize, offset)

	rows, err := r.pool.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feedback", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Feedback, 0)

	for rows.Next() {
		var f domain.Feedback
		var comment, source sql.NullString

		if scanErr := rows.Scan(&f.ID, &f.OwnerID, &f.CustomerID, &f.Type, &f.Score, &comment, &source, &f.CreatedAt); scanErr != nil {
			return nil, 0, scanErr
		}

		f.Comment = comment.String
		f.Source = source.String
		items = append(items, f)
	}

	return items, total, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.FeedbackInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_feedback (owner_id, customer_id, type, score, comment, source)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.CustomerID, input.Type, input.Score,
		nullableString(input.Comment), nullableString(input.Source),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create feedback", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
