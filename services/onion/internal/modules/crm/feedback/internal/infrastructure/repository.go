package infrastructure

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/crm/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/feedback/internal/infrastructure/persistence/sqlc"
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
// CustomerExistsInBusiness reports whether the customer id belongs to the
// business. crm_feedback has no FK, so the write path must verify ownership
// itself or its scoped list will disclose other tenants' customer names.
func (r *Repository) CustomerExistsInBusiness(ctx context.Context, businessID, customerID int64) bool {
	var exists bool
	if err := r.pool.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM crm_customers WHERE id = ? AND business_id = ?)",
		customerID, businessID,
	).Scan(&exists); err != nil {
		logger.Ctx(ctx).Error("feedback customer ownership lookup failed", logger.Error(err))
		return false
	}
	return exists
}

func (r *Repository) List(ctx context.Context, businessID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	whereConds := feedbackWhereConds("", businessID, filters)

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

	// cross-schema read from crm_customers — resolves a display name for each feedback row
	listQuery, listArgs, err := squirrel.Select(
		"f.id", "f.owner_id", "f.business_id", "f.customer_id",
		"COALESCE(TRIM(CONCAT_WS(' ', c.first_name, c.last_name)), '')", "f.type", "f.score", "f.comment", "f.source", "f.created_at",
	).
		From("crm_feedback f").
		LeftJoin("crm_customers c ON c.id = f.customer_id").
		Where(feedbackWhereConds("f.", businessID, filters)).
		OrderBy("f.created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
			customerName  string
			typeStr       string
			score         sql.NullInt32
			comment       sql.NullString
			source        sql.NullString
			createdAt     sql.NullTime
		)

		if scanErr := rows.Scan(&id, &ownerRowID, &businessRowID, &customerID, &customerName, &typeStr, &score, &comment, &source, &createdAt); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan feedback row", logger.Error(scanErr))

			return nil, 0, scanErr
		}

		f := domain.Feedback{
			ID:           id,
			OwnerID:      ownerRowID,
			BusinessID:   businessRowID,
			CustomerID:   customerID,
			CustomerName: customerName,
			Type:         typeStr,
			Score:        int(score.Int32),
			Comment:      comment.String,
			Source:       source.String,
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

// feedbackWhereConds builds the WHERE clause shared by the count and list
// queries. prefix is the crm_feedback table alias (with trailing dot) to
// disambiguate columns once the list query joins crm_customers, which also
// has a business_id column; the count query passes "" since it queries
// crm_feedback unaliased and unjoined.
func feedbackWhereConds(prefix string, businessID int64, filters domain.ListFilters) squirrel.And {
	conds := squirrel.And{squirrel.Eq{prefix + "business_id": businessID}}

	if filters.Type != "" {
		conds = append(conds, squirrel.Eq{prefix + "type": filters.Type})
	}

	if filters.Source != "" {
		conds = append(conds, squirrel.Eq{prefix + "source": filters.Source})
	}

	if filters.MinScoreSet {
		conds = append(conds, squirrel.GtOrEq{prefix + "score": filters.MinScore})
	}

	if filters.MaxScoreSet {
		conds = append(conds, squirrel.LtOrEq{prefix + "score": filters.MaxScore})
	}

	return conds
}

func (r *Repository) Create(ctx context.Context, input domain.FeedbackInput) (int64, error) {
	result, err := r.queries.CreateFeedback(ctx, sqlc.CreateFeedbackParams{
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		CustomerID: input.CustomerID,
		Type:       sqlc.CrmFeedbackType(input.Type),
		Score:      sql.NullInt32{Int32: int32(input.Score), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Comment:    sql.NullString{String: input.Comment, Valid: input.Comment != ""},
		Source:     sql.NullString{String: input.Source, Valid: input.Source != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create feedback", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}
