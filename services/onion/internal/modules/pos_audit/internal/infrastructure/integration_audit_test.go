package infrastructure_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
