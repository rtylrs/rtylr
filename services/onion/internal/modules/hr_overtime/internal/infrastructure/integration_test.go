package infrastructure_test

import (
	"context"
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
	if err != domain.ErrOvertimeRequestNotFound {
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
	if err != nil && err != domain.ErrOvertimeConfigNotFound {
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

	ids, err := repo.ListAllConfiguredOwnerIDs(ctx)
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
