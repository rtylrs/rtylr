package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	poscashv1 "github.com/voxire/voxie/gen/go/pos/cash/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "pos cash view", App: authorization.AppPOS, Permission: "pos:cash:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewCashAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListCashMovements(ctx, nil)
			return response, err
		}},
		{Name: "pos cash drop", App: authorization.AppPOS, Permission: "pos:cash:drop", Call: func(ctx context.Context) (any, error) {
			response, err := NewCashAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CashDrop(ctx, &poscashv1.CashDropRequest{BranchId: testutil.Int64Ptr(3001), TerminalId: testutil.Int64Ptr(1), SessionId: testutil.Int64Ptr(7210), Amount: testutil.Float64Ptr(1), Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos cash pickup", App: authorization.AppPOS, Permission: "pos:cash:pickup", Call: func(ctx context.Context) (any, error) {
			response, err := NewCashAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CashPickup(ctx, &poscashv1.CashPickupRequest{BranchId: testutil.Int64Ptr(3001), TerminalId: testutil.Int64Ptr(1), SessionId: testutil.Int64Ptr(7210), Amount: testutil.Float64Ptr(1), Reason: "authorization probe"})
			return response, err
		}},
		{Name: "pos cash approve", App: authorization.AppPOS, Permission: "pos:cash:approve", Call: func(ctx context.Context) (any, error) {
			response, err := NewCashAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CashAdjust(ctx, &poscashv1.CashAdjustRequest{BranchId: testutil.Int64Ptr(3001), TerminalId: testutil.Int64Ptr(1), SessionId: testutil.Int64Ptr(7210), Amount: testutil.Float64Ptr(1), Reason: "authorization probe"})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
