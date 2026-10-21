package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store/storefronts/internal/domain"
	"github.com/voxire/onion/internal/modules/store/storefronts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStorefronts_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStorefronts_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.BusinessID != testutil.BusinessID {
			t.Errorf("expected business_id %d, got %d", testutil.BusinessID, s.BusinessID)
		}
	}
}

func TestStorefronts_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{
		Name: "List Appear Store",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "OldStoreName"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StorefrontInput{Name: "NewStoreName"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, s := range items {
		if s.ID == id && s.Name != "NewStoreName" {
			t.Errorf("expected name NewStoreName, got %s", s.Name)
		}
	}
}

func TestStorefronts_Update_Domain(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{
		Name: "DomainTest", Domain: "old.example.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StorefrontInput{
		Name: "DomainTest", Domain: "new.example.com",
	}); err != nil {
		t.Fatalf("Update domain: %v", err)
	}
}

func TestStorefronts_Update_IsPublished(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{
		Name: "DraftStore", IsPublished: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StorefrontInput{
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
		t.Errorf("expected 0 for unknown business, got %d", len(items))
	}
}

func TestStorefronts_Create_WithLogoURL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{
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
		id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: name})
		if err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
		if id <= 0 {
			t.Errorf("store %s: expected positive ID", name)
		}
	}
}

func TestStorefronts_List_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestStorefronts_Create_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "StoredNameStorefront"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: want})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if s.Name != want {
				t.Errorf("Name = %q, want %q", s.Name, want)
			}
			return
		}
	}
	t.Errorf("storefront id=%d not found", id)
}

func TestStorefronts_Create_DomainIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "domain-stored.example.com"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{
		Name:   "DomainStored",
		Domain: want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if s.Domain != want {
				t.Errorf("Domain = %q, want %q", s.Domain, want)
			}
			return
		}
	}
	t.Errorf("storefront id=%d not found", id)
}

func TestStorefronts_Create_IsPublishedStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{
		Name:        "PublishedStored",
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if !s.IsPublished {
				t.Error("expected IsPublished=true, got false")
			}
			return
		}
	}
	t.Errorf("storefront id=%d not found", id)
}

func TestStorefronts_Update_LogoURL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "LogoUpdateStore"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StorefrontInput{
		Name:    "LogoUpdateStore",
		LogoURL: "https://cdn.example.com/new-logo.png",
	}); err != nil {
		t.Fatalf("Update logo: %v", err)
	}
}

func TestStorefronts_Update_ThemeID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "ThemeIDStore"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StorefrontInput{
		Name:    "ThemeIDStore",
		ThemeID: 5,
	}); err != nil {
		t.Fatalf("Update themeID: %v", err)
	}
}

func TestStorefronts_List_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	before, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "CountStore"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestStorefronts_Create_ZeroOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for business_id=0, got %d", len(items))
	}
}

func TestStorefronts_Update_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "UpdateNoErrorStore"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(context.Background(), testutil.BusinessID, id, domain.StorefrontInput{Name: "UpdateNoErrorStore2"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestStorefronts_Create_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id1, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "UniqueA"})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "UniqueB"})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestStorefronts_List_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _ = repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "IDCheckStore"})

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, s := range items {
		if s.ID == 0 {
			t.Errorf("items[%d].ID is zero", i)
		}
	}
}

func TestStorefronts_Update_ReflectedInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StorefrontInput{Name: "ReflectStore"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StorefrontInput{
		Name: "ReflectStoreUpdated",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if s.Name != "ReflectStoreUpdated" {
				t.Errorf("Name = %q, want ReflectStoreUpdated", s.Name)
			}
			return
		}
	}
	t.Errorf("storefront id=%d not found after update", id)
}
