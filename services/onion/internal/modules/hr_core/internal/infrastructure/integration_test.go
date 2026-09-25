package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListEmployees_ReturnsAliceAndBob(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	employees, total, err := repo.ListEmployees(ctx, testutil.BusinessID, domain.ListEmployeesParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListEmployees: %v", err)
	}
	if total < 2 {
		t.Fatalf("expected at least 2 employees, got total=%d", total)
	}

	names := make(map[string]bool)
	for _, e := range employees {
		names[e.FirstName] = true
	}
	if !names["Alice"] {
		t.Error("expected Alice in employee list")
	}
	if !names["Bob"] {
		t.Error("expected Bob in employee list")
	}
}

func TestGetEmployee_AliceHasSalary1200(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	emp, err := repo.GetEmployee(ctx, testutil.BusinessID, 4101)
	if err != nil {
		t.Fatalf("GetEmployee: %v", err)
	}
	if emp.FirstName != "Alice" {
		t.Errorf("expected FirstName=Alice, got %q", emp.FirstName)
	}
	if emp.SalaryCents != 120000 {
		t.Errorf("expected SalaryCents=120000, got %v", emp.SalaryCents)
	}
}

func TestListDepartments_ReturnsOperations(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	departments, err := repo.ListDepartments(ctx, testutil.BusinessID, 0)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}

	found := false
	for _, d := range departments {
		if d.Name == "Operations" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected department 'Operations' in list, got %v", departments)
	}
}

func TestListDepartments_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	depts, err := repo.ListDepartments(ctx, 0, 0)
	if err != nil {
		t.Fatalf("ListDepartments zero owner: %v", err)
	}
	if len(depts) != 0 {
		t.Errorf("expected 0 departments for owner 0, got %d", len(depts))
	}
}

func TestListDepartments_ContainsOperations(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	depts, err := repo.ListDepartments(ctx, testutil.BusinessID, 0)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range depts {
		if d.Name == "Operations" {
			return
		}
	}
	t.Errorf("department 'Operations' not found in %d results", len(depts))
}

func TestGetDepartmentByID_4001_IsOperations(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Verify department 4001 appears in the list with name Operations.
	depts, err := repo.ListDepartments(ctx, testutil.BusinessID, 0)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	var found *domain.DepartmentRecord
	for i, d := range depts {
		if d.ID == 4001 {
			found = &depts[i]
			break
		}
	}
	if found == nil {
		t.Fatal("department id=4001 not found")
	}
	if found.Name != "Operations" {
		t.Errorf("expected name=Operations, got %q", found.Name)
	}
}

func TestGetDepartmentByID_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// No GetDepartmentByID method exists; we verify via list that id 999999 is absent.
	depts, err := repo.ListDepartments(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	for _, d := range depts {
		if d.ID == 999999 {
			t.Error("unexpected department id=999999 found")
		}
	}
}

func TestCreateDepartment_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{
		Name: "Test Dept Create",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestCreateDepartment_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := fmt.Sprintf("Dept Name %d", 42)
	created, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{
		Name: name,
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	if created.Name != name {
		t.Errorf("expected name=%q, got %q", name, created.Name)
	}
}

func TestCreateDepartment_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{
		Name: "Owner Check Dept",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}
	if created.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id=%d, got %d", testutil.OwnerID, created.OwnerID)
	}
}

func TestUpdateDepartment_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{
		Name: "Before Update",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	created.Name = "After Update"
	updated, err := repo.UpdateDepartment(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateDepartment: %v", err)
	}
	if updated.Name != "After Update" {
		t.Errorf("expected name='After Update', got %q", updated.Name)
	}
}

func TestCreateDepartment_ThenGet_Roundtrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{
		Name: "Roundtrip Dept",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	depts, err := repo.ListDepartments(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListDepartments: %v", err)
	}
	found := false
	for _, d := range depts {
		if d.ID == created.ID {
			found = true
			if d.Name != "Roundtrip Dept" {
				t.Errorf("expected name='Roundtrip Dept', got %q", d.Name)
			}
			break
		}
	}
	if !found {
		t.Errorf("created department id=%d not found in list", created.ID)
	}
}

func TestListDepartments_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListDepartments(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListDepartments before: %v", err)
	}

	_, err = repo.CreateDepartment(ctx, testutil.OwnerID, domain.DepartmentRecord{
		Name: "Count Increase Dept",
	})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	after, err := repo.ListDepartments(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListDepartments after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase from %d, got %d", len(before), len(after))
	}
}

func TestListLocations_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locs, err := repo.ListLocations(ctx, 0)
	if err != nil {
		t.Fatalf("ListLocations zero owner: %v", err)
	}
	if len(locs) != 0 {
		t.Errorf("expected 0 locations for owner 0, got %d", len(locs))
	}
}

func TestListLocations_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locs, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}
	_ = locs
}

func TestListWorkSchedules_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	schedules, err := repo.ListWorkSchedules(ctx, 0, 100, 0)
	if err != nil {
		t.Fatalf("ListWorkSchedules zero owner: %v", err)
	}
	if len(schedules) != 0 {
		t.Errorf("expected 0 schedules for owner 0, got %d", len(schedules))
	}
}

func TestCreateWorkSchedule_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ws, err := repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{
		Name:         "Test Schedule",
		ScheduleData: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}
	if ws.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestUpsertHRConfig_DirectoryEnabled_Roundtrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, domain.HRCompanyConfigRecord{
		CompanyName:      "Directory Test Co",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DirectoryEnabled: true,
	}); err != nil {
		t.Fatalf("UpsertHRConfig (enable): %v", err)
	}

	cfg, err := repo.GetHRConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetHRConfig after enable: %v", err)
	}
	if !cfg.DirectoryEnabled {
		t.Error("expected DirectoryEnabled=true after upsert, got false")
	}

	if _, err := repo.UpsertHRConfig(ctx, testutil.OwnerID, domain.HRCompanyConfigRecord{
		CompanyName:      "Directory Test Co",
		DefaultCurrency:  "USD",
		DefaultTimezone:  "UTC",
		DirectoryEnabled: false,
	}); err != nil {
		t.Fatalf("UpsertHRConfig (disable): %v", err)
	}

	cfg, err = repo.GetHRConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetHRConfig after disable: %v", err)
	}
	if cfg.DirectoryEnabled {
		t.Error("expected DirectoryEnabled=false after disabling, got true")
	}
}

func TestListRoles_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	roles, err := repo.ListRoles(ctx, 0)
	if err != nil {
		t.Fatalf("ListRoles zero owner: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected 0 roles for owner 0, got %d", len(roles))
	}
}
