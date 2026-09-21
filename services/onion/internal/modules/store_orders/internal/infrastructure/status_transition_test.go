package infrastructure

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
)

// Confirming an order deducts stock; allowing a return to pending would let
// the next confirmation deduct the same quantity again.
func TestValidateOrderStatusTransitionRejectsReturnToPending(t *testing.T) {
	for _, current := range []string{"confirmed", "processing", "shipped", "delivered", "completed", "cancelled", "refunded"} {
		if err := validateOrderStatusTransition(current, "pending"); !errors.Is(err, domain.ErrInvalidStatus) {
			t.Errorf("%s -> pending error = %v, want ErrInvalidStatus", current, err)
		}
	}
}

func TestValidateOrderStatusTransitionAllowsLifecycleFlow(t *testing.T) {
	tests := []struct {
		current string
		next    string
		wantErr bool
	}{
		{"pending", "pending", false},
		{"pending", "confirmed", false},
		{"pending", "cancelled", false},
		{"pending", "refunded", false},
		{"pending", "shipped", true},
		{"confirmed", "confirmed", false},
		{"confirmed", "processing", false},
		{"confirmed", "shipped", false},
		{"confirmed", "delivered", false},
		{"confirmed", "completed", false},
		{"confirmed", "cancelled", false},
		{"confirmed", "refunded", false},
		{"processing", "shipped", false},
		{"processing", "confirmed", false},
		{"shipped", "delivered", false},
		{"delivered", "completed", false},
		{"completed", "cancelled", false},
		{"cancelled", "confirmed", false},
		{"cancelled", "pending", true},
		{"refunded", "confirmed", false},
		{"refunded", "pending", true},
		{"pending", "unknown", true},
	}

	for _, tc := range tests {
		err := validateOrderStatusTransition(tc.current, tc.next)
		if tc.wantErr {
			if !errors.Is(err, domain.ErrInvalidStatus) {
				t.Errorf("%s -> %s error = %v, want ErrInvalidStatus", tc.current, tc.next, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s -> %s error = %v, want nil", tc.current, tc.next, err)
		}
	}
}
