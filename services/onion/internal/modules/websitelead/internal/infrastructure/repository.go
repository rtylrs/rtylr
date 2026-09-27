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

// PendingSubscriber is a row still missing from the Resend audience.
type PendingSubscriber struct {
	ID    int64
	Email string
}

// ListUnsyncedSubscribers returns rows never mirrored into the Resend
// audience that have failed fewer than maxAttempts times, oldest first.
func (r *Repository) ListUnsyncedSubscribers(ctx context.Context, maxAttempts, limit int) ([]PendingSubscriber, error) {
	rows, err := r.queries.ListUnsyncedNewsletterSubscribers(ctx, sqlc.ListUnsyncedNewsletterSubscribersParams{
		ResendSyncAttempts: int32(maxAttempts),
		Limit:              int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]PendingSubscriber, 0, len(rows))
	for _, row := range rows {
		out = append(out, PendingSubscriber{ID: row.ID, Email: row.Email})
	}
	return out, nil
}

func (r *Repository) MarkSubscriberSynced(ctx context.Context, id int64, contactID string) error {
	return r.queries.MarkNewsletterSubscriberSynced(ctx, sqlc.MarkNewsletterSubscriberSyncedParams{
		ResendContactID: sql.NullString{String: contactID, Valid: contactID != ""},
		ID:              id,
	})
}

func (r *Repository) MarkSubscriberSyncFailed(ctx context.Context, id int64, reason string) error {
	if len(reason) > 255 {
		reason = reason[:255]
	}
	return r.queries.MarkNewsletterSubscriberSyncFailed(ctx, sqlc.MarkNewsletterSubscriberSyncFailedParams{
		ResendSyncError: sql.NullString{String: reason, Valid: reason != ""},
		ID:              id,
	})
}
