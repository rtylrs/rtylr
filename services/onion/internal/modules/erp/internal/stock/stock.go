// Package stock owns ERP inventory movements and their ledger transaction.
package stock

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	stocksqlc "github.com/voxire/onion/internal/modules/erp/internal/stock/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrInsufficientStock is returned when a movement would drive a location's
// on-hand below zero and the movement does not permit negative stock.
var ErrInsufficientStock = errors.New("insufficient stock at location")

var ErrWarehouseNotFound = errors.New("warehouse not found")

// DBTX is satisfied by *sql.DB and *sql.Tx. Stock mutations must be passed a
// *sql.Tx so they commit atomically with the caller's other writes.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// EntryType enumerates the ledger movement kinds.
type EntryType string

const (
	EntryPurchase    EntryType = "purchase"
	EntrySale        EntryType = "sale"
	EntryReturn      EntryType = "return"
	EntryAdjustment  EntryType = "adjustment"
	EntryWaste       EntryType = "waste"
	EntryCount       EntryType = "count"
	EntryTransferIn  EntryType = "transfer_in"
	EntryTransferOut EntryType = "transfer_out"
	EntryOpening     EntryType = "opening"
)

// Movement describes one signed stock change at a single location. ERP rows
// carry the business ID in owner_id as well as business_id, so BusinessID
// fills both.
type Movement struct {
	BusinessID    int64
	LocationID    int64
	ProductID     int64
	EntryType     EntryType
	Change        float64 // signed: +receive, -sale
	ReferenceID   *int64
	ReferenceType string
	Channel       string // "erp" | "pos" (provenance only)
	Note          string
	CreatedBy     *int64
	EmployeeID    *int64
	TransferGroup *int64
	AllowNegative bool
	// UnitCost values inbound stock for the FIFO cost layers. Nil uses the
	// product's cost price.
	UnitCost *float64
}

func fmtQty(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func parseQty(ctx context.Context, s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Ctx(ctx).Error("stock_core: failed to parse decimal", logger.String("value", s), logger.Error(err))
		return 0, err
	}
	return v, nil
}

func queries(tx DBTX) *stocksqlc.Queries {
	return stocksqlc.New(tx)
}

// DefaultLocationID returns the business's default warehouse location, creating
// a native one on first use. Concurrency-safe via the (owner_id, source_table,
// source_id) unique key on stock_locations, where owner_id is the business ID.
func DefaultLocationID(ctx context.Context, tx DBTX, businessID int64) (int64, error) {
	q := queries(tx)

	id, err := q.GetDefaultWarehouseLocationID(ctx, businessID)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup default location: %w", err)
	}

	// Insert-or-ignore on the unique key, then re-read (handles a concurrent creator).
	if err := q.InsertDefaultWarehouseLocation(ctx, stocksqlc.InsertDefaultWarehouseLocationParams{
		OwnerID:    businessID,
		BusinessID: businessID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: create default location: %w", err)
	}
	id, err = q.GetDefaultWarehouseLocationID(ctx, businessID)
	if err != nil {
		return 0, fmt.Errorf("stock_core: re-read default location: %w", err)
	}
	return id, nil
}

// POSBranchLocationID maps a branch the caller already confirmed belongs to
// businessID; it does that read before opening its ERP transaction.
func POSBranchLocationID(ctx context.Context, tx DBTX, businessID, posLocationID int64) (int64, error) {
	if posLocationID == 0 {
		return DefaultLocationID(ctx, tx, businessID)
	}
	q := queries(tx)
	id, err := q.GetPOSBranchLocationID(ctx, stocksqlc.GetPOSBranchLocationIDParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		SourceID:   posLocationID,
	})
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup pos branch location: %w", err)
	}

	if err = q.InsertPOSBranchLocation(ctx, stocksqlc.InsertPOSBranchLocationParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		SourceID:   posLocationID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: create pos branch location: %w", err)
	}
	id, err = q.GetPOSBranchLocationID(ctx, stocksqlc.GetPOSBranchLocationIDParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		SourceID:   posLocationID,
	})
	if err != nil {
		return 0, fmt.Errorf("stock_core: re-read pos branch location: %w", err)
	}
	return id, nil
}

// WarehouseLocationID resolves (creating on first use) the stock_locations row
// for an erp_warehouse of businessID, refusing a warehouse of any other
// business. Zero or unknown warehouse IDs are rejected; callers must
// explicitly request DefaultLocationID when default routing is intended.
func WarehouseLocationID(ctx context.Context, tx DBTX, businessID, warehouseID int64) (int64, error) {
	if warehouseID <= 0 {
		return 0, fmt.Errorf("%w: %d", ErrWarehouseNotFound, warehouseID)
	}

	q := queries(tx)

	whBusiness, err := q.GetWarehouseBusinessID(ctx, warehouseID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: %d", ErrWarehouseNotFound, warehouseID)
	}
	if err != nil {
		return 0, fmt.Errorf("stock_core: read warehouse business: %w", err)
	}
	if whBusiness != businessID {
		return 0, fmt.Errorf("stock_core: warehouse %d belongs to business %d, not %d (refusing to attribute stock)", warehouseID, whBusiness, businessID)
	}

	id, err := q.GetWarehouseLocationID(ctx, stocksqlc.GetWarehouseLocationIDParams{
		OwnerID:  businessID,
		SourceID: warehouseID,
	})
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup warehouse location: %w", err)
	}

	if err = q.InsertWarehouseLocation(ctx, stocksqlc.InsertWarehouseLocationParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		ID:         warehouseID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: create warehouse location: %w", err)
	}
	id, err = q.GetWarehouseLocationID(ctx, stocksqlc.GetWarehouseLocationIDParams{
		OwnerID:  businessID,
		SourceID: warehouseID,
	})
	if err != nil {
		return 0, fmt.Errorf("stock_core: re-read warehouse location: %w", err)
	}
	return id, nil
}
