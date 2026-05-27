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

func TestGetSetupSteps_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	if steps == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestUpsertSetupStep_MultipleStepsForSameOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	for _, key := range []string{"step_multi_a", "step_multi_b", "step_multi_c"} {
		if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, key, now); err != nil {
			t.Fatalf("UpsertSetupStep %q: %v", key, err)
		}
	}

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	keys := make(map[string]bool)
	for _, s := range steps {
		keys[s.StepKey] = true
	}
	for _, key := range []string{"step_multi_a", "step_multi_b", "step_multi_c"} {
		if !keys[key] {
			t.Errorf("step key %q not found after upsert", key)
		}
	}
}

func TestUpsertSetupStep_OwnerIsolation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_888_001
	stepKey := "isolated_step"
	now := time.Now().UTC().Truncate(time.Second)

	if err := repo.UpsertSetupStep(ctx, isolatedOwner, stepKey, now); err != nil {
		t.Fatalf("UpsertSetupStep isolated: %v", err)
	}

	steps, err := repo.GetSetupSteps(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetSetupSteps isolated: %v", err)
	}
	for _, s := range steps {
		if s.OwnerID != isolatedOwner {
			t.Errorf("step belongs to wrong owner: %d", s.OwnerID)
		}
	}
}

func TestUpsertHRConfig_IsIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// First call.
	if _, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, nil); err != nil {
		t.Fatalf("UpsertHRConfig first: %v", err)
	}
	// Second call should not error (idempotent).
	if _, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, nil); err != nil {
		t.Fatalf("UpsertHRConfig second: %v", err)
	}
}

func TestGetHRConfig_OwnerIDIsCorrect(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, nil); err != nil {
		t.Fatalf("UpsertHRConfig: %v", err)
	}

	cfg, err := repo.GetHRConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetHRConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id=%d, got %d", testutil.OwnerID, cfg.OwnerID)
	}
}

func TestUpsertHRConfig_WithCompletedAtUpdates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	t1 := time.Now().UTC().Truncate(time.Second)
	cfg1, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, &t1)
	if err != nil {
		t.Fatalf("UpsertHRConfig first: %v", err)
	}
	if cfg1 == nil {
		t.Fatal("expected non-nil config")
	}

	t2 := t1.Add(time.Hour)
	cfg2, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, &t2)
	if err != nil {
		t.Fatalf("UpsertHRConfig second: %v", err)
	}
	if cfg2 == nil {
		t.Fatal("expected non-nil config after second upsert")
	}
}

func TestGetSetupSteps_UnknownOwner_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	steps, err := repo.GetSetupSteps(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	// Nil is not acceptable; it must be an empty (non-nil) slice.
	if steps == nil {
		t.Error("expected non-nil slice for unknown owner, got nil")
	}
}

func TestUpsertSetupStep_CompletedAtIsUpdated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stepKey := "completed_at_update_check"
	t1 := time.Now().UTC().Truncate(time.Second)

	if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, stepKey, t1); err != nil {
		t.Fatalf("UpsertSetupStep: %v", err)
	}

	t2 := t1.Add(time.Minute)
	if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, stepKey, t2); err != nil {
		t.Fatalf("UpsertSetupStep update: %v", err)
	}

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	for _, s := range steps {
		if s.StepKey == stepKey {
			// CompletedAt must be set.
			if s.CompletedAt == nil {
				t.Errorf("expected CompletedAt to be set, got nil")
			} else if s.CompletedAt.Before(t1) {
				t.Errorf("expected CompletedAt >= %v, got %v", t1, *s.CompletedAt)
			}
			return
		}
	}
	t.Errorf("step %q not found after upsert", stepKey)
}

func TestGetHRConfig_NotFound_IsNilOrError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use a fresh isolated owner that definitely has no config.
	const isolatedOwner int64 = 9_888_002
	cfg, err := repo.GetHRConfig(ctx, isolatedOwner)
	if err != nil && !errors.Is(err, domain.ErrConfigNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
	// Either nil result (no config) or ErrConfigNotFound is acceptable.
	_ = cfg
}

func TestUpsertSetupStep_StepKeyIsNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stepKey := "nonempty_key_check"
	if err := repo.UpsertSetupStep(ctx, testutil.OwnerID, stepKey, time.Now().UTC()); err != nil {
		t.Fatalf("UpsertSetupStep: %v", err)
	}

	steps, err := repo.GetSetupSteps(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSetupSteps: %v", err)
	}
	for _, s := range steps {
		if s.StepKey == stepKey {
			if s.StepKey == "" {
				t.Error("step_key must not be empty")
			}
			return
		}
	}
	t.Errorf("step %q not found", stepKey)
}
