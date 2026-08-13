package infrastructure

import "testing"

func TestNullableDecimalPreservesDatabaseDecimalText(t *testing.T) {
	got := nullableDecimal([]byte("12.34"))
	if !got.Valid || got.String != "12.34" {
		t.Fatalf("nullableDecimal = %#v, want valid 12.34", got)
	}
}
