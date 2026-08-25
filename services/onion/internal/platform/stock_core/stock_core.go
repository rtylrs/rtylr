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

// ApplyMovementTx applies one movement inside the caller's transaction: it
// row-locks (and creates if absent) the (owner, location, product) stock_levels
// row, adjusts on_hand by Change, rejects a negative result unless
// AllowNegative, writes the ledger row, and re-syncs the product's scalar
// cache to the owner-wide SUM of on_hand. Returns the before/after on_hand at
// the location. MUST be called with a *sql.Tx.
func ApplyMovementTx(ctx context.Context, tx DBTX, m Movement) (before, after float64, err error) {
	q := queries(tx)

	// Ensure the level row exists so FOR UPDATE has something to lock.
	if err = q.EnsureStockLevelRow(ctx, stocksqlc.EnsureStockLevelRowParams{
		OwnerID:    m.OwnerID,
		BusinessID: movementBusinessID(m),
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	}); err != nil {
		return 0, 0, fmt.Errorf("stock_core: ensure level row: %w", err)
	}

	onHandStr, err := q.LockStockLevelOnHand(ctx, stocksqlc.LockStockLevelOnHandParams{
		OwnerID:    m.OwnerID,
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	})
	if err != nil {
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
	if _, _, err = applyKnown(ctx, q, m, before, after); err != nil {
		return 0, 0, err
	}
	if m.EntryType == EntrySale && m.Change < 0 {
		if err := consumeCostLayersTx(ctx, tx, m, -m.Change); err != nil {
			return 0, 0, err
		}
	}
	return before, after, nil
}

// consumeCostLayersTx depletes receipt layers in FIFO order for a sale. Stock
// movement and valuation updates share the caller's transaction, so a failed
// layer update rolls back the corresponding stock movement as well. Missing
// layers are tolerated for legacy stock that predates cost-layer tracking.
func consumeCostLayersTx(ctx context.Context, tx DBTX, m Movement, quantity float64) error {
	if quantity <= 0 {
		return nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, quantity_remaining
		FROM inventory_cost_layers
		WHERE owner_id = ?
		  AND business_id = ?
		  AND product_id = ?
		  AND location_id = ?
		  AND quantity_remaining > 0
		ORDER BY received_at ASC, id ASC
		FOR UPDATE`,
		m.OwnerID, movementBusinessID(m), m.ProductID, m.LocationID,
	)
	if err != nil {
		return fmt.Errorf("stock_core: lock cost layers: %w", err)
	}
	defer rows.Close()

	type costLayer struct {
		id        int64
		available float64
	}
	layers := make([]costLayer, 0)
	for rows.Next() {
		var layerID int64
		var layerQuantity string
		if err := rows.Scan(&layerID, &layerQuantity); err != nil {
			return fmt.Errorf("stock_core: scan cost layer: %w", err)
		}
		available, err := parseQty(ctx, layerQuantity)
		if err != nil {
			return err
		}
		layers = append(layers, costLayer{id: layerID, available: available})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("stock_core: iterate cost layers: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("stock_core: close cost layers: %w", err)
	}
	available := make([]float64, len(layers))
	for i, layer := range layers {
		available[i] = layer.available
	}
	consumption := fifoLayerConsumption(available, quantity)
	for i, consumed := range consumption {
		if consumed <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE inventory_cost_layers
			SET quantity_remaining = quantity_remaining - ?
			WHERE id = ?`, consumed, layers[i].id); err != nil {
			return fmt.Errorf("stock_core: deplete cost layer %d: %w", layers[i].id, err)
		}
	}
	return nil
}

func fifoLayerConsumption(available []float64, quantity float64) []float64 {
	consumption := make([]float64, len(available))
	if quantity <= 0 {
		return consumption
	}
	remaining := quantity
	for i, layerQuantity := range available {
		if remaining <= 0 {
			break
		}
		if layerQuantity <= 0 {
			continue
		}
		consumed := layerQuantity
		if consumed > remaining {
			consumed = remaining
		}
		consumption[i] = consumed
		remaining -= consumed
	}
	return consumption
}

// applyKnown writes the level update, ledger row, and scalar-cache sync given a
// pre-locked level row and its before/after values. m.Change must equal
// after-before. Used by both ApplyMovementTx and SetLevelTx.
func applyKnown(ctx context.Context, q *stocksqlc.Queries, m Movement, before, after float64) (float64, float64, error) {
	if err := q.UpdateStockLevelOnHand(ctx, stocksqlc.UpdateStockLevelOnHandParams{
		OnHand:     fmtQty(after),
		OwnerID:    m.OwnerID,
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	}); err != nil {
		return 0, 0, fmt.Errorf("stock_core: update level: %w", err)
	}

	channel := m.Channel
	if channel == "" {
		channel = "erp"
	}
	if err := q.InsertStockLedgerEntry(ctx, stocksqlc.InsertStockLedgerEntryParams{
		OwnerID:        m.OwnerID,
		BusinessID:     movementBusinessID(m),
		LocationID:     m.LocationID,
		ProductID:      m.ProductID,
		EntryType:      stocksqlc.StockLedgerEntryType(m.EntryType),
		QuantityChange: fmtQty(m.Change),
		QuantityBefore: fmtQty(before),
		QuantityAfter:  fmtQty(after),
		ReferenceID:    nullInt(m.ReferenceID),
		ReferenceType:  nullStr(m.ReferenceType),
		TransferGroup:  nullInt(m.TransferGroup),
		Channel:        stocksqlc.StockLedgerChannel(channel),
		Note:           nullStr(m.Note),
		EmployeeID:     nullInt(m.EmployeeID),
		CreatedBy:      nullInt(m.CreatedBy),
	}); err != nil {
		return 0, 0, fmt.Errorf("stock_core: insert ledger: %w", err)
	}

	// Keep the legacy scalar synced by applying the SAME delta, NOT by recomputing
	// SUM(on_hand). An incremental add takes a write lock on the erp_products row,
	// which serializes concurrent same-product movements, and adds to the latest
	// committed value — so it's immune to REPEATABLE-READ snapshot staleness (a
	// SUM subquery could read a stale snapshot and clobber a concurrent write).
	// The scalar therefore stays == SUM(on_hand) by construction.
	if err := q.SyncProductStockScalar(ctx, stocksqlc.SyncProductStockScalarParams{
		StockQuantity: fmtQty(after - before),
		OwnerID:       m.OwnerID,
		ID:            m.ProductID,
	}); err != nil {
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

	q := queries(tx)

	if err = q.EnsureStockLevelRow(ctx, stocksqlc.EnsureStockLevelRowParams{
		OwnerID:    m.OwnerID,
		BusinessID: movementBusinessID(m),
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	}); err != nil {
		return 0, 0, fmt.Errorf("stock_core: ensure level row: %w", err)
	}

	onHandStr, err := q.LockStockLevelOnHand(ctx, stocksqlc.LockStockLevelOnHandParams{
		OwnerID:    m.OwnerID,
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("stock_core: lock level row: %w", err)
	}

	before, err = parseQty(ctx, onHandStr)
	if err != nil {
		return 0, 0, err
	}
	delta = target - before
	m.Change = delta
	if _, _, err = applyKnown(ctx, q, m, before, target); err != nil {
		return before, 0, err
	}
	return before, delta, nil
}

// SetTotalTx sets the product's owner-wide stock total to target by applying
// the variance to one location. Every stock-level row for the product is
// locked in location order before the product scalar is updated, which keeps
// absolute stock edits from clobbering concurrent movements or acquiring locks
// in the opposite order from ApplyMovementTx.
func SetTotalTx(ctx context.Context, tx DBTX, m Movement, target float64) (before, delta float64, err error) {
	if target < 0 && !m.AllowNegative {
		return 0, 0, ErrInsufficientStock
	}
	if m.LocationID <= 0 {
		return 0, 0, fmt.Errorf("stock_core: location is required for total adjustment")
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO stock_levels (owner_id, business_id, location_id, product_id, on_hand)
		VALUES (?, ?, ?, ?, 0)
		ON DUPLICATE KEY UPDATE on_hand = on_hand`,
		m.OwnerID, movementBusinessID(m), m.LocationID, m.ProductID,
	); err != nil {
		return 0, 0, fmt.Errorf("stock_core: ensure total level row: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT location_id, on_hand
		FROM stock_levels
		WHERE owner_id = ? AND product_id = ?
		ORDER BY location_id
		FOR UPDATE`, m.OwnerID, m.ProductID)
	if err != nil {
		return 0, 0, fmt.Errorf("stock_core: lock total levels: %w", err)
	}
	defer rows.Close()

	var currentLocation float64
	locationFound := false
	for rows.Next() {
		var locationID int64
		var onHand string
		if err := rows.Scan(&locationID, &onHand); err != nil {
			return 0, 0, fmt.Errorf("stock_core: scan total level: %w", err)
		}
		quantity, parseErr := parseQty(ctx, onHand)
		if parseErr != nil {
			return 0, 0, parseErr
		}
		before += quantity
		if locationID == m.LocationID {
			currentLocation = quantity
			locationFound = true
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("stock_core: iterate total levels: %w", err)
	}
	if !locationFound {
		return 0, 0, fmt.Errorf("stock_core: total adjustment location %d was not created", m.LocationID)
	}

	// Confirm the product exists before inserting a ledger entry. This query is
	// intentionally after the level locks so all stock writers use level →
	// product lock order.
	var productID int64
	if err := tx.QueryRowContext(ctx,
		"SELECT id FROM erp_products WHERE owner_id = ? AND id = ? FOR UPDATE",
		m.OwnerID, m.ProductID,
	).Scan(&productID); err != nil {
		return 0, 0, fmt.Errorf("stock_core: lock product for total adjustment: %w", err)
	}

	delta = target - before
	if delta == 0 {
		return before, 0, nil
	}

	m.Change = delta
	if _, _, err = applyKnown(ctx, queries(tx), m, currentLocation, currentLocation+delta); err != nil {
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
