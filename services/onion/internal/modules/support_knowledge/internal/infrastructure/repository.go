package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/domain"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Article, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := r.queries.CountArticles(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count articles", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListArticles(ctx, sqlc.ListArticlesParams{
		BusinessID: businessID,
		Limit:      int32(pageSize),
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list articles", logger.Error(err))
		return nil, 0, err
	}

	items := make([]domain.Article, 0, len(rows))
	for _, row := range rows {
		items = append(items, articleFromRow(row))
	}
	return items, int(total), nil
}

func (r *Repository) Get(ctx context.Context, businessID, articleID int64, slug, title string) (domain.Article, error) {
	var (
		row sqlc.SupportArticle
		err error
	)
	switch {
	case articleID > 0:
		row, err = r.queries.GetArticleByID(ctx, sqlc.GetArticleByIDParams{BusinessID: businessID, ID: articleID})
	case strings.TrimSpace(slug) != "":
		row, err = r.queries.GetArticleBySlug(ctx, sqlc.GetArticleBySlugParams{BusinessID: businessID, Slug: strings.TrimSpace(slug)})
	default:
		row, err = r.queries.GetArticleByTitle(ctx, sqlc.GetArticleByTitleParams{BusinessID: businessID, Title: strings.TrimSpace(title)})
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ErrArticleNotFound
		}
		logger.Ctx(ctx).Error("Failed to get support article", logger.Error(err))
		return domain.Article{}, err
	}
	return articleFromRow(row), nil
}

func (r *Repository) Create(ctx context.Context, input domain.ArticleInput) (int64, error) {
	result, err := r.queries.CreateArticle(ctx, sqlc.CreateArticleParams{
		BusinessID:  input.BusinessID,
		Title:       input.Title,
		Slug:        input.Slug,
		Body:        input.Body,
		Category:    sql.NullString{String: input.Category, Valid: input.Category != ""},
		IsPublished: sql.NullBool{Bool: input.IsPublished, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create article", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.ArticleInput) error {
	err := r.queries.UpdateArticle(ctx, sqlc.UpdateArticleParams{
		Title:       input.Title,
		Slug:        input.Slug,
		Body:        input.Body,
		Category:    sql.NullString{String: input.Category, Valid: input.Category != ""},
		IsPublished: sql.NullBool{Bool: input.IsPublished, Valid: true},
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update article", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteArticle(ctx, sqlc.DeleteArticleParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete article", logger.Error(err))
		return err
	}
	return nil
}

func articleFromRow(row sqlc.SupportArticle) domain.Article {
	a := domain.Article{
		ID:         row.ID,
		BusinessID: row.BusinessID,
		Title:      row.Title,
		Slug:       row.Slug,
		Body:       row.Body,
	}
	if row.Category.Valid {
		a.Category = row.Category.String
	}
	if row.IsPublished.Valid {
		a.IsPublished = row.IsPublished.Bool
	}
	if row.Views.Valid {
		a.Views = int(row.Views.Int32)
	}
	if row.CreatedAt.Valid {
		a.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		a.UpdatedAt = row.UpdatedAt.Time
	}
	return a
}
