package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/store_products/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.StoreProduct, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT
			sp.id,
			sp.owner_id,
			sp.product_id,
			COALESCE(ep.name, sp.title) AS title,
			sp.slug,
			COALESCE(ep.description, sp.description) AS description,
			COALESCE(epi.images, sp.images, JSON_ARRAY()) AS images,
			COALESCE(ep.sale_price, sp.price) AS price,
			COALESCE(ep.compare_at_price, sp.compare_at_price) AS compare_at_price,
			sp.is_published,
			sp.sort_order,
			sp.created_at,
			sp.updated_at
		FROM store_products sp
		LEFT JOIN erp_products ep
		  ON ep.owner_id = sp.owner_id
		 AND ep.id = sp.product_id
		LEFT JOIN (
			SELECT owner_id, product_id, JSON_ARRAYAGG(url) AS images
			FROM erp_product_images
			GROUP BY owner_id, product_id
		) epi
		  ON epi.owner_id = sp.owner_id
		 AND epi.product_id = sp.product_id
		WHERE sp.owner_id = ?
		ORDER BY sp.sort_order ASC
	`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store products", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.StoreProduct, 0)
	for rows.Next() {
		item, err := scanStoreProduct(ctx, rows)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to scan store product", logger.Error(err))
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.StoreProductInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin store product transaction", logger.Error(err))
		return 0, err
	}
	defer tx.Rollback()

	productID, err := syncCatalogProduct(ctx, tx, ownerID, input.ProductID, input)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sync catalog product for storefront", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateStoreProduct(ctx, sqlc.CreateStoreProductParams{
		OwnerID:        ownerID,
		ProductID:      sql.NullInt64{Int64: productID, Valid: productID != 0},
		Title:          input.Title,
		Slug:           sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Images:         nullJSON(input.Images),
		Price:          strconv.FormatFloat(input.Price, 'f', 2, 64),
		CompareAtPrice: nullDecimal(input.CompareAtPrice),
		IsPublished:    sql.NullBool{Bool: input.IsPublished, Valid: true},
		SortOrder:      sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create store product", logger.Error(err))
		return 0, err
	}
	id, _ := result.LastInsertId()
	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit store product transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.StoreProductInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin store product update transaction", logger.Error(err))
		return err
	}
	defer tx.Rollback()

	productID := input.ProductID
	if productID == 0 {
		var linkedProductID sql.NullInt64
		if err := tx.QueryRowContext(ctx,
			`SELECT product_id FROM store_products WHERE owner_id = ? AND id = ? LIMIT 1`,
			ownerID, id,
		).Scan(&linkedProductID); err != nil {
			logger.Ctx(ctx).Error("Failed to load existing store product link", logger.Error(err))
			return err
		}
		if linkedProductID.Valid {
			productID = linkedProductID.Int64
		}
	}

	productID, err = syncCatalogProduct(ctx, tx, ownerID, productID, input)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sync catalog product for storefront", logger.Error(err))
		return err
	}

	err = r.queries.WithTx(tx).UpdateStoreProduct(ctx, sqlc.UpdateStoreProductParams{
		ProductID:      sql.NullInt64{Int64: productID, Valid: productID != 0},
		Title:          input.Title,
		Slug:           sql.NullString{String: input.Slug, Valid: input.Slug != ""},
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Images:         nullJSON(input.Images),
		Price:          strconv.FormatFloat(input.Price, 'f', 2, 64),
		CompareAtPrice: nullDecimal(input.CompareAtPrice),
		IsPublished:    sql.NullBool{Bool: input.IsPublished, Valid: true},
		SortOrder:      sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
		OwnerID:        ownerID,
		ID:             id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update store product", logger.Error(err))
		return err
	}
	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit store product update transaction", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteStoreProduct(ctx, sqlc.DeleteStoreProductParams{OwnerID: ownerID, ID: id})
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

func scanStoreProduct(ctx context.Context, rows *sql.Rows) (domain.StoreProduct, error) {
	var (
		p              domain.StoreProduct
		productID      sql.NullInt64
		slug           sql.NullString
		description    sql.NullString
		images         []byte
		price          string
		compareAtPrice sql.NullString
		isPublished    sql.NullBool
		sortOrder      sql.NullInt32
		createdAt      sql.NullTime
		updatedAt      sql.NullTime
	)
	err := rows.Scan(
		&p.ID,
		&p.OwnerID,
		&productID,
		&p.Title,
		&slug,
		&description,
		&images,
		&price,
		&compareAtPrice,
		&isPublished,
		&sortOrder,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.StoreProduct{}, err
	}
	if productID.Valid {
		p.ProductID = productID.Int64
	}
	if slug.Valid {
		p.Slug = slug.String
	}
	if description.Valid {
		p.Description = description.String
	}
	if len(images) != 0 {
		p.Images = string(images)
	}
	p.Price = parseMonetary(ctx, price)
	if compareAtPrice.Valid {
		p.CompareAtPrice = parseMonetary(ctx, compareAtPrice.String)
	}
	if isPublished.Valid {
		p.IsPublished = isPublished.Bool
	}
	if sortOrder.Valid {
		p.SortOrder = int(sortOrder.Int32)
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}
	return p, nil
}

func syncCatalogProduct(ctx context.Context, tx *sql.Tx, ownerID, productID int64, input domain.StoreProductInput) (int64, error) {
	if productID != 0 {
		var existingID int64
		if err := tx.QueryRowContext(ctx,
			`SELECT id FROM erp_products WHERE owner_id = ? AND id = ? LIMIT 1`,
			ownerID, productID,
		).Scan(&existingID); err != nil {
			return 0, err
		}

		_, err := tx.ExecContext(ctx,
			`UPDATE erp_products
			 SET name = ?,
			     description = ?,
			     sale_price = ?,
			     compare_at_price = ?,
			     is_active = TRUE
			 WHERE owner_id = ? AND id = ?`,
			input.Title,
			sql.NullString{String: input.Description, Valid: input.Description != ""},
			strconv.FormatFloat(input.Price, 'f', 2, 64),
			strconv.FormatFloat(input.CompareAtPrice, 'f', 2, 64),
			ownerID,
			existingID,
		)
		if err != nil {
			return 0, err
		}
		if err := syncCatalogImages(ctx, tx, ownerID, existingID, input.Images); err != nil {
			return 0, err
		}
		return existingID, nil
	}

	result, err := tx.ExecContext(ctx,
		`INSERT INTO erp_products
		 (owner_id, name, description, sale_price, compare_at_price, is_active, track_inventory)
		 VALUES (?, ?, ?, ?, ?, TRUE, TRUE)`,
		ownerID,
		input.Title,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
		strconv.FormatFloat(input.Price, 'f', 2, 64),
		strconv.FormatFloat(input.CompareAtPrice, 'f', 2, 64),
	)
	if err != nil {
		return 0, err
	}
	productID, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := syncCatalogImages(ctx, tx, ownerID, productID, input.Images); err != nil {
		return 0, err
	}
	return productID, nil
}

type productImageInput struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

func syncCatalogImages(ctx context.Context, tx *sql.Tx, ownerID, productID int64, imagesJSON string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM erp_product_images WHERE owner_id = ? AND product_id = ?`,
		ownerID,
		productID,
	)
	if err != nil {
		return err
	}

	images := parseImageInputs(ctx, imagesJSON)
	for i, image := range images {
		url := strings.TrimSpace(image.URL)
		if url == "" {
			continue
		}
		alt := strings.TrimSpace(image.Alt)
		_, err = tx.ExecContext(ctx,
			`INSERT INTO erp_product_images
			 (owner_id, product_id, url, alt, position, is_primary)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			ownerID,
			productID,
			url,
			sql.NullString{String: alt, Valid: alt != ""},
			i,
			i == 0,
		)
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
