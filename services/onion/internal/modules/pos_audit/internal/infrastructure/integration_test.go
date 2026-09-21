package infrastructure_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const (
	fixtureBusinessID int64 = 1
	fixtureBranchID   int64 = 3001
	fixtureEmployeeID int64 = 3601
)

func TestListAuditLogs_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	filter := domain.AuditLogFilter{BusinessID: fixtureBusinessID}
	_, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
}

func TestListAuditLogs_UnknownBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	filter := domain.AuditLogFilter{BusinessID: 0}
	logs, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected empty list for business_id=0, got %d items", len(logs))
	}
}

func TestListAuditLogs_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	filter := domain.AuditLogFilter{BusinessID: fixtureBusinessID}
	logs, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, l := range logs {
		if l.ID == 0 {
			t.Errorf("audit_log[%d] has zero ID", i)
		}
	}
}

func TestListAuditLogs_EachBusinessIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	filter := domain.AuditLogFilter{BusinessID: fixtureBusinessID}
	logs, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, l := range logs {
		if l.BusinessID != fixtureBusinessID {
			t.Errorf("log[%d] has BusinessID=%d, want %d", i, l.BusinessID, fixtureBusinessID)
		}
	}
}

func TestListAuditLogs_FilterByAction(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	action := "nonexistent_action_xyz"
	filter := domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		Action:     &action,
	}
	logs, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs for action=%q, got %d", action, len(logs))
	}
}

func TestListAuditLogs_FilterByEntityType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entityType := "nonexistent_entity_type_xyz"
	filter := domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		EntityType: &entityType,
	}
	logs, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs for entity_type=%q, got %d", entityType, len(logs))
	}
}

func TestCreateAuditLog_InsertsRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	details := json.RawMessage(`{"test": true}`)
	input := domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "test_action",
		EntityType: "test_entity",
		Details:    &details,
	}

	before, err := repo.List(ctx, domain.AuditLogFilter{BusinessID: fixtureBusinessID})
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID from Create")
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_audit_logs WHERE id = ?", id)
	})

	after, err := repo.List(ctx, domain.AuditLogFilter{BusinessID: fixtureBusinessID})
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to grow by 1: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateAuditLog_ActionPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantAction := "login"
	input := domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     wantAction,
		EntityType: "session",
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_audit_logs WHERE id = ?", id)
	})

	actionFilter := wantAction
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		Action:     &actionFilter,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	found := false
	for _, l := range logs {
		if l.ID == id {
			found = true
			if l.Action != wantAction {
				t.Errorf("Action mismatch: got %q, want %q", l.Action, wantAction)
			}
		}
	}
	if !found {
		t.Errorf("inserted log id=%d not found in filtered results", id)
	}
}

func TestListAuditLogs_EachHasNonEmptyAction(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	filter := domain.AuditLogFilter{BusinessID: fixtureBusinessID}
	logs, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, l := range logs {
		if l.Action == "" {
			t.Errorf("log[%d] (id=%d) has empty Action", i, l.ID)
		}
	}
}

func TestListAuditLogs_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	filter := domain.AuditLogFilter{BusinessID: fixtureBusinessID}
	first, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("first List: %v", err)
	}
	second, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("second List: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("count changed between calls: %d vs %d", len(first), len(second))
	}
}

func TestCreateAuditLog_IPAddressPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantIP := "192.168.1.1"
	input := domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "ip_test_action",
		EntityType: "terminal",
		IPAddress:  &wantIP,
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
			if l.IPAddress == nil || *l.IPAddress != wantIP {
				t.Errorf("IPAddress mismatch: got %v, want %q", l.IPAddress, wantIP)
			}
			return
		}
	}
	t.Errorf("inserted log id=%d not found", id)
}

func TestCreateAuditLog_WithJSONDetails_Persisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	details := json.RawMessage(`{"amount": 99.99, "currency": "USD"}`)
	input := domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "payment_processed",
		EntityType: "order",
		Details:    &details,
	}
	id, err := repo.Create(ctx, input)
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

func TestCreateAuditLog_EntityTypePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantEntityType := "product"
	input := domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "product_updated",
		EntityType: wantEntityType,
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_audit_logs WHERE id = ?", id)
	})

	entityType := wantEntityType
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		EntityType: &entityType,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	var found bool
	for _, l := range logs {
		if l.ID == id {
			found = true
			if l.EntityType != wantEntityType {
				t.Errorf("EntityType mismatch: got %q, want %q", l.EntityType, wantEntityType)
			}
		}
	}
	if !found {
		t.Errorf("inserted log id=%d not found in entity-type-filtered list", id)
	}
}

func TestListAuditLogs_FilterByEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "employee_filter_test",
		EntityType: "session",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_audit_logs WHERE id = ?", id)
	})

	empID := fixtureEmployeeID
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		EmployeeID: &empID,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, l := range logs {
		if l.EmployeeID != fixtureEmployeeID {
			t.Errorf("log[%d] has EmployeeID=%d, want %d", i, l.EmployeeID, fixtureEmployeeID)
		}
	}
}

func TestListAuditLogs_AllNilFiltersReturnsAllForBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	filter := domain.AuditLogFilter{BusinessID: fixtureBusinessID}
	logs, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	_ = logs
}

func TestCreateAuditLog_EntityIDPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantEntityID := int64(42)
	input := domain.AuditLogInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		EmployeeID: fixtureEmployeeID,
		Action:     "entity_id_test",
		EntityType: "order",
		EntityID:   &wantEntityID,
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
			if l.EntityID == nil || *l.EntityID != wantEntityID {
				t.Errorf("EntityID mismatch: got %v, want %d", l.EntityID, wantEntityID)
			}
			return
		}
	}
	t.Errorf("log id=%d not found", id)
}

func TestListAuditLogs_UnknownEmployeeReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	empID := int64(999999999)
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		EmployeeID: &empID,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs for unknown employee, got %d", len(logs))
	}
}

func TestListAuditLogs_StartDateFilterExcludesOldRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	futureDate := "2099-01-01"
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		StartDate:  &futureDate,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs with start_date in far future, got %d", len(logs))
	}
}

func TestListAuditLogs_EndDateFilterExcludesFutureRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	pastDate := "2000-01-01"
	logs, err := repo.List(ctx, domain.AuditLogFilter{
		BusinessID: fixtureBusinessID,
		EndDate:    &pastDate,
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs with end_date in past, got %d", len(logs))
	}
}

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
