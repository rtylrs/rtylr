package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
	"github.com/voxire/onion/internal/modules/categories/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListCategories(ctx context.Context, businessID int64) ([]domain.Category, error) {
	rows, err := r.queries.ListCategories(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list categories", logger.Error(err))

		return nil, err
	}

	categories := make([]domain.Category, 0, len(rows))

	for _, row := range rows {
		var c domain.Category
		c.ID = row.ID
		c.OwnerID = row.OwnerID
		c.BusinessID = row.BusinessID
		c.Name = row.Name

		if row.Description.Valid {
			c.Description = row.Description.String
		}

		if row.CreatedAt.Valid {
			c.CreatedAt = row.CreatedAt.Time
		}

		if row.UpdatedAt.Valid {
			c.UpdatedAt = row.UpdatedAt.Time
		}

		categories = append(categories, c)
	}

	return categories, nil
}

func (r *Repository) CreateCategory(ctx context.Context, businessID int64, input domain.CategoryInput) (int64, error) {
	result, err := r.queries.CreateCategory(ctx, sqlc.CreateCategoryParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        input.Name,
		Description: nullString(input.Description),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create category", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, businessID, id int64, input domain.CategoryInput) error {
	err := r.queries.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		Name:        input.Name,
		Description: nullString(input.Description),
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update category", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteCategory(ctx, sqlc.DeleteCategoryParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete category", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
