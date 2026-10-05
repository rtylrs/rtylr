package infrastructure

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
)

type productRowScanner interface {
	Scan(...any) error
}

func productFromSQLRow(row productRowScanner) (domain.Product, error) {
	var product domain.Product
	var categoryID, supplierID sql.NullInt64
	var sku, barcode, description, unit, imageURL sql.NullString
	var lowStockThreshold, reorderQuantity sql.NullFloat64
	if err := row.Scan(
		&product.ID,
		&product.OwnerID,
		&product.BusinessID,
		&categoryID,
		&supplierID,
		&product.Name,
		&sku,
		&barcode,
		&description,
		&unit,
		&product.CostPrice,
		&product.SalePrice,
		&product.StockQuantity,
		&lowStockThreshold,
		&reorderQuantity,
		&product.IsActive,
		&product.TrackInventory,
		&imageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		return domain.Product{}, err
	}

	if categoryID.Valid {
		product.CategoryID = categoryID.Int64
	}
	if supplierID.Valid {
		product.SupplierID = supplierID.Int64
	}
	if sku.Valid {
		product.SKU = sku.String
	}
	if barcode.Valid {
		product.Barcode = barcode.String
	}
	if description.Valid {
		product.Description = description.String
	}
	if unit.Valid {
		product.Unit = unit.String
	}
	if lowStockThreshold.Valid {
		product.LowStockThreshold = lowStockThreshold.Float64
	}
	if reorderQuantity.Valid {
		product.ReorderQuantity = reorderQuantity.Float64
	}
	if imageURL.Valid {
		product.ImageURL = imageURL.String
	}
	return product, nil
}
