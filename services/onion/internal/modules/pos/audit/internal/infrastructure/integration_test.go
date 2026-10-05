package infrastructure_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/audit/internal/infrastructure"
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
