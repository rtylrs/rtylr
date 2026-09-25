package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListUnbilledBillableTimeEntries_IsPaginatedAndTenantScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	insert := func(businessID int64, date string, duration int, projectID, taskID *int64, isBillable bool, status string, invoiceID *int64) int64 {
		result, err := db.ExecContext(ctx, `
			INSERT INTO hr_time_entries
				(owner_id, business_id, employee_id, project_id, task_id, date,
				 duration_minutes, description, is_billable, status, source, billable_invoice_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			businessID, businessID, 4101, projectID, taskID, date, duration,
			"billable test entry", isBillable, status, "flow", invoiceID,
		)
		if err != nil {
			t.Fatalf("insert time entry: %v", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("read inserted time entry id: %v", err)
		}
		return id
	}

	projectID := int64(7001)
	taskID := int64(8001)
	first := insert(testutil.BusinessID, "2026-08-01", 45, &projectID, &taskID, true, "approved", nil)
	second := insert(testutil.BusinessID, "2026-08-02", 60, nil, nil, true, "approved", nil)
	third := insert(testutil.BusinessID, "2026-08-03", 15, nil, nil, true, "approved", nil)
	insert(testutil.BusinessID, "2026-08-04", 90, nil, nil, false, "approved", nil)
	insert(testutil.BusinessID, "2026-08-05", 90, nil, nil, true, "draft", nil)
	invoicedID := int64(9001)
	insert(testutil.BusinessID, "2026-08-06", 90, nil, nil, true, "approved", &invoicedID)
	foreign := insert(testutil.OtherBusinessID, "2026-08-01", 120, nil, nil, true, "approved", nil)
	t.Cleanup(func() {
		for _, id := range []int64{first, second, third, foreign} {
			if _, err := db.ExecContext(ctx, "DELETE FROM hr_time_entries WHERE id = ?", id); err != nil {
				t.Errorf("cleanup time entry %d: %v", id, err)
			}
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM hr_time_entries WHERE business_id = ? AND date BETWEEN ? AND ? AND description = ?", testutil.BusinessID, "2026-08-04", "2026-08-06", "billable test entry"); err != nil {
			t.Errorf("cleanup non-eligible time entries: %v", err)
		}
	})

	entries, err := repo.ListUnbilledBillableTimeEntries(ctx, testutil.BusinessID, 2, 0)
	if err != nil {
		t.Fatalf("ListUnbilledBillableTimeEntries: %v", err)
	}
	total, totalMinutes, err := repo.CountUnbilledBillableTimeEntries(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CountUnbilledBillableTimeEntries: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if totalMinutes != 120 {
		t.Fatalf("total minutes = %d, want 120", totalMinutes)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want page size 2", len(entries))
	}
	if entries[0].ID != first || entries[1].ID != second {
		t.Errorf("entry order = [%d, %d], want [%d, %d]", entries[0].ID, entries[1].ID, first, second)
	}
	if entries[0].ProjectID == nil || *entries[0].ProjectID != projectID {
		t.Errorf("project_id = %v, want %d", entries[0].ProjectID, projectID)
	}
	if entries[0].TaskID == nil || *entries[0].TaskID != taskID {
		t.Errorf("task_id = %v, want %d", entries[0].TaskID, taskID)
	}
	if entries[0].DurationMinutes != 45 || entries[1].DurationMinutes != 60 {
		t.Errorf("durations = [%d, %d], want [45, 60]", entries[0].DurationMinutes, entries[1].DurationMinutes)
	}

	secondPage, err := repo.ListUnbilledBillableTimeEntries(ctx, testutil.BusinessID, 2, 2)
	if err != nil {
		t.Fatalf("ListUnbilledBillableTimeEntries second page: %v", err)
	}
	secondTotal, _, err := repo.CountUnbilledBillableTimeEntries(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CountUnbilledBillableTimeEntries second page: %v", err)
	}
	if secondTotal != total {
		t.Errorf("second page total = %d, want %d", secondTotal, total)
	}
	if len(secondPage) != 1 || secondPage[0].ID != third {
		t.Errorf("second page = %#v, want only entry %d", secondPage, third)
	}
}
