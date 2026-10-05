package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
