package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hub security manage", App: authorization.AppHub, Permission: "hub:security:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewSecurityAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetSecurityState(ctx, nil)
			return response, err
		}},
	})
}
