package stock

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	stocksqlc "github.com/voxire/onion/internal/modules/erp/internal/stock/persistence/sqlc"
)

// ApplyMovementTx applies one movement inside the caller's transaction: it
// row-locks (and creates if absent) the (business, location, product) stock_levels
// row, adjusts on_hand by Change, rejects a negative result unless
// AllowNegative, writes the ledger row, and re-syncs the product's scalar
// cache to the business-wide SUM of on_hand. Inbound stock opens a cost layer and
// outbound stock depletes layers in FIFO order. Returns the before/after
// on_hand at the location. MUST be called with a *sql.Tx.
func ApplyMovementTx(ctx context.Context, tx DBTX, m Movement) (before, after float64, err error) {
	before, after, _, applied, err := applyMovementLevelTx(ctx, tx, m)
	if err != nil || !applied {
		return before, after, err
	}
	if m.Change < 0 {
		if _, err := consumeCostLayersTx(ctx, tx, m, -m.Change); err != nil {
			return 0, 0, err
		}
		return before, after, nil
	}
	if err := openCostLayersTx(ctx, tx, m, inboundLayerQuantity(before, after), nil); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}

// applyMovementLevelTx is ApplyMovementTx without valuation. applied is false
// when the movement's reference was already ledgered, in which case nothing
// was written and ledgerID is the existing row.
func applyMovementLevelTx(ctx context.Context, tx DBTX, m Movement) (before, after float64, ledgerID int64, applied bool, err error) {
	q := queries(tx)

	// Ensure the level row exists so FOR UPDATE has something to lock.
	if err = q.EnsureStockLevelRow(ctx, stocksqlc.EnsureStockLevelRowParams{
		OwnerID:    m.BusinessID,
		BusinessID: m.BusinessID,
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	}); err != nil {
		return 0, 0, 0, false, fmt.Errorf("stock_core: ensure level row: %w", err)
	}

	onHandStr, err := q.LockStockLevelOnHand(ctx, stocksqlc.LockStockLevelOnHandParams{
		OwnerID:    m.BusinessID,
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	})
	if err != nil {
		return 0, 0, 0, false, fmt.Errorf("stock_core: lock level row: %w", err)
	}

	before, err = parseQty(ctx, onHandStr)
	if err != nil {
		return 0, 0, 0, false, err
	}
	if m.ReferenceID != nil && m.ReferenceType != "" {
		var existingID int64
		err = tx.QueryRowContext(ctx, `
			SELECT id
			FROM stock_ledger
			WHERE owner_id = ? AND reference_type = ? AND reference_id = ?
			  AND entry_type = ? AND product_id = ? AND location_id = ?
			FOR UPDATE`,
			m.BusinessID, m.ReferenceType, *m.ReferenceID, string(m.EntryType), m.ProductID, m.LocationID,
		).Scan(&existingID)
		if err == nil {
			return before, before, existingID, false, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, 0, 0, false, fmt.Errorf("stock_core: check movement reference: %w", err)
		}
	}

	after = before + m.Change
	if after < 0 && !m.AllowNegative {
		return before, before, 0, false, ErrInsufficientStock
	}
	if ledgerID, err = applyKnown(ctx, q, m, before, after); err != nil {
		return 0, 0, 0, false, err
	}
	return before, after, ledgerID, true, nil
}

// TransferTx moves quantity of one product from out.LocationID to
// in.LocationID. Both level rows are locked in location order before either
// leg touches the product row, so concurrent transfers in opposite directions
// cannot deadlock. The FIFO layers consumed at the source are re-opened at the
// destination with their original unit costs, so a transfer never changes the
// valuation of the stock it moves.
func TransferTx(ctx context.Context, tx DBTX, out, in Movement) (outLedgerID, inLedgerID int64, err error) {
	if out.ProductID != in.ProductID || out.BusinessID != in.BusinessID {
		return 0, 0, fmt.Errorf("stock_core: transfer legs must move the same product for the same business")
	}
	if out.LocationID == in.LocationID {
		return 0, 0, fmt.Errorf("stock_core: transfer source and destination must differ")
	}
	if out.Change >= 0 || in.Change != -out.Change {
		return 0, 0, fmt.Errorf("stock_core: transfer legs must move the same positive quantity")
	}

	q := queries(tx)
	legs := []Movement{out, in}
	if in.LocationID < out.LocationID {
		legs = []Movement{in, out}
	}
	for _, leg := range legs {
		if err := q.EnsureStockLevelRow(ctx, stocksqlc.EnsureStockLevelRowParams{
			OwnerID:    leg.BusinessID,
			BusinessID: leg.BusinessID,
			LocationID: leg.LocationID,
			ProductID:  leg.ProductID,
		}); err != nil {
			return 0, 0, fmt.Errorf("stock_core: ensure transfer level row: %w", err)
		}
		if _, err := q.LockStockLevelOnHand(ctx, stocksqlc.LockStockLevelOnHandParams{
			OwnerID:    leg.BusinessID,
			LocationID: leg.LocationID,
			ProductID:  leg.ProductID,
		}); err != nil {
			return 0, 0, fmt.Errorf("stock_core: lock transfer level row: %w", err)
		}
	}

	_, _, outLedgerID, outApplied, err := applyMovementLevelTx(ctx, tx, out)
	if err != nil {
		return 0, 0, err
	}
	var carried []CostSlice
	if outApplied {
		if carried, err = consumeCostLayersTx(ctx, tx, out, -out.Change); err != nil {
			return 0, 0, err
		}
	}
	inBefore, inAfter, inLedgerID, inApplied, err := applyMovementLevelTx(ctx, tx, in)
	if err != nil {
		return 0, 0, err
	}
	if inApplied {
		if err = openCostLayersTx(ctx, tx, in, inboundLayerQuantity(inBefore, inAfter), carried); err != nil {
			return 0, 0, err
		}
	}
	return outLedgerID, inLedgerID, nil
}

// CostSlice is a quantity drawn from one cost layer at that layer's unit cost.
type CostSlice struct {
	Quantity float64
	UnitCost float64
}

// consumeCostLayersTx depletes receipt layers in FIFO order for a negative
// stock movement and returns what it drew from each layer. Stock movement and
// valuation updates share the caller's transaction, so a failed layer update
// rolls back the corresponding stock movement as well. Missing layers are
// tolerated for legacy stock that predates cost-layer tracking.
func consumeCostLayersTx(ctx context.Context, tx DBTX, m Movement, quantity float64) ([]CostSlice, error) {
	if quantity <= 0 {
		return nil, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, quantity_remaining, unit_cost
		FROM inventory_cost_layers
		WHERE owner_id = ?
		  AND business_id = ?
		  AND product_id = ?
		  AND location_id = ?
		  AND quantity_remaining > 0
		ORDER BY received_at ASC, id ASC
		FOR UPDATE`,
		m.BusinessID, m.BusinessID, m.ProductID, m.LocationID,
	)
	if err != nil {
		return nil, fmt.Errorf("stock_core: lock cost layers: %w", err)
	}
	defer rows.Close()

	type costLayer struct {
		id        int64
		available float64
		unitCost  float64
	}
	layers := make([]costLayer, 0)
	for rows.Next() {
		var layerID int64
		var layerQuantity, layerCost string
		if err := rows.Scan(&layerID, &layerQuantity, &layerCost); err != nil {
			return nil, fmt.Errorf("stock_core: scan cost layer: %w", err)
		}
		available, err := parseQty(ctx, layerQuantity)
		if err != nil {
			return nil, err
		}
		unitCost, err := parseQty(ctx, layerCost)
		if err != nil {
			return nil, err
		}
		layers = append(layers, costLayer{id: layerID, available: available, unitCost: unitCost})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("stock_core: iterate cost layers: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("stock_core: close cost layers: %w", err)
	}
	available := make([]float64, len(layers))
	for i, layer := range layers {
		available[i] = layer.available
	}
	consumption := fifoLayerConsumption(available, quantity)
	slices := make([]CostSlice, 0, len(consumption))
	for i, consumed := range consumption {
		if consumed <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE inventory_cost_layers
			SET quantity_remaining = quantity_remaining - ?
			WHERE id = ?`, consumed, layers[i].id); err != nil {
			return nil, fmt.Errorf("stock_core: deplete cost layer %d: %w", layers[i].id, err)
		}
		slices = append(slices, CostSlice{Quantity: consumed, UnitCost: layers[i].unitCost})
	}
	return slices, nil
}

// inboundLayerQuantity is the part of an inbound movement that lands in
// positive stock. Units that only pay back a negative balance were sold
// before they were received and have no remaining value.
func inboundLayerQuantity(before, after float64) float64 {
	return math.Max(after, 0) - math.Max(before, 0)
}

// openCostLayersTx records quantity of inbound stock at the location. The
// carried slices (a transfer's source layers) are used first, at their own
// unit costs; whatever they don't cover is valued at m.UnitCost, or at the
// product's cost price when the movement has none.
func openCostLayersTx(ctx context.Context, tx DBTX, m Movement, quantity float64, carried []CostSlice) error {
	if quantity <= 0 {
		return nil
	}
	referenceType := m.ReferenceType
	if referenceType == "" {
		referenceType = string(m.EntryType)
	}
	insert := func(qty, unitCost float64) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inventory_cost_layers
			(owner_id, business_id, product_id, location_id, quantity_remaining, unit_cost, reference_id, reference_type)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			m.BusinessID, m.BusinessID, m.ProductID, m.LocationID,
			fmtQty(qty), fmtQty(unitCost), nullInt(m.ReferenceID), referenceType,
		); err != nil {
			return fmt.Errorf("stock_core: open cost layer: %w", err)
		}
		return nil
	}

	remaining := quantity
	for _, slice := range carried {
		if remaining <= 0 {
			return nil
		}
		qty := math.Min(slice.Quantity, remaining)
		if qty <= 0 {
			continue
		}
		if err := insert(qty, slice.UnitCost); err != nil {
			return err
		}
		remaining -= qty
	}
	if remaining <= 0 {
		return nil
	}

	unitCost := 0.0
	if m.UnitCost != nil {
		unitCost = *m.UnitCost
	} else {
		var cost sql.NullString
		err := tx.QueryRowContext(ctx,
			"SELECT cost_price FROM erp_products WHERE business_id = ? AND id = ?",
			m.BusinessID, m.ProductID,
		).Scan(&cost)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("stock_core: read product cost: %w", err)
		}
		if cost.Valid {
			if unitCost, err = parseQty(ctx, cost.String); err != nil {
				return err
			}
		}
	}
	if unitCost < 0 {
		unitCost = 0
	}
	return insert(remaining, unitCost)
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
func applyKnown(ctx context.Context, q *stocksqlc.Queries, m Movement, before, after float64) (int64, error) {
	if err := q.UpdateStockLevelOnHand(ctx, stocksqlc.UpdateStockLevelOnHandParams{
		OnHand:     fmtQty(after),
		OwnerID:    m.BusinessID,
		LocationID: m.LocationID,
		ProductID:  m.ProductID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: update level: %w", err)
	}

	channel := m.Channel
	if channel == "" {
		channel = "erp"
	}
	result, err := q.InsertStockLedgerEntry(ctx, stocksqlc.InsertStockLedgerEntryParams{
		OwnerID:        m.BusinessID,
		BusinessID:     m.BusinessID,
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
	})
	if err != nil {
		return 0, fmt.Errorf("stock_core: insert ledger: %w", err)
	}
	ledgerID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("stock_core: read ledger id: %w", err)
	}

	// Keep the legacy scalar synced by applying the SAME delta, NOT by recomputing
	// SUM(on_hand). An incremental add takes a write lock on the erp_products row,
	// which serializes concurrent same-product movements, and adds to the latest
	// committed value — so it's immune to REPEATABLE-READ snapshot staleness (a
	// SUM subquery could read a stale snapshot and clobber a concurrent write).
	// The scalar therefore stays == SUM(on_hand) by construction.
	if err := q.SyncProductStockScalar(ctx, stocksqlc.SyncProductStockScalarParams{
		StockQuantity: fmtQty(after - before),
		BusinessID:    m.BusinessID,
		ID:            m.ProductID,
	}); err != nil {
		return 0, fmt.Errorf("stock_core: sync scalar cache: %w", err)
	}

	return ledgerID, nil
}
