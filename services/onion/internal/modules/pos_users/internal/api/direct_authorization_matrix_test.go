package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_users/internal/application"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "pos employees manage", App: authorization.AppPOS, Permission: "pos:employees:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewUserAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListPOSUsers(ctx, nil)
			return response, err
		}},
	})
}
