package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "finance bank feeds view", App: authorization.AppFinance, Permission: "finance:bank_feeds:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ListRequest{})
			return response, err
		}},
	})
}
