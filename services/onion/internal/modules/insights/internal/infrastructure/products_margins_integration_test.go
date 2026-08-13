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
