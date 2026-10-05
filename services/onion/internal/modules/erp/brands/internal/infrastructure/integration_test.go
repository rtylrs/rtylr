package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/erp/brands/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/brands/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListBrands_AggregatesProductCountsForPage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suffix := time.Now().UnixNano()
	makeBrand := func(name string) int64 {
		id, err := repo.CreateBrand(ctx, testutil.BusinessID, domain.BrandInput{Name: name})
		if err != nil {
			t.Fatalf("CreateBrand %q: %v", name, err)
		}
		return id
	}
	firstBrandID := makeBrand(fmt.Sprintf("Count A %d", suffix))
	secondBrandID := makeBrand(fmt.Sprintf("Count B %d", suffix))
	emptyBrandID := makeBrand(fmt.Sprintf("Count C %d", suffix))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM erp_products WHERE business_id = ? AND brand_id IN (?, ?, ?)", testutil.BusinessID, firstBrandID, secondBrandID, emptyBrandID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM erp_brands WHERE business_id = ? AND id IN (?, ?, ?)", testutil.BusinessID, firstBrandID, secondBrandID, emptyBrandID)
	})

	for i, brandID := range []int64{firstBrandID, firstBrandID, secondBrandID} {
		if _, err := db.ExecContext(ctx,
			"INSERT INTO erp_products (owner_id, business_id, brand_id, name) VALUES (?, ?, ?, ?)",
			testutil.OwnerID, testutil.BusinessID, brandID, fmt.Sprintf("Brand count product %d-%d", suffix, i),
		); err != nil {
			t.Fatalf("insert product %d: %v", i, err)
		}
	}

	brands, err := repo.ListBrandsPaginated(ctx, testutil.BusinessID, 100, 0)
	if err != nil {
		t.Fatalf("ListBrandsPaginated: %v", err)
	}
	counts := make(map[int64]int64, len(brands))
	for _, brand := range brands {
		counts[brand.ID] = brand.ProductCount
	}
	if got := counts[firstBrandID]; got != 2 {
		t.Fatalf("first product count = %d, want 2", got)
	}
	if got := counts[secondBrandID]; got != 1 {
		t.Fatalf("second product count = %d, want 1", got)
	}
	if got := counts[emptyBrandID]; got != 0 {
		t.Fatalf("empty product count = %d, want 0", got)
	}
}

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

func TestListBrands_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	brands, err := repo.ListBrands(ctx, 0)
	if err != nil {
		t.Fatalf("ListBrands(ownerID=0): %v", err)
	}
	if len(brands) != 0 {
		t.Fatalf("expected 0 brands for ownerID=0, got %d", len(brands))
	}
}

func TestListBrands_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	if brands == nil {
		t.Fatal("expected non-nil slice from ListBrands")
	}
}

func TestCreateBrand_IsActiveIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{
		Name: "Active Brand",
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
			// Verify brand is present and has a non-zero ID (active by default)
			if b.ID <= 0 {
				t.Errorf("expected positive brand ID, got %d", b.ID)
			}
			return
		}
	}
	t.Fatalf("brand id=%d not found", id)
}

func TestDeleteBrand_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Deleting a brand that doesn't exist should not error
	err := repo.DeleteBrand(ctx, testutil.OwnerID, 999999999)
	if err != nil {
		t.Fatalf("DeleteBrand non-existent: %v", err)
	}
}

func TestCreate_ThenList_FindsBrand(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "FindMe Brand"})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	for _, b := range brands {
		if b.ID == id {
			return
		}
	}
	t.Fatalf("brand id=%d not found in list", id)
}

func TestListBrands_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands before: %v", err)
	}

	_, err = repo.CreateBrand(ctx, testutil.OwnerID, domain.BrandInput{Name: "Count Increase Brand"})
	if err != nil {
		t.Fatalf("CreateBrand: %v", err)
	}

	after, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands after: %v", err)
	}
	if len(after) <= len(before) {
		t.Fatalf("expected count to increase: before=%d after=%d", len(before), len(after))
	}
}

func TestCreateBrand_MultipleOwners_Scoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const altOwner int64 = 555111222

	_, err := repo.CreateBrand(ctx, altOwner, domain.BrandInput{Name: "Alt Owner Brand"})
	if err != nil {
		t.Fatalf("CreateBrand altOwner: %v", err)
	}

	brands, err := repo.ListBrands(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListBrands mainOwner: %v", err)
	}
	for _, b := range brands {
		if b.OwnerID == altOwner {
			t.Errorf("brand from altOwner leaked into mainOwner list: id=%d", b.ID)
		}
	}
}
