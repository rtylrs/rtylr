package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetPerformanceConfig_NilOrNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetPerformanceConfig(ctx, testutil.OwnerID)
	if err != nil && !errors.Is(err, domain.ErrPerformanceConfigNotFound) {
		t.Fatalf("GetPerformanceConfig unexpected error: %v", err)
	}
	if cfg != nil && cfg.OwnerID != testutil.OwnerID {
		t.Errorf("config belongs to wrong owner: %d", cfg.OwnerID)
	}
}

func TestListReviewCycles_TotalMatchesRecordBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycles, total, err := repo.ListReviewCycles(ctx, testutil.OwnerID, 10, 0)
	if err != nil {
		t.Fatalf("ListReviewCycles: %v", err)
	}
	if int64(len(cycles)) > total {
		t.Errorf("returned %d cycles but total is %d", len(cycles), total)
	}
}

func TestCreateReviewCycle_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Name Storage Check",
		CycleType:   "quarterly",
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-03-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Name Storage Check")
	})
	if created.Name != "Name Storage Check" {
		t.Errorf("expected name='Name Storage Check', got %q", created.Name)
	}
}

func TestCreateReviewCycle_CycleTypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Type Check Cycle",
		CycleType:   "annual",
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-12-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Type Check Cycle")
	})
	if created.CycleType != "annual" {
		t.Errorf("expected cycle_type='annual', got %q", created.CycleType)
	}
}

func TestGetReviewCycle_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "GetByID Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-06-30",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "GetByID Cycle")
	})

	got, err := repo.GetReviewCycle(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetReviewCycle: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected id=%d, got %d", created.ID, got.ID)
	}
	if got.Name != "GetByID Cycle" {
		t.Errorf("expected name='GetByID Cycle', got %q", got.Name)
	}
}

func TestUpdateReviewCycleStatus_Changes(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Status Update Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2026-07-01",
		PeriodEnd:   "2026-09-30",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Status Update Cycle")
	})

	if err := repo.UpdateReviewCycleStatus(ctx, testutil.OwnerID, created.ID, "closed"); err != nil {
		t.Fatalf("UpdateReviewCycleStatus: %v", err)
	}

	got, err := repo.GetReviewCycle(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetReviewCycle after status update: %v", err)
	}
	if got.Status != "closed" {
		t.Errorf("expected status='closed', got %q", got.Status)
	}
}

func TestCreatePerformanceReview_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Review Create Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-12-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Review Create Cycle")
	})

	err = repo.CreatePerformanceReview(ctx, testutil.OwnerID, 4101, cycle.ID, 4101, "self")
	if err != nil {
		t.Fatalf("CreatePerformanceReview: %v", err)
	}
}

func TestListPerformanceReviews_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Count Increase Review Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2027-01-01",
		PeriodEnd:   "2027-03-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Count Increase Review Cycle")
	})

	_, total1, err := repo.ListPerformanceReviews(ctx, testutil.OwnerID, cycle.ID, 100, 0)
	if err != nil {
		t.Fatalf("ListPerformanceReviews before: %v", err)
	}

	if err := repo.CreatePerformanceReview(ctx, testutil.OwnerID, 4101, cycle.ID, 4101, "self"); err != nil {
		t.Fatalf("CreatePerformanceReview: %v", err)
	}

	_, total2, err := repo.ListPerformanceReviews(ctx, testutil.OwnerID, cycle.ID, 100, 0)
	if err != nil {
		t.Fatalf("ListPerformanceReviews after: %v", err)
	}
	if total2 <= total1 {
		t.Errorf("expected total to increase from %d, got %d", total1, total2)
	}
}
