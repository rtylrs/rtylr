package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/templates/internal/domain"
	"github.com/voxire/onion/internal/modules/templates/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Template, error) {
	rows, err := r.queries.ListTemplates(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list templates", logger.Error(err))

		return nil, err
	}

	items := make([]domain.Template, 0, len(rows))

	for _, row := range rows {
		items = append(items, templateFromRow(row))
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, input domain.TemplateInput) (int64, error) {
	result, err := r.queries.CreateTemplate(ctx, sqlc.CreateTemplateParams{
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		Name:       input.Name,
		Type:       sqlc.CrmTemplatesType(input.Type),
		Subject:    sql.NullString{String: input.Subject, Valid: input.Subject != ""},
		Body:       input.Body,
		Variables:  json.RawMessage(input.Variables),
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create template", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TemplateUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	result, err := r.queries.UpdateTemplate(ctx, sqlc.UpdateTemplateParams{
		Name:       input.Name,
		Type:       sqlc.CrmTemplatesType(input.Type),
		Subject:    sql.NullString{String: input.Subject, Valid: input.Subject != ""},
		Body:       input.Body,
		Variables:  json.RawMessage(input.Variables),
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
		ID:         input.ID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update template", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, templateID int64) error {
	result, err := r.queries.DeleteTemplate(ctx, sqlc.DeleteTemplateParams{
		ID:         templateID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete template", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}

func templateFromRow(row sqlc.CrmTemplate) domain.Template {
	t := domain.Template{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Name:       row.Name,
		Type:       string(row.Type),
		Subject:    row.Subject.String,
		Body:       row.Body,
		IsActive:   row.IsActive.Bool,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}

	if row.Variables != nil {
		t.Variables = []byte(row.Variables)
	} else {
		t.Variables = []byte("[]")
	}

	return t
}
