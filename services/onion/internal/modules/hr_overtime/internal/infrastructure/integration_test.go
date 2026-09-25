package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListOvertimeRequests_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListOvertimeRequests(ctx, testutil.OwnerID, domain.ListOvertimeRequestsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListOvertimeRequests: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = records
}

func TestListOvertimeRequests_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListOvertimeRequests(ctx, testutil.OwnerID, domain.ListOvertimeRequestsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListOvertimeRequests: %v", err)
	}
	for _, rec := range records {
		if rec.OwnerID != testutil.OwnerID {
			t.Errorf("record %d belongs to wrong owner: %d", rec.ID, rec.OwnerID)
		}
	}
}

func TestListOvertimeRequests_TotalMatchesRecordBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListOvertimeRequests(ctx, testutil.OwnerID, domain.ListOvertimeRequestsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListOvertimeRequests: %v", err)
	}
	if int64(len(records)) > total {
		t.Errorf("returned %d records but total is %d", len(records), total)
	}
}

func TestListOvertimeRequests_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListOvertimeRequests(ctx, 0, domain.ListOvertimeRequestsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 || len(records) != 0 {
		t.Errorf("expected empty result for unknown owner, got total=%d records=%d", total, len(records))
	}
}

func TestGetOvertimeRequest_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetOvertimeRequest(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrOvertimeRequestNotFound) {
		t.Errorf("expected ErrOvertimeRequestNotFound, got %v", err)
	}
}

func TestCreateOvertimeRequest_Succeeds(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	req := domain.OvertimeRequestRecord{
		EmployeeID:           4101,
		Date:                 "2026-03-20",
		PlannedExtraMinutes:  60,
		Reason:               "Deadline crunch",
		Status:               "pending",
		CurrentApprovalLevel: 1,
	}

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, req)
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	if created.PlannedExtraMinutes != 60 {
		t.Errorf("expected 60 planned minutes, got %d", created.PlannedExtraMinutes)
	}
}

func TestGetOvertimeRequest_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	req := domain.OvertimeRequestRecord{
		EmployeeID:           4102,
		Date:                 "2026-03-21",
		PlannedExtraMinutes:  90,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	}

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, req)
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}

	got, err := repo.GetOvertimeRequest(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetOvertimeRequest: %v", err)
	}
	if got.EmployeeID != 4102 {
		t.Errorf("expected employee 4102, got %d", got.EmployeeID)
	}
	if got.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", got.Status)
	}
}

func TestGetOvertimeConfig_NilOrNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetOvertimeConfig(ctx, testutil.OwnerID)
	// If config does not exist it returns ErrOvertimeConfigNotFound.
	if err != nil && !errors.Is(err, domain.ErrOvertimeConfigNotFound) {
		t.Fatalf("GetOvertimeConfig unexpected error: %v", err)
	}
	if cfg != nil && cfg.OwnerID != testutil.OwnerID {
		t.Errorf("config belongs to wrong owner: %d", cfg.OwnerID)
	}
}

func TestListFlaggedForDate_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.ListFlaggedForDate(ctx, testutil.OwnerID, "2026-03-20")
	if err != nil {
		t.Fatalf("ListFlaggedForDate: %v", err)
	}
	_ = records
}

func TestListAllConfiguredOwnerIDs_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ids, err := repo.ListAllConfiguredOwnerIDs(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListAllConfiguredOwnerIDs: %v", err)
	}
	_ = ids
}

func TestListOvertimeRequests_RecordFieldsPopulated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListOvertimeRequests(ctx, testutil.OwnerID, domain.ListOvertimeRequestsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListOvertimeRequests: %v", err)
	}
	for _, rec := range records {
		if rec.ID == 0 {
			t.Error("record has zero ID")
		}
		if rec.Date == "" {
			t.Errorf("record %d has empty Date", rec.ID)
		}
		if rec.Status == "" {
			t.Errorf("record %d has empty Status", rec.ID)
		}
	}
}

func TestCreateOvertimeRequest_EmployeeIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, domain.OvertimeRequestRecord{
		EmployeeID:           4101,
		Date:                 "2026-04-01",
		PlannedExtraMinutes:  45,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}
	if created.EmployeeID != 4101 {
		t.Errorf("expected employee_id=4101, got %d", created.EmployeeID)
	}
}

func TestCreateOvertimeRequest_HoursIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, domain.OvertimeRequestRecord{
		EmployeeID:           4102,
		Date:                 "2026-04-02",
		PlannedExtraMinutes:  120,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}
	if created.PlannedExtraMinutes != 120 {
		t.Errorf("expected 120 planned minutes, got %d", created.PlannedExtraMinutes)
	}
}

func TestCreateOvertimeRequest_DateIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, domain.OvertimeRequestRecord{
		EmployeeID:           4101,
		Date:                 "2026-04-15",
		PlannedExtraMinutes:  30,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}
	if created.Date != "2026-04-15" {
		t.Errorf("expected date='2026-04-15', got %q", created.Date)
	}
}

func TestCreateOvertimeRequest_StatusIsPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, domain.OvertimeRequestRecord{
		EmployeeID:           4101,
		Date:                 "2026-05-01",
		PlannedExtraMinutes:  60,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}
	if created.Status != "pending" {
		t.Errorf("expected status='pending', got %q", created.Status)
	}
}

func TestApproveOvertimeRequest_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, domain.OvertimeRequestRecord{
		EmployeeID:           4101,
		Date:                 "2026-05-10",
		PlannedExtraMinutes:  90,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}

	if err := repo.ApproveOvertimeRequest(ctx, testutil.OwnerID, created.ID, 4101, 1.5); err != nil {
		t.Fatalf("ApproveOvertimeRequest: %v", err)
	}

	got, err := repo.GetOvertimeRequest(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetOvertimeRequest after approve: %v", err)
	}
	if got.Status != "approved" {
		t.Errorf("expected status='approved', got %q", got.Status)
	}
}

func TestRejectOvertimeRequest_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateOvertimeRequest(ctx, testutil.OwnerID, domain.OvertimeRequestRecord{
		EmployeeID:           4102,
		Date:                 "2026-05-15",
		PlannedExtraMinutes:  60,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}

	if err := repo.RejectOvertimeRequest(ctx, testutil.OwnerID, created.ID, "insufficient coverage"); err != nil {
		t.Fatalf("RejectOvertimeRequest: %v", err)
	}

	got, err := repo.GetOvertimeRequest(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetOvertimeRequest after reject: %v", err)
	}
	if got.Status != "rejected" {
		t.Errorf("expected status='rejected', got %q", got.Status)
	}
}

func TestUpsertOvertimeConfig_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertOvertimeConfig(ctx, testutil.OwnerID, domain.OvertimeConfigRecord{
		DefaultMultiplier:     1.5,
		DailyThresholdMinutes: 480,
	})
	if err != nil {
		t.Fatalf("UpsertOvertimeConfig: %v", err)
	}

	cfg, err := repo.GetOvertimeConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetOvertimeConfig after upsert: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id=%d, got %d", testutil.OwnerID, cfg.OwnerID)
	}
}

func TestUpsertOvertimeConfig_MultiplierIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertOvertimeConfig(ctx, testutil.OwnerID, domain.OvertimeConfigRecord{
		DefaultMultiplier:     2.0,
		DailyThresholdMinutes: 600,
	})
	if err != nil {
		t.Fatalf("UpsertOvertimeConfig: %v", err)
	}

	cfg, err := repo.GetOvertimeConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetOvertimeConfig: %v", err)
	}
	if cfg.DefaultMultiplier != 2.0 {
		t.Errorf("expected multiplier=2.0, got %v", cfg.DefaultMultiplier)
	}
}

func TestCreateFlaggedOvertimeRequest_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.CreateFlaggedOvertimeRequest(ctx, testutil.OwnerID, 4101, "2026-06-01", 30)
	if err != nil {
		t.Fatalf("CreateFlaggedOvertimeRequest: %v", err)
	}
}

func TestListFlaggedForDate_AfterCreate_ContainsRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	date := "2026-06-20"
	if err := repo.CreateFlaggedOvertimeRequest(ctx, testutil.OwnerID, 4101, date, 45); err != nil {
		t.Fatalf("CreateFlaggedOvertimeRequest: %v", err)
	}

	records, err := repo.ListFlaggedForDate(ctx, testutil.OwnerID, date)
	if err != nil {
		t.Fatalf("ListFlaggedForDate: %v", err)
	}
	found := false
	for _, r := range records {
		if r.EmployeeID == 4101 && r.Date == date {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected flagged record for employee 4101 on %q", date)
	}
}

func TestListOvertimeRequests_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total1, err := repo.ListOvertimeRequests(ctx, testutil.OwnerID, domain.ListOvertimeRequestsParams{Limit: 200, Offset: 0})
	if err != nil {
		t.Fatalf("ListOvertimeRequests before: %v", err)
	}

	_, err = repo.CreateOvertimeRequest(ctx, testutil.OwnerID, domain.OvertimeRequestRecord{
		EmployeeID:           4101,
		Date:                 "2026-07-01",
		PlannedExtraMinutes:  75,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateOvertimeRequest: %v", err)
	}

	_, total2, err := repo.ListOvertimeRequests(ctx, testutil.OwnerID, domain.ListOvertimeRequestsParams{Limit: 200, Offset: 0})
	if err != nil {
		t.Fatalf("ListOvertimeRequests after: %v", err)
	}
	if total2 <= total1 {
		t.Errorf("expected total to increase from %d, got %d", total1, total2)
	}
}
