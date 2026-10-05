package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func baseParams() domain.ReportParams {
	return domain.ReportParams{
		DateFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}
}

func TestGenerateHeadcountReport_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateHeadcountReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateHeadcountReport: %v", err)
	}
	_ = rows
}

func TestGenerateHeadcountReport_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateHeadcountReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateHeadcountReport: %v", err)
	}
	// Each row must have a "department" key.
	for i, row := range rows {
		if _, ok := row["department"]; !ok {
			t.Errorf("row %d missing 'department' key", i)
		}
	}
}

func TestGenerateAttendanceSummaryReport_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateAttendanceSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateAttendanceSummaryReport: %v", err)
	}
	_ = rows
}

func TestGenerateAttendanceSummaryReport_HasExpectedKeys(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateAttendanceSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateAttendanceSummaryReport: %v", err)
	}
	want := []string{"employee_id", "employee_name", "department", "total_records", "days_present"}
	for i, row := range rows {
		for _, key := range want {
			if _, ok := row[key]; !ok {
				t.Errorf("row %d missing key %q", i, key)
			}
		}
	}
}

func TestGenerateLeaveSummaryReport_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateLeaveSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateLeaveSummaryReport: %v", err)
	}
	_ = rows
}

func TestGeneratePayrollSummaryReport_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GeneratePayrollSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GeneratePayrollSummaryReport: %v", err)
	}
	_ = rows
}

func TestGeneratePayrollSummaryReport_HasPayrollKeys(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GeneratePayrollSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GeneratePayrollSummaryReport: %v", err)
	}
	want := []string{"employee_id", "employee_name", "total_net_pay", "total_gross_pay"}
	for i, row := range rows {
		for _, key := range want {
			if _, ok := row[key]; !ok {
				t.Errorf("row %d missing key %q", i, key)
			}
		}
	}
}

func TestGenerateOvertimeSummaryReport_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateOvertimeSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateOvertimeSummaryReport: %v", err)
	}
	_ = rows
}

func TestGenerateTimesheetSummaryReport_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateTimesheetSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateTimesheetSummaryReport: %v", err)
	}
	_ = rows
}

func TestGenerateTimesheetSummaryReport_HasTimesheetKeys(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateTimesheetSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateTimesheetSummaryReport: %v", err)
	}
	want := []string{"employee_id", "employee_name", "total_timesheets", "total_logged_minutes"}
	for i, row := range rows {
		for _, key := range want {
			if _, ok := row[key]; !ok {
				t.Errorf("row %d missing key %q", i, key)
			}
		}
	}
}

func TestGenerateHeadcountReport_WithDepartmentFilter_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	deptID := int64(4001)
	params := baseParams()
	params.DepartmentID = &deptID

	rows, err := repo.GenerateHeadcountReport(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("GenerateHeadcountReport with dept filter: %v", err)
	}
	_ = rows
}

func TestGenerateAttendanceSummaryReport_WithEmployeeFilter_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	empID := int64(4101)
	params := baseParams()
	params.EmployeeID = &empID

	rows, err := repo.GenerateAttendanceSummaryReport(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("GenerateAttendanceSummaryReport with employee filter: %v", err)
	}
	_ = rows
}

func TestGeneratePayrollSummaryReport_NarrowDateRange_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	params := domain.ReportParams{
		DateFrom: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	rows, err := repo.GeneratePayrollSummaryReport(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("GeneratePayrollSummaryReport narrow range: %v", err)
	}
	_ = rows
}

func TestGenerateHeadcountReport_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateHeadcountReport(ctx, 0, baseParams())
	if err != nil {
		t.Fatalf("GenerateHeadcountReport zero owner: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for owner 0, got %d", len(rows))
	}
}

func TestGenerateAttendanceSummaryReport_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateAttendanceSummaryReport(ctx, 0, baseParams())
	if err != nil {
		t.Fatalf("GenerateAttendanceSummaryReport zero owner: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for owner 0, got %d", len(rows))
	}
}

func TestGenerateLeaveSummaryReport_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateLeaveSummaryReport(ctx, 0, baseParams())
	if err != nil {
		t.Fatalf("GenerateLeaveSummaryReport zero owner: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for owner 0, got %d", len(rows))
	}
}

func TestGeneratePayrollSummaryReport_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GeneratePayrollSummaryReport(ctx, 0, baseParams())
	if err != nil {
		t.Fatalf("GeneratePayrollSummaryReport zero owner: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for owner 0, got %d", len(rows))
	}
}

func TestGenerateOvertimeSummaryReport_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateOvertimeSummaryReport(ctx, 0, baseParams())
	if err != nil {
		t.Fatalf("GenerateOvertimeSummaryReport zero owner: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for owner 0, got %d", len(rows))
	}
}

func TestGenerateTimesheetSummaryReport_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateTimesheetSummaryReport(ctx, 0, baseParams())
	if err != nil {
		t.Fatalf("GenerateTimesheetSummaryReport zero owner: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for owner 0, got %d", len(rows))
	}
}

func TestGenerateLeaveSummaryReport_HasExpectedKeys(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateLeaveSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateLeaveSummaryReport: %v", err)
	}
	want := []string{"employee_id", "employee_name", "leave_type", "used_days"}
	for i, row := range rows {
		for _, key := range want {
			if _, ok := row[key]; !ok {
				t.Errorf("row %d missing key %q", i, key)
			}
		}
	}
}

func TestGenerateOvertimeSummaryReport_HasExpectedKeys(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GenerateOvertimeSummaryReport(ctx, testutil.OwnerID, baseParams())
	if err != nil {
		t.Fatalf("GenerateOvertimeSummaryReport: %v", err)
	}
	want := []string{"employee_id", "employee_name", "approved_overtime_minutes"}
	for i, row := range rows {
		for _, key := range want {
			if _, ok := row[key]; !ok {
				t.Errorf("row %d missing key %q", i, key)
			}
		}
	}
}

func TestGenerateHeadcountReport_WithEmployeeFilter_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	empID := int64(4101)
	params := baseParams()
	params.EmployeeID = &empID

	rows, err := repo.GenerateHeadcountReport(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("GenerateHeadcountReport with employee filter: %v", err)
	}
	_ = rows
}

func TestGenerateLeaveSummaryReport_WithDepartmentFilter_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	deptID := int64(4001)
	params := baseParams()
	params.DepartmentID = &deptID

	rows, err := repo.GenerateLeaveSummaryReport(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("GenerateLeaveSummaryReport with dept filter: %v", err)
	}
	_ = rows
}

func TestGenerateOvertimeSummaryReport_WithEmployeeFilter_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	empID := int64(4102)
	params := baseParams()
	params.EmployeeID = &empID

	rows, err := repo.GenerateOvertimeSummaryReport(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("GenerateOvertimeSummaryReport with employee filter: %v", err)
	}
	_ = rows
}

func TestGenerateTimesheetSummaryReport_WithDepartmentFilter_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	deptID := int64(4001)
	params := baseParams()
	params.DepartmentID = &deptID

	rows, err := repo.GenerateTimesheetSummaryReport(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("GenerateTimesheetSummaryReport with dept filter: %v", err)
	}
	_ = rows
}
