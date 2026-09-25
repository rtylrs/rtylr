package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
