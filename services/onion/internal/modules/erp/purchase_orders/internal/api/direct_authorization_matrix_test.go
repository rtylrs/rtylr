package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	erppurchaseorderv1 "github.com/voxire/voxie/gen/go/erp/purchase_order/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "erp purchasing view", App: authorization.AppERP, Permission: "erp:purchasing:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewPurchaseOrderAPI(newPurchaseOrderTestService(testutil.DirectAuthorizationDB(ctx))).ListPurchaseOrders(ctx, nil)
			return response, err
		}},
		{Name: "erp purchasing create", App: authorization.AppERP, Permission: "erp:purchasing:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewPurchaseOrderAPI(newPurchaseOrderTestService(testutil.DirectAuthorizationDB(ctx))).CreatePurchaseOrder(ctx, &erppurchaseorderv1.CreatePurchaseOrderRequest{SupplierId: testutil.Int64Ptr(5101), PoNumber: "AUTH-PROBE", Items: []*erppurchaseorderv1.CreatePurchaseOrderItemRequest{{ProductId: testutil.Int64Ptr(5201), QuantityOrdered: 1, UnitCost: 5}}})
			return response, err
		}},
		{Name: "erp purchasing approve", App: authorization.AppERP, Permission: "erp:purchasing:approve", Call: func(ctx context.Context) (any, error) {
			response, err := NewPurchaseOrderAPI(newPurchaseOrderTestService(testutil.DirectAuthorizationDB(ctx))).ApprovePurchaseOrder(ctx, &erppurchaseorderv1.ApprovePurchaseOrderRequest{Id: 7301})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
