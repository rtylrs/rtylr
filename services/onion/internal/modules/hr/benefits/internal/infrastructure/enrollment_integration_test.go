package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/benefits/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/benefits/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// newPlan creates a benefit plan in the given scope and registers its cleanup.
func newPlan(t *testing.T, db *sql.DB, repo *infrastructure.Repository, businessID int64, name string) int64 {
	t.Helper()
	ctx := context.Background()

	id, err := repo.Create(ctx, businessID, domain.BenefitPlanInput{
		Name: name, Type: "health", IsActive: true,
	})
	if err != nil {
		t.Fatalf("create plan %q: %v", name, err)
	}
	t.Cleanup(func() {
		if _, delErr := db.ExecContext(ctx, "DELETE FROM hr_benefit_enrollments WHERE plan_id = ?", id); delErr != nil {
			t.Logf("cleanup enrollments for plan %d: %v", id, delErr)
		}
		if _, delErr := db.ExecContext(ctx, "DELETE FROM hr_benefit_plans WHERE id = ?", id); delErr != nil {
			t.Logf("cleanup plan %d: %v", id, delErr)
		}
	})
	return id
}

func TestCreateEnrollment_ReturnsActiveEnrollment(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const employeeID int64 = 810001
	planID := newPlan(t, db, repo, testutil.OwnerID, "Enroll Active Plan")

	got, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeID, planID)
	if err != nil {
		t.Fatalf("CreateEnrollment: %v", err)
	}
	if got.ID <= 0 {
		t.Errorf("ID = %d, want positive", got.ID)
	}
	if got.EmployeeID != employeeID {
		t.Errorf("EmployeeID = %d, want %d", got.EmployeeID, employeeID)
	}
	if got.PlanID != planID {
		t.Errorf("PlanID = %d, want %d", got.PlanID, planID)
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
	if got.EnrolledAt.IsZero() {
		t.Error("EnrolledAt is zero, want a timestamp")
	}
	if got.CancelledAt != nil {
		t.Errorf("CancelledAt = %v, want nil for a new enrollment", got.CancelledAt)
	}
}

func TestCreateEnrollment_AppearsInEmployeeList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const employeeID int64 = 810002
	planID := newPlan(t, db, repo, testutil.OwnerID, "Enroll Listed Plan")

	created, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeID, planID)
	if err != nil {
		t.Fatalf("CreateEnrollment: %v", err)
	}

	items, total, err := repo.ListEnrollmentsByEmployee(ctx, testutil.OwnerID, employeeID, 50, 0)
	if err != nil {
		t.Fatalf("ListEnrollmentsByEmployee: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("expected the created enrollment id=%d in list, got %+v", created.ID, items)
	}
}

// A second employee's enrollments must never leak into the first employee's list.
func TestListEnrollments_OtherEmployeeIsNotVisible(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const employeeA int64 = 810003
	const employeeB int64 = 810004
	planID := newPlan(t, db, repo, testutil.OwnerID, "Enroll Scoping Plan")

	if _, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeA, planID); err != nil {
		t.Fatalf("CreateEnrollment employeeA: %v", err)
	}

	items, total, err := repo.ListEnrollmentsByEmployee(ctx, testutil.OwnerID, employeeB, 50, 0)
	if err != nil {
		t.Fatalf("ListEnrollmentsByEmployee employeeB: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0 for an employee with no enrollments", total)
	}
	if len(items) != 0 {
		t.Fatalf("employeeB sees %d enrollment(s) belonging to employeeA", len(items))
	}
}

// Enrolling twice in the same plan is a conflict, not a silent no-op.
func TestCreateEnrollment_DuplicateReturnsErrAlreadyEnrolled(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const employeeID int64 = 810005
	planID := newPlan(t, db, repo, testutil.OwnerID, "Enroll Duplicate Plan")

	if _, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeID, planID); err != nil {
		t.Fatalf("first CreateEnrollment: %v", err)
	}

	_, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeID, planID)
	if !errors.Is(err, domain.ErrAlreadyEnrolled) {
		t.Fatalf("second CreateEnrollment error = %v, want ErrAlreadyEnrolled", err)
	}

	_, total, listErr := repo.ListEnrollmentsByEmployee(ctx, testutil.OwnerID, employeeID, 50, 0)
	if listErr != nil {
		t.Fatalf("ListEnrollmentsByEmployee: %v", listErr)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1 - a rejected duplicate must not create a second row", total)
	}
}

func TestCreateEnrollment_UnknownPlanReturnsPlanNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreateEnrollment(ctx, testutil.OwnerID, 810006, 999999999)
	if !errors.Is(err, domain.ErrBenefitPlanNotFound) {
		t.Fatalf("error = %v, want ErrBenefitPlanNotFound", err)
	}
}

// A plan owned by another business must not be enrollable from this scope.
func TestCreateEnrollment_PlanFromOtherBusinessReturnsPlanNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const otherBusinessID int64 = 987654
	planID := newPlan(t, db, repo, otherBusinessID, "Foreign Business Plan")

	_, err := repo.CreateEnrollment(ctx, testutil.OwnerID, 810007, planID)
	if !errors.Is(err, domain.ErrBenefitPlanNotFound) {
		t.Fatalf("error = %v, want ErrBenefitPlanNotFound for a cross-tenant plan", err)
	}
}

func TestListEnrollments_NewestFirst(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const employeeID int64 = 810008
	firstPlan := newPlan(t, db, repo, testutil.OwnerID, "Order Plan One")
	secondPlan := newPlan(t, db, repo, testutil.OwnerID, "Order Plan Two")

	first, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeID, firstPlan)
	if err != nil {
		t.Fatalf("CreateEnrollment first: %v", err)
	}
	second, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeID, secondPlan)
	if err != nil {
		t.Fatalf("CreateEnrollment second: %v", err)
	}

	items, _, err := repo.ListEnrollmentsByEmployee(ctx, testutil.OwnerID, employeeID, 50, 0)
	if err != nil {
		t.Fatalf("ListEnrollmentsByEmployee: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 enrollments, got %d", len(items))
	}
	// Enrolled-at timestamps can land in the same second, so id breaks the tie
	// and the newer row must still lead.
	if items[0].ID != second.ID || items[1].ID != first.ID {
		t.Errorf("order = [%d %d], want newest first [%d %d]",
			items[0].ID, items[1].ID, second.ID, first.ID)
	}
}

func TestListEnrollments_LimitAndOffsetPaginate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const employeeID int64 = 810009
	for i, name := range []string{"Page Plan A", "Page Plan B", "Page Plan C"} {
		planID := newPlan(t, db, repo, testutil.OwnerID, name)
		if _, err := repo.CreateEnrollment(ctx, testutil.OwnerID, employeeID, planID); err != nil {
			t.Fatalf("CreateEnrollment %d: %v", i, err)
		}
	}

	page1, total, err := repo.ListEnrollmentsByEmployee(ctx, testutil.OwnerID, employeeID, 2, 0)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3 (the full count, not the page size)", total)
	}
	if len(page1) != 2 {
		t.Errorf("page 1 length = %d, want 2", len(page1))
	}

	page2, _, err := repo.ListEnrollmentsByEmployee(ctx, testutil.OwnerID, employeeID, 2, 2)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(page2) != 1 {
		t.Errorf("page 2 length = %d, want 1", len(page2))
	}
}
