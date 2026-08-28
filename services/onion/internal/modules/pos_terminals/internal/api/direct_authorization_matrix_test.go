package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "pos terminals manage", App: authorization.AppPOS, Permission: "pos:terminals:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewTerminalAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListTerminals(ctx, nil)
			return response, err
		}},
	})
}
