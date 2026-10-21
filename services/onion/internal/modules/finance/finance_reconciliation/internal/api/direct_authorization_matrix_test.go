package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "finance reconciliation manage", App: authorization.AppFinance, Permission: "finance:reconciliation:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewReconciliationAPI(newReconciliationTestService(testutil.DirectAuthorizationDB(ctx))).GetOverview(ctx, nil)
			return response, err
		}},
	})
}
