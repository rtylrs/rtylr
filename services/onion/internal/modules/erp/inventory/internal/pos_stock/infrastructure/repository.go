package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	resolveOwner func(context.Context, int64) (int64, error)
	branchOwner  func(context.Context, int64, int64) (int64, error)
	queries      *sqlc.Queries
	pool         *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db), pool: db}
}

func (r *Repository) ListMovements(ctx context.Context, businessID, _, branchID, productID int64) ([]domain.StockMovement, error) {
	rows, err := r.queries.ListPOSStockMovements(ctx, sqlc.ListPOSStockMovementsParams{
		BusinessID: businessID,
		SourceID:   branchID,
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
			BusinessID: businessID,
			BranchID:   branchID,
			ProductID:  row.ProductID,
			Type:       posMovementType(string(row.EntryType)),
			Quantity:   sqlutil.DecimalToFloat(row.QuantityChange),
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
		if row.CreatedAt.Valid {
			m.CreatedAt = row.CreatedAt.Time
		}
		movements = append(movements, m)
	}
	return movements, nil
}

func (r *Repository) ReceiveStock(ctx context.Context, input domain.ReceiveStockInput) (int64, error) {
	return r.applySingleMovement(ctx, movementInput{
		BusinessID:  input.BusinessID,
		OwnerID:     input.OwnerID,
		BranchID:    input.BranchID,
		ProductID:   input.ProductID,
		EntryType:   stock.EntryPurchase,
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
		EntryType:   stock.EntryAdjustment,
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
		EntryType:   stock.EntryWaste,
		DisplayType: "waste",
		Change:      -input.Quantity,
		Note:        stringPtrValue(input.Note),
		EmployeeID:  input.EmployeeID,
	})
}

func (r *Repository) TransferStock(ctx context.Context, input domain.TransferStockInput) (int64, error) {
	ownerID, err := r.ownerIDForBusiness(ctx, input.BusinessID)
	if err != nil {
		return 0, err
	}
	input.OwnerID = ownerID
	for _, branchID := range []int64{input.FromBranchID, input.ToBranchID} {
		if err := r.validateBranch(ctx, input.BusinessID, ownerID, branchID); err != nil {
			return 0, err
		}
	}

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
		EntryType:   stock.EntryTransferOut,
		DisplayType: "transfer_out",
		Change:      -input.Quantity,
		Note:        stringPtrValue(input.Note),
		EmployeeID:  input.EmployeeID,
	})
	if err != nil {
		return 0, err
	}
	transferGroup := outID
	qtx := r.queries.WithTx(tx)
	if err = qtx.UpdateStockLedgerTransferGroup(ctx, sqlc.UpdateStockLedgerTransferGroupParams{
		TransferGroup: sql.NullInt64{Int64: transferGroup, Valid: true},
		ID:            outID,
		BusinessID:    input.BusinessID,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to mark POS transfer out group", logger.Error(err))
		return 0, err
	}

	inID, err := r.applyMovementTx(ctx, tx, movementInput{
		BusinessID:     input.BusinessID,
		OwnerID:        input.OwnerID,
		BranchID:       input.ToBranchID,
		ProductID:      input.ProductID,
		EntryType:      stock.EntryTransferIn,
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
	return r.applySingleMovement(ctx, movementInput{
		BusinessID:    businessID,
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
	_, err := r.applySingleMovement(ctx, movementInput{
		BusinessID:  businessID,
		BranchID:    branchID,
		ProductID:   productID,
		EntryType:   stock.EntryAdjustment,
		DisplayType: "adjust",
		Change:      quantityDelta,
	})
	return err
}

func (r *Repository) SetInventoryQuantity(ctx context.Context, businessID, branchID, productID int64, quantity float64) error {
	ownerID, err := r.ownerIDForBusiness(ctx, businessID)
	if err != nil {
		return err
	}
	if err := r.validateBranch(ctx, businessID, ownerID, branchID); err != nil {
		return err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin POS stock count", logger.Error(err))
		return err
	}
	defer rollbackOnError(tx, &err)

	if err = r.ensureInventoryProduct(ctx, tx, businessID, productID); err != nil {
		return err
	}
	locationID, err := stock.POSBranchLocationID(ctx, tx, ownerID, businessID, branchID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve POS branch stock location", logger.Error(err))
		return err
	}
	_, _, err = stock.SetLevelTx(ctx, tx, stock.Movement{
		OwnerID:       ownerID,
		BusinessID:    businessID,
		LocationID:    locationID,
		ProductID:     productID,
		EntryType:     stock.EntryCount,
		ReferenceType: "pos_stock_count",
		Channel:       "pos",
		AllowNegative: false,
	}, quantity)
	if errors.Is(err, stock.ErrInsufficientStock) {
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

func (r *Repository) ListBranchInventory(ctx context.Context, businessID, _, branchID int64) ([]domain.BranchInventory, error) {
	rows, err := r.queries.ListBranchInventory(ctx, sqlc.ListBranchInventoryParams{
		SourceID:   branchID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list branch inventory", logger.Error(err))
		return nil, err
	}
	return mapBranchInventoryRows(ctx, rows, businessID, branchID)
}

func (r *Repository) ListLowStockItems(ctx context.Context, businessID, _, branchID int64) ([]domain.BranchInventory, error) {
	rows, err := r.queries.ListLowStockBranchInventory(ctx, sqlc.ListLowStockBranchInventoryParams{
		SourceID:   branchID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list low stock items", logger.Error(err))
		return nil, err
	}
	converted := make([]sqlc.ListBranchInventoryRow, len(rows))
	for i, row := range rows {
		converted[i] = sqlc.ListBranchInventoryRow(row)
	}
	return mapBranchInventoryRows(ctx, converted, businessID, branchID)
}

func (r *Repository) GetInventoryItem(ctx context.Context, businessID, _, branchID, productID int64) (domain.BranchInventory, error) {
	rows, err := r.queries.GetBranchInventoryItem(ctx, sqlc.GetBranchInventoryItemParams{
		SourceID:   branchID,
		BusinessID: businessID,
		ID:         productID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory item", logger.Error(err))
		return domain.BranchInventory{}, err
	}
	converted := make([]sqlc.ListBranchInventoryRow, len(rows))
	for i, row := range rows {
		converted[i] = sqlc.ListBranchInventoryRow(row)
	}
	items, err := mapBranchInventoryRows(ctx, converted, businessID, branchID)
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
	EntryType      stock.EntryType
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
	input.OwnerID, err = r.ownerIDForBusiness(ctx, input.BusinessID)
	if err != nil {
		return 0, err
	}
	if err = r.validateBranch(ctx, input.BusinessID, input.OwnerID, input.BranchID); err != nil {
		return 0, err
	}

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
	if err := r.ensureInventoryProduct(ctx, tx, input.BusinessID, input.ProductID); err != nil {
		return 0, err
	}

	locationID, err := stock.POSBranchLocationID(ctx, tx, input.OwnerID, input.BusinessID, input.BranchID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve POS branch stock location", logger.Error(err))
		return 0, err
	}

	referenceType := input.ReferenceType
	if referenceType == "" {
		referenceType = "pos_stock_movement"
	}
	_, _, err = stock.ApplyMovementTx(ctx, tx, stock.Movement{
		OwnerID:       input.OwnerID,
		BusinessID:    input.BusinessID,
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
	if errors.Is(err, stock.ErrInsufficientStock) {
		return 0, domain.ErrInsufficientStock
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to apply POS stock movement", logger.Error(err))
		return 0, err
	}

	ledgerID, err := lastStockLedgerID(ctx, r.queries.WithTx(tx), input.BusinessID, locationID, input.ProductID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read POS stock ledger id", logger.Error(err))
		return 0, err
	}
	return ledgerID, nil
}

func (r *Repository) ensureInventoryProduct(ctx context.Context, tx *sql.Tx, businessID, productID int64) error {
	_, err := r.queries.WithTx(tx).InventoryProductExists(ctx, sqlc.InventoryProductExistsParams{
		BusinessID: businessID,
		ID:         productID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrInventoryNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to validate POS stock product", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) WithBusinessOwnerResolver(resolve func(context.Context, int64) (int64, error)) *Repository {
	r.resolveOwner = resolve
	return r
}

func (r *Repository) WithBranchOwnerReader(read func(context.Context, int64, int64) (int64, error)) *Repository {
	r.branchOwner = read
	return r
}

func (r *Repository) ownerIDForBusiness(ctx context.Context, businessID int64) (int64, error) {
	if r.resolveOwner == nil {
		return 0, fmt.Errorf("inventory requires a business owner resolver")
	}
	ownerID, err := r.resolveOwner(ctx, businessID)
	if err != nil {
		return 0, err
	}
	if ownerID <= 0 {
		return 0, fmt.Errorf("inventory received an invalid business owner")
	}
	return ownerID, nil
}

func (r *Repository) validateBranch(ctx context.Context, businessID, ownerID, branchID int64) error {
	if branchID == 0 {
		return nil
	}
	if r.branchOwner == nil {
		return fmt.Errorf("inventory requires a POS branch reader")
	}
	branchOwner, err := r.branchOwner(ctx, businessID, branchID)
	if err != nil {
		return err
	}
	if branchOwner != ownerID {
		return fmt.Errorf("POS branch %d is outside business %d", branchID, businessID)
	}
	return nil
}

func mapBranchInventoryRows(_ context.Context, rows []sqlc.ListBranchInventoryRow, businessID, branchID int64) ([]domain.BranchInventory, error) {
	items := make([]domain.BranchInventory, 0, len(rows))
	for _, row := range rows {
		item := domain.BranchInventory{
			ID:         row.ID,
			ProductID:  row.ProductID,
			BusinessID: businessID,
			BranchID:   branchID,
			Name:       row.Name,
			Quantity:   sqlutil.DecimalToFloat(row.Quantity),
		}
		if row.Sku.Valid {
			item.SKU = row.Sku.String
		}
		if row.Unit.Valid && strings.TrimSpace(row.Unit.String) != "" {
			item.Unit = row.Unit.String
		} else {
			item.Unit = "units"
		}
		if row.ReorderPoint.Valid {
			rp := sqlutil.DecimalToFloat(row.ReorderPoint.String)
			item.ReorderPoint = &rp
		}
		if row.ReorderQuantity.Valid {
			rq := sqlutil.DecimalToFloat(row.ReorderQuantity.String)
			item.ReorderQuantity = &rq
		}
		if row.LowStockThreshold.Valid {
			threshold := sqlutil.DecimalToFloat(row.LowStockThreshold.String)
			item.LowStockThreshold = &threshold
		}
		if row.UpdatedAt.Valid {
			item.UpdatedAt = row.UpdatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

func lastStockLedgerID(ctx context.Context, qtx *sqlc.Queries, businessID, locationID, productID int64) (int64, error) {
	return qtx.LastStockLedgerID(ctx, sqlc.LastStockLedgerIDParams{
		BusinessID: businessID,
		LocationID: locationID,
		ProductID:  productID,
	})
}

func rollbackOnError(tx *sql.Tx, err *error) {
	if err != nil && *err != nil {
		_ = tx.Rollback()
	}
}

func posMovementType(entryType string) string {
	switch entryType {
	case string(stock.EntryPurchase):
		return "receive"
	case string(stock.EntryAdjustment):
		return "adjust"
	default:
		return entryType
	}
}

func entryTypeForMovement(movementType string) stock.EntryType {
	switch movementType {
	case "receive":
		return stock.EntryPurchase
	case "adjust":
		return stock.EntryAdjustment
	case "waste":
		return stock.EntryWaste
	case "transfer_in":
		return stock.EntryTransferIn
	case "transfer_out":
		return stock.EntryTransferOut
	case "sale":
		return stock.EntrySale
	case "return":
		return stock.EntryReturn
	default:
		return stock.EntryAdjustment
	}
}

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
