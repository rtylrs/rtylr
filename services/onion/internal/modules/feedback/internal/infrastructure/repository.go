package infrastructure

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"

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
//
// dynamic-query: squirrel
func (r *Repository) List(ctx context.Context, businessID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	whereConds := squirrel.And{squirrel.Eq{"business_id": businessID}}

	if filters.Type != "" {
		whereConds = append(whereConds, squirrel.Eq{"type": filters.Type})
	}

	if filters.Source != "" {
		whereConds = append(whereConds, squirrel.Eq{"source": filters.Source})
	}

	if filters.MinScoreSet {
		whereConds = append(whereConds, squirrel.GtOrEq{"score": filters.MinScore})
	}

	if filters.MaxScoreSet {
		whereConds = append(whereConds, squirrel.LtOrEq{"score": filters.MaxScore})
	}

	countQuery, countArgs, err := squirrel.Select("COUNT(*)").
		From("crm_feedback").
		Where(whereConds).
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build count feedback query",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)

		return nil, 0, err
	}

	var total int64

	if err := r.pool.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count feedback",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)

		return nil, 0, err
	}

	listQuery, listArgs, err := squirrel.Select(
		"id", "owner_id", "business_id", "customer_id", "type", "score", "comment", "source", "created_at",
	).
		From("crm_feedback").
		Where(whereConds).
		OrderBy("created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list feedback query",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)

		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feedback",
			logger.Error(err),
			logger.Any("business_id", businessID),
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
			id            int64
			ownerRowID    int64
			businessRowID int64
			customerID    int64
			typeStr       string
			score         sql.NullInt32
			comment       sql.NullString
			source        sql.NullString
			createdAt     sql.NullTime
		)

		if scanErr := rows.Scan(&id, &ownerRowID, &businessRowID, &customerID, &typeStr, &score, &comment, &source, &createdAt); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan feedback row", logger.Error(scanErr))

			return nil, 0, scanErr
		}

		f := domain.Feedback{
			ID:         id,
			OwnerID:    ownerRowID,
			BusinessID: businessRowID,
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
		BusinessID: input.BusinessID,
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
