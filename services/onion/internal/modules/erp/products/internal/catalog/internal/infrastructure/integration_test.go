package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestMenu_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = items
}

func TestMenu_List_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestMenu_List_WrongBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.List(context.Background(), 999999)
	if err != nil {
		t.Fatalf("List wrong business: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for unknown business, got %d", len(items))
	}
}

func TestMenu_List_AllBelongToBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, item := range items {
		if item.BusinessID != testutil.BusinessID {
			t.Errorf("items[%d].BusinessID = %d, want %d", i, item.BusinessID, testutil.BusinessID)
		}
	}
}

func TestMenu_List_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, item := range items {
		if item.ID == 0 {
			t.Errorf("items[%d].ID is zero", i)
		}
	}
}

func TestMenu_List_AllHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, item := range items {
		if item.Name == "" {
			t.Errorf("items[%d].Name is empty", i)
		}
	}
}

func TestMenu_List_AllHaveNonNegativeSalePrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, item := range items {
		if item.SalePrice < 0 {
			t.Errorf("items[%d].SalePrice = %f, want >= 0", i, item.SalePrice)
		}
	}
}

func TestMenu_QuickCreate_ReturnsNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.QuickCreate(ctx, testutil.BusinessID, domain.MenuItemInput{
		Name:      "Integration Test Item",
		SKU:       "INT-001",
		SalePrice: 9.99,
	})
	if err != nil {
		t.Fatalf("QuickCreate: %v", err)
	}
	if id == 0 {
		t.Fatal("QuickCreate returned ID=0")
	}
}

func TestMenu_QuickCreate_ThenListContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.QuickCreate(ctx, testutil.BusinessID, domain.MenuItemInput{
		Name:      "List Check Item",
		SalePrice: 5.0,
	})
	if err != nil {
		t.Fatalf("QuickCreate: %v", err)
	}

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after create: %v", err)
	}
	found := false
	for _, item := range items {
		if item.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("new item id=%d not found in list", id)
	}
}

func TestMenu_QuickCreate_NameStoredCorrectly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const itemName = "Name Check Burger"
	id, err := repo.QuickCreate(ctx, testutil.BusinessID, domain.MenuItemInput{
		Name:      itemName,
		SalePrice: 12.50,
	})
	if err != nil {
		t.Fatalf("QuickCreate: %v", err)
	}

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			if item.Name != itemName {
				t.Errorf("Name = %q, want %q", item.Name, itemName)
			}
			return
		}
	}
	t.Errorf("item id=%d not found", id)
}

func TestMenu_List_AtLeastOneItemAfterCreate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.QuickCreate(ctx, testutil.BusinessID, domain.MenuItemInput{
		Name:      "Ensure Non-Empty",
		SalePrice: 1.0,
	})
	if err != nil {
		t.Fatalf("QuickCreate: %v", err)
	}

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one item after create")
	}
}

func TestMenu_QuickCreate_ZeroPriceAllowed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.QuickCreate(ctx, testutil.BusinessID, domain.MenuItemInput{
		Name:      "Free Item",
		SalePrice: 0,
	})
	if err != nil {
		t.Fatalf("QuickCreate zero price: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}
}
