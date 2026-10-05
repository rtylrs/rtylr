package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	"github.com/voxire/onion/internal/modules/erp/categories/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/categories/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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
		categories = append(categories, categoryFromRow(row))
	}

	return categories, nil
}

// ListCategoriesPaginated returns a single page of categories for the
// business, ordered by name. Limit and offset are applied as-is; callers
// clamp them.
func (r *Repository) ListCategoriesPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Category, error) {
	rows, err := r.queries.ListCategoriesPaginated(ctx, sqlc.ListCategoriesPaginatedParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list categories page", logger.Error(err))

		return nil, err
	}

	categories := make([]domain.Category, 0, len(rows))

	for _, row := range rows {
		categories = append(categories, categoryFromRow(row))
	}

	return categories, nil
}

func (r *Repository) CountCategories(ctx context.Context, businessID int64) (int64, error) {
	total, err := r.queries.CountCategories(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count categories", logger.Error(err))
		return 0, err
	}
	return total, nil
}

// categoryFromRow maps a sqlc erp_categories row to the domain model,
// collapsing nullable columns to their zero value when NULL.
func categoryFromRow(row sqlc.ErpCategory) domain.Category {
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

	return c
}

func (r *Repository) CreateCategory(ctx context.Context, businessID int64, input domain.CategoryInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	result, err := qtx.CreateCategory(ctx, sqlc.CreateCategoryParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        input.Name,
		Description: sqlutil.NullString(input.Description),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create category", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "category", EntityID: id, Action: "created", OldValues: nil, NewValues: categoryInputSnapshot(input)}); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, businessID, id int64, input domain.CategoryInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	err = qtx.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		Name:        input.Name,
		Description: sqlutil.NullString(input.Description),
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update category", logger.Error(err))

		return err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "category", EntityID: id, Action: "updated", OldValues: nil, NewValues: categoryInputSnapshot(input)}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteCategory(ctx context.Context, businessID, id int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	err = qtx.DeleteCategory(ctx, sqlc.DeleteCategoryParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete category", logger.Error(err))

		return err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "category", EntityID: id, Action: "deleted", OldValues: nil, NewValues: nil}); err != nil {
		return err
	}
	return tx.Commit()
}
