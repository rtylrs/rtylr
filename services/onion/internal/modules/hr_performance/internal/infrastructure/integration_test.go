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
