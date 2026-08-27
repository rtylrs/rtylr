package api

import (
	"context"
	"testing"

	erpstocktransferv1 "github.com/voxire/voxie/gen/go/erp/stock_transfer/v1"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/application"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "erp inventory transfer", App: authorization.AppERP, Permission: "erp:inventory:transfer", Call: func(ctx context.Context) (any, error) {
			response, err := NewStockTransferAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CreateStockTransfer(ctx, &erpstocktransferv1.CreateStockTransferRequest{FromWarehouseId: 1, ToWarehouseId: 2, Items: []*erpstocktransferv1.CreateStockTransferItemRequest{{ProductId: 5201, Quantity: 1}}})
			return response, err
		}},
	})
}
