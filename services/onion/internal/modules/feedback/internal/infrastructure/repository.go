package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/feedback/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
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

	var total int64
	var err error

	if feedbackType != "" {
		total, err = r.queries.CountFeedbackByType(ctx, sqlc.CountFeedbackByTypeParams{
			OwnerID: ownerID,
			Type:    sqlc.CrmFeedbackType(feedbackType),
		})
	} else {
		total, err = r.queries.CountFeedback(ctx, ownerID)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to count feedback", logger.Error(err))

		return nil, 0, err
	}

	var rows []sqlc.CrmFeedback

	if feedbackType != "" {
		rows, err = r.queries.ListFeedbackByType(ctx, sqlc.ListFeedbackByTypeParams{
			OwnerID: ownerID,
			Type:    sqlc.CrmFeedbackType(feedbackType),
			Limit:   int32(pageSize),
			Offset:  int32(offset),
		})
	} else {
		rows, err = r.queries.ListFeedback(ctx, sqlc.ListFeedbackParams{
			OwnerID: ownerID,
			Limit:   int32(pageSize),
			Offset:  int32(offset),
		})
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feedback", logger.Error(err))

		return nil, 0, err
	}

	items := make([]domain.Feedback, 0, len(rows))

	for _, row := range rows {
		items = append(items, domain.Feedback{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			CustomerID: row.CustomerID,
			Type:       string(row.Type),
			Score:      int(row.Score.Int32),
			Comment:    row.Comment.String,
			Source:     row.Source.String,
			CreatedAt:  row.CreatedAt.Time,
		})
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
