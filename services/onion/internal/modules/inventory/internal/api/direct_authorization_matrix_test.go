package api

import (
	"context"
	"testing"

	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "erp inventory view", App: authorization.AppERP, Permission: "erp:inventory:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewInventoryAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetInventoryReport(ctx, nil)
			return response, err
		}},
		{Name: "erp inventory adjust", App: authorization.AppERP, Permission: "erp:inventory:adjust", Call: func(ctx context.Context) (any, error) {
			response, err := NewInventoryAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).AdjustInventory(ctx, &erpinventoryv1.AdjustInventoryRequest{ProductId: testutil.Int64Ptr(5201), Change: testutil.Float64Ptr(1), Type: "adjustment"})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
