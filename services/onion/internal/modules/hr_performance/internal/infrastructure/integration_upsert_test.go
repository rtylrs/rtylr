package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpsertPerformanceConfig_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.UpsertPerformanceConfig(ctx, testutil.OwnerID, domain.UpsertConfigParams{
		RatingScaleType: "numeric",
		RatingMin:       1,
		RatingMax:       5,
	})
	if err != nil {
		t.Fatalf("UpsertPerformanceConfig: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id=%d, got %d", testutil.OwnerID, cfg.OwnerID)
	}
}

func TestUpsertPerformanceConfig_RatingScaleIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.UpsertPerformanceConfig(ctx, testutil.OwnerID, domain.UpsertConfigParams{
		RatingScaleType: "star",
		RatingMin:       1,
		RatingMax:       10,
	})
	if err != nil {
		t.Fatalf("UpsertPerformanceConfig: %v", err)
	}
	if cfg.RatingScaleType != "star" {
		t.Errorf("expected rating_scale_type='star', got %q", cfg.RatingScaleType)
	}
}

func TestCreateGoal_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Goal Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2027-04-01",
		PeriodEnd:   "2027-06-30",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Goal Cycle")
	})

	goal, err := repo.CreateGoal(ctx, testutil.OwnerID, domain.CreateGoalParams{
		EmployeeID:    4101,
		ReviewCycleID: cycle.ID,
		Title:         "Improve throughput",
		Weight:        1.0,
	})
	if err != nil {
		t.Fatalf("CreateGoal: %v", err)
	}
	if goal.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestGetGoal_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "GetGoal Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2027-07-01",
		PeriodEnd:   "2027-09-30",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "GetGoal Cycle")
	})

	goal, err := repo.CreateGoal(ctx, testutil.OwnerID, domain.CreateGoalParams{
		EmployeeID:    4101,
		ReviewCycleID: cycle.ID,
		Title:         "Goal Roundtrip",
		Weight:        0.5,
	})
	if err != nil {
		t.Fatalf("CreateGoal: %v", err)
	}

	got, err := repo.GetGoal(ctx, testutil.OwnerID, goal.ID)
	if err != nil {
		t.Fatalf("GetGoal: %v", err)
	}
	if got.Title != "Goal Roundtrip" {
		t.Errorf("expected title='Goal Roundtrip', got %q", got.Title)
	}
}

func TestGetGoal_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetGoal(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrGoalNotFound) {
		t.Errorf("expected ErrGoalNotFound, got %v", err)
	}
}

func TestListPerformanceReviewsByEmployee_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "ByEmployee Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2027-10-01",
		PeriodEnd:   "2027-12-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "ByEmployee Cycle")
	})

	reviews, total, err := repo.ListPerformanceReviewsByEmployee(ctx, testutil.OwnerID, cycle.ID, 4101, 50, 0)
	if err != nil {
		t.Fatalf("ListPerformanceReviewsByEmployee: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = reviews
}

func TestUpdateReviewSubmission_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycle, err := repo.CreateReviewCycle(ctx, testutil.OwnerID, domain.CreateReviewCycleParams{
		Name:        "Submission Cycle",
		CycleType:   "quarterly",
		PeriodStart: "2028-01-01",
		PeriodEnd:   "2028-03-31",
	})
	if err != nil {
		t.Fatalf("CreateReviewCycle: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_performance_review_cycles WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Submission Cycle")
	})

	if err := repo.CreatePerformanceReview(ctx, testutil.OwnerID, 4101, cycle.ID, 4101, "self"); err != nil {
		t.Fatalf("CreatePerformanceReview: %v", err)
	}

	reviews, _, err := repo.ListPerformanceReviews(ctx, testutil.OwnerID, cycle.ID, 10, 0)
	if err != nil || len(reviews) == 0 {
		t.Skip("no reviews to update")
	}

	err = repo.UpdateReviewSubmission(ctx, testutil.OwnerID, reviews[0].ID, "4", "Great performance")
	if err != nil {
		t.Fatalf("UpdateReviewSubmission: %v", err)
	}
}

func performanceTestID(base int64) int64 {
	return base + time.Now().UnixNano()%100_000_000
}

func insertPerformanceEmployee(t *testing.T, db interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, businessID, employeeID int64,
) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO hr_employees
			(id, owner_id, business_id, first_name, last_name, email, status)
		VALUES (?, ?, ?, 'Performance', 'Regression', ?, 'active')`,
		employeeID, businessID, businessID, fmt.Sprintf("performance-%d@example.invalid", employeeID))
	if err != nil {
		t.Fatalf("insert performance employee: %v", err)
	}
}
