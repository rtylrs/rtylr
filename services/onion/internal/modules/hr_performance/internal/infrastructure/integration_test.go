package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListReviewCycles_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycles, total, err := repo.ListReviewCycles(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListReviewCycles: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = cycles
}

func TestListReviewCycles_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycles, _, err := repo.ListReviewCycles(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListReviewCycles: %v", err)
	}
	for _, c := range cycles {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("cycle %d belongs to wrong owner: %d", c.ID, c.OwnerID)
		}
	}
}

func TestListReviewCycles_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycles, total, err := repo.ListReviewCycles(ctx, 0, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 || len(cycles) != 0 {
		t.Errorf("expected empty result for unknown owner, got total=%d cycles=%d", total, len(cycles))
	}
}

func TestGetReviewCycle_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetReviewCycle(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrReviewCycleNotFound) {
		t.Errorf("expected ErrReviewCycleNotFound, got %v", err)
	}
}

func TestCreateAndGetReviewCycle_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Q1 2026 Review",
		CycleType:   "quarterly",
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-03-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Q1 2026 Review")
	})
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	if created.Name != "Q1 2026 Review" {
		t.Errorf("name mismatch: got %q", created.Name)
	}
	if created.Status == "" {
		t.Error("expected non-empty status on new cycle")
	}
}

func TestCreateReviewCycle_DateFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Date Check Cycle",
		CycleType:   "annual",
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-12-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Date Check Cycle")
	})
	if created.PeriodStart != "2026-01-01" {
		t.Errorf("PeriodStart mismatch: got %q", created.PeriodStart)
	}
	if created.PeriodEnd != "2026-12-31" {
		t.Errorf("PeriodEnd mismatch: got %q", created.PeriodEnd)
	}
}

func TestGetPerformanceReview_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetPerformanceReview(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrReviewNotFound) {
		t.Errorf("expected ErrReviewNotFound, got %v", err)
	}
}

func TestListPerformanceReviews_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create a cycle to query against.
	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Review List Test Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-06-30",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Review List Test Cycle")
	})

	reviews, total, err := repo.ListPerformanceReviews(ctx, testutil.OwnerID, cycle.ID, 50, 0)
	if err != nil {
		t.Fatalf("ListPerformanceReviews: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = reviews
}

func TestListActiveEmployeeIDs_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ids, err := repo.ListActiveEmployeeIDs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveEmployeeIDs: %v", err)
	}
	_ = ids
}

func TestListActiveEmployeeIDs_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ids, err := repo.ListActiveEmployeeIDs(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 IDs for unknown owner, got %d", len(ids))
	}
}

func TestListGoals_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Goals List Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2026-07-01",
		PeriodEnd:   "2026-09-30",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Goals List Cycle")
	})

	goals, total, err := repo.ListGoals(ctx, testutil.OwnerID, 4101, cycle.ID, 50, 0)
	if err != nil {
		t.Fatalf("ListGoals: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = goals
}
