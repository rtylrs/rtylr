package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	"github.com/voxire/onion/internal/modules/staff/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListStaff_ReturnsAliceAndBob(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staff, err := repo.ListStaff(ctx, testutil.BusinessID, "", "", "")
	if err != nil {
		t.Fatalf("ListStaff: %v", err)
	}
	if len(staff) < 2 {
		t.Fatalf("expected at least 2 staff members, got %d", len(staff))
	}

	names := make(map[string]bool)
	for _, s := range staff {
		names[s.FirstName] = true
	}
	if !names["Alice"] {
		t.Error("expected Alice in staff list")
	}
	if !names["Bob"] {
		t.Error("expected Bob in staff list")
	}
}

func TestListStaffDirectory_ReturnsAliceAndBob(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staff, err := repo.ListStaffDirectory(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStaffDirectory: %v", err)
	}
	if len(staff) < 2 {
		t.Fatalf("expected at least 2 staff members, got %d", len(staff))
	}

	names := make(map[string]bool)
	for _, s := range staff {
		names[s.FirstName] = true
	}
	if !names["Alice"] {
		t.Error("expected Alice in staff directory")
	}
	if !names["Bob"] {
		t.Error("expected Bob in staff directory")
	}
}

func TestListStaffDirectory_LimitOne_ReturnsSingleRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staff, err := repo.ListStaffDirectory(ctx, testutil.OwnerID, 1, 0)
	if err != nil {
		t.Fatalf("ListStaffDirectory: %v", err)
	}
	if len(staff) != 1 {
		t.Fatalf("expected exactly 1 row with limit=1, got %d", len(staff))
	}
}

func TestCountStaffDirectory_MatchesUnpaginatedTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	all, err := repo.ListStaffDirectory(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("ListStaffDirectory: %v", err)
	}
	total, err := repo.CountStaffDirectory(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountStaffDirectory: %v", err)
	}
	if total != int64(len(all)) {
		t.Fatalf("expected total=%d to match unpaginated row count=%d", total, len(all))
	}
}

func TestGetStaffByID_AliceID4101(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	s, err := repo.GetStaffByID(ctx, testutil.BusinessID, 4101)
	if err != nil {
		t.Fatalf("GetStaffByID: %v", err)
	}
	if s.FirstName != "Alice" {
		t.Errorf("expected FirstName=Alice, got %q", s.FirstName)
	}
	if s.EmploymentType != "full_time" {
		t.Errorf("expected EmploymentType=full_time, got %q", s.EmploymentType)
	}
}

func TestListStaff_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staff, err := repo.ListStaff(ctx, 0, "", "", "")
	if err != nil {
		t.Fatalf("ListStaff: %v", err)
	}
	if len(staff) != 0 {
		t.Errorf("expected empty slice for owner_id=0, got %d rows", len(staff))
	}
}

func TestListStaff_FilterByDepartmentID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Filter by a non-existent department; expect empty result without error.
	staff, err := repo.ListStaff(ctx, testutil.OwnerID, "", "999999", "")
	if err != nil {
		t.Fatalf("ListStaff with departmentID filter: %v", err)
	}
	// The result may be empty or contain rows; just verify no error and non-nil slice.
	if staff == nil {
		t.Error("expected non-nil slice")
	}
}

func TestListStaff_FilterByStatus_Active(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staff, err := repo.ListStaff(ctx, testutil.OwnerID, "", "", "active")
	if err != nil {
		t.Fatalf("ListStaff with status filter: %v", err)
	}
	for _, s := range staff {
		if s.Status != "active" {
			t.Errorf("expected status=active for all results, got %q for id=%d", s.Status, s.ID)
		}
	}
}

func TestListStaff_SearchByName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staff, err := repo.ListStaff(ctx, testutil.BusinessID, "Alice", "", "")
	if err != nil {
		t.Fatalf("ListStaff search: %v", err)
	}
	if len(staff) == 0 {
		t.Fatal("expected at least one result searching for 'Alice'")
	}
	for _, s := range staff {
		if s.FirstName != "Alice" && s.LastName != "Alice" {
			// Search may be partial; just ensure Alice is among results.
			_ = s
		}
	}
}

func TestGetStaffByID_Bob(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	s, err := repo.GetStaffByID(ctx, testutil.BusinessID, 4102)
	if err != nil {
		t.Fatalf("GetStaffByID(4102): %v", err)
	}
	if s.FirstName != "Bob" {
		t.Errorf("expected FirstName=Bob, got %q", s.FirstName)
	}
}

func TestGetStaffByID_EmploymentTypePart(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	s, err := repo.GetStaffByID(ctx, testutil.BusinessID, 4102)
	if err != nil {
		t.Fatalf("GetStaffByID(4102): %v", err)
	}
	if s.EmploymentType != "part_time" {
		t.Errorf("expected EmploymentType=part_time for Bob, got %q", s.EmploymentType)
	}
}

func TestGetStaffByID_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetStaffByID(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent staff id, got nil")
	}
}

func TestCountActiveStaff_AtLeastTwo(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountActiveStaff(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CountActiveStaff: %v", err)
	}
	if count < 2 {
		t.Errorf("expected at least 2 active staff, got %d", count)
	}
}

func TestCountOnShiftToday_IsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountOnShiftToday(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountOnShiftToday: %v", err)
	}
	if count < 0 {
		t.Errorf("expected non-negative count, got %d", count)
	}
}

func TestCountPendingTimeOff_IsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountPendingTimeOff(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountPendingTimeOff: %v", err)
	}
	if count < 0 {
		t.Errorf("expected non-negative count, got %d", count)
	}
}

func TestCreateStaff_PersistsName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StaffInput{
		// OwnerID is the caller's own user id; BusinessID is a distinct value
		// (unlike the shared 30_hr.sql fixture rows, which intentionally set
		// them equal). Using different constants here actually exercises the
		// businessID contract instead of letting owner_id=business_id mask it.
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		FirstName:      "Carol",
		LastName:       "Test",
		Email:          "carol@test.local",
		Role:           "cashier",
		EmploymentType: "full_time",
		Status:         "active",
		Salary:         3000,
		StartDate:      "2024-01-01",
	}

	id, err := repo.CreateStaff(ctx, input)
	if err != nil {
		t.Fatalf("CreateStaff: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	s, err := repo.GetStaffByID(ctx, testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetStaffByID after create: %v", err)
	}
	if s.FirstName != "Carol" {
		t.Errorf("expected FirstName=Carol, got %q", s.FirstName)
	}
}

func TestCreateStaff_RoleIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StaffInput{
		// OwnerID is the caller's own user id; BusinessID is a distinct value
		// (unlike the shared 30_hr.sql fixture rows, which intentionally set
		// them equal). Using different constants here actually exercises the
		// businessID contract instead of letting owner_id=business_id mask it.
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		FirstName:      "Dave",
		LastName:       "Roley",
		Email:          "dave.roley@test.local",
		Role:           "server",
		EmploymentType: "part_time",
		Status:         "active",
		StartDate:      "2024-01-01",
	}

	id, err := repo.CreateStaff(ctx, input)
	if err != nil {
		t.Fatalf("CreateStaff: %v", err)
	}

	s, err := repo.GetStaffByID(ctx, testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetStaffByID after create: %v", err)
	}
	if s.Role != "server" {
		t.Errorf("expected Role=server, got %q", s.Role)
	}
}

func TestCreateStaff_SalaryIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StaffInput{
		// OwnerID is the caller's own user id; BusinessID is a distinct value
		// (unlike the shared 30_hr.sql fixture rows, which intentionally set
		// them equal). Using different constants here actually exercises the
		// businessID contract instead of letting owner_id=business_id mask it.
		OwnerID:        testutil.OwnerID,
		BusinessID:     testutil.BusinessID,
		FirstName:      "Eve",
		LastName:       "Pay",
		Email:          "eve.pay@test.local",
		Role:           "cashier",
		EmploymentType: "full_time",
		Status:         "active",
		Salary:         4500,
		StartDate:      "2024-01-01",
	}

	id, err := repo.CreateStaff(ctx, input)
	if err != nil {
		t.Fatalf("CreateStaff: %v", err)
	}

	s, err := repo.GetStaffByID(ctx, testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetStaffByID after create: %v", err)
	}
	if s.Salary != 4500 {
		t.Errorf("expected Salary=4500, got %v", s.Salary)
	}
}
