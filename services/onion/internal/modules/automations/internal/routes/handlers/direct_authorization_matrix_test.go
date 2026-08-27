package handlers

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/application"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "crm automations manage", App: authorization.AppCRM, Permission: "crm:automations:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListAutomationsHandler{Service: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, nil)
			return response, err
		}},
	})
}
