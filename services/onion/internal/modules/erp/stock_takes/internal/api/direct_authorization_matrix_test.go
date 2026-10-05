package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/erp/stock_takes/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	erpstocktakev1 "github.com/voxire/voxie/gen/go/erp/stock_take/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "erp inventory approve", App: authorization.AppERP, Permission: "erp:inventory:approve", Call: func(ctx context.Context) (any, error) {
			response, err := NewStockTakeAPI(application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))).UpdateStockTake(ctx, &erpstocktakev1.UpdateStockTakeRequest{Id: 7303, Body: &erpstocktakev1.UpdateStockTakeBody{Status: "completed", Items: []*erpstocktakev1.UpdateStockTakeItemBody{{ProductId: 5201, ExpectedQty: 50, CountedQty: 50}}}})
			return response, err
		}},
	})
}
