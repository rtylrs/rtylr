package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/modules/websitelead/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db), db: db}
}

func (r *Repository) Create(ctx context.Context, input domain.LeadInput) (int64, error) {
	result, err := r.queries.CreateWebsiteLead(ctx, sqlc.CreateWebsiteLeadParams{
		Name:    input.Name,
		Email:   input.Email,
		Phone:   sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		Company: sql.NullString{String: input.Company, Valid: input.Company != ""},
		Service: sql.NullString{String: input.Service, Valid: input.Service != ""},
		Project: sql.NullString{String: input.Project, Valid: input.Project != ""},
		Budget:  sql.NullString{String: input.Budget, Valid: input.Budget != ""},
		Source:  sql.NullString{String: input.Source, Valid: input.Source != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert website lead", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) CreateSubscriber(ctx context.Context, input domain.SubscriberInput) (int64, error) {
	result, err := r.queries.CreateNewsletterSubscriber(ctx, sqlc.CreateNewsletterSubscriberParams{
		Email:        input.Email,
		Source:       input.Source,
		PostSlug:     sql.NullString{String: input.PostSlug, Valid: input.PostSlug != ""},
		PostTitle:    sql.NullString{String: input.PostTitle, Valid: input.PostTitle != ""},
		PostCategory: sql.NullString{String: input.PostCategory, Valid: input.PostCategory != ""},
		Language:     input.Language,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert newsletter subscriber", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}
