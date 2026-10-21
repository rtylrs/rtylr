package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpdateLeaveTypePatchReturnsTheUpdatedScopedRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	service := &Service{repo: repo}
	ctx := context.Background()
	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name: fmt.Sprintf("service patch %d", time.Now().UnixNano()), IsPaid: true,
		RequiresApproval: true, AccrualRule: domain.AccrualRuleMonthly, AccrualAmount: 1.5,
		MaxCarryoverDays: 4, MinNoticeDays: 2, RequiresDocumentationAfterDays: 6, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create leave type: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id = ?", lt.ID) })

	name := "Service renamed"
	updated, err := service.UpdateLeaveTypePatch(ctx, testutil.OwnerID, lt.ID, domain.LeaveTypePatch{Name: &name})
	if err != nil {
		t.Fatalf("patch leave type: %v", err)
	}
	if updated == nil || updated.ID != lt.ID || updated.Name != name || !updated.IsPaid ||
		updated.AccrualAmount != 1.5 || updated.RequiresDocumentationAfterDays != 6 {
		t.Fatalf("updated record = %+v, want scoped name patch with preserved fields", updated)
	}
	if _, err := service.UpdateLeaveTypePatch(ctx, testutil.OwnerID, 999999999, domain.LeaveTypePatch{Name: &name}); !errors.Is(err, domain.ErrLeaveTypeNotFound) {
		t.Fatalf("missing patch error = %v, want ErrLeaveTypeNotFound", err)
	}
}
