package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/support/canned/internal/domain"
	"github.com/voxire/onion/internal/modules/support/canned/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.CannedResponse, error) {
	rows, err := r.queries.ListCannedResponses(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list canned responses", logger.Error(err))
		return nil, err
	}
	items := make([]domain.CannedResponse, 0, len(rows))
	for _, row := range rows {
		c := domain.CannedResponse{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Body:       row.Body,
		}
		if row.Category.Valid {
			c.Category = row.Category.String
		}
		if row.Shortcut.Valid {
			c.Shortcut = row.Shortcut.String
		}
		if row.CreatedAt.Valid {
			c.CreatedAt = row.CreatedAt.Time
		}
		items = append(items, c)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, input domain.CannedResponseInput) (int64, error) {
	result, err := r.queries.CreateCannedResponse(ctx, sqlc.CreateCannedResponseParams{
		BusinessID: input.BusinessID,
		Name:       input.Name,
		Body:       input.Body,
		Category:   sql.NullString{String: input.Category, Valid: input.Category != ""},
		Shortcut:   sql.NullString{String: input.Shortcut, Valid: input.Shortcut != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create canned response", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.CannedResponseInput) error {
	err := r.queries.UpdateCannedResponse(ctx, sqlc.UpdateCannedResponseParams{
		Name:       input.Name,
		Body:       input.Body,
		Category:   sql.NullString{String: input.Category, Valid: input.Category != ""},
		Shortcut:   sql.NullString{String: input.Shortcut, Valid: input.Shortcut != ""},
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update canned response", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteCannedResponse(ctx, sqlc.DeleteCannedResponseParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete canned response", logger.Error(err))
		return err
	}
	return nil
}
