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
}, businessID, employeeID int64) {
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

func TestListGoals_OptionalFiltersAndPagination(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := performanceTestID(7_000_000_000)
	employeeID := performanceTestID(8_000_000_000)
	otherEmployeeID := employeeID + 1
	insertPerformanceEmployee(t, db, businessID, employeeID)
	insertPerformanceEmployee(t, db, businessID, otherEmployeeID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_goals WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_cycles WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id IN (?, ?)", employeeID, otherEmployeeID)
	})

	cycle, err := repo.CreateReviewCycle(ctx, businessID, domain.CreateReviewCycleParams{
		Name:        fmt.Sprintf("Goal pagination %d", businessID),
		CycleType:   domain.CycleTypeAnnual,
		PeriodStart: "2029-01-01",
		PeriodEnd:   "2029-12-31",
	})
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := repo.CreateGoal(ctx, businessID, domain.CreateGoalParams{
			EmployeeID:    employeeID,
			ReviewCycleID: cycle.ID,
			Title:         fmt.Sprintf("Goal %d", i),
			Weight:        float64(i + 1),
		}); err != nil {
			t.Fatalf("create goal %d: %v", i, err)
		}
	}
	if _, err := repo.CreateGoal(ctx, businessID, domain.CreateGoalParams{
		EmployeeID:    otherEmployeeID,
		ReviewCycleID: cycle.ID,
		Title:         "Other employee goal",
		Weight:        1,
	}); err != nil {
		t.Fatalf("create other employee goal: %v", err)
	}

	all, total, err := repo.ListGoals(ctx, businessID, 0, cycle.ID, 2, 0)
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if total != 4 || len(all) != 2 {
		t.Fatalf("first page = %d rows, total=%d; want 2 rows and total 4", len(all), total)
	}
	second, total, err := repo.ListGoals(ctx, businessID, 0, cycle.ID, 2, 2)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if total != 4 || len(second) != 2 {
		t.Fatalf("second page = %d rows, total=%d; want 2 rows and total 4", len(second), total)
	}
	byEmployee, total, err := repo.ListGoals(ctx, businessID, employeeID, 0, 100, 0)
	if err != nil {
		t.Fatalf("list employee-only: %v", err)
	}
	if total != 3 || len(byEmployee) != 3 {
		t.Fatalf("employee-only result = %d rows, total=%d; want 3", len(byEmployee), total)
	}
	byCycle, total, err := repo.ListGoals(ctx, businessID, 0, cycle.ID, 100, 0)
	if err != nil {
		t.Fatalf("list cycle-only: %v", err)
	}
	if total != 4 || len(byCycle) != 4 {
		t.Fatalf("cycle-only result = %d rows, total=%d; want 4", len(byCycle), total)
	}
}

func TestCreateGoal_RejectsForeignEmployeeAndCycle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := performanceTestID(7_100_000_000)
	foreignBusinessID := performanceTestID(7_200_000_000)
	employeeID := performanceTestID(8_100_000_000)
	foreignEmployeeID := employeeID + 1
	insertPerformanceEmployee(t, db, businessID, employeeID)
	insertPerformanceEmployee(t, db, foreignBusinessID, foreignEmployeeID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_goals WHERE business_id IN (?, ?)", businessID, foreignBusinessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_cycles WHERE business_id IN (?, ?)", businessID, foreignBusinessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id IN (?, ?)", employeeID, foreignEmployeeID)
	})

	cycle, err := repo.CreateReviewCycle(ctx, businessID, domain.CreateReviewCycleParams{
		Name:        fmt.Sprintf("Goal ownership %d", businessID),
		CycleType:   domain.CycleTypeAnnual,
		PeriodStart: "2030-01-01",
		PeriodEnd:   "2030-12-31",
	})
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	foreignCycle, err := repo.CreateReviewCycle(ctx, foreignBusinessID, domain.CreateReviewCycleParams{
		Name:        fmt.Sprintf("Foreign goal ownership %d", foreignBusinessID),
		CycleType:   domain.CycleTypeAnnual,
		PeriodStart: "2030-01-01",
		PeriodEnd:   "2030-12-31",
	})
	if err != nil {
		t.Fatalf("create foreign cycle: %v", err)
	}

	_, err = repo.CreateGoal(ctx, businessID, domain.CreateGoalParams{
		EmployeeID: foreignEmployeeID, ReviewCycleID: cycle.ID, Title: "foreign employee", Weight: 1,
	})
	if !errors.Is(err, domain.ErrEmployeeNotFound) {
		t.Fatalf("foreign employee error = %v, want ErrEmployeeNotFound", err)
	}
	_, err = repo.CreateGoal(ctx, businessID, domain.CreateGoalParams{
		EmployeeID: employeeID, ReviewCycleID: foreignCycle.ID, Title: "foreign cycle", Weight: 1,
	})
	if !errors.Is(err, domain.ErrReviewCycleNotFound) {
		t.Fatalf("foreign cycle error = %v, want ErrReviewCycleNotFound", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM hr_performance_review_goals WHERE business_id = ?", businessID).Scan(&count); err != nil {
		t.Fatalf("count goals: %v", err)
	}
	if count != 0 {
		t.Fatalf("foreign references created %d goals", count)
	}
}
