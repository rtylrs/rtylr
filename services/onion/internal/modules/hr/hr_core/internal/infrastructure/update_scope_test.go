package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// missingRowID is an id no fixture row uses.
const missingRowID int64 = 9999999

// Regression: the update statements were generated as :exec, so an UPDATE that
// matched no row returned nil and the handler echoed the request as an updated
// record with 200. Every update must report not-found instead.
func TestUpdateLocation_MissingID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.UpdateLocation(context.Background(), testutil.OwnerID, domain.LocationRecord{
		ID: missingRowID, Name: "Ghost Location",
	})
	if !errors.Is(err, domain.ErrLocationNotFound) {
		t.Fatalf("UpdateLocation(missing) error = %v, want ErrLocationNotFound", err)
	}
}

func TestUpdateDepartment_MissingID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.UpdateDepartment(context.Background(), testutil.OwnerID, domain.DepartmentRecord{
		ID: missingRowID, Name: "Ghost Department",
	})
	if !errors.Is(err, domain.ErrDepartmentNotFound) {
		t.Fatalf("UpdateDepartment(missing) error = %v, want ErrDepartmentNotFound", err)
	}
}

func TestUpdateWorkSchedule_MissingID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.UpdateWorkSchedule(context.Background(), testutil.OwnerID, domain.WorkScheduleRecord{
		ID: missingRowID, Name: "Ghost Schedule", ScheduleData: `{}`,
	})
	if !errors.Is(err, domain.ErrWorkScheduleNotFound) {
		t.Fatalf("UpdateWorkSchedule(missing) error = %v, want ErrWorkScheduleNotFound", err)
	}
}

func TestUpdateRole_MissingID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.UpdateRole(context.Background(), testutil.OwnerID, domain.RoleRecord{
		ID: missingRowID, Name: "Ghost Role", Permissions: `{}`,
	})
	if !errors.Is(err, domain.ErrRoleNotFound) {
		t.Fatalf("UpdateRole(missing) error = %v, want ErrRoleNotFound", err)
	}
}

// Regression: a write scoped to another tenant's id matched zero rows because
// business_id is part of the WHERE clause, but the caller still got a success
// response. The four updates must fail closed instead of pretending to write.
func TestUpdate_CrossTenantID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	loc, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{Name: "Scope Location", Country: "TC"})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	dept, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{Name: "Scope Department", LocationID: loc.ID})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	ws, err := repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{Name: "Scope Schedule", ScheduleData: `{}`})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}
	role, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{Name: "Scope Role", Permissions: `{}`})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_departments WHERE id = ?", dept.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_locations WHERE id = ?", loc.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_work_schedules WHERE id = ?", ws.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_roles WHERE id = ?", role.ID)
	})

	if _, err := repo.UpdateLocation(ctx, testutil.OtherBusinessID, domain.LocationRecord{ID: loc.ID, Name: "Hijacked Location"}); !errors.Is(err, domain.ErrLocationNotFound) {
		t.Errorf("cross-tenant UpdateLocation error = %v, want ErrLocationNotFound", err)
	}
	if _, err := repo.UpdateDepartment(ctx, testutil.OtherBusinessID, domain.DepartmentRecord{ID: dept.ID, Name: "Hijacked Department"}); !errors.Is(err, domain.ErrDepartmentNotFound) {
		t.Errorf("cross-tenant UpdateDepartment error = %v, want ErrDepartmentNotFound", err)
	}
	if _, err := repo.UpdateWorkSchedule(ctx, testutil.OtherBusinessID, domain.WorkScheduleRecord{ID: ws.ID, Name: "Hijacked Schedule", ScheduleData: `{}`}); !errors.Is(err, domain.ErrWorkScheduleNotFound) {
		t.Errorf("cross-tenant UpdateWorkSchedule error = %v, want ErrWorkScheduleNotFound", err)
	}
	if _, err := repo.UpdateRole(ctx, testutil.OtherBusinessID, domain.RoleRecord{ID: role.ID, Name: "Hijacked Role", Permissions: `{}`}); !errors.Is(err, domain.ErrRoleNotFound) {
		t.Errorf("cross-tenant UpdateRole error = %v, want ErrRoleNotFound", err)
	}

	// The rejected writes must not have changed the caller's rows.
	got, err := repo.GetLocation(ctx, testutil.OwnerID, loc.ID)
	if err != nil {
		t.Fatalf("GetLocation after rejected cross-tenant update: %v", err)
	}
	if got.Name != "Scope Location" {
		t.Errorf("cross-tenant update mutated the row: name = %q, want 'Scope Location'", got.Name)
	}
	roles, err := repo.ListRoles(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListRoles after rejected cross-tenant update: %v", err)
	}
	for _, r := range roles {
		if r.ID == role.ID && r.Name != "Scope Role" {
			t.Errorf("cross-tenant update mutated the role: name = %q, want 'Scope Role'", r.Name)
		}
	}
}

// Regression: ListDepartments accepted a location_id filter, included it in the
// service cache key, and then dropped it before the query, so a client that
// filtered by location received every department of the business.
func TestListDepartments_LocationFilter_ReturnsOnlyThatLocation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locA, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{Name: "Filter Location A", Country: "TC"})
	if err != nil {
		t.Fatalf("CreateLocation A: %v", err)
	}
	locB, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{Name: "Filter Location B", Country: "TC"})
	if err != nil {
		t.Fatalf("CreateLocation B: %v", err)
	}
	deptA, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{Name: "Filter Dept A", LocationID: locA.ID})
	if err != nil {
		t.Fatalf("CreateDepartment A: %v", err)
	}
	deptB, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{Name: "Filter Dept B", LocationID: locB.ID})
	if err != nil {
		t.Fatalf("CreateDepartment B: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_departments WHERE id IN (?, ?)", deptA.ID, deptB.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_locations WHERE id IN (?, ?)", locA.ID, locB.ID)
	})

	filtered, err := repo.ListDepartments(ctx, testutil.OwnerID, locA.ID)
	if err != nil {
		t.Fatalf("ListDepartments(location=%d): %v", locA.ID, err)
	}

	var sawA, sawB bool
	for _, d := range filtered {
		switch d.ID {
		case deptA.ID:
			sawA = true
		case deptB.ID:
			sawB = true
		}
	}
	if !sawA {
		t.Errorf("own department %d missing, so the absence below proves nothing", deptA.ID)
	}
	if sawB {
		t.Errorf("location filter ignored: department %d belongs to location %d", deptB.ID, locB.ID)
	}

	// A zero location keeps meaning "all locations".
	unfiltered, err := repo.ListDepartments(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListDepartments(location=0): %v", err)
	}
	sawA, sawB = false, false
	for _, d := range unfiltered {
		switch d.ID {
		case deptA.ID:
			sawA = true
		case deptB.ID:
			sawB = true
		}
	}
	if !sawA || !sawB {
		t.Errorf("ListDepartments(0) must keep returning every location's departments, got A=%v B=%v", sawA, sawB)
	}
}

// MySQL reports rows *changed*, not rows matched, so a PUT that stores exactly
// the values the row already holds affects zero rows. That must stay a success:
// only a genuinely absent row is a 404. This pins the zero-affected probe for
// every update that checks it.
func TestUpdate_UnchangedValues_IsNotReportedMissing(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	loc, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{Name: "Noop Location", Country: "TC"})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	dept, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{Name: "Noop Department", LocationID: loc.ID})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	ws, err := repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{Name: "Noop Schedule", ScheduleData: `{}`})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}
	role, err := repo.CreateRole(ctx, testutil.OwnerID, domain.RoleRecord{Name: "Noop Role", Permissions: `{}`})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_departments WHERE id = ?", dept.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_locations WHERE id = ?", loc.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_work_schedules WHERE id = ?", ws.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_roles WHERE id = ?", role.ID)
	})

	if _, err := repo.UpdateLocation(ctx, testutil.OwnerID, *loc); err != nil {
		t.Errorf("unchanged UpdateLocation = %v, want nil", err)
	}
	if _, err := repo.UpdateDepartment(ctx, testutil.OwnerID, *dept); err != nil {
		t.Errorf("unchanged UpdateDepartment = %v, want nil", err)
	}
	if _, err := repo.UpdateWorkSchedule(ctx, testutil.OwnerID, *ws); err != nil {
		t.Errorf("unchanged UpdateWorkSchedule = %v, want nil", err)
	}
	if _, err := repo.UpdateRole(ctx, testutil.OwnerID, *role); err != nil {
		t.Errorf("unchanged UpdateRole = %v, want nil", err)
	}
}

// Regression: an unknown step_key hit an UPDATE that matched no row, and the
// follow-up lookup surfaced sql.ErrNoRows as a 500. It must read as not-found.
func TestCompleteSetupStep_UnknownKey_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.CompleteSetupStep(context.Background(), testutil.OwnerID, "no-such-step-key")
	if !errors.Is(err, domain.ErrSetupStepNotFound) {
		t.Fatalf("CompleteSetupStep(unknown key) error = %v, want ErrSetupStepNotFound", err)
	}
}
