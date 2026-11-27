package infrastructure

import (
	"database/sql"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos/receipts/internal/infrastructure/persistence/sqlc"
)

func TestMapReceiptRow_FallsBackForUnpaidOnlineOrder(t *testing.T) {
	created := time.Date(2026, 11, 27, 10, 30, 0, 0, time.UTC)
	got := mapReceiptRow(
		42,
		"ORD-20261127-0001",
		sql.NullString{String: "Guest", Valid: true},
		sql.NullString{String: "12.50", Valid: true},
		sqlc.NullPosOrdersPaymentMethod{},
		sql.NullTime{},
		sql.NullTime{Time: created, Valid: true},
		1,
	)

	if !got.PaidAt.Equal(created) {
		t.Fatalf("expected PaidAt to fall back to CreatedAt %v, got %v", created, got.PaidAt)
	}
	if got.PaymentMethod != "online" {
		t.Fatalf("expected PaymentMethod fallback 'online', got %q", got.PaymentMethod)
	}
	if got.Total != 12.5 {
		t.Fatalf("expected Total 12.5, got %v", got.Total)
	}
}
