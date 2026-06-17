package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/stock_core"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}

func (r *Repository) ListMovements(ctx context.Context, businessID, ownerID, branchID, productID int64) ([]domain.StockMovement, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT
			l.id,
			l.product_id,
			l.entry_type,
			l.quantity_change,
			l.reference_id,
			l.reference_type,
			l.note,
			l.employee_id,
			l.created_at
		FROM stock_ledger l
		INNER JOIN stock_locations loc
		  ON loc.id = l.location_id
		 AND loc.owner_id = l.owner_id
		WHERE l.owner_id = ?
		  AND loc.source_table = 'pos_locations'
		  AND loc.source_id = ?
		  AND l.product_id = ?
		  AND l.channel = 'pos'
		ORDER BY l.created_at DESC, l.id DESC
	`, ownerID, branchID, productID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock movements", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	movements := make([]domain.StockMovement, 0)
	for rows.Next() {
		var (
			m             domain.StockMovement
			entryType     string
			quantity      string
			referenceID   sql.NullInt64
			referenceType sql.NullString
			note          sql.NullString
			employeeID    sql.NullInt64
			createdAt     sql.NullTime
		)
		if err := rows.Scan(
			&m.ID,
			&m.ProductID,
			&entryType,
			&quantity,
			&referenceID,
			&referenceType,
			&note,
			&employeeID,
			&createdAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock movement", logger.Error(err))
			return nil, err
		}

		m.BusinessID = businessID
		m.BranchID = branchID
		m.Type = posMovementType(entryType)
		m.Quantity = decimalToFloat(quantity)
		if referenceID.Valid {
			m.ReferenceID = &referenceID.Int64
		}
		if referenceType.Valid {
			m.ReferenceType = &referenceType.String
		}
		if note.Valid {
			m.Note = &note.String
		}
		if employeeID.Valid {
			m.EmployeeID = &employeeID.Int64
		}
		if createdAt.Valid {
			m.CreatedAt = createdAt.Time
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		logger.Ctx(ctx).Error("Failed to iterate stock movements", logger.Error(err))
		return nil, err
	}
	return movements, nil
}

func (r *Repository) ReceiveStock(ctx context.Context, input domain.ReceiveStockInput) (int64, error) {
	return r.applySingleMovement(ctx, movementInput{
		BusinessID:  input.BusinessID,
		OwnerID:     input.OwnerID,
		BranchID:    input.BranchID,
		ProductID:   input.ProductID,
		EntryType:   stock_core.EntryPurchase,
		DisplayType: "receive",
		Change:      input.Quantity,
		Note:        stringPtrValue(input.Note),
		EmployeeID:  input.EmployeeID,
	})
}

func (r *Repository) AdjustStock(ctx context.Context, input domain.AdjustStockInput) (int64, error) {
	return r.applySingleMovement(ctx, movementInput{
		BusinessID:  input.BusinessID,
		OwnerID:     input.OwnerID,
		BranchID:    input.BranchID,
		ProductID:   input.ProductID,
		EntryType:   stock_core.EntryAdjustment,
		DisplayType: "adjust",
		Change:      input.Quantity,
		Note:        stringPtrValue(input.Note),
		EmployeeID:  input.EmployeeID,
	})
}

func (r *Repository) RecordWaste(ctx context.Context, input domain.WasteStockInput) (int64, error) {
	return r.applySingleMovement(ctx, movementInput{
		BusinessID:  input.BusinessID,
		OwnerID:     input.OwnerID,
		BranchID:    input.BranchID,
		ProductID:   input.ProductID,
		EntryType:   stock_core.EntryWaste,
		DisplayType: "waste",
		Change:      -input.Quantity,
		Note:        stringPtrValue(input.Note),
		EmployeeID:  input.EmployeeID,
	})
}

func (r *Repository) TransferStock(ctx context.Context, input domain.TransferStockInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin POS stock transfer", logger.Error(err))
		return 0, err
	}
	defer rollbackOnError(tx, &err)

	outID, err := r.applyMovementTx(ctx, tx, movementInput{
		BusinessID:  input.BusinessID,
		OwnerID:     input.OwnerID,
		BranchID:    input.FromBranchID,
		ProductID:   input.ProductID,
		EntryType:   stock_core.EntryTransferOut,
		DisplayType: "transfer_out",
		Change:      -input.Quantity,
		Note:        stringPtrValue(input.Note),
		EmployeeID:  input.EmployeeID,
	})
	if err != nil {
		return 0, err
	}
	transferGroup := outID
	if _, err = tx.ExecContext(ctx, `UPDATE stock_ledger SET transfer_group = ? WHERE id = ? AND owner_id = ?`, transferGroup, outID, input.OwnerID); err != nil {
		logger.Ctx(ctx).Error("Failed to mark POS transfer out group", logger.Error(err))
		return 0, err
	}

	inID, err := r.applyMovementTx(ctx, tx, movementInput{
		BusinessID:     input.BusinessID,
		OwnerID:        input.OwnerID,
		BranchID:       input.ToBranchID,
		ProductID:      input.ProductID,
		EntryType:      stock_core.EntryTransferIn,
		DisplayType:    "transfer_in",
		Change:         input.Quantity,
		Note:           stringPtrValue(input.Note),
		EmployeeID:     input.EmployeeID,
		TransferGroup:  &transferGroup,
		AllowNegative:  false,
		ReferenceType:  "pos_stock_transfer",
		ReferenceID:    &outID,
		SkipProductLog: true,
	})
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit POS stock transfer", logger.Error(err))
		return 0, err
	}
	return inID, nil
}

func (r *Repository) CreateMovement(ctx context.Context, businessID, branchID, productID int64, movementType string, quantity float64, referenceID *int64, referenceType, note *string, employeeID *int64) (int64, error) {
	entryType := entryTypeForMovement(movementType)
	ownerID := r.ownerIDForBusiness(ctx, businessID)
	return r.applySingleMovement(ctx, movementInput{
		BusinessID:    businessID,
		OwnerID:       ownerID,
		BranchID:      branchID,
		ProductID:     productID,
		EntryType:     entryType,
		DisplayType:   movementType,
		Change:        quantity,
		ReferenceID:   referenceID,
		ReferenceType: stringPtrValue(referenceType),
		Note:          stringPtrValue(note),
		EmployeeID:    employeeID,
	})
}

func (r *Repository) UpdateInventory(ctx context.Context, businessID, branchID, productID int64, quantityDelta float64) error {
	ownerID := r.ownerIDForBusiness(ctx, businessID)
	_, err := r.applySingleMovement(ctx, movementInput{
		BusinessID:  businessID,
		OwnerID:     ownerID,
		BranchID:    branchID,
		ProductID:   productID,
		EntryType:   stock_core.EntryAdjustment,
		DisplayType: "adjust",
		Change:      quantityDelta,
	})
	return err
}

func (r *Repository) SetInventoryQuantity(ctx context.Context, businessID, branchID, productID int64, quantity float64) error {
	ownerID := r.ownerIDForBusiness(ctx, businessID)
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin POS stock count", logger.Error(err))
		return err
	}
	defer rollbackOnError(tx, &err)

	if err = r.ensureInventoryProduct(ctx, tx, ownerID, productID); err != nil {
		return err
	}
	locationID, err := stock_core.POSBranchLocationID(ctx, tx, ownerID, branchID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve POS branch stock location", logger.Error(err))
		return err
	}
	_, _, err = stock_core.SetLevelTx(ctx, tx, stock_core.Movement{
		OwnerID:       ownerID,
		LocationID:    locationID,
		ProductID:     productID,
		EntryType:     stock_core.EntryCount,
		ReferenceType: "pos_stock_count",
		Channel:       "pos",
		AllowNegative: false,
	}, quantity)
	if errors.Is(err, stock_core.ErrInsufficientStock) {
		return domain.ErrInsufficientStock
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to set POS branch inventory quantity", logger.Error(err))
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit POS stock count", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ListBranchInventory(ctx context.Context, businessID, ownerID, branchID int64) ([]domain.BranchInventory, error) {
	rows, err := r.pool.QueryContext(ctx, inventorySelectSQL(`
		WHERE p.owner_id = ?
		  AND p.is_active = TRUE
		  AND p.track_inventory = TRUE
		ORDER BY p.name
	`), branchID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list branch inventory", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanBranchInventoryRows(ctx, rows, businessID, branchID)
}

func (r *Repository) ListLowStockItems(ctx context.Context, businessID, ownerID, branchID int64) ([]domain.BranchInventory, error) {
	rows, err := r.pool.QueryContext(ctx, inventorySelectSQL(`
		WHERE p.owner_id = ?
		  AND p.is_active = TRUE
		  AND p.track_inventory = TRUE
		  AND COALESCE(sl.reorder_point, p.low_stock_threshold, 0) > 0
		  AND COALESCE(sl.on_hand, 0) <= COALESCE(sl.reorder_point, p.low_stock_threshold, 0)
		ORDER BY COALESCE(sl.on_hand, 0) ASC, p.name
	`), branchID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list low stock items", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanBranchInventoryRows(ctx, rows, businessID, branchID)
}

func (r *Repository) GetInventoryItem(ctx context.Context, businessID, ownerID, branchID, productID int64) (domain.BranchInventory, error) {
	rows, err := r.pool.QueryContext(ctx, inventorySelectSQL(`
		WHERE p.owner_id = ?
		  AND p.id = ?
		  AND p.is_active = TRUE
		  AND p.track_inventory = TRUE
		LIMIT 1
	`), branchID, ownerID, productID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory item", logger.Error(err))
		return domain.BranchInventory{}, err
	}
	defer rows.Close()

	items, err := scanBranchInventoryRows(ctx, rows, businessID, branchID)
	if err != nil {
		return domain.BranchInventory{}, err
	}
	if len(items) == 0 {
		return domain.BranchInventory{}, domain.ErrInventoryNotFound
	}
	return items[0], nil
}

type movementInput struct {
	BusinessID     int64
	OwnerID        int64
	BranchID       int64
	ProductID      int64
	EntryType      stock_core.EntryType
	DisplayType    string
	Change         float64
	ReferenceID    *int64
	ReferenceType  string
	Note           string
	EmployeeID     *int64
	TransferGroup  *int64
	AllowNegative  bool
	SkipProductLog bool
}

func (r *Repository) applySingleMovement(ctx context.Context, input movementInput) (id int64, err error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin POS stock movement", logger.Error(err))
		return 0, err
	}
	defer rollbackOnError(tx, &err)

	id, err = r.applyMovementTx(ctx, tx, input)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit POS stock movement", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) applyMovementTx(ctx context.Context, tx *sql.Tx, input movementInput) (int64, error) {
	if err := r.ensureInventoryProduct(ctx, tx, input.OwnerID, input.ProductID); err != nil {
		return 0, err
	}

	locationID, err := stock_core.POSBranchLocationID(ctx, tx, input.OwnerID, input.BranchID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve POS branch stock location", logger.Error(err))
		return 0, err
	}

	referenceType := input.ReferenceType
	if referenceType == "" {
		referenceType = "pos_stock_movement"
	}
	_, _, err = stock_core.ApplyMovementTx(ctx, tx, stock_core.Movement{
		OwnerID:       input.OwnerID,
		LocationID:    locationID,
		ProductID:     input.ProductID,
		EntryType:     input.EntryType,
		Change:        input.Change,
		ReferenceID:   input.ReferenceID,
		ReferenceType: referenceType,
		Channel:       "pos",
		Note:          input.Note,
		EmployeeID:    input.EmployeeID,
		CreatedBy:     input.EmployeeID,
		TransferGroup: input.TransferGroup,
		AllowNegative: input.AllowNegative,
	})
	if errors.Is(err, stock_core.ErrInsufficientStock) {
		return 0, domain.ErrInsufficientStock
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to apply POS stock movement", logger.Error(err))
		return 0, err
	}

	ledgerID, err := lastStockLedgerID(ctx, tx, input.OwnerID, locationID, input.ProductID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read POS stock ledger id", logger.Error(err))
		return 0, err
	}
	return ledgerID, nil
}

func (r *Repository) ensureInventoryProduct(ctx context.Context, tx *sql.Tx, ownerID, productID int64) error {
	var exists int
	err := tx.QueryRowContext(ctx, `
		SELECT 1
		FROM erp_products
		WHERE owner_id = ?
		  AND id = ?
		  AND is_active = TRUE
		  AND track_inventory = TRUE
		LIMIT 1
	`, ownerID, productID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrInventoryNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to validate POS stock product", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ownerIDForBusiness(ctx context.Context, businessID int64) int64 {
	if businessID <= 0 {
		return businessID
	}
	var ownerID int64
	err := r.pool.QueryRowContext(ctx, `
		SELECT owner_id
		FROM businesses
		WHERE id = ?
		  AND is_active = TRUE
		LIMIT 1
	`, businessID).Scan(&ownerID)
	if err != nil || ownerID <= 0 {
		return businessID
	}
	return ownerID
}

func inventorySelectSQL(whereClause string) string {
	return `
		SELECT
			p.id AS id,
			p.id AS product_id,
			p.name,
			p.sku,
			p.unit,
			COALESCE(sl.on_hand, 0) AS quantity,
			COALESCE(sl.reorder_point, p.low_stock_threshold) AS reorder_point,
			COALESCE(sl.reorder_quantity, p.reorder_quantity) AS reorder_quantity,
			p.low_stock_threshold,
			COALESCE(sl.updated_at, p.updated_at) AS updated_at
		FROM erp_products p
		LEFT JOIN stock_locations loc
		  ON loc.owner_id = p.owner_id
		 AND loc.source_table = 'pos_locations'
		 AND loc.source_id = ?
		LEFT JOIN stock_levels sl
		  ON sl.owner_id = p.owner_id
		 AND sl.location_id = loc.id
		 AND sl.product_id = p.id
	` + whereClause
}

func scanBranchInventoryRows(ctx context.Context, rows *sql.Rows, businessID, branchID int64) ([]domain.BranchInventory, error) {
	items := make([]domain.BranchInventory, 0)
	for rows.Next() {
		var (
			item              domain.BranchInventory
			sku               sql.NullString
			unit              sql.NullString
			quantity          string
			reorderPoint      sql.NullString
			reorderQuantity   sql.NullString
			lowStockThreshold sql.NullString
			updatedAt         sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Name,
			&sku,
			&unit,
			&quantity,
			&reorderPoint,
			&reorderQuantity,
			&lowStockThreshold,
			&updatedAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan branch inventory", logger.Error(err))
			return nil, err
		}
		item.BusinessID = businessID
		item.BranchID = branchID
		if sku.Valid {
			item.SKU = sku.String
		}
		if unit.Valid && strings.TrimSpace(unit.String) != "" {
			item.Unit = unit.String
		} else {
			item.Unit = "units"
		}
		item.Quantity = decimalToFloat(quantity)
		if reorderPoint.Valid {
			rp := decimalToFloat(reorderPoint.String)
			item.ReorderPoint = &rp
		}
		if reorderQuantity.Valid {
			rq := decimalToFloat(reorderQuantity.String)
			item.ReorderQuantity = &rq
		}
		if lowStockThreshold.Valid {
			threshold := decimalToFloat(lowStockThreshold.String)
			item.LowStockThreshold = &threshold
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		logger.Ctx(ctx).Error("Failed to iterate branch inventory", logger.Error(err))
		return nil, err
	}
	return items, nil
}

func lastStockLedgerID(ctx context.Context, tx *sql.Tx, ownerID, locationID, productID int64) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM stock_ledger
		WHERE owner_id = ?
		  AND location_id = ?
		  AND product_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, ownerID, locationID, productID).Scan(&id)
	return id, err
}

func rollbackOnError(tx *sql.Tx, err *error) {
	if err != nil && *err != nil {
		_ = tx.Rollback()
	}
}

func posMovementType(entryType string) string {
	switch entryType {
	case string(stock_core.EntryPurchase):
		return "receive"
	case string(stock_core.EntryAdjustment):
		return "adjust"
	default:
		return entryType
	}
}

func entryTypeForMovement(movementType string) stock_core.EntryType {
	switch movementType {
	case "receive":
		return stock_core.EntryPurchase
	case "adjust":
		return stock_core.EntryAdjustment
	case "waste":
		return stock_core.EntryWaste
	case "transfer_in":
		return stock_core.EntryTransferIn
	case "transfer_out":
		return stock_core.EntryTransferOut
	case "sale":
		return stock_core.EntrySale
	case "return":
		return stock_core.EntryReturn
	default:
		return stock_core.EntryAdjustment
	}
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
		logger.Global().Warn("failed to parse decimal from database",
			logger.String("value", str), logger.Error(err))
		return 0
	}
	return f
}

func floatToDecimal(f float64) string {
	return fmt.Sprintf("%.3f", f)
}

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
