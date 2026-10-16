package application

import (
	"context"
	"errors"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestSubmitTimesheetForCaller_OnlyOwnerOrManagerMaySubmit(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const businessID int64 = 9100791
	const ownerUserID, ownerEmployeeID int64 = 4301, 4311
	const peerUserID, peerEmployeeID int64 = 4302, 4312
	const managerUserID, managerEmployeeID int64 = 4303, 4313

	t.Cleanup(func() {
		for _, stmt := range []string{
			"DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(context.Background(), stmt, businessID); err != nil {
				t.Logf("cleanup %q: %v", stmt, err)
			}
		}
	})

	permissions := newW2PermissionProvider()
	permissions.employeeByUser[ownerUserID] = ownerEmployeeID
	permissions.employeeByUser[peerUserID] = peerEmployeeID
	permissions.employeeByUser[managerUserID] = managerEmployeeID
	permissions.setPermission(managerEmployeeID, domain.PermissionApproveTimeOff, true)
	service := NewService(ctx, moduleinfra.NewRepository(db), singleLevelApprovalChain{}, permissions)

	ts, err := service.GetOrCreateTimesheet(ctx, businessID, ownerEmployeeID, "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("GetOrCreateTimesheet: %v", err)
	}

	if _, err := service.SubmitTimesheetForCaller(ctx, businessID, peerUserID, ts.ID); !errors.Is(err, domain.ErrTimesheetAccessDenied) {
		t.Fatalf("peer submit error = %v, want ErrTimesheetAccessDenied", err)
	}
	got, err := service.GetTimesheet(ctx, businessID, ts.ID)
	if err != nil {
		t.Fatalf("GetTimesheet: %v", err)
	}
	if got.Status != domain.TimesheetStatusDraft {
		t.Fatalf("denied submit changed status to %q", got.Status)
	}

	if _, err := service.SubmitTimesheetForCaller(ctx, businessID, ownerUserID, ts.ID); err != nil {
		t.Fatalf("owner submit: %v", err)
	}
	if _, err := service.SubmitTimesheetForCaller(ctx, businessID, managerUserID, ts.ID); !errors.Is(err, domain.ErrTimesheetAlreadySubmitted) {
		t.Fatalf("manager resubmit error = %v, want ErrTimesheetAlreadySubmitted", err)
	}
}
