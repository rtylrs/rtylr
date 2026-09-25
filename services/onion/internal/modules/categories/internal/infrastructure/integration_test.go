package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
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

func TestListCategories_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	categories, err := repo.ListCategories(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(categories) != 0 {
		t.Fatalf("expected 0 categories for owner 0, got %d", len(categories))
	}
}

func TestListCategories_ReturnsAtLeastTwo(t *testing.T) {
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
}

func TestCreateCategory_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCategory(ctx, testutil.OwnerID, domain.CategoryInput{
		Name:        "Test Create Category",
		Description: "A test category",
	})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateCategory_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := "Stored Name Category"
	_, err := repo.CreateCategory(ctx, testutil.OwnerID, domain.CategoryInput{Name: name})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	found := false
	for _, c := range categories {
		if c.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("category %q not found after create", name)
	}
}

func TestUpdateCategory_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCategory(ctx, testutil.OwnerID, domain.CategoryInput{Name: "Before Update"})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	if err := repo.UpdateCategory(ctx, testutil.OwnerID, id, domain.CategoryInput{Name: "After Update"}); err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	found := false
	for _, c := range categories {
		if c.ID == id {
			if c.Name != "After Update" {
				t.Errorf("Name = %q, want %q", c.Name, "After Update")
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("category id=%d not found after update", id)
	}
}

func TestDeleteCategory_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCategory(ctx, testutil.OwnerID, domain.CategoryInput{Name: "To Be Deleted"})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	if err := repo.DeleteCategory(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	for _, c := range categories {
		if c.ID == id {
			t.Errorf("category id=%d still present after delete", id)
		}
	}
}

func TestListCategories_ContainsBeverages(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	for _, c := range categories {
		if c.Name == "Beverages" {
			return
		}
	}
	t.Errorf("fixture category 'Beverages' not found")
}

func TestListCategories_ContainsMains(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	for _, c := range categories {
		if c.Name == "Mains" {
			return
		}
	}
	t.Errorf("fixture category 'Mains' not found")
}

func TestListCategories_OwnerIDScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	for i, c := range categories {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("categories[%d].OwnerID = %d, want %d", i, c.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreateCategory_DuplicateNameAllowed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.CategoryInput{Name: "Duplicate Category"}

	id1, err := repo.CreateCategory(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("first CreateCategory: %v", err)
	}

	id2, err := repo.CreateCategory(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("second CreateCategory: %v", err)
	}

	if id1 == id2 {
		t.Errorf("expected different IDs for duplicate names, got same id=%d", id1)
	}
}

func TestDeleteCategory_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Deleting a non-existent ID should not error (noop).
	err := repo.DeleteCategory(ctx, testutil.OwnerID, 999999999)
	if err != nil {
		t.Errorf("DeleteCategory non-existent: unexpected error: %v", err)
	}
}

func TestUpdateCategory_NonExistentOwnerFails(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCategory(ctx, testutil.OwnerID, domain.CategoryInput{Name: "Owner Scoped Update"})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	// Update using a wrong owner should be a noop (0 rows affected) — no error is expected but the
	// category should remain unchanged.
	_ = repo.UpdateCategory(ctx, 999999, id, domain.CategoryInput{Name: "Should Not Change"})

	categories, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}

	for _, c := range categories {
		if c.ID == id && c.Name == "Should Not Change" {
			t.Errorf("category was updated by wrong owner")
		}
	}
}
