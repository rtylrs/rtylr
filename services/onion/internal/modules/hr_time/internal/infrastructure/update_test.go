package infrastructure_test

import (
	"context"
	"database/sql"
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

func TestMarkEntryExported_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entry, err := repo.CreateTimeEntry(ctx, testutil.OwnerID, domain.TimeEntryRecord{
		EmployeeID:      4101,
		Date:            "2026-04-15",
		DurationMinutes: 60,
		IsBillable:      true,
		Status:          domain.EntryStatusApproved,
	})
	if err != nil {
		t.Fatalf("CreateTimeEntry: %v", err)
	}

	if err := repo.MarkEntryExported(ctx, testutil.OwnerID, entry.ID, 9001); err != nil {
		t.Errorf("MarkEntryExported: %v", err)
	}
}

func TestExportApprovedBillableEntries_MarksOneBatch(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	createdIDs := make([]int64, 0, 2)
	for _, description := range []string{"atomic export one", "atomic export two"} {
		entry, err := repo.CreateTimeEntry(ctx, testutil.OwnerID, domain.TimeEntryRecord{
			EmployeeID:      4101,
			Date:            "2026-04-02",
			DurationMinutes: 60,
			Description:     description,
			IsBillable:      true,
			Status:          "approved",
		})
		if err != nil {
			t.Fatalf("CreateTimeEntry: %v", err)
		}
		createdIDs = append(createdIDs, entry.ID)
	}

	batchID := int64(91001)
	exported, err := repo.ExportApprovedBillableEntries(ctx, testutil.OwnerID, batchID)
	if err != nil {
		t.Fatalf("ExportApprovedBillableEntries: %v", err)
	}
	if len(exported) < len(createdIDs) {
		t.Fatalf("exported %d entries, want at least %d", len(exported), len(createdIDs))
	}

	for _, id := range createdIDs {
		var invoiceID, exportBatchID sql.NullInt64
		if err := db.QueryRowContext(ctx,
			"SELECT billable_invoice_id, export_batch_id FROM hr_time_entries WHERE id = ? AND business_id = ?",
			id, testutil.OwnerID,
		).Scan(&invoiceID, &exportBatchID); err != nil {
			t.Fatalf("read export markers for entry %d: %v", id, err)
		}
		if invoiceID.Valid {
			t.Errorf("entry %d billable_invoice_id = %d, want NULL before invoicing", id, invoiceID.Int64)
		}
		if !exportBatchID.Valid || exportBatchID.Int64 != batchID {
			t.Errorf("entry %d export_batch_id = %v, want %d", id, exportBatchID, batchID)
		}
	}
}
