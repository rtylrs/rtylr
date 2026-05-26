package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	"github.com/voxire/onion/internal/modules/benefits/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_Benefits_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestList_Benefits_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestCreate_BenefitPlan_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BenefitPlanInput{
		Name:         "Health Plan A",
		Type:         "health",
		Provider:     "BlueCross",
		CostEmployer: 500.00,
		CostEmployee: 150.00,
		IsActive:     true,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreate_BenefitPlan_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BenefitPlanInput{
		Name:     "Dental Plan B",
		Type:     "dental",
		IsActive: true,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, bp := range items {
		if bp.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created benefit plan id=%d not found in list", id)
	}
}

func TestCreate_BenefitPlan_DefaultTypeOther(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Type="" should default to "other"
	input := domain.BenefitPlanInput{
		Name:     "No Type Plan",
		IsActive: true,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, bp := range items {
		if bp.ID == id {
			if bp.Type != "other" {
				t.Errorf("Type = %q, want %q", bp.Type, "other")
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestList_Benefits_ItemsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{Name: "owner-check", Type: "health", IsActive: true})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, bp := range items {
		if bp.OwnerID != testutil.OwnerID {
			t.Errorf("items[%d].OwnerID = %d, want %d", i, bp.OwnerID, testutil.OwnerID)
		}
	}
}

func TestList_Benefits_ItemsHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{Name: "name-check", Type: "dental", IsActive: true})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, bp := range items {
		if bp.Name == "" {
			t.Errorf("items[%d].Name is empty", i)
		}
	}
}

func TestList_Benefits_ItemsHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{Name: "id-check", Type: "vision", IsActive: true})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, bp := range items {
		if bp.ID == 0 {
			t.Errorf("items[%d].ID is zero", i)
		}
	}
}

func TestCreate_BenefitPlan_CostRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BenefitPlanInput{
		Name:         "Cost Round Trip",
		Type:         "health",
		CostEmployer: 400.00,
		CostEmployee: 100.00,
		IsActive:     true,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, bp := range items {
		if bp.ID == id {
			if bp.CostEmployer != 400.00 {
				t.Errorf("CostEmployer = %f, want 400.00", bp.CostEmployer)
			}
			if bp.CostEmployee != 100.00 {
				t.Errorf("CostEmployee = %f, want 100.00", bp.CostEmployee)
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_IsActiveFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name:     "Inactive Plan",
		Type:     "other",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, bp := range items {
		if bp.ID == id {
			if bp.IsActive {
				t.Errorf("expected IsActive=false, got true")
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_WithProvider(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name:     "Provider Plan",
		Type:     "dental",
		Provider: "Aetna",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, bp := range items {
		if bp.ID == id {
			if bp.Provider != "Aetna" {
				t.Errorf("Provider = %q, want %q", bp.Provider, "Aetna")
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}
