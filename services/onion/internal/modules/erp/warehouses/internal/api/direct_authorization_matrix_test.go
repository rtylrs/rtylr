package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/erp/warehouses/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "erp warehouses manage", App: authorization.AppERP, Permission: "erp:warehouses:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewWarehouseAPI(application.NewService(
				moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)),
			)).ListWarehouses(ctx, nil)
			return response, err
		}},
	})
}
