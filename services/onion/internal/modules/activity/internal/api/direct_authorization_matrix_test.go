package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/activity/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/activity/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hub audit view", App: authorization.AppHub, Permission: "hub:audit:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewActivityAPI(application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))).ListActivity(ctx, nil)
			return response, err
		}},
	})
}
