package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/kpis/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hub dashboard view", App: authorization.AppHub, Permission: "hub:dashboard:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewDashboardAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetKPIs(ctx, nil)
			return response, err
		}},
	})
}
