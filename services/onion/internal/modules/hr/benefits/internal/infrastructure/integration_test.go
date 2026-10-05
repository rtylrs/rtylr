package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/benefits/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/benefits/internal/infrastructure"
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

func TestCreate_BenefitPlan_TypeHealth(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name: "Health Basic", Type: "health", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create health: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			if bp.Type != "health" {
				t.Errorf("Type = %q, want health", bp.Type)
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_TypeDental(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name: "Dental Plus", Type: "dental", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create dental: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			if bp.Type != "dental" {
				t.Errorf("Type = %q, want dental", bp.Type)
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_TypeVision(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name: "Vision Care", Type: "vision", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create vision: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			if bp.Type != "vision" {
				t.Errorf("Type = %q, want vision", bp.Type)
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_CleanupRemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name: "Cleanup Test Plan", Type: "other", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id); err != nil {
		t.Fatalf("cleanup delete: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			t.Fatalf("deleted benefit plan id=%d still appears in list", id)
		}
	}
}

func TestCreate_BenefitPlan_WrongOwnerNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name: "Owner Check Plan", Type: "health", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			t.Fatalf("plan created for owner %d should not appear for owner 999999", testutil.OwnerID)
		}
	}
}

func TestCreate_BenefitPlan_ZeroCosts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name:         "Zero Cost Plan",
		Type:         "other",
		CostEmployer: 0.0,
		CostEmployee: 0.0,
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("Create zero costs: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			if bp.CostEmployer != 0.0 {
				t.Errorf("CostEmployer = %f, want 0.0", bp.CostEmployer)
			}
			if bp.CostEmployee != 0.0 {
				t.Errorf("CostEmployee = %f, want 0.0", bp.CostEmployee)
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_HasCreatedAt(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name: "CreatedAt Check", Type: "health", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			if bp.CreatedAt.IsZero() {
				t.Error("expected non-zero CreatedAt")
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_MultipleUniqueIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{Name: "Plan X1", Type: "health", IsActive: true})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id1) })

	id2, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{Name: "Plan X2", Type: "dental", IsActive: true})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id2) })

	if id1 == id2 {
		t.Errorf("expected unique IDs, got id1=%d id2=%d", id1, id2)
	}
}

func TestList_Benefits_CountIncrementsAfterCreate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name: "Count Test Plan", Type: "vision", IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}

	if len(after) != len(before)+1 {
		t.Errorf("expected count to increase by 1: before=%d after=%d", len(before), len(after))
	}
}

func TestCreate_BenefitPlan_LargeCostValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name:         "Large Cost Plan",
		Type:         "health",
		CostEmployer: 9999.99,
		CostEmployee: 4999.99,
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("Create large cost: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			if bp.CostEmployer != 9999.99 {
				t.Errorf("CostEmployer = %f, want 9999.99", bp.CostEmployer)
			}
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}

func TestCreate_BenefitPlan_EmptyProvider(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BenefitPlanInput{
		Name:     "No Provider Plan",
		Type:     "other",
		Provider: "",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create empty provider: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, bp := range items {
		if bp.ID == id {
			return
		}
	}
	t.Fatalf("benefit plan id=%d not found", id)
}
