package infrastructure

import (
	"context"
	"database/sql"
	"strings"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/feedback/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool:    db,
		queries: sqlc.New(db),
	}
}

// List returns feedback entries for an owner, applying optional filters by
// type, source, and score range. Always filters by owner_id.
func (r *Repository) List(ctx context.Context, ownerID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "owner_id = ?")
	args = append(args, ownerID)

	if filters.Type != "" {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, filters.Type)
	}

	if filters.Source != "" {
		whereClauses = append(whereClauses, "source = ?")
		args = append(args, filters.Source)
	}

	if filters.MinScoreSet {
		whereClauses = append(whereClauses, "score >= ?")
		args = append(args, filters.MinScore)
	}

	if filters.MaxScoreSet {
		whereClauses = append(whereClauses, "score <= ?")
		args = append(args, filters.MaxScore)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int64

	if err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM crm_feedback WHERE "+whereSQL, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count feedback",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
		)

		return nil, 0, err
	}

	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, pageSize, offset)

	rows, err := r.pool.QueryContext(ctx,
		"SELECT id, owner_id, customer_id, type, score, comment, source, created_at "+
			"FROM crm_feedback WHERE "+whereSQL+
			" ORDER BY created_at DESC LIMIT ? OFFSET ?", listArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feedback",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
		)

		return nil, 0, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("Failed to close feedback rows", logger.Error(closeErr))
		}
	}()

	items := make([]domain.Feedback, 0, pageSize)

	for rows.Next() {
		var (
			id         int64
			ownerRowID int64
			customerID int64
			typeStr    string
			score      sql.NullInt32
			comment    sql.NullString
			source     sql.NullString
			createdAt  sql.NullTime
		)

		if scanErr := rows.Scan(&id, &ownerRowID, &customerID, &typeStr, &score, &comment, &source, &createdAt); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan feedback row", logger.Error(scanErr))

			return nil, 0, scanErr
		}

		f := domain.Feedback{
			ID:         id,
			OwnerID:    ownerRowID,
			CustomerID: customerID,
			Type:       typeStr,
			Score:      int(score.Int32),
			Comment:    comment.String,
			Source:     source.String,
		}

		if createdAt.Valid {
			f.CreatedAt = createdAt.Time
		}

		items = append(items, f)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, rowsErr
	}

	return items, int(total), nil
}

func (r *Repository) Create(ctx context.Context, input domain.FeedbackInput) (int64, error) {
	result, err := r.queries.CreateFeedback(ctx, sqlc.CreateFeedbackParams{
		OwnerID:    input.OwnerID,
		CustomerID: input.CustomerID,
		Type:       sqlc.CrmFeedbackType(input.Type),
		Score:      sql.NullInt32{Int32: int32(input.Score), Valid: true},
		Comment:    sql.NullString{String: input.Comment, Valid: input.Comment != ""},
		Source:     sql.NullString{String: input.Source, Valid: input.Source != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create feedback", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}
