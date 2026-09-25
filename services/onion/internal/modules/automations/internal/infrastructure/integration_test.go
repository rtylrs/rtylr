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
