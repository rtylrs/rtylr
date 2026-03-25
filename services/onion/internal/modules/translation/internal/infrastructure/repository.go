package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/voxire/onion/internal/modules/translation/internal/domain"
	"github.com/voxire/onion/internal/modules/translation/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
	}
}

func (r *Repository) GetTranslation(ctx context.Context, keyText string) (*domain.Translation, bool) {
	row, err := r.queries.GetTranslation(ctx, keyText)
	if err != nil {
		return nil, false
	}

	var translation domain.Translation

	translation.FromRow(row)

	return &translation, true
}

func (r *Repository) UpsertTranslation(ctx context.Context, input domain.TranslationInput) bool {
	err := r.queries.AddTranslation(ctx, sqlc.AddTranslationParams{
		KeyText: input.KeyText,
		En:      pgtype.Text{String: input.En, Valid: true},
		Ar:      pgtype.Text{String: input.Ar, Valid: true},
		Fr:      pgtype.Text{String: input.Fr, Valid: true},
	})

	return err == nil
}
