package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/careers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetActiveCareers_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	// Result may be empty if no postings exist; the call itself must succeed.
	_ = careers
}

func TestGetActiveCareers_NeverReturnsNilOnEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	if careers == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestGetActiveCareers_ActivePostingsHaveTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	for _, c := range careers {
		if c.Title == "" {
			t.Errorf("career id=%d has empty Title", c.Id)
		}
		if !c.IsActive {
			t.Errorf("career id=%d returned but IsActive=false", c.Id)
		}
	}
}

func TestGetActiveCareers_IdsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	for _, c := range careers {
		if c.Id <= 0 {
			t.Errorf("expected positive ID, got %d", c.Id)
		}
	}
}

func TestGetActiveCareers_StableAcrossMultipleCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	second, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("expected stable result: first=%d second=%d", len(first), len(second))
	}
}
