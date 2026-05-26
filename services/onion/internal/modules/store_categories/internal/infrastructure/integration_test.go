package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store_categories/internal/domain"
	"github.com/voxire/onion/internal/modules/store_categories/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreCategories_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
}

func TestStoreCategories_List_ReturnsOwnerData(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, c.OwnerID)
		}
	}
}

func TestStoreCategories_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.StoreCategoryInput{
		Name:        "Integration Test Category",
		Slug:        "integration-test-cat",
		Description: "Created by integration test",
		SortOrder:   99,
	}
	id, err := repo.Create(context.Background(), testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() {
		_ = repo.Delete(context.Background(), testutil.OwnerID, id)
	})
}

func TestStoreCategories_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.StoreCategoryInput{Name: "ListTestCategory", Slug: "list-test-category"}
	id, err := repo.Create(context.Background(), testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, c := range items {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created category id=%d not found in list", id)
	}
}

func TestStoreCategories_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreCategoryInput{Name: "BeforeUpdate", Slug: "before-update"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.StoreCategoryInput{Name: "AfterUpdate", Slug: "after-update"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, c := range items {
		if c.ID == id && c.Name != "AfterUpdate" {
			t.Errorf("expected name AfterUpdate, got %s", c.Name)
		}
	}
}

func TestStoreCategories_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreCategoryInput{Name: "ToDelete", Slug: "to-delete"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, c := range items {
		if c.ID == id {
			t.Errorf("deleted category id=%d still appears in list", id)
		}
	}
}

func TestStoreCategories_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestStoreCategories_Create_WithParentID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	parentID, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreCategoryInput{Name: "Parent", Slug: "parent-cat"})
	if err != nil {
		t.Fatalf("Create parent: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, parentID) })

	childID, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreCategoryInput{
		Name: "Child", Slug: "child-cat", ParentID: parentID,
	})
	if err != nil {
		t.Fatalf("Create child: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, childID) })

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, c := range items {
		if c.ID == childID && c.ParentID != parentID {
			t.Errorf("expected parent_id %d, got %d", parentID, c.ParentID)
		}
	}
}

func TestStoreCategories_Update_DoesNotAffectOtherOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreCategoryInput{Name: "OwnerScoped", Slug: "owner-scoped"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	// Update with different ownerID should silently affect 0 rows, not error
	_ = repo.Update(context.Background(), 9999999, id, domain.StoreCategoryInput{Name: "Hijacked", Slug: "hijacked"})

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, c := range items {
		if c.ID == id && c.Name == "Hijacked" {
			t.Error("cross-owner update succeeded unexpectedly")
		}
	}
}

func TestStoreCategories_Create_WithImageURL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreCategoryInput{
		Name:     "WithImage",
		Slug:     "with-image",
		ImageURL: "https://example.com/image.png",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
