package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/store_themes/internal/domain"
	"github.com/voxire/onion/internal/modules/store_themes/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct{ queries *sqlc.Queries; pool *sql.DB }
func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool(); return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Theme, error) {
	rows, err := r.queries.ListThemes(ctx, ownerID)
	if err != nil { logger.Ctx(ctx).Error("Failed to list themes", logger.Error(err)); return nil, err }
	items := make([]domain.Theme, 0, len(rows))
	for _, row := range rows { items = append(items, fromRow(row)) }
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.ThemeInput) (int64, error) {
	result, err := r.queries.CreateTheme(ctx, sqlc.CreateThemeParams{
		OwnerID: ownerID, Name: input.Name,
		PrimaryColor: sql.NullString{String: input.PrimaryColor, Valid: input.PrimaryColor != ""},
		Font: sql.NullString{String: input.Font, Valid: input.Font != ""},
		HeaderStyle: sql.NullString{String: input.HeaderStyle, Valid: input.HeaderStyle != ""},
		FooterStyle: sql.NullString{String: input.FooterStyle, Valid: input.FooterStyle != ""},
		CustomCss: sql.NullString{String: input.CustomCSS, Valid: input.CustomCSS != ""},
	})
	if err != nil { logger.Ctx(ctx).Error("Failed to create theme", logger.Error(err)); return 0, err }
	id, _ := result.LastInsertId(); return id, nil
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.ThemeInput) error {
	err := r.queries.UpdateTheme(ctx, sqlc.UpdateThemeParams{
		Name: input.Name,
		PrimaryColor: sql.NullString{String: input.PrimaryColor, Valid: input.PrimaryColor != ""},
		Font: sql.NullString{String: input.Font, Valid: input.Font != ""},
		HeaderStyle: sql.NullString{String: input.HeaderStyle, Valid: input.HeaderStyle != ""},
		FooterStyle: sql.NullString{String: input.FooterStyle, Valid: input.FooterStyle != ""},
		CustomCss: sql.NullString{String: input.CustomCSS, Valid: input.CustomCSS != ""},
		OwnerID: ownerID, ID: id,
	})
	if err != nil { logger.Ctx(ctx).Error("Failed to update theme", logger.Error(err)); return err }
	return nil
}

func fromRow(row sqlc.StoreTheme) domain.Theme {
	t := domain.Theme{ID: row.ID, OwnerID: row.OwnerID, Name: row.Name}
	if row.PrimaryColor.Valid { t.PrimaryColor = row.PrimaryColor.String }
	if row.Font.Valid { t.Font = row.Font.String }
	if row.HeaderStyle.Valid { t.HeaderStyle = row.HeaderStyle.String }
	if row.FooterStyle.Valid { t.FooterStyle = row.FooterStyle.String }
	if row.CustomCss.Valid { t.CustomCSS = row.CustomCss.String }
	if row.CreatedAt.Valid { t.CreatedAt = row.CreatedAt.Time }
	if row.UpdatedAt.Valid { t.UpdatedAt = row.UpdatedAt.Time }
	return t
}
