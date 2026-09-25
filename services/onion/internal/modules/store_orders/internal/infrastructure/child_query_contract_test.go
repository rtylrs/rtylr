package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestStoreOrderItemsHaveServerBound(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/store_orders.sql")
	if err != nil {
		t.Fatalf("read store-order queries: %v", err)
	}
	if !strings.Contains(string(query), "ORDER BY store_order_items.id ASC\nLIMIT 500") {
		t.Fatal("store-order item reads must have a server-side bound")
	}
}
