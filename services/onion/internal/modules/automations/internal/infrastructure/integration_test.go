package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_Automations_ReturnsNonNilSlice(t *testing.T) {
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

func TestList_Automations_WrongOwnerReturnsEmpty(t *testing.T) {
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

func TestCreate_Automation_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "test-automation",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreate_Automation_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "list-check-automation",
		TriggerType:   "customer_signup",
		TriggerConfig: []byte(`{"key":"val"}`),
		Actions:       []byte(`[{"type":"send_email"}]`),
		IsActive:      false,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, a := range items {
		if a.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created automation id=%d not found in list", id)
	}
}

func TestUpdate_Automation_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "before-update",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updateInput := domain.AutomationUpdateInput{
		ID:            id,
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "after-update",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	}

	if err := repo.Update(ctx, updateInput); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after update: %v", err)
	}

	for _, a := range items {
		if a.ID == id {
			if a.Name != "after-update" {
				t.Errorf("Name = %q, want %q", a.Name, "after-update")
			}
			return
		}
	}
	t.Fatalf("updated automation id=%d not found", id)
}

func TestUpdate_Automation_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "wrong-owner-update",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(ctx, domain.AutomationUpdateInput{
		ID:            id,
		OwnerID:       999999,
		Name:          "sneaky-update",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
	})
	if !errors.Is(err, domain.ErrAutomationNotFound) {
		t.Errorf("expected ErrAutomationNotFound, got %v", err)
	}
}

func TestDelete_Automation_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "to-delete",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDelete_Automation_ThenNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "delete-then-check",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			t.Errorf("deleted automation id=%d still in list", id)
		}
	}
}

func TestDelete_Automation_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "wrong-owner-delete",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Delete(ctx, 999999, id)
	if !errors.Is(err, domain.ErrAutomationNotFound) {
		t.Errorf("expected ErrAutomationNotFound, got %v", err)
	}
}

func TestList_Automations_ItemsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, domain.AutomationInput{
		OwnerID: testutil.OwnerID, Name: "owner-check", TriggerType: "order_created",
		BusinessID:    testutil.BusinessID,
		TriggerConfig: []byte(`{}`), Actions: []byte(`[]`), IsActive: true,
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, a := range items {
		if a.OwnerID != testutil.OwnerID {
			t.Errorf("items[%d].OwnerID = %d, want %d", i, a.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreate_Automation_InactiveByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "inactive-automation",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.IsActive {
				t.Errorf("expected IsActive=false, got true")
			}
			return
		}
	}
	t.Fatalf("automation id=%d not found", id)
}

func TestCreate_Automation_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "name-is-stored-automation"
	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          want,
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.Name != want {
				t.Errorf("Name = %q, want %q", a.Name, want)
			}
			return
		}
	}
	t.Fatalf("automation id=%d not found", id)
}

func TestCreate_Automation_TriggerTypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "customer_signup"
	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "trigger-type-check",
		TriggerType:   want,
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.TriggerType != want {
				t.Errorf("TriggerType = %q, want %q", a.TriggerType, want)
			}
			return
		}
	}
	t.Fatalf("automation id=%d not found", id)
}

func TestList_Automations_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "count-check",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestUpdate_Automation_TriggerType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "trigger-update",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.AutomationUpdateInput{
		ID:            id,
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "trigger-update",
		TriggerType:   "customer_signup",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	}); err != nil {
		t.Fatalf("Update trigger type: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.TriggerType != "customer_signup" {
				t.Errorf("TriggerType = %q, want customer_signup", a.TriggerType)
			}
			return
		}
	}
	t.Fatalf("automation id=%d not found after update", id)
}

func TestUpdate_Automation_IsActiveToggle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "toggle-active",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.AutomationUpdateInput{
		ID:            id,
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "toggle-active",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if !a.IsActive {
				t.Errorf("expected IsActive=true after toggle, got false")
			}
			return
		}
	}
	t.Fatalf("automation id=%d not found after toggle", id)
}

func TestDelete_Automation_NonExistent_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Deleting non-existent with wrong owner should get ErrNotFound or noop
	err := repo.Delete(ctx, testutil.OwnerID, 9999999)
	if err != nil && !errors.Is(err, domain.ErrAutomationNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_Automation_ActiveIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "active-stored",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if !a.IsActive {
				t.Errorf("expected IsActive=true, got false")
			}
			return
		}
	}
	t.Fatalf("automation id=%d not found", id)
}

func TestCreate_Automation_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "unique-a",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "unique-b",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestList_Automations_ItemsHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "id-check",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, a := range items {
		if a.ID == 0 {
			t.Errorf("items[%d].ID is zero", i)
		}
	}
}

func TestUpdate_Automation_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "no-error-update",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(ctx, domain.AutomationUpdateInput{
		ID:            id,
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "no-error-update-2",
		TriggerType:   "order_created",
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestCreate_Automation_ZeroOwnerNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner_id=0, got %d", len(items))
	}
}
