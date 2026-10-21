package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hr schedules manage", App: authorization.AppHR, Permission: "hr:schedules:manage", Call: func(ctx context.Context) (any, error) {
			repo := moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx))
			response, err := (&CreateShiftHandler{Service: application.NewService(ctx, repo, nil, nil, repo)}).Serve(ctx, &CreateShiftRequest{LocationID: 1, Name: "authorization probe", StartTime: "09:00", EndTime: "17:00"})
			return response, err
		}},
	})
}
