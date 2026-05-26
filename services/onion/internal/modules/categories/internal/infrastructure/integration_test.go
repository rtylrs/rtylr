package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/categories/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListCategories_ReturnsFixtureRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(categories) < 2 {
		t.Fatalf("expected at least 2 categories, got %d", len(categories))
	}

	var foundBeverages, foundMains bool
	for _, c := range categories {
		switch c.Name {
		case "Beverages":
			foundBeverages = true
		case "Mains":
			foundMains = true
		}
	}
	if !foundBeverages {
		t.Errorf("expected category 'Beverages', not found in %+v", categories)
	}
	if !foundMains {
		t.Errorf("expected category 'Mains', not found in %+v", categories)
	}
}
