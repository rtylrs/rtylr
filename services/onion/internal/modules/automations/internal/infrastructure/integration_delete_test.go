package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
