package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// ResumeOrder replaces the editable lines of a held/pending order in one
// transaction. The order row is locked and tenant-scoped before any delete, so
// a stale or foreign id cannot replace another business's items.
func (r *Repository) ResumeOrder(ctx context.Context, businessID, staffID, orderID int64, input domain.OrderInput) error {
	if len(input.Items) == 0 {
		return errors.New("order must have at least one item")
	}
	for _, item := range input.Items {
		if item.Quantity <= 0 || item.UnitPrice < 0 {
			return errors.New("order item quantity and price must be valid")
		}
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var status, orderType string
	var paymentStatus, existingDiscount, locationID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT status, type, payment_status, discount_amount, COALESCE(CAST(location_id AS CHAR), '')
		FROM pos_orders
		WHERE business_id = ? AND id = ?
		FOR UPDATE`, businessID, orderID).Scan(
		&status, &orderType, &paymentStatus, &existingDiscount, &locationID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}
	if status != string(domain.OrderStatusHeld) &&
		status != string(domain.OrderStatusPending) &&
		status != string(domain.OrderStatusConfirmed) {
		return domain.ErrOrderNotEditable
	}
	if paymentStatus.Valid && paymentStatus.String != string(domain.PaymentStatusUnpaid) {
		return domain.ErrOrderNotEditable
	}
	if input.LocationID <= 0 && locationID.Valid {
		if parsed, parseErr := strconv.ParseInt(locationID.String, 10, 64); parseErr == nil {
			input.LocationID = parsed
		}
	}
	if input.Type == "" {
		input.Type = domain.OrderType(orderType)
	}
	if !domain.IsOrderType(input.Type) {
		return domain.ErrOrderNotEditable
	}

	qtx := r.queries.WithTx(tx)
	resolvedItems, err := r.resolveOrderItems(ctx, qtx, businessID, staffID, input)
	if err != nil {
		return err
	}
	totals := domain.ComputeTotals(resolvedItems)
	discount := sqlutil.DecimalToFloat(existingDiscount.String)
	if discount < 0 {
		discount = 0
	}
	if discount > totals.Subtotal {
		discount = totals.Subtotal
	}
	taxAmount := totals.TaxAmount
	if totals.Subtotal > 0 {
		taxAmount *= 1 - discount/totals.Subtotal
	}
	total := totals.Subtotal - discount + taxAmount
	if total < 0 {
		total = 0
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM pos_order_items WHERE order_id = ?`, orderID); err != nil {
		return err
	}
	for _, item := range resolvedItems {
		lineSubtotal := item.UnitPrice * item.Quantity
		lineDiscount := lineSubtotal * item.DiscountPct / 100
		if item.DiscountAmount > 0 {
			remaining := lineSubtotal - lineDiscount
			if item.DiscountAmount < remaining {
				lineDiscount += item.DiscountAmount
			} else {
				lineDiscount = lineSubtotal
			}
		}
		lineAfterDiscount := lineSubtotal - lineDiscount
		lineTotal := lineAfterDiscount + lineAfterDiscount*item.TaxRate/100
		modifierJSON, marshalErr := json.Marshal(item.ModifierIDs)
		if marshalErr != nil {
			return marshalErr
		}
		if _, err = qtx.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID:        orderID,
			MenuItemID:     sqlutil.NullInt64(item.MenuItemID),
			Name:           item.Name,
			Sku:            sqlutil.NullString(item.SKU),
			Quantity:       floatToDecimal3(item.Quantity),
			UnitPrice:      sqlutil.FloatToDecimal(item.UnitPrice),
			DiscountPct:    sql.NullString{String: sqlutil.FloatToDecimal(item.DiscountPct), Valid: true},
			DiscountAmount: sql.NullString{String: sqlutil.FloatToDecimal(lineDiscount), Valid: true},
			TaxRate:        sql.NullString{String: sqlutil.FloatToDecimal(item.TaxRate), Valid: true},
			Total:          sqlutil.FloatToDecimal(lineTotal),
			Notes:          sqlutil.NullString(item.Notes),
			Modifiers:      modifierJSON,
			ComboID:        sqlutil.NullInt64(item.ComboID),
		}); err != nil {
			return err
		}
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = 'pending', type = ?, customer_id = ?, customer_name = ?,
		    subtotal = ?, tax_amount = ?, total = ?, notes = ?, kitchen_notes = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE business_id = ? AND id = ?`,
		string(input.Type), sqlutil.NullInt64(input.CustomerID), sqlutil.NullString(input.CustomerName),
		sqlutil.FloatToDecimal(totals.Subtotal), sqlutil.FloatToDecimal(taxAmount), sqlutil.FloatToDecimal(total),
		sqlutil.NullString(input.Notes), sqlutil.NullString(input.KitchenNotes), businessID, orderID,
	); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *Repository) resolveOrderItems(ctx context.Context, qtx *sqlc.Queries, businessID, staffID int64, input domain.OrderInput) ([]domain.OrderItemInput, error) {
	items := make([]domain.OrderItemInput, len(input.Items))
	copy(items, input.Items)

	needsDiscountCapability := false
	for _, item := range items {
		if item.DiscountPct > 0 || item.DiscountAmount > 0 {
			needsDiscountCapability = true
			break
		}
	}

	var maxDiscountPct float64
	var canApplyDiscount bool
	if needsDiscountCapability {
		capability, err := qtx.GetPOSUserCapabilities(ctx, sqlc.GetPOSUserCapabilitiesParams{
			BusinessID:   businessID,
			UserID:       staffID,
			LocationID:   input.LocationID,
			Column4:      input.LocationID,
			LocationID_2: input.LocationID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, domain.ErrDiscountNotAllowed
			}
			return nil, err
		}
		canApplyDiscount = capability.CanApplyDiscount.Valid && capability.CanApplyDiscount.Bool
		maxDiscountPct = sqlutil.DecimalToFloat(capability.MaxDiscountPct.String)
		if maxDiscountPct > 100 {
			maxDiscountPct = 100
		}
	}

	for i := range items {
		item := &items[i]
		if item.DiscountPct < 0 || item.DiscountPct > 100 {
			return nil, domain.ErrDiscountExceedsLimit
		}
		if item.DiscountPct > 0 || item.DiscountAmount > 0 {
			if !canApplyDiscount {
				return nil, domain.ErrDiscountNotAllowed
			}
			if item.DiscountPct > maxDiscountPct {
				return nil, domain.ErrDiscountExceedsLimit
			}
			if item.DiscountAmount > 0 && maxDiscountPct > 0 && item.UnitPrice > 0 {
				flatPercent := item.DiscountAmount / (item.UnitPrice * item.Quantity) * 100
				if flatPercent > maxDiscountPct {
					return nil, domain.ErrDiscountExceedsLimit
				}
			}
		}

		modifierTotal, err := r.resolveModifierTotal(ctx, qtx, businessID, item.ModifierIDs)
		if err != nil {
			return nil, err
		}

		if item.ComboID > 0 {
			if item.MenuItemID > 0 {
				return nil, domain.ErrPricedItemNotFound
			}
			combo, comboErr := qtx.GetPOSComboForOrder(ctx, sqlc.GetPOSComboForOrderParams{
				BusinessID: businessID,
				ID:         item.ComboID,
			})
			if comboErr != nil {
				if errors.Is(comboErr, sql.ErrNoRows) {
					return nil, domain.ErrPricedItemNotFound
				}
				return nil, comboErr
			}
			item.Name = combo.Name
			item.SKU = ""
			item.UnitPrice = sqlutil.DecimalToFloat(combo.Price) + modifierTotal
			item.TaxRate = 0
		} else {
			if item.MenuItemID <= 0 {
				return nil, domain.ErrPricedItemNotFound
			}
			product, productErr := r.GetPublicMenuItem(ctx, businessID, item.MenuItemID)
			if productErr != nil {
				if errors.Is(productErr, domain.ErrPublicMenuItemNotFound) {
					return nil, domain.ErrPricedItemNotFound
				}
				return nil, productErr
			}
			item.Name = product.Name
			item.SKU = product.SKU
			item.UnitPrice = product.UnitPrice + modifierTotal
			item.TaxRate = product.TaxRate
		}
		if item.UnitPrice < 0 {
			return nil, domain.ErrPricedItemNotFound
		}
	}

	return items, nil
}

func (r *Repository) resolveModifierTotal(ctx context.Context, qtx *sqlc.Queries, businessID int64, modifierIDs []int64) (float64, error) {
	if len(modifierIDs) == 0 {
		return 0, nil
	}
	seen := make(map[int64]struct{}, len(modifierIDs))
	for _, id := range modifierIDs {
		if id <= 0 {
			return 0, domain.ErrPricedItemNotFound
		}
		if _, exists := seen[id]; exists {
			return 0, domain.ErrPricedItemNotFound
		}
		seen[id] = struct{}{}
	}

	row, err := qtx.GetPOSModifierTotal(ctx, sqlc.GetPOSModifierTotalParams{
		BusinessID:  businessID,
		ModifierIds: modifierIDs,
	})
	if err != nil {
		return 0, err
	}
	if row.ModifierCount != int64(len(modifierIDs)) {
		return 0, domain.ErrPricedItemNotFound
	}
	return interfaceToFloat(row.PriceDelta), nil
}
