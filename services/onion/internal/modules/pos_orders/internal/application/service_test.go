package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
)

// ProcessRefund's input guards run before any repository call, so we can
// exercise them with a zero-value Service and verify they reject malformed
// input before touching the database.

func TestProcessRefund_RejectsMissingOrderID(t *testing.T) {
	s := &Service{}
	if _, err := s.ProcessRefund(context.Background(), 1, 1, domain.RefundInput{Amount: 5}); err == nil {
		t.Fatalf("expected error for missing order id")
	} else if !strings.Contains(err.Error(), "order id") {
		t.Fatalf("expected order id error, got: %v", err)
	}
}

func TestProcessRefund_RejectsNonPositiveAmount(t *testing.T) {
	s := &Service{}
	for _, amount := range []float64{0, -1, -0.01} {
		if _, err := s.ProcessRefund(context.Background(), 1, 1, domain.RefundInput{OrderID: 7, Amount: amount}); err == nil {
			t.Fatalf("expected error for amount %v", amount)
		}
	}
}
