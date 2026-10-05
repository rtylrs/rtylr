package stock

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	stocksqlc "github.com/voxire/onion/internal/modules/erp/internal/stock/persistence/sqlc"
)

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
	if m.ReferenceID != nil && m.ReferenceType != "" {
		var existingID int64
		err = tx.QueryRowContext(ctx, `
			SELECT id
			FROM stock_ledger
			WHERE owner_id = ? AND reference_type = ? AND reference_id = ?
			  AND entry_type = ? AND product_id = ? AND location_id = ?
			FOR UPDATE`,
			m.OwnerID, m.ReferenceType, *m.ReferenceID, string(m.EntryType), m.ProductID, m.LocationID,
		).Scan(&existingID)
		if err == nil {
			return before, before, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("stock_core: check movement reference: %w", err)
		}
	}

	after = before + m.Change
	if after < 0 && !m.AllowNegative {
		return before, before, ErrInsufficientStock
	}
	if _, _, err = applyKnown(ctx, q, m, before, after); err != nil {
		return 0, 0, err
	}
	if m.Change < 0 {
		if err := consumeCostLayersTx(ctx, tx, m, -m.Change); err != nil {
			return 0, 0, err
		}
	}
	return before, after, nil
}

// consumeCostLayersTx depletes receipt layers in FIFO order for a negative
// stock movement. Stock movement and valuation updates share the caller's
// transaction, so a failed layer update rolls back the corresponding stock
// movement as well. Missing layers are tolerated for legacy stock that
// predates cost-layer tracking.
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
