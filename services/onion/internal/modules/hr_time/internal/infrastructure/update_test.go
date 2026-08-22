package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpdateTimeEntry_ReflectedOnGet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateTimeEntry(ctx, testutil.OwnerID, domain.TimeEntryRecord{
		EmployeeID:      4101,
		Date:            "2026-04-01",
		DurationMinutes: 120,
		Description:     "initial",
	})
	if err != nil {
		t.Fatalf("CreateTimeEntry: %v", err)
	}

	created.DurationMinutes = 240
	created.Description = "updated description"
	updated, err := repo.UpdateTimeEntry(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateTimeEntry: %v", err)
	}
	if updated.DurationMinutes != 240 {
		t.Errorf("expected 240 minutes, got %d", updated.DurationMinutes)
	}
}

func TestSubmitTimesheet_StatusChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ts, err := repo.UpsertTimesheet(ctx, testutil.OwnerID, domain.TimesheetRecord{
		EmployeeID:  4101,
		PeriodStart: "2026-03-01",
		PeriodEnd:   "2026-03-31",
		Status:      domain.TimesheetStatusDraft,
	})
	if err != nil {
		t.Fatalf("UpsertTimesheet: %v", err)
	}

	submitted, err := repo.SubmitTimesheet(ctx, testutil.OwnerID, ts.ID)
	if err != nil {
		t.Fatalf("SubmitTimesheet: %v", err)
	}
	if submitted.Status != domain.TimesheetStatusSubmitted {
		t.Errorf("expected status %q, got %q", domain.TimesheetStatusSubmitted, submitted.Status)
	}
}

func TestRejectTimesheet_StatusChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ts, err := repo.UpsertTimesheet(ctx, testutil.OwnerID, domain.TimesheetRecord{
		EmployeeID:  4102,
		PeriodStart: "2026-02-01",
		PeriodEnd:   "2026-02-28",
		Status:      domain.TimesheetStatusSubmitted,
	})
	if err != nil {
		t.Fatalf("UpsertTimesheet: %v", err)
	}

	rejected, err := repo.RejectTimesheet(ctx, testutil.OwnerID, ts.ID, "incomplete hours")
	if err != nil {
		t.Fatalf("RejectTimesheet: %v", err)
	}
	if rejected.Status != domain.TimesheetStatusRejected {
		t.Errorf("expected status %q, got %q", domain.TimesheetStatusRejected, rejected.Status)
	}
}

func TestCreateApprovalAction_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ts, err := repo.UpsertTimesheet(ctx, testutil.OwnerID, domain.TimesheetRecord{
		EmployeeID:  4101,
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-01-31",
		Status:      domain.TimesheetStatusSubmitted,
	})
	if err != nil {
		t.Fatalf("UpsertTimesheet: %v", err)
	}

	if err := repo.CreateApprovalAction(ctx, testutil.OwnerID, domain.ApprovalActionRecord{
		TimesheetID: ts.ID,
		ApproverID:  4101,
		Action:      domain.ApprovalActionApprove,
		Level:       1,
		Comment:     "looks good",
	}); err != nil {
		t.Errorf("CreateApprovalAction: %v", err)
	}
}
