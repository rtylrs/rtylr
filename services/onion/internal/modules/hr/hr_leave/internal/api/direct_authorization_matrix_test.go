package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_leave/internal/infrastructure"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hr leave request", App: authorization.AppHR, Permission: "hr:time_off:request", Call: func(ctx context.Context) (any, error) {
			response, err := (&CreateLeaveRequestHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)), nil, hrcore.NewModule(ctx, testutil.DirectAuthorizationDB(ctx)))}).Serve(ctx, &CreateLeaveRequestRequest{EmployeeID: 4101, LeaveTypeID: 7203, StartDate: "2026-11-10", EndDate: "2026-11-11", TotalDays: 2})
			return response, err
		}},
		{Name: "hr leave approve", App: authorization.AppHR, Permission: "hr:time_off:approve", Call: func(ctx context.Context) (any, error) {
			response, err := (&ApproveLeaveHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)), nil, hrcore.NewModule(ctx, testutil.DirectAuthorizationDB(ctx)))}).Serve(ctx, &ApproveLeaveRequest{RequestID: 7205, Approved: true})
			return response, err
		}},
	})
}
