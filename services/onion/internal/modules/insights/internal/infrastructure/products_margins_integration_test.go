package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetMarginDistribution_ClassifiesBoundariesInDatabase(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := time.Now().UTC().UnixNano()
	prices := [][2]float64{
		{101, 100}, // <0%
		{100, 100}, // 0%
		{90, 100},  // 10%
		{75, 100},  // 25%
		{50, 100},  // 50%
		{25, 100},  // 75%
		{0, 0},     // excluded zero-price product
	}
	for i, price := range prices {
		_, err := db.ExecContext(ctx,
			`INSERT INTO erp_products (owner_id, business_id, name, cost_price, sale_price, is_active)
			 VALUES (?, ?, ?, ?, ?, TRUE)`,
			businessID, businessID, fmt.Sprintf("margin-boundary-%d", i), price[0], price[1])
		if err != nil {
			t.Fatalf("insert product %d: %v", i, err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_products WHERE business_id = ?", businessID)
	})

	rows, err := infrastructure.NewRepository(db).GetMarginDistribution(ctx, businessID)
	if err != nil {
		t.Fatalf("GetMarginDistribution: %v", err)
	}
	if len(rows) != 6 {
		t.Fatalf("bucket count = %d, want 6", len(rows))
	}
	for _, row := range rows {
		if row.ProductCount != 1 {
			t.Errorf("bucket %q count = %d, want 1", row.Label, row.ProductCount)
		}
	}
}

// TestGetProductMarginRows_ReturnsUnitsAndMinorUnitRevenue pins B55-R04: the
// DECIMAL SUMs must reach the response as the sold quantity and as cents,
// matching cost_minor_units in the same message.
func TestGetProductMarginRows_ReturnsUnitsAndMinorUnitRevenue(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := time.Now().UTC().UnixNano()

	res, err := db.ExecContext(ctx,
		`INSERT INTO erp_products (owner_id, business_id, name, cost_price, sale_price, is_active)
		 VALUES (?, ?, 'margin-row-product', 4.00, 10.00, TRUE)`,
		businessID, businessID)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	productID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("product id: %v", err)
	}

	res, err = db.ExecContext(ctx,
		`INSERT INTO pos_orders (owner_id, business_id, order_number, total, payment_status)
		 VALUES (?, ?, 'margin-row-order', 25.00, 'paid')`,
		businessID, businessID)
	if err != nil {
		t.Fatalf("insert order: %v", err)
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("order id: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO pos_order_items (order_id, menu_item_id, name, quantity, unit_price, total)
		 VALUES (?, ?, 'margin-row-product', 2.500, 10.00, 25.00)`,
		orderID, productID); err != nil {
		t.Fatalf("insert order item: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM pos_order_items WHERE order_id = ?", orderID)
		_, _ = db.ExecContext(ctx, "DELETE FROM pos_orders WHERE id = ?", orderID)
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_products WHERE id = ?", productID)
	})

	rows, err := infrastructure.NewRepository(db).GetProductMarginRows(ctx, businessID)
	if err != nil {
		t.Fatalf("GetProductMarginRows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.UnitsSold != 2.5 {
		t.Errorf("UnitsSold = %v, want 2.5", row.UnitsSold)
	}
	if row.RevenueMinor != 2500 {
		t.Errorf("RevenueMinor = %d, want 2500 (SUM(total) in cents)", row.RevenueMinor)
	}
	if row.CostMinorUnits == nil || *row.CostMinorUnits != 400 {
		t.Errorf("CostMinorUnits = %v, want 400", row.CostMinorUnits)
	}
}
