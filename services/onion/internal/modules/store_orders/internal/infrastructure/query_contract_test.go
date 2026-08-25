package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestStoreOrderLegacyListHasServerBoundAndStableOrder(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/store_orders.sql")
	if err != nil {
		t.Fatalf("read store-order queries: %v", err)
	}
	sql := string(query)
	if !strings.Contains(sql, "ORDER BY created_at DESC, id DESC LIMIT 500") {
		t.Fatal("legacy store-order list must have a server-side bound and unique ordering")
	}
}

func TestStoreOrderWritesDoNotAcceptClientMoney(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/store_orders.sql")
	if err != nil {
		t.Fatalf("read store-order queries: %v", err)
	}
	sql := string(query)
	if strings.Contains(sql, "SET customer_email = ?, customer_name = ?, customer_phone = ?, status = ?, subtotal = ?") {
		t.Fatal("order updates must not overwrite server-derived monetary columns")
	}
	if !strings.Contains(sql, "GetStoreProductPricing") || !strings.Contains(sql, "sale_price") {
		t.Fatal("order creation must have a server-side catalog pricing query")
	}
}
