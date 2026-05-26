package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	"github.com/voxire/onion/internal/modules/brands/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListBrands_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	if brands == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestListBrands_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	brands, err := repo.ListBrands(ctx, 999999)
	if err != nil {
		t.Fatalf("ListBrands wrong owner: %v", err)
	}
	if len(brands) != 0 {
		t.Fatalf("expected 0 brands for unknown owner, got %d", len(brands))
	}
}

func TestCreateBrand_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BrandInput{
		Name:        "Test Brand",
		LogoURL:     "https://example.com/logo.png",
		Description: "A test brand",
	}

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateBrand_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BrandInput{
		Name: "List Check Brand",
	}

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}

	found := false
	for _, b := range brands {
		if b.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created brand id=%d not found in list", id)
	}
}

func TestListBrands_ItemsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "OwnerID Check Brand"})

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	for i, b := range brands {
		if b.OwnerID != testutil.OwnerID {
			t.Errorf("brands[%d].OwnerID = %d, want %d", i, b.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListBrands_ItemsHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "Name Check Brand"})

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	for i, b := range brands {
		if b.Name == "" {
			t.Errorf("brands[%d].Name is empty", i)
		}
	}
}

func TestListBrands_ItemsHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "ID Check Brand"})

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	for i, b := range brands {
		if b.ID == 0 {
			t.Errorf("brands[%d].ID is zero", i)
		}
	}
}

func TestUpdateBrand_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "Before Update"})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	if err := repo.UpdateBrand(ctx, testutil.OwnerID, id, domain.BrandInput{Name: "After Update"}); err != nil {
		t.Fatalf("UpdateBrand: %v", err)
	}

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands after update: %v", err)
	}

	for _, b := range brands {
		if b.ID == id {
			if b.Name != "After Update" {
				t.Errorf("Name = %q, want %q", b.Name, "After Update")
			}
			return
		}
	}
	t.Fatalf("updated brand id=%d not found", id)
}

func TestDeleteBrand_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "To Delete Brand"})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	if err := repo.DeleteBrand(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteBrand: %v", err)
	}
}

func TestDeleteBrand_ThenNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "Delete Then Check Brand"})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	if err := repo.DeleteBrand(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteBrand: %v", err)
	}

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands after delete: %v", err)
	}
	for _, b := range brands {
		if b.ID == id {
			t.Errorf("deleted brand id=%d still in list", id)
		}
	}
}

func TestCreateBrand_DescriptionRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{
		Name:        "Description Brand",
		Description: "A detailed description",
	})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}

	for _, b := range brands {
		if b.ID == id {
			if b.Description != "A detailed description" {
				t.Errorf("Description = %q, want %q", b.Description, "A detailed description")
			}
			return
		}
	}
	t.Fatalf("brand id=%d not found", id)
}

func TestCreateBrand_LogoURLRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	logoURL := "https://cdn.example.com/brand-logo.png"
	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{
		Name:    "Logo Brand",
		LogoURL: logoURL,
	})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}

	for _, b := range brands {
		if b.ID == id {
			if b.LogoURL != logoURL {
				t.Errorf("LogoURL = %q, want %q", b.LogoURL, logoURL)
			}
			return
		}
	}
	t.Fatalf("brand id=%d not found", id)
}

func TestUpdateBrand_WrongOwnerNoMatch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "Cross Owner Brand"})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	// Update with wrong owner - no error expected (0 rows affected silently)
	_ = repo.UpdateBrand(ctx, 999999, id, domain.BrandInput{Name: "Sneaky Update"})

	// Original brand name should be unchanged
	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	for _, b := range brands {
		if b.ID == id {
			if b.Name != "Cross Owner Brand" {
				t.Errorf("Name was mutated by wrong-owner update: got %q", b.Name)
			}
			return
		}
	}
	t.Fatalf("brand id=%d not found after wrong-owner update attempt", id)
}
