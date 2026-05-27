package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetSetupSteps_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	_ = steps
}

func TestGetSetupSteps_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	for _, s := range steps {
		if s.OwnerID != testutil.OwnerID {
			t.Errorf("step %d has owner_id %d, want %d", s.ID, s.OwnerID, testutil.OwnerID)
		}
	}
}

func TestUpsertSetupStep_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stepKey := "test_step_idempotent"
	completedAt := time.Now().UTC().Truncate(time.Second)

	if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, stepKey, completedAt); err != nil {
		t.Fatalf("UpsertSetupStep first call: %v", err)
	}
	// Second call with updated time should not error (ON DUPLICATE KEY UPDATE).
	if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, stepKey, completedAt.Add(time.Minute)); err != nil {
		t.Fatalf("UpsertSetupStep second call: %v", err)
	}
}

func TestUpsertSetupStep_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stepKey := "test_step_appears_in_list"
	completedAt := time.Now().UTC().Truncate(time.Second)

	if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, stepKey, completedAt); err != nil {
		t.Fatalf("UpsertSetupStep: %v", err)
	}

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps after upsert: %v", err)
	}
	found := false
	for _, s := range steps {
		if s.StepKey == stepKey {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("upserted step key %q not found in GetSetupSteps result", stepKey)
	}
}

func TestGetHRConfig_NotFoundOrPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetHRConfig(ctx, testutil.OwnerID)
	if err != nil && !errors.Is(err, domain.ErrConfigNotFound) {
		t.Fatalf("GetHRConfig unexpected error: %v", err)
	}
	// Either not found (nil) or a valid config -- both are fine.
	_ = cfg
}

func TestUpsertHRConfig_CreatesOrUpdates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("UpsertHRConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("UpsertHRConfig returned nil config")
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("UpsertHRConfig owner_id = %d, want %d", cfg.OwnerID, testutil.OwnerID)
	}
}

func TestUpsertHRConfig_WithCompletedAt(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	cfg, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, &now)
	if err != nil {
		t.Fatalf("UpsertHRConfig with completedAt: %v", err)
	}
	if cfg == nil {
		t.Fatal("UpsertHRConfig returned nil config")
	}
}

func TestGetHRConfig_AfterUpsert_ReturnsConfig(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, nil); err != nil {
		t.Fatalf("UpsertHRConfig setup: %v", err)
	}

	cfg, err := repo.GetHRConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetHRConfig after upsert: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("GetHRConfig owner_id = %d, want %d", cfg.OwnerID, testutil.OwnerID)
	}
}

func TestGetSetupSteps_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	steps, err := repo.GetSetupSteps(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetSetupSteps unknown owner: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("expected 0 steps for unknown owner, got %d", len(steps))
	}
}

func TestUpsertSetupStep_StepKeyPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stepKey := "test_step_key_present"
	completedAt := time.Now().UTC().Truncate(time.Second)

	if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, stepKey, completedAt); err != nil {
		t.Fatalf("UpsertSetupStep: %v", err)
	}

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	for _, s := range steps {
		if s.StepKey == stepKey {
			if s.StepKey == "" {
				t.Error("step key is empty")
			}
			return
		}
	}
	t.Errorf("step key %q not found", stepKey)
}
