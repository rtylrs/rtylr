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
	_, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
}

func TestStoreCategories_List_ReturnsOwnerData(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() {
		_ = repo.Delete(context.Background(), testutil.BusinessID, id)
	})
}

func TestStoreCategories_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.StoreCategoryInput{Name: "ListTestCategory", Slug: "list-test-category"}
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, err := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{Name: "BeforeUpdate", Slug: "before-update"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StoreCategoryInput{Name: "AfterUpdate", Slug: "after-update"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, c := range items {
		if c.ID == id && c.Name != "AfterUpdate" {
			t.Errorf("expected name AfterUpdate, got %s", c.Name)
		}
	}
}

func TestStoreCategories_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{Name: "ToDelete", Slug: "to-delete"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.BusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
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
	parentID, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{Name: "Parent", Slug: "parent-cat"})
	if err != nil {
		t.Fatalf("Create parent: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, parentID) })

	childID, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "Child", Slug: "child-cat", ParentID: parentID,
	})
	if err != nil {
		t.Fatalf("Create child: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, childID) })

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, c := range items {
		if c.ID == childID && c.ParentID != parentID {
			t.Errorf("expected parent_id %d, got %d", parentID, c.ParentID)
		}
	}
}

func TestStoreCategories_Update_DoesNotAffectOtherOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{Name: "OwnerScoped", Slug: "owner-scoped"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	// Update with different businessID should silently affect 0 rows, not error
	_ = repo.Update(context.Background(), 9999999, id, domain.StoreCategoryInput{Name: "Hijacked", Slug: "hijacked"})

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, c := range items {
		if c.ID == id && c.Name == "Hijacked" {
			t.Error("cross-business update succeeded unexpectedly")
		}
	}
}

func TestStoreCategories_Create_WithImageURL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name:     "WithImage",
		Slug:     "with-image",
		ImageURL: "https://example.com/image.png",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestListStoreCategories_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestListStoreCategories_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestCreateStoreCategory_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "PersistCat",
		Slug: "persist-cat",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })
}

func TestCreateStoreCategory_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "NameStoredCat",
		Slug: "name-stored-cat",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id {
			if c.Name != "NameStoredCat" {
				t.Errorf("expected name NameStoredCat, got %s", c.Name)
			}
			return
		}
	}
	t.Errorf("category id=%d not found in list", id)
}

func TestUpdateStoreCategory_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "OldCatName",
		Slug: "old-cat-name",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	if err := repo.Update(ctx, testutil.BusinessID, id, domain.StoreCategoryInput{
		Name: "NewCatName",
		Slug: "new-cat-name",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id && c.Name != "NewCatName" {
			t.Errorf("expected name NewCatName, got %s", c.Name)
		}
	}
}

func TestDeleteStoreCategory_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "DeleteCat",
		Slug: "delete-cat-row",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.BusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id {
			t.Errorf("deleted category id=%d still in list", id)
		}
	}
}

func TestCreate_ThenList_FindsCategory(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "FindMeCat",
		Slug: "find-me-cat",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id {
			return
		}
	}
	t.Errorf("category id=%d not found after create", id)
}

func TestDeleteStoreCategory_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, testutil.BusinessID, 99999999); err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestListStoreCategories_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "CountCat",
		Slug: "count-cat",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	after, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateStoreCategory_DescriptionIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name:        "DescCat",
		Slug:        "desc-cat",
		Description: "A test category with description",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id {
			if c.Description != "A test category with description" {
				t.Errorf("expected description stored, got %s", c.Description)
			}
			return
		}
	}
	t.Errorf("category id=%d not found in list", id)
}

func TestUpdateStoreCategory_ChangesDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name:        "DescUpdateCat",
		Slug:        "desc-update-cat",
		Description: "Old description",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	if err := repo.Update(ctx, testutil.BusinessID, id, domain.StoreCategoryInput{
		Name:        "DescUpdateCat",
		Slug:        "desc-update-cat",
		Description: "New description",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id && c.Description != "New description" {
			t.Errorf("expected description New description, got %s", c.Description)
		}
	}
}

func TestListStoreCategories_MultipleOwners_Scoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.StoreCategoryInput{
		Name: "ScopedCat",
		Slug: "scoped-cat",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.BusinessID, id) })

	other, err := repo.List(ctx, 9999999)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	for _, c := range other {
		if c.ID == id {
			t.Errorf("category id=%d leaked to another owner", id)
		}
	}
}
