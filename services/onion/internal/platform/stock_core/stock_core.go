// Package stock_core is the single write primitive for on-hand inventory,
// shared by every channel (ERP modules and POS) so a movement on one side is
// immediately visible to the other. It owns three tables — stock_locations,
// stock_levels (per-location on_hand, the row-lock target), and stock_ledger
// (append-only) — and keeps the legacy erp_products.stock_quantity scalar in
// sync as a denormalized cache (= SUM of on_hand across the product's
// locations) so existing readers keep working until the read cutover.
//
// It is infrastructure-only (no HTTP surface) and uses raw parameterized SQL
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
)

// ErrInsufficientStock is returned when a movement would drive a location's
// on-hand below zero and the movement does not permit negative stock.
var ErrInsufficientStock = errors.New("insufficient stock at location")

// DBTX is satisfied by *sql.DB and *sql.Tx. Stock mutations must be passed a
// *sql.Tx so they commit atomically with the caller's other writes.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
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

// DefaultLocationID returns the tenant's default warehouse location, creating a
// native one on first use. Concurrency-safe via the (owner_id, source_table,
// source_id) unique key on stock_locations.
func DefaultLocationID(ctx context.Context, tx DBTX, ownerID int64) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM stock_locations
		 WHERE owner_id = ? AND kind = 'warehouse' AND source_table = 'native' AND source_id = 0
		 LIMIT 1`, ownerID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup default location: %w", err)
	}

	// Insert-or-ignore on the unique key, then re-read (handles a concurrent creator).
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
		 VALUES (?, 0, 'warehouse', 'native', 0, 'Default Warehouse', TRUE, TRUE)
		 ON DUPLICATE KEY UPDATE id = id`, ownerID); err != nil {
		return 0, fmt.Errorf("stock_core: create default location: %w", err)
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM stock_locations
		 WHERE owner_id = ? AND kind = 'warehouse' AND source_table = 'native' AND source_id = 0
		 LIMIT 1`, ownerID).Scan(&id); err != nil {
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

	var locOwner int64
	err := tx.QueryRowContext(ctx, `SELECT owner_id FROM pos_locations WHERE id = ?`, posLocationID).Scan(&locOwner)
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

	var id int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM stock_locations
		 WHERE owner_id = ? AND source_table = 'pos_locations' AND source_id = ? LIMIT 1`,
		ownerID, posLocationID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup pos branch location: %w", err)
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
		 VALUES (?, 0, 'pos_branch', 'pos_locations', ?, 'POS Branch', FALSE, TRUE)
		 ON DUPLICATE KEY UPDATE source_id = VALUES(source_id)`,
		ownerID, posLocationID); err != nil {
		return 0, fmt.Errorf("stock_core: create pos branch location: %w", err)
	}
	if err = tx.QueryRowContext(ctx,
		`SELECT id FROM stock_locations
		 WHERE owner_id = ? AND source_table = 'pos_locations' AND source_id = ? LIMIT 1`,
		ownerID, posLocationID).Scan(&id); err != nil {
		return 0, fmt.Errorf("stock_core: re-read pos branch location: %w", err)
	}
	return id, nil
}

// WarehouseLocationID resolves (creating on first use) the stock_locations row
// for an erp_warehouse, scoped to ownerID, with the same fail-closed owner
// guard as POSBranchLocationID. A zero/unknown warehouseID falls back to the
// tenant's default warehouse.
func WarehouseLocationID(ctx context.Context, tx DBTX, ownerID, warehouseID int64) (int64, error) {
	if warehouseID == 0 {
		return DefaultLocationID(ctx, tx, ownerID)
	}

	var whOwner int64
	err := tx.QueryRowContext(ctx, `SELECT owner_id FROM erp_warehouses WHERE id = ?`, warehouseID).Scan(&whOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultLocationID(ctx, tx, ownerID)
	}
	if err != nil {
		return 0, fmt.Errorf("stock_core: read warehouse owner: %w", err)
	}
	if whOwner != ownerID {
		return 0, fmt.Errorf("stock_core: warehouse %d belongs to owner %d, not %d (refusing to attribute stock)", warehouseID, whOwner, ownerID)
	}

	var id int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM stock_locations
		 WHERE owner_id = ? AND source_table = 'erp_warehouses' AND source_id = ? LIMIT 1`,
		ownerID, warehouseID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("stock_core: lookup warehouse location: %w", err)
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
		 SELECT ?, 0, 'warehouse', 'erp_warehouses', id, name, FALSE, is_active FROM erp_warehouses WHERE id = ?
		 ON DUPLICATE KEY UPDATE source_id = VALUES(source_id)`,
		ownerID, warehouseID); err != nil {
		return 0, fmt.Errorf("stock_core: create warehouse location: %w", err)
	}
	if err = tx.QueryRowContext(ctx,
		`SELECT id FROM stock_locations
		 WHERE owner_id = ? AND source_table = 'erp_warehouses' AND source_id = ? LIMIT 1`,
		ownerID, warehouseID).Scan(&id); err != nil {
		return 0, fmt.Errorf("stock_core: re-read warehouse location: %w", err)
	}
	return id, nil
}

// ApplyMovementTx applies one movement inside the caller's transaction: it
// row-locks (and creates if absent) the (owner, location, product) stock_levels
// row, adjusts on_hand by Change, rejects a negative result unless
// AllowNegative, writes the ledger row, and re-syncs the product's scalar
// cache to the owner-wide SUM of on_hand. Returns the before/after on_hand at
// the location. MUST be called with a *sql.Tx.
func ApplyMovementTx(ctx context.Context, tx DBTX, m Movement) (before, after float64, err error) {
	// Ensure the level row exists so FOR UPDATE has something to lock.
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO stock_levels (owner_id, location_id, product_id, on_hand)
		 VALUES (?, ?, ?, 0) ON DUPLICATE KEY UPDATE on_hand = on_hand`,
		m.OwnerID, m.LocationID, m.ProductID); err != nil {
		return 0, 0, fmt.Errorf("stock_core: ensure level row: %w", err)
	}

	var onHandStr string
	if err = tx.QueryRowContext(ctx,
		`SELECT on_hand FROM stock_levels
		 WHERE owner_id = ? AND location_id = ? AND product_id = ? FOR UPDATE`,
		m.OwnerID, m.LocationID, m.ProductID).Scan(&onHandStr); err != nil {
		return 0, 0, fmt.Errorf("stock_core: lock level row: %w", err)
	}

	before, err = parseQty(ctx, onHandStr)
	if err != nil {
		return 0, 0, err
	}

	after = before + m.Change
	if after < 0 && !m.AllowNegative {
		return before, before, ErrInsufficientStock
	}
	if _, _, err = applyKnown(ctx, tx, m, before, after); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}

// applyKnown writes the level update, ledger row, and scalar-cache sync given a
// pre-locked level row and its before/after values. m.Change must equal
// after-before. Used by both ApplyMovementTx and SetLevelTx.
func applyKnown(ctx context.Context, tx DBTX, m Movement, before, after float64) (float64, float64, error) {
	if _, err := tx.ExecContext(ctx,
		`UPDATE stock_levels SET on_hand = ?
		 WHERE owner_id = ? AND location_id = ? AND product_id = ?`,
		fmtQty(after), m.OwnerID, m.LocationID, m.ProductID); err != nil {
		return 0, 0, fmt.Errorf("stock_core: update level: %w", err)
	}

	channel := m.Channel
	if channel == "" {
		channel = "erp"
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO stock_ledger
		 (owner_id, location_id, product_id, entry_type, quantity_change, quantity_before, quantity_after,
		  reference_id, reference_type, transfer_group, channel, note, employee_id, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.OwnerID, m.LocationID, m.ProductID, string(m.EntryType),
		fmtQty(m.Change), fmtQty(before), fmtQty(after),
		nullInt(m.ReferenceID), nullStr(m.ReferenceType), nullInt(m.TransferGroup),
		channel, nullStr(m.Note), nullInt(m.EmployeeID), nullInt(m.CreatedBy)); err != nil {
		return 0, 0, fmt.Errorf("stock_core: insert ledger: %w", err)
	}

	// Keep the legacy scalar synced by applying the SAME delta, NOT by recomputing
	// SUM(on_hand). An incremental add takes a write lock on the erp_products row,
	// which serializes concurrent same-product movements, and adds to the latest
	// committed value — so it's immune to REPEATABLE-READ snapshot staleness (a
	// SUM subquery could read a stale snapshot and clobber a concurrent write).
	// The scalar therefore stays == SUM(on_hand) by construction.
	if _, err := tx.ExecContext(ctx,
		`UPDATE erp_products SET stock_quantity = stock_quantity + ?
		 WHERE owner_id = ? AND id = ?`,
		fmtQty(after-before), m.OwnerID, m.ProductID); err != nil {
		return 0, 0, fmt.Errorf("stock_core: sync scalar cache: %w", err)
	}

	return before, after, nil
}

// SetLevelTx sets a location's on_hand to an absolute target (e.g. a stock-take
// physical count), ledgering the signed variance. Same transactional + cache-sync
// guarantees as ApplyMovementTx. A negative target is rejected.
func SetLevelTx(ctx context.Context, tx DBTX, m Movement, target float64) (before, delta float64, err error) {
	if target < 0 && !m.AllowNegative {
		return 0, 0, ErrInsufficientStock
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO stock_levels (owner_id, location_id, product_id, on_hand)
		 VALUES (?, ?, ?, 0) ON DUPLICATE KEY UPDATE on_hand = on_hand`,
		m.OwnerID, m.LocationID, m.ProductID); err != nil {
		return 0, 0, fmt.Errorf("stock_core: ensure level row: %w", err)
	}
	var onHandStr string
	if err = tx.QueryRowContext(ctx,
		`SELECT on_hand FROM stock_levels
		 WHERE owner_id = ? AND location_id = ? AND product_id = ? FOR UPDATE`,
		m.OwnerID, m.LocationID, m.ProductID).Scan(&onHandStr); err != nil {
		return 0, 0, fmt.Errorf("stock_core: lock level row: %w", err)
	}
	before, err = parseQty(ctx, onHandStr)
	if err != nil {
		return 0, 0, err
	}
	delta = target - before
	m.Change = delta
	if _, _, err = applyKnown(ctx, tx, m, before, target); err != nil {
		return before, 0, err
	}
	return before, delta, nil
}

func nullInt(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
