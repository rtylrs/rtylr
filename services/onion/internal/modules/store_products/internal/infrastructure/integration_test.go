package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	"github.com/voxire/onion/internal/modules/store_products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreProducts_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreProducts_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, p.OwnerID)
		}
	}
}

func TestStoreProducts_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreProductInput{
		Title:       "Integration Test Product",
		Slug:        "integration-test-product",
		Price:       29.99,
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
}

func TestStoreProducts_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreProductInput{
		Title: "ListAppear Product",
		Slug:  "list-appear-product",
		Price: 9.99,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, p := range items {
		if p.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created product id=%d not found in list", id)
	}
}

func TestStoreProducts_Update_ChangesTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreProductInput{
		Title: "OldTitle", Slug: "old-title", Price: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.StoreProductInput{
		Title: "NewTitle", Slug: "new-title", Price: 5.0,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, p := range items {
		if p.ID == id && p.Title != "NewTitle" {
			t.Errorf("expected title NewTitle, got %s", p.Title)
		}
	}
}

func TestStoreProducts_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreProductInput{
		Title: "DeleteMe", Slug: "delete-me", Price: 1.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, p := range items {
		if p.ID == id {
			t.Errorf("deleted product id=%d still in list", id)
		}
	}
}

func TestStoreProducts_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(items))
	}
}

func TestStoreProducts_Create_WithCompareAtPrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreProductInput{
		Title:          "Sale Product",
		Slug:           "sale-product",
		Price:          49.99,
		CompareAtPrice: 79.99,
		IsPublished:    true,
	})
	if err != nil {
		t.Fatalf("Create with compare price: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStoreProducts_Update_PublishedToggle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreProductInput{
		Title: "DraftProduct", Slug: "draft-product", Price: 10.0, IsPublished: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.StoreProductInput{
		Title: "DraftProduct", Slug: "draft-product", Price: 10.0, IsPublished: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestStoreProducts_Create_WithDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StoreProductInput{
		Title:       "Described Product",
		Slug:        "described-product",
		Description: "This is a detailed description of the product.",
		Price:       15.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
