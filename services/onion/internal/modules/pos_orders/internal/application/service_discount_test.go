package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
)

func TestApplyDiscount_RejectsZeroOrderID(t *testing.T) {
	s := &Service{}
	err := s.ApplyDiscount(context.Background(), 1, domain.DiscountInput{
		DiscountType:  "pct",
		DiscountValue: 10,
	})
	if err == nil {
		t.Fatal("expected error for zero order id")
	}
}

func TestApplyDiscount_RejectsBlankType(t *testing.T) {
	s := &Service{}
	err := s.ApplyDiscount(context.Background(), 1, domain.DiscountInput{
		OrderID:       5,
		DiscountValue: 10,
	})
	if err == nil {
		t.Fatal("expected error for blank discount type")
	}
}

func TestApplyDiscount_RejectsNegativeValue(t *testing.T) {
	s := &Service{}
	err := s.ApplyDiscount(context.Background(), 1, domain.DiscountInput{
		OrderID:       5,
		DiscountType:  "pct",
		DiscountValue: -5,
	})
	if err == nil {
		t.Fatal("expected error for negative discount value")
	}
}
