package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/domain"
)

func TestPriceOrderInput_NegativeShippingCost(t *testing.T) {
	r := NewRepository(nil)
	input := domain.StoreOrderInput{
		ShippingCost: -5.00,
		Items: []domain.OrderItemInput{
			{ProductID: 1, Quantity: 1},
		},
	}
	_, err := r.priceOrderInput(context.Background(), nil, 1, input)
	if !errors.Is(err, domain.ErrInvalidAmount) {
		t.Fatalf("priceOrderInput with negative shipping = %v, want ErrInvalidAmount", err)
	}
}
