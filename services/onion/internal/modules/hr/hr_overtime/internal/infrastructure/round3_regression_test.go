package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateOvertimeRequest_ReturnsDatabaseTimestamps(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9210001

	created, err := repo.CreateOvertimeRequest(ctx, businessID, domain.OvertimeRequestRecord{
		EmployeeID:          92101,
		Date:                "2026-11-20",
		PlannedExtraMinutes: 45,
		Status:              domain.OvertimeStatusPending,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_overtime_requests WHERE business_id = ?", businessID)
	})
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("create returned zero timestamps: created=%v updated=%v", created.CreatedAt, created.UpdatedAt)
	}
	if created.CreatedAt.Unix() <= 0 || created.UpdatedAt.Unix() <= 0 {
		t.Fatalf("create returned invalid Unix timestamps: created=%d updated=%d", created.CreatedAt.Unix(), created.UpdatedAt.Unix())
	}

	stored, err := repo.GetOvertimeRequest(ctx, businessID, created.ID)
	if err != nil {
		t.Fatalf("get after create: %v", err)
	}
	if !stored.CreatedAt.Equal(created.CreatedAt) || !stored.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("returned timestamps differ from stored row: returned=(%v,%v) stored=(%v,%v)", created.CreatedAt, created.UpdatedAt, stored.CreatedAt, stored.UpdatedAt)
	}
}

func TestRejectOvertimeRequest_PersistsRejectionReason(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9210002

	created, err := repo.CreateOvertimeRequest(ctx, businessID, domain.OvertimeRequestRecord{
		EmployeeID:          92201,
		Date:                "2026-11-21",
		PlannedExtraMinutes: 60,
		Status:              domain.OvertimeStatusPending,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_overtime_requests WHERE business_id = ?", businessID)
	})

	const reason = "Insufficient coverage"
	if err := repo.RejectOvertimeRequest(ctx, businessID, created.ID, reason); err != nil {
		t.Fatalf("reject: %v", err)
	}
	stored, err := repo.GetOvertimeRequest(ctx, businessID, created.ID)
	if err != nil {
		t.Fatalf("get after reject: %v", err)
	}
	if stored.Status != domain.OvertimeStatusRejected || stored.RejectionReason != reason {
		t.Fatalf("rejected row = status %q reason %q, want rejected/%q", stored.Status, stored.RejectionReason, reason)
	}
}

func TestListOvertimeRequests_AppliesEmployeeAndStatusFiltersToRowsAndCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9210003

	for _, input := range []struct {
		employeeID int64
		date       string
		status     string
	}{
		{employeeID: 92301, date: "2026-11-22", status: domain.OvertimeStatusPending},
		{employeeID: 92302, date: "2026-11-23", status: domain.OvertimeStatusApproved},
	} {
		if _, err := repo.CreateOvertimeRequest(ctx, businessID, domain.OvertimeRequestRecord{
			EmployeeID:          input.employeeID,
			Date:                input.date,
			PlannedExtraMinutes: 30,
			Status:              input.status,
		}); err != nil {
			t.Fatalf("seed %+v: %v", input, err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_overtime_requests WHERE business_id = ?", businessID)
	})

	employeeID := int64(92301)
	status := domain.OvertimeStatusPending
	records, total, err := repo.ListOvertimeRequests(ctx, businessID, domain.ListOvertimeRequestsParams{
		EmployeeID: &employeeID,
		Status:     &status,
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("filtered list: %v", err)
	}
	if total != 1 || len(records) != 1 {
		t.Fatalf("filtered page/count = %d/%d, want 1/1", len(records), total)
	}
	if records[0].EmployeeID != employeeID || records[0].Status != status {
		t.Fatalf("filtered row = employee %d status %q, want %d/%q", records[0].EmployeeID, records[0].Status, employeeID, status)
	}

	otherStatus := domain.OvertimeStatusPending
	otherEmployee := int64(92302)
	records, total, err = repo.ListOvertimeRequests(ctx, businessID, domain.ListOvertimeRequestsParams{
		EmployeeID: &otherEmployee,
		Status:     &otherStatus,
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("non-matching filtered list: %v", err)
	}
	if total != 0 || len(records) != 0 {
		t.Fatalf("non-matching page/count = %d/%d, want 0/0", len(records), total)
	}
}
