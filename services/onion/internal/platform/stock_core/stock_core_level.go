package stock_core

import (
	"context"
	"database/sql"
	"fmt"

	stocksqlc "github.com/voxire/onion/internal/platform/stock_core/persistence/sqlc"
)

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
	if delta < 0 {
		if err := consumeCostLayersTx(ctx, tx, m, -delta); err != nil {
			return before, 0, err
		}
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
	if currentLocation+delta < 0 && !m.AllowNegative {
		return before, 0, ErrInsufficientStock
	}

	m.Change = delta
	if _, _, err = applyKnown(ctx, queries(tx), m, currentLocation, currentLocation+delta); err != nil {
		return before, 0, err
	}
	if delta < 0 {
		if err := consumeCostLayersTx(ctx, tx, m, -delta); err != nil {
			return before, 0, err
		}
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
