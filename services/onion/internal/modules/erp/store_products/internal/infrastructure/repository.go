package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/store_products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.StoreProduct, error) {
	// cross-schema read from erp_products + erp_product_images — catalog enrichment
	rows, err := r.queries.ListStoreProductsWithCatalog(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store products", logger.Error(err))
		return nil, err
	}

	items := make([]domain.StoreProduct, 0, len(rows))
	for _, row := range rows {
		item, err := storeProductFromCatalogRow(ctx, row)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to map store product", logger.Error(err))
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.StoreProductInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin store product transaction", logger.Error(err))
		return 0, err
	}
	defer tx.Rollback()

	productID, err := syncCatalogProduct(ctx, r.queries.WithTx(tx), ownerID, businessID, input.ProductID, input)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sync catalog product for storefront", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateStoreProduct(ctx, sqlc.CreateStoreProductParams{
		OwnerID:        ownerID,
		BusinessID:     businessID,
		ProductID:      sql.NullInt64{Int64: productID, Valid: productID != 0},
		Title:          input.Title,
		Slug:           sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Images:         nullJSON(input.Images),
		Price:          strconv.FormatFloat(input.Price, 'f', 2, 64),
		CompareAtPrice: nullDecimal(input.CompareAtPrice),
		IsPublished:    sql.NullBool{Bool: input.IsPublished, Valid: true},
		SortOrder:      sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create store product", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read store product insert id", logger.Error(err))
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit store product transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.StoreProductInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin store product update transaction", logger.Error(err))
		return err
	}
	defer tx.Rollback()

	// Verify the storefront row exists before touching the shared ERP catalog:
	// a stale caller (row deleted elsewhere, stale list) otherwise rewrites
	// erp_products and then reports success from a zero-row update.
	linkedProductID, err := r.queries.WithTx(tx).GetStoreProductLinkedID(ctx, sqlc.GetStoreProductLinkedIDParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrStoreProductNotFound
		}
		logger.Ctx(ctx).Error("Failed to load existing store product link", logger.Error(err))
		return err
	}

	productID := input.ProductID
	if productID == 0 && linkedProductID.Valid {
		productID = linkedProductID.Int64
	}

	productID, err = syncCatalogProduct(ctx, r.queries.WithTx(tx), businessID, businessID, productID, input)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sync catalog product for storefront", logger.Error(err))
		return err
	}

	affected, err := r.queries.WithTx(tx).UpdateStoreProduct(ctx, sqlc.UpdateStoreProductParams{
		ProductID:      sql.NullInt64{Int64: productID, Valid: productID != 0},
		Title:          input.Title,
		Slug:           sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Images:         nullJSON(input.Images),
		Price:          strconv.FormatFloat(input.Price, 'f', 2, 64),
		CompareAtPrice: nullDecimal(input.CompareAtPrice),
		IsPublished:    sql.NullBool{Bool: input.IsPublished, Valid: true},
		SortOrder:      sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BusinessID:     businessID,
		ID:             id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update store product", logger.Error(err))
		return err
	}
	if affected == 0 {
		// Raced with a concurrent delete after the existence check. Returning
		// here rolls the whole transaction back, including the catalog write,
		// so the caller is never told an update happened that did not.
		return domain.ErrStoreProductNotFound
	}
	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit store product update transaction", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteStoreProduct(ctx, sqlc.DeleteStoreProductParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete store product", logger.Error(err))
		return err
	}
	return nil
}

func parseMonetary(ctx context.Context, s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to parse monetary value", logger.String("value", s), logger.Error(err))
		return 0
	}
	return v
}

func storeProductFromCatalogRow(ctx context.Context, row sqlc.ListStoreProductsWithCatalogRow) (domain.StoreProduct, error) {
	p := domain.StoreProduct{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Title:      row.Title,
	}
	if row.ProductID.Valid {
		p.ProductID = row.ProductID.Int64
	}
	if row.Slug.Valid {
		p.Slug = row.Slug.String
	}
	if row.Description.Valid {
		p.Description = row.Description.String
	}
	if len(row.Images) != 0 {
		p.Images = string(row.Images)
	}
	p.Price = parseMonetary(ctx, row.Price)
	if row.CompareAtPrice != "" {
		p.CompareAtPrice = parseMonetary(ctx, row.CompareAtPrice)
	}
	if row.IsPublished.Valid {
		p.IsPublished = row.IsPublished.Bool
	}
	if row.SortOrder.Valid {
		p.SortOrder = int(row.SortOrder.Int32)
	}
	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		p.UpdatedAt = row.UpdatedAt.Time
	}
	return p, nil
}

func syncCatalogProduct(ctx context.Context, q *sqlc.Queries, ownerID, businessID, productID int64, input domain.StoreProductInput) (int64, error) {
	if productID != 0 {
		_, err := q.GetCatalogProductByID(ctx, sqlc.GetCatalogProductByIDParams{
			BusinessID: businessID,
			ID:         productID,
		})
		if err != nil {
			return 0, err
		}

		err = q.UpdateCatalogProduct(ctx, sqlc.UpdateCatalogProductParams{
			Name:           input.Title,
			Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
			SalePrice:      strconv.FormatFloat(input.Price, 'f', 2, 64),
			CompareAtPrice: strconv.FormatFloat(input.CompareAtPrice, 'f', 2, 64),
			BusinessID:     businessID,
			ID:             productID,
		})
		if err != nil {
			return 0, err
		}
		if err := syncCatalogImages(ctx, q, ownerID, businessID, productID, input.Images); err != nil {
			return 0, err
		}
		return productID, nil
	}

	result, err := q.InsertCatalogProduct(ctx, sqlc.InsertCatalogProductParams{
		OwnerID:        ownerID,
		BusinessID:     businessID,
		Name:           input.Title,
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		SalePrice:      strconv.FormatFloat(input.Price, 'f', 2, 64),
		CompareAtPrice: strconv.FormatFloat(input.CompareAtPrice, 'f', 2, 64),
	})
	if err != nil {
		return 0, err
	}
	productID, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := syncCatalogImages(ctx, q, ownerID, businessID, productID, input.Images); err != nil {
		return 0, err
	}
	return productID, nil
}

type productImageInput struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

func syncCatalogImages(ctx context.Context, q *sqlc.Queries, ownerID, businessID, productID int64, imagesJSON string) error {
	if err := q.DeleteCatalogProductImages(ctx, sqlc.DeleteCatalogProductImagesParams{
		BusinessID: businessID,
		ProductID:  productID,
	}); err != nil {
		return err
	}

	images := parseImageInputs(ctx, imagesJSON)
	for i, image := range images {
		url := strings.TrimSpace(image.URL)
		if url == "" {
			continue
		}
		alt := strings.TrimSpace(image.Alt)
		err := q.InsertCatalogProductImage(ctx, sqlc.InsertCatalogProductImageParams{
			OwnerID:    ownerID,
			BusinessID: businessID,
			ProductID:  productID,
			Url:        url,
			Alt:        sql.NullString{String: alt, Valid: alt != ""},
			Position:   int32(i),
			IsPrimary:  i == 0,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func parseImageInputs(ctx context.Context, imagesJSON string) []productImageInput {
	raw := strings.TrimSpace(imagesJSON)
	if raw == "" {
		return nil
	}

	var objects []productImageInput
	if err := json.Unmarshal([]byte(raw), &objects); err == nil {
		return objects
	}

	var urls []string
	if err := json.Unmarshal([]byte(raw), &urls); err == nil {
		images := make([]productImageInput, 0, len(urls))
		for _, url := range urls {
			images = append(images, productImageInput{URL: url})
		}
		return images
	}

	logger.Ctx(ctx).Warn("Failed to parse store product images JSON")
	return nil
}

func nullJSON(s string) []byte {
	if s == "" {
		return nil
	}
	return []byte(s)
}

func nullDecimal(v float64) sql.NullString {
	return sql.NullString{String: strconv.FormatFloat(v, 'f', 2, 64), Valid: true}
}
