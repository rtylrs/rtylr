package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func productFromRow(row sqlc.GetProductRow) domain.Product {
	var p domain.Product
	p.ID = row.ID
	p.OwnerID = row.OwnerID
	p.BusinessID = row.BusinessID
	p.Name = row.Name

	if row.CategoryID.Valid {
		p.CategoryID = row.CategoryID.Int64
	}

	if row.SupplierID.Valid {
		p.SupplierID = row.SupplierID.Int64
	}

	if row.Sku.Valid {
		p.SKU = row.Sku.String
	}

	if row.Barcode.Valid {
		p.Barcode = row.Barcode.String
	}

	if row.Description.Valid {
		p.Description = row.Description.String
	}

	if row.Unit.Valid {
		p.Unit = row.Unit.String
	}

	p.CostPrice = sqlutil.ParseDecimal(row.CostPrice)
	p.SalePrice = sqlutil.ParseDecimal(row.SalePrice)
	p.StockQuantity = sqlutil.ParseDecimal(row.StockQuantity)

	if row.LowStockThreshold.Valid {
		p.LowStockThreshold = sqlutil.ParseDecimal(row.LowStockThreshold.String)
	}

	if row.ReorderQuantity.Valid {
		p.ReorderQuantity = sqlutil.ParseDecimal(row.ReorderQuantity.String)
	}

	if row.IsActive.Valid {
		p.IsActive = row.IsActive.Bool
	}

	if row.TrackInventory.Valid {
		p.TrackInventory = row.TrackInventory.Bool
	}

	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		p.UpdatedAt = row.UpdatedAt.Time
	}

	if row.BrandID.Valid {
		p.BrandID = row.BrandID.Int64
	}
	p.CompareAtPrice = sqlutil.ParseDecimal(row.CompareAtPrice)
	p.TaxRate = sqlutil.ParseDecimal(row.TaxRate)
	p.Currency = row.Currency
	p.Weight = sqlutil.ParseDecimal(row.Weight)
	p.WeightUnit = row.WeightUnit
	p.Length = sqlutil.ParseDecimal(row.Length)
	p.Width = sqlutil.ParseDecimal(row.Width)
	p.Height = sqlutil.ParseDecimal(row.Height)
	p.DimensionUnit = row.DimensionUnit
	if row.CountryOfOrigin.Valid {
		p.CountryOfOrigin = row.CountryOfOrigin.String
	}
	if row.HsCode.Valid {
		p.HSCode = row.HsCode.String
	}
	p.AllowBackorder = row.AllowBackorder
	p.LeadTimeDays = int(row.LeadTimeDays)

	return p
}

func productFromBarcodeRow(row sqlc.GetProductByBarcodeRow) domain.Product {
	var p domain.Product
	p.ID = row.ID
	p.OwnerID = row.OwnerID
	p.BusinessID = row.BusinessID
	p.Name = row.Name

	if row.CategoryID.Valid {
		p.CategoryID = row.CategoryID.Int64
	}

	if row.SupplierID.Valid {
		p.SupplierID = row.SupplierID.Int64
	}

	if row.Sku.Valid {
		p.SKU = row.Sku.String
	}

	if row.Barcode.Valid {
		p.Barcode = row.Barcode.String
	}

	if row.Description.Valid {
		p.Description = row.Description.String
	}

	if row.Unit.Valid {
		p.Unit = row.Unit.String
	}

	p.CostPrice = sqlutil.ParseDecimal(row.CostPrice)
	p.SalePrice = sqlutil.ParseDecimal(row.SalePrice)
	p.StockQuantity = sqlutil.ParseDecimal(row.StockQuantity)

	if row.LowStockThreshold.Valid {
		p.LowStockThreshold = sqlutil.ParseDecimal(row.LowStockThreshold.String)
	}

	if row.ReorderQuantity.Valid {
		p.ReorderQuantity = sqlutil.ParseDecimal(row.ReorderQuantity.String)
	}

	if row.IsActive.Valid {
		p.IsActive = row.IsActive.Bool
	}

	if row.TrackInventory.Valid {
		p.TrackInventory = row.TrackInventory.Bool
	}

	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		p.UpdatedAt = row.UpdatedAt.Time
	}

	if row.BrandID.Valid {
		p.BrandID = row.BrandID.Int64
	}
	p.CompareAtPrice = sqlutil.ParseDecimal(row.CompareAtPrice)
	p.TaxRate = sqlutil.ParseDecimal(row.TaxRate)
	p.Currency = row.Currency
	p.Weight = sqlutil.ParseDecimal(row.Weight)
	p.WeightUnit = row.WeightUnit
	p.Length = sqlutil.ParseDecimal(row.Length)
	p.Width = sqlutil.ParseDecimal(row.Width)
	p.Height = sqlutil.ParseDecimal(row.Height)
	p.DimensionUnit = row.DimensionUnit
	if row.CountryOfOrigin.Valid {
		p.CountryOfOrigin = row.CountryOfOrigin.String
	}
	if row.HsCode.Valid {
		p.HSCode = row.HsCode.String
	}
	p.AllowBackorder = row.AllowBackorder
	p.LeadTimeDays = int(row.LeadTimeDays)

	return p
}

// ListProductAttributes returns the custom attributes for a product, ordered.
func (r *Repository) ListProductAttributes(ctx context.Context, ownerID, productID int64) ([]domain.ProductAttribute, error) {
	rows, err := r.queries.ListProductAttributes(ctx, sqlc.ListProductAttributesParams{
		BusinessID: ownerID,
		ProductID:  productID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductAttribute, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ProductAttribute{
			ID:       row.ID,
			Name:     row.Name,
			Value:    row.Value,
			Position: int(row.Position),
		})
	}
	return out, nil
}

// ReplaceProductAttributes replaces all of a product's custom attributes with
// the given set (delete-then-insert). Blank-named entries are skipped.
func (r *Repository) ReplaceProductAttributes(ctx context.Context, ownerID, productID int64, attrs []domain.ProductAttributeInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	if _, err := qtx.GetProductForUpdate(ctx, sqlc.GetProductForUpdateParams{BusinessID: ownerID, ID: productID}); err != nil {
		return err
	}
	oldRows, err := qtx.ListProductAttributes(ctx, sqlc.ListProductAttributesParams{BusinessID: ownerID, ProductID: productID})
	if err != nil {
		return err
	}
	old := make([]productAttributeSnapshot, 0, len(oldRows))
	for _, row := range oldRows {
		old = append(old, productAttributeSnapshot{ID: row.ID, Name: row.Name, Value: row.Value, Position: int(row.Position)})
	}
	newValues := make([]productAttributeInputSnapshot, 0, len(attrs))

	if err := qtx.DeleteProductAttributes(ctx, sqlc.DeleteProductAttributesParams{
		BusinessID: ownerID,
		ProductID:  productID,
	}); err != nil {
		return err
	}
	for i, a := range attrs {
		if a.Name == "" {
			continue
		}
		newValues = append(newValues, productAttributeInputSnapshot(a))
		if err := qtx.InsertProductAttribute(ctx, sqlc.InsertProductAttributeParams{
			OwnerID:    ownerID,
			BusinessID: ownerID,
			ProductID:  productID,
			Name:       a.Name,
			Value:      a.Value,
			Position:   int32(i),
		}); err != nil {
			return err
		}
	}
	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: ownerID, EntityType: "product_attribute", EntityID: productID, Action: "replaced", OldValues: old, NewValues: newValues}); err != nil {
		return err
	}
	return tx.Commit()
}

var nullString = sqlutil.NullString

// defaultStr returns def when s is empty — for NOT NULL columns with a sensible
// default (currency, units) where the caller may not supply a value.
func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

var nullInt64 = sqlutil.NullInt64

func int64Ptr(v int64) *int64 {
	return &v
}

func nullStringFromFloat(v float64) sql.NullString {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	return sql.NullString{String: s, Valid: true}
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
