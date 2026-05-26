package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLocations_ReturnsFixtureRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locations, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locations) == 0 {
		t.Fatal("expected fixture rows, got none")
	}

	var found bool
	for _, loc := range locations {
		if loc.ID == 3001 && loc.Name == "Test Branch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected location id=3001 named 'Test Branch', got %+v", locations)
	}
}

func TestGetOrder_ReturnsFixturePaidOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	order, err := repo.GetOrder(ctx, testutil.OwnerID, 3301)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order == nil {
		t.Fatal("expected order, got nil")
	}
	if string(order.Status) != "paid" {
		t.Errorf("expected status='paid', got %q", order.Status)
	}
	const wantSubtotal = 150.00
	if order.Subtotal != wantSubtotal {
		t.Errorf("expected subtotal=%.2f, got %.2f", wantSubtotal, order.Subtotal)
	}
}

func TestListOrders_ReturnsBothFixtureOrders(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// sessionID=3201 has 2 orders in fixtures
	orders, err := repo.ListOrders(ctx, testutil.OwnerID, 3201, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) < 2 {
		t.Fatalf("expected at least 2 orders, got %d", len(orders))
	}
}
