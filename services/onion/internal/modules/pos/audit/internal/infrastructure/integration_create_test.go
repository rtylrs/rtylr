package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/audit/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateAuditLog_BranchIDPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "branch_id_check",
		EntityType: "branch",
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_audit_logs WHERE id = ?", id)
	})

	logs, err := repo.List(ctx, domain.AuditLogFilter{BusinessID: fixtureBusinessID})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, l := range logs {
		if l.ID == id {
			if l.BranchID != fixtureBranchID {
				t.Errorf("BranchID mismatch: got %d, want %d", l.BranchID, fixtureBranchID)
			}
			return
		}
	}
	t.Errorf("log id=%d not found", id)
}

func TestCreateAuditLog_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "positive_id_check",
		EntityType: "misc",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_audit_logs WHERE id = ?", id)
	})
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestListAuditLogs_EndDateIncludesRowsCreatedLaterThatDay(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Insert directly so the row lands at 09:00 on the end date; a date-only
	// upper bound parsed at midnight would drop it (B55-R06).
	const action = "end_date_inclusive_test"
	res, err := db.ExecContext(ctx, `
		INSERT INTO pos_audit_log (business_id, branch_id, employee_id, action, entity_type, created_at)
		VALUES (?, ?, ?, ?, 'misc', '2026-03-15 09:00:00')`,
		fixtureBusinessID, fixtureBranchID, fixtureEmployeeID, action)
	if err != nil {
		t.Fatalf("insert audit row: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("audit row id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM pos_audit_log WHERE id = ?", id)
	})

	actionFilter := action
	end := "2026-03-15"
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		Action:     &actionFilter,
		EndDate:    &end,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	found := false
	for _, l := range logs {
		if l.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("row created at 09:00 on end_date=%s was excluded from the audit list", end)
	}

	previousDay := "2026-03-14"
	logs, err = repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		Action:     &actionFilter,
		EndDate:    &previousDay,
	})
	if err != nil {
		t.Fatalf("List error for previous day: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("end_date=%s returned %d rows created the next day", previousDay, len(logs))
	}
}

func TestListAuditLogs_DateRangeIncludesCreatedRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "date_range_test",
		EntityType: "misc",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_audit_logs WHERE id = ?", id)
	})

	start := "2026-01-01"
	end := "2099-12-31"
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		StartDate:  &start,
		EndDate:    &end,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	var found bool
	for _, l := range logs {
		if l.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("inserted log id=%d not found in wide date-range filter", id)
	}
}
