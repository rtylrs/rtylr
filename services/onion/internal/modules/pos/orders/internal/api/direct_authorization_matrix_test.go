package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	posordersdiscountsv1 "github.com/voxire/voxie/gen/go/pos/orders/discounts/v1"
	posordersrefundsv1 "github.com/voxire/voxie/gen/go/pos/orders/refunds/v1"
	posordersv1 "github.com/voxire/voxie/gen/go/pos/orders/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "pos orders view", App: authorization.AppPOS, Permission: "pos:orders:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewOrderAPI(newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))).ListOrders(ctx, nil)
			return response, err
		}},
		{Name: "pos orders create", App: authorization.AppPOS, Permission: "pos:orders:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewOrderAPI(newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))).CreateOrder(ctx, &posordersv1.CreateOrderRequest{
				SessionId:  3201,
				TableId:    3101,
				LocationId: 3001,
				Type:       "dine_in",
				Items:      []*posordersv1.CreateOrderItemInput{{MenuItemId: 5201, Name: "Burger", Quantity: 1}},
			})
			return response, err
		}},
		{Name: "pos orders void", App: authorization.AppPOS, Permission: "pos:orders:void", Call: func(ctx context.Context) (any, error) {
			response, err := (&VoidOrderHandler{Service: newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &VoidOrderRequest{OrderID: 7211, Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos orders void approve", App: authorization.AppPOS, Permission: "pos:orders:void:approve", Call: func(ctx context.Context) (any, error) {
			response, err := (&ApproveVoidOrderHandler{Service: newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &VoidOrderApprovalRequest{OrderID: 7212, Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos orders refund", App: authorization.AppPOS, Permission: "pos:orders:refund", Call: func(ctx context.Context) (any, error) {
			response, err := NewRefundAPI(newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))).ProcessRefund(ctx, &posordersrefundsv1.ProcessRefundRequest{OrderId: 7213, Amount: testutil.Float64Ptr(1), Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos orders refund approve", App: authorization.AppPOS, Permission: "pos:orders:refund:approve", Call: func(ctx context.Context) (any, error) {
			response, err := (&ApproveRefundOrderHandler{Service: newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ProcessRefundRequest{OrderID: 7214, Amount: 1, Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos orders void override", App: authorization.AppPOS, Permission: "pos:orders:void:override", Call: func(ctx context.Context) (any, error) {
			response, err := (&OverrideVoidOrderHandler{Service: newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &VoidOrderApprovalRequest{OrderID: 7215, Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos orders refund override", App: authorization.AppPOS, Permission: "pos:orders:refund:override", Call: func(ctx context.Context) (any, error) {
			response, err := (&OverrideRefundOrderHandler{Service: newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ProcessRefundRequest{OrderID: 7216, Amount: 1, Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos orders discount", App: authorization.AppPOS, Permission: "pos:orders:discount", Call: func(ctx context.Context) (any, error) {
			response, err := NewDiscountAPI(newOrderTestService(ctx, testutil.DirectAuthorizationDB(ctx))).DiscountOrder(ctx, &posordersdiscountsv1.DiscountOrderRequest{OrderId: 7218, DiscountType: "percentage", DiscountValue: testutil.Float64Ptr(5), Reason: "authorization probe"})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
