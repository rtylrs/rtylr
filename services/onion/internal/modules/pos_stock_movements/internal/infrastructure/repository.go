package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListMovements(ctx context.Context, businessID, branchID, productID int64) ([]domain.StockMovement, error) {
	rows, err := r.queries.ListStockMovements(ctx, sqlc.ListStockMovementsParams{
		BusinessID: businessID,
		BranchID:   branchID,
		ProductID:  productID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock movements", logger.Error(err))
		return nil, err
	}

	movements := make([]domain.StockMovement, 0, len(rows))
	for _, row := range rows {
		m := domain.StockMovement{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			BranchID:   row.BranchID,
			ProductID:  row.ProductID,
			Type:       string(row.Type),
			Quantity:   decimalToFloat(row.Quantity),
			CreatedAt:  row.CreatedAt.Time,
		}
		if row.ReferenceID.Valid {
			m.ReferenceID = &row.ReferenceID.Int64
		}
		if row.ReferenceType.Valid {
			m.ReferenceType = &row.ReferenceType.String
		}
		if row.Note.Valid {
			m.Note = &row.Note.String
		}
		if row.EmployeeID.Valid {
			m.EmployeeID = &row.EmployeeID.Int64
		}
		movements = append(movements, m)
	}
	return movements, nil
}

func (r *Repository) CreateMovement(ctx context.Context, businessID, branchID, productID int64, movementType string, quantity float64, referenceID *int64, referenceType, note *string, employeeID *int64) (int64, error) {
	var refID sql.NullInt64
	if referenceID != nil {
		refID = sql.NullInt64{Int64: *referenceID, Valid: true}
	}
	var refType sql.NullString
	if referenceType != nil {
		refType = sql.NullString{String: *referenceType, Valid: true}
	}
	var n sql.NullString
	if note != nil {
		n = sql.NullString{String: *note, Valid: true}
	}
	var empID sql.NullInt64
	if employeeID != nil {
		empID = sql.NullInt64{Int64: *employeeID, Valid: true}
	}

	result, err := r.queries.CreateStockMovement(ctx, sqlc.CreateStockMovementParams{
		BusinessID:    businessID,
		BranchID:      branchID,
		ProductID:     productID,
		Type:          sqlc.PosStockMovementsType(movementType),
		Quantity:      floatToDecimal(quantity),
		ReferenceID:   refID,
		ReferenceType: refType,
		Note:          n,
		EmployeeID:    empID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create stock movement", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateInventory(ctx context.Context, businessID, branchID, productID int64, quantityDelta float64) error {
	_, err := r.queries.UpsertBranchInventory(ctx, sqlc.UpsertBranchInventoryParams{
		BusinessID: businessID,
		BranchID:   branchID,
		ProductID:  productID,
		Quantity:   floatToDecimal(quantityDelta),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update branch inventory", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) SetInventoryQuantity(ctx context.Context, businessID, branchID, productID int64, quantity float64) error {
	_, err := r.queries.SetBranchInventoryQuantity(ctx, sqlc.SetBranchInventoryQuantityParams{
		BusinessID: businessID,
		BranchID:   branchID,
		ProductID:  productID,
		Quantity:   floatToDecimal(quantity),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to set branch inventory quantity", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ListBranchInventory(ctx context.Context, businessID, branchID int64) ([]domain.BranchInventory, error) {
	rows, err := r.queries.ListBranchInventory(ctx, sqlc.ListBranchInventoryParams{
		BusinessID: businessID,
		BranchID:   branchID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list branch inventory", logger.Error(err))
		return nil, err
	}

	items := make([]domain.BranchInventory, 0, len(rows))
	for _, row := range rows {
		item := domain.BranchInventory{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			BranchID:   row.BranchID,
			ProductID:  row.ProductID,
			Quantity:   decimalToFloat(row.Quantity),
			UpdatedAt:  row.UpdatedAt.Time,
		}
		if row.ReorderPoint.Valid {
			rp := decimalToFloat(row.ReorderPoint.String)
			item.ReorderPoint = &rp
		}
		if row.ReorderQuantity.Valid {
			rq := decimalToFloat(row.ReorderQuantity.String)
			item.ReorderQuantity = &rq
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) ListLowStockItems(ctx context.Context, businessID, branchID int64) ([]domain.BranchInventory, error) {
	rows, err := r.queries.ListLowStockItems(ctx, sqlc.ListLowStockItemsParams{
		BusinessID: businessID,
		BranchID:   branchID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list low stock items", logger.Error(err))
		return nil, err
	}

	items := make([]domain.BranchInventory, 0, len(rows))
	for _, row := range rows {
		item := domain.BranchInventory{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			BranchID:   row.BranchID,
			ProductID:  row.ProductID,
			Quantity:   decimalToFloat(row.Quantity),
			UpdatedAt:  row.UpdatedAt.Time,
		}
		if row.ReorderPoint.Valid {
			rp := decimalToFloat(row.ReorderPoint.String)
			item.ReorderPoint = &rp
		}
		if row.ReorderQuantity.Valid {
			rq := decimalToFloat(row.ReorderQuantity.String)
			item.ReorderQuantity = &rq
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetInventoryItem(ctx context.Context, businessID, branchID, productID int64) (domain.BranchInventory, error) {
	row, err := r.queries.GetBranchInventoryItem(ctx, sqlc.GetBranchInventoryItemParams{
		BusinessID: businessID,
		BranchID:   branchID,
		ProductID:  productID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.BranchInventory{}, domain.ErrInventoryNotFound
		}
		logger.Ctx(ctx).Error("Failed to get inventory item", logger.Error(err))
		return domain.BranchInventory{}, err
	}
	item := domain.BranchInventory{
		ID:         row.ID,
		BusinessID: row.BusinessID,
		BranchID:   row.BranchID,
		ProductID:  row.ProductID,
		Quantity:   decimalToFloat(row.Quantity),
		UpdatedAt:  row.UpdatedAt.Time,
	}
	if row.ReorderPoint.Valid {
		rp := decimalToFloat(row.ReorderPoint.String)
		item.ReorderPoint = &rp
	}
	if row.ReorderQuantity.Valid {
		rq := decimalToFloat(row.ReorderQuantity.String)
		item.ReorderQuantity = &rq
	}
	return item, nil
}

func decimalToFloat(s any) float64 {
	var str string
	switch v := s.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return 0
	}
	if str == "" {
		return 0
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		// Don't silently coerce corrupt decimal data to 0 — surface it.
		logger.Global().Warn("failed to parse decimal from database",
			logger.String("value", str), logger.Error(err))
		return 0
	}
	return f
}

func floatToDecimal(f float64) string {
	return fmt.Sprintf("%.3f", f)
}
