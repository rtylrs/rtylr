package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListSuppliers_ReturnsFixtureRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	suppliers, _, err := repo.ListSuppliers(context.Background(), testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}
	if len(suppliers) == 0 {
		t.Fatal("expected at least one supplier, got none")
	}

	found := false
	for _, s := range suppliers {
		if s.Name == "Test Supplier" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find 'Test Supplier' in results, got: %v", suppliers)
	}
}

func TestGetSupplier_ReturnsCorrectNameAndActiveStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// List all and find supplier 5101
	suppliers, _, err := repo.ListSuppliers(context.Background(), testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	var found bool
	for _, s := range suppliers {
		if s.ID == 5101 {
			found = true
			if s.Name != "Test Supplier" {
				t.Errorf("expected name='Test Supplier', got %q", s.Name)
			}
			if !s.IsActive {
				t.Errorf("expected is_active=true for supplier 5101, got false")
			}
			break
		}
	}
	if !found {
		t.Fatal("supplier id=5101 not found in listing for owner 1001")
	}
}
