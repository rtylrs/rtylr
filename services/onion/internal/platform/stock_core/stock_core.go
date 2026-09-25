// Package stock_core is the single write primitive for on-hand inventory,
// shared by every channel (ERP modules and POS) so a movement on one side is
// immediately visible to the other. It owns three tables — stock_locations,
// stock_levels (per-location on_hand, the row-lock target), and stock_ledger
// (append-only) — and keeps the legacy erp_products.stock_quantity scalar in
// sync as a denormalized cache (= SUM of on_hand across the product's
// locations) so existing readers keep working until the read cutover.
//
// It is infrastructure-only (no HTTP surface) and uses sqlc-generated queries
// against a DBTX, so callers pass their own *sql.Tx and the stock change
// commits atomically with the caller's own writes (PO receive, POS payment …).
package stock_core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/platform/logger"
	stocksqlc "github.com/voxire/onion/internal/platform/stock_core/persistence/sqlc"
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

// Movement describes one signed stock change at a single location.
type Movement struct {
	OwnerID       int64
	BusinessID    int64 // denormalized tenant id; defaults to OwnerID for legacy callers
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
}

func fmtQty(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func movementBusinessID(m Movement) int64 {
	if m.BusinessID != 0 {
		return m.BusinessID
	}
	return m.OwnerID
}

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

// DefaultLocationID returns the tenant's default warehouse location, creating a
// native one on first use. Concurrency-safe via the (owner_id, source_table,
// source_id) unique key on stock_locations.
func DefaultLocationID(ctx context.Context, tx DBTX, ownerID int64) (int64, error) {
	q := queries(tx)

	id, err := q.GetDefaultWarehouseLocationID(ctx, ownerID)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup default location: %w", err)
	}

	// Insert-or-ignore on the unique key, then re-read (handles a concurrent creator).
	if err := q.InsertDefaultWarehouseLocation(ctx, stocksqlc.InsertDefaultWarehouseLocationParams{
		OwnerID:    ownerID,
		BusinessID: ownerID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: create default location: %w", err)
	}
	id, err = q.GetDefaultWarehouseLocationID(ctx, ownerID)
	if err != nil {
		return 0, fmt.Errorf("stock_core: re-read default location: %w", err)
	}
	return id, nil
}

// POSBranchLocationID resolves (creating on first use) the stock_locations row
// for a POS branch (pos_locations.id), scoped to ownerID. It is the
// tenant-isolation boundary for POS stock: if the pos_location belongs to a
// different owner it FAILS CLOSED (returns an error) rather than attributing
// stock to the wrong tenant. A zero/unknown posLocationID falls back to the
// tenant's default warehouse.
func POSBranchLocationID(ctx context.Context, tx DBTX, ownerID, posLocationID int64) (int64, error) {
	if posLocationID == 0 {
		return DefaultLocationID(ctx, tx, ownerID)
	}

	q := queries(tx)

	locOwner, err := q.GetPOSLocationOwnerID(ctx, posLocationID)
	if errors.Is(err, sql.ErrNoRows) {
		// Unknown location — don't guess a branch; use the default warehouse.
		return DefaultLocationID(ctx, tx, ownerID)
	}
	if err != nil {
		return 0, fmt.Errorf("stock_core: read pos_location owner: %w", err)
	}
	if locOwner != ownerID {
		return 0, fmt.Errorf("stock_core: pos_location %d belongs to owner %d, not %d (refusing to attribute stock)", posLocationID, locOwner, ownerID)
	}

	id, err := q.GetPOSBranchLocationID(ctx, stocksqlc.GetPOSBranchLocationIDParams{
		OwnerID:  ownerID,
		SourceID: posLocationID,
	})
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup pos branch location: %w", err)
	}

	if err = q.InsertPOSBranchLocation(ctx, stocksqlc.InsertPOSBranchLocationParams{
		OwnerID:    ownerID,
		BusinessID: ownerID,
		SourceID:   posLocationID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: create pos branch location: %w", err)
	}
	id, err = q.GetPOSBranchLocationID(ctx, stocksqlc.GetPOSBranchLocationIDParams{
		OwnerID:  ownerID,
		SourceID: posLocationID,
	})
	if err != nil {
		return 0, fmt.Errorf("stock_core: re-read pos branch location: %w", err)
	}
	return id, nil
}

// WarehouseLocationID resolves (creating on first use) the stock_locations row
// for an erp_warehouse, scoped to ownerID, with the same fail-closed owner
// guard as POSBranchLocationID. Zero or unknown warehouse IDs are rejected;
// callers must explicitly request DefaultLocationID when default routing is
// intended.
func WarehouseLocationID(ctx context.Context, tx DBTX, ownerID, warehouseID int64) (int64, error) {
	if warehouseID <= 0 {
		return 0, fmt.Errorf("%w: %d", ErrWarehouseNotFound, warehouseID)
	}

	q := queries(tx)

	whOwner, err := q.GetWarehouseOwnerID(ctx, warehouseID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: %d", ErrWarehouseNotFound, warehouseID)
	}
	if err != nil {
		return 0, fmt.Errorf("stock_core: read warehouse owner: %w", err)
	}
	if whOwner != ownerID {
		return 0, fmt.Errorf("stock_core: warehouse %d belongs to owner %d, not %d (refusing to attribute stock)", warehouseID, whOwner, ownerID)
	}

	id, err := q.GetWarehouseLocationID(ctx, stocksqlc.GetWarehouseLocationIDParams{
		OwnerID:  ownerID,
		SourceID: warehouseID,
	})
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup warehouse location: %w", err)
	}

	if err = q.InsertWarehouseLocation(ctx, stocksqlc.InsertWarehouseLocationParams{
		OwnerID:    ownerID,
		BusinessID: ownerID,
		ID:         warehouseID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: create warehouse location: %w", err)
	}
	id, err = q.GetWarehouseLocationID(ctx, stocksqlc.GetWarehouseLocationIDParams{
		OwnerID:  ownerID,
		SourceID: warehouseID,
	})
	if err != nil {
		return 0, fmt.Errorf("stock_core: re-read warehouse location: %w", err)
	}
	return id, nil
}
