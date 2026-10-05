package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) UpdateProduct(ctx context.Context, ownerID, id int64, input domain.ProductInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	locked, err := qtx.GetProductForUpdate(ctx, sqlc.GetProductForUpdateParams{BusinessID: ownerID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrProductNotFound
	}
	if err != nil {
		return err
	}
	old := productFromRow(sqlc.GetProductRow(locked))
	input = domain.MergeProductUpdate(old, input)

	if err := validateProductReferences(ctx, qtx, ownerID, input); err != nil {
		return err
	}

	if input.StockQuantitySet {
		locationID, err := stock.DefaultLocationID(ctx, tx, ownerID)
		if err != nil {
			return err
		}
		if _, _, err := stock.SetTotalTx(ctx, tx, stock.Movement{
			OwnerID:       ownerID,
			LocationID:    locationID,
			ProductID:     id,
			EntryType:     stock.EntryAdjustment,
			ReferenceType: "product",
			Channel:       "erp",
			Note:          "product stock edit",
		}, input.StockQuantity); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductNotFound
			}
			return err
		}
	}

	err = qtx.UpdateProduct(ctx, sqlc.UpdateProductParams{
		CategoryID:        nullInt64(input.CategoryID),
		SupplierID:        nullInt64(input.SupplierID),
		Name:              input.Name,
		Sku:               nullString(input.SKU),
		Barcode:           nullString(input.Barcode),
		Description:       nullString(input.Description),
		Unit:              nullString(input.Unit),
		CostPrice:         strconv.FormatFloat(input.CostPrice, 'f', -1, 64),
		SalePrice:         strconv.FormatFloat(input.SalePrice, 'f', -1, 64),
		BrandID:           nullInt64(input.BrandID),
		CompareAtPrice:    strconv.FormatFloat(input.CompareAtPrice, 'f', -1, 64),
		TaxRate:           strconv.FormatFloat(input.TaxRate, 'f', -1, 64),
		Currency:          defaultStr(input.Currency, "USD"),
		Weight:            strconv.FormatFloat(input.Weight, 'f', -1, 64),
		WeightUnit:        defaultStr(input.WeightUnit, "kg"),
		Length:            strconv.FormatFloat(input.Length, 'f', -1, 64),
		Width:             strconv.FormatFloat(input.Width, 'f', -1, 64),
		Height:            strconv.FormatFloat(input.Height, 'f', -1, 64),
		DimensionUnit:     defaultStr(input.DimensionUnit, "cm"),
		CountryOfOrigin:   nullString(input.CountryOfOrigin),
		HsCode:            nullString(input.HSCode),
		AllowBackorder:    input.AllowBackorder,
		LeadTimeDays:      int32(input.LeadTimeDays), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		LowStockThreshold: nullStringFromFloat(input.LowStockThreshold),
		ReorderQuantity:   nullStringFromFloat(input.ReorderQuantity),
		IsActive:          sql.NullBool{Bool: input.IsActive, Valid: true},
		TrackInventory:    sql.NullBool{Bool: input.TrackInventory, Valid: true},
		BusinessID:        ownerID,
		ID:                id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update product", logger.Error(err))

		return err
	}

	if input.ImageURLSet || (input.PresentFields == nil && strings.TrimSpace(input.ImageURL) != "") {
		if err := r.replacePrimaryProductImage(ctx, qtx, ownerID, id, input.Name, input.ImageURL); err != nil {
			return err
		}
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: ownerID, EntityType: "product", EntityID: id, Action: "updated", OldValues: productSnapshot(old), NewValues: productInputSnapshot(input)}); err != nil {
		return err
	}
	return tx.Commit()
}

func validateProductReferences(ctx context.Context, qtx *sqlc.Queries, ownerID int64, input domain.ProductInput) error {
	if input.CategoryID > 0 {
		if _, err := qtx.GetCategoryReference(ctx, sqlc.GetCategoryReferenceParams{BusinessID: ownerID, ID: input.CategoryID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductReferenceNotFound
			}
			return err
		}
	}
	if input.SupplierID > 0 {
		if _, err := qtx.GetSupplierReference(ctx, sqlc.GetSupplierReferenceParams{BusinessID: ownerID, ID: input.SupplierID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductReferenceNotFound
			}
			return err
		}
	}
	if input.BrandID > 0 {
		if _, err := qtx.GetBrandReference(ctx, sqlc.GetBrandReferenceParams{BusinessID: ownerID, ID: input.BrandID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrProductReferenceNotFound
			}
			return err
		}
	}
	return nil
}

func (r *Repository) replacePrimaryProductImage(ctx context.Context, qtx *sqlc.Queries, ownerID, productID int64, alt, imageURL string) error {
	imageURL = strings.TrimSpace(imageURL)
	if err := qtx.DeletePrimaryProductImage(ctx, sqlc.DeletePrimaryProductImageParams{
		BusinessID: ownerID,
		ProductID:  productID,
	}); err != nil {
		return err
	}
	if imageURL == "" {
		return nil
	}
	return qtx.InsertPrimaryProductImage(ctx, sqlc.InsertPrimaryProductImageParams{
		OwnerID:    ownerID,
		BusinessID: ownerID,
		ProductID:  productID,
		Url:        imageURL,
		Alt:        nullString(alt),
	})
}

func (r *Repository) applyProductStockTotal(
	ctx context.Context,
	tx *sql.Tx,
	ownerID int64,
	productID int64,
	currentTotal float64,
	targetTotal float64,
	entryType stock.EntryType,
	note string,
) error {
	locationID, err := stock.DefaultLocationID(ctx, tx, ownerID)
	if err != nil {
		return err
	}

	_, _, err = stock.ApplyMovementTx(ctx, tx, stock.Movement{
		OwnerID:       ownerID,
		LocationID:    locationID,
		ProductID:     productID,
		EntryType:     entryType,
		Change:        targetTotal - currentTotal,
		ReferenceID:   int64Ptr(productID),
		ReferenceType: "product",
		Channel:       "erp",
		Note:          note,
	})
	if err != nil {
		return fmt.Errorf("sync product stock: %w", err)
	}
	return nil
}

func (r *Repository) DeleteProduct(ctx context.Context, ownerID, id int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	locked, err := qtx.GetProductForUpdate(ctx, sqlc.GetProductForUpdateParams{BusinessID: ownerID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrProductNotFound
	}
	if err != nil {
		return err
	}
	old := productFromRow(sqlc.GetProductRow(locked))

	err = qtx.DeleteProduct(ctx, sqlc.DeleteProductParams{
		BusinessID: ownerID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete product", logger.Error(err))

		return err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: ownerID, EntityType: "product", EntityID: id, Action: "deleted", OldValues: productSnapshot(old)}); err != nil {
		return err
	}
	return tx.Commit()
}

// CountProducts returns the total number of products for the given owner,
// optionally filtered by category. Pass categoryID=0 to count all categories.
func (r *Repository) CountProducts(ctx context.Context, ownerID int64, categoryID int) (int64, error) {
	if categoryID > 0 {
		count, err := r.queries.CountProductsByCategory(ctx, sqlc.CountProductsByCategoryParams{
			BusinessID: ownerID,
			CategoryID: sql.NullInt64{Int64: int64(categoryID), Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to count products by category", logger.Error(err))
			return 0, err
		}
		return count, nil
	}

	count, err := r.queries.CountAllProducts(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count products", logger.Error(err))
		return 0, err
	}
	return count, nil
}
