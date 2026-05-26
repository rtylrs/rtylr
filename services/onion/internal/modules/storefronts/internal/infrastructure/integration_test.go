package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/storefronts/internal/domain"
	"github.com/voxire/onion/internal/modules/storefronts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStorefronts_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStorefronts_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, s.OwnerID)
		}
	}
}

func TestStorefronts_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StorefrontInput{
		Name:        "My Test Store",
		Domain:      "teststore.example.com",
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestStorefronts_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StorefrontInput{
		Name: "List Appear Store",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, s := range items {
		if s.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created storefront id=%d not found in list", id)
	}
}

func TestStorefronts_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StorefrontInput{Name: "OldStoreName"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.StorefrontInput{Name: "NewStoreName"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, s := range items {
		if s.ID == id && s.Name != "NewStoreName" {
			t.Errorf("expected name NewStoreName, got %s", s.Name)
		}
	}
}

func TestStorefronts_Update_Domain(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StorefrontInput{
		Name: "DomainTest", Domain: "old.example.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.StorefrontInput{
		Name: "DomainTest", Domain: "new.example.com",
	}); err != nil {
		t.Fatalf("Update domain: %v", err)
	}
}

func TestStorefronts_Update_IsPublished(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StorefrontInput{
		Name: "DraftStore", IsPublished: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.StorefrontInput{
		Name: "DraftStore", IsPublished: true,
	}); err != nil {
		t.Fatalf("Update publish: %v", err)
	}
}

func TestStorefronts_List_UnknownOwner_Empty(t *testing.T) {
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

func TestStorefronts_Create_WithLogoURL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StorefrontInput{
		Name:    "LogoStore",
		LogoURL: "https://example.com/logo.png",
	})
	if err != nil {
		t.Fatalf("Create with logo: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStorefronts_Create_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	names := []string{"Store Alpha", "Store Beta", "Store Gamma"}
	for _, name := range names {
		id, err := repo.Create(context.Background(), testutil.OwnerID, domain.StorefrontInput{Name: name})
		if err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
		if id <= 0 {
			t.Errorf("store %s: expected positive ID", name)
		}
	}
}
