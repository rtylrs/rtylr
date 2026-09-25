package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreate_Automation_BirthdayTrigger(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "Birthday Greet",
		TriggerType:   domain.TriggerDateBirthday,
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[{"type":"send_email"}]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
}

func TestCreate_Automation_InactivityTrigger(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "Re-engagement",
		TriggerType:   domain.TriggerInactivity,
		TriggerConfig: []byte(`{"days":30}`),
		Actions:       []byte(`[{"type":"send_sms"}]`),
		IsActive:      false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
}

func TestCreate_Automation_LoyaltyTierUpTrigger(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "Tier Up Congrats",
		TriggerType:   domain.TriggerLoyaltyTierUp,
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[{"type":"send_email"},{"type":"award_points","amount":50}]`),
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
}

func TestDelete_AfterCreate_NotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "Extra Delete Check",
		TriggerType:   domain.TriggerOrderCreated,
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	list, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range list {
		if a.ID == id {
			t.Errorf("found deleted automation id=%d in list", id)
		}
	}
}

func TestUpdate_Automation_ChangesIsActiveFromTrueToFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AutomationInput{
		OwnerID:       testutil.OwnerID,
		BusinessID:    testutil.BusinessID,
		Name:          "Active-to-Inactive",
		TriggerType:   domain.TriggerOrderCreated,
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
		Name:          "Active-to-Inactive",
		TriggerType:   domain.TriggerOrderCreated,
		TriggerConfig: []byte(`{}`),
		Actions:       []byte(`[]`),
		IsActive:      false,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	list, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range list {
		if a.ID == id {
			if a.IsActive {
				t.Errorf("expected IsActive=false after update, got true")
			}
			return
		}
	}
	t.Fatalf("automation id=%d not found", id)
}
