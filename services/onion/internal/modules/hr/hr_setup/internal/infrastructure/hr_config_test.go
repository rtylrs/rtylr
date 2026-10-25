package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_setup/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

// Setup completion belongs to the business, not to whoever finished setup.
func TestHRConfigIsKeyedByBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := NewRepository(db)
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM hr_configs WHERE business_id = ?", testutil.OtherBusinessID); err != nil {
			t.Errorf("cleanup hr_configs: %v", err)
		}
	})

	completed := time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC)
	if _, err := repo.UpsertHRConfig(ctx, testutil.OtherBusinessID, &completed); err != nil {
		t.Fatalf("UpsertHRConfig: %v", err)
	}

	cfg, err := repo.GetHRConfig(ctx, testutil.OtherBusinessID)
	if err != nil {
		t.Fatalf("GetHRConfig: %v", err)
	}
	if cfg.SetupCompletedAt == nil || !cfg.SetupCompletedAt.Equal(completed) {
		t.Fatalf("setup completion = %v, want %v", cfg.SetupCompletedAt, completed)
	}
	if _, err := repo.GetHRConfig(ctx, testutil.OtherOwnerID); !errors.Is(err, domain.ErrConfigNotFound) {
		t.Fatalf("owner user ID resolved a config: %v", err)
	}

	if _, err := repo.UpsertHRConfig(ctx, testutil.OtherBusinessID, nil); err != nil {
		t.Fatalf("second UpsertHRConfig: %v", err)
	}
	var rows int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM hr_configs WHERE business_id = ?", testutil.OtherBusinessID).Scan(&rows); err != nil {
		t.Fatalf("count: %v", err)
	}
	if rows != 1 {
		t.Fatalf("business has %d config rows, want 1", rows)
	}
}
