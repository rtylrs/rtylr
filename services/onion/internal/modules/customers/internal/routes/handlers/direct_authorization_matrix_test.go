package handlers

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "crm customers export", App: authorization.AppCRM, Permission: "crm:customers:export", Call: func(ctx context.Context) (any, error) {
			response, err := (&ExportCustomerHandler{CRMService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ExportCustomerRequest{CustomerID: 5301})
			return response, err
		}},
		{Name: "crm loyalty manage", App: authorization.AppCRM, Permission: "crm:loyalty:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&UpdateLoyaltyConfigHandler{CRMService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &UpdateLoyaltyConfigRequest{IsEnabled: true, PointsPerCurrencyUnit: 1, CurrencyPerPoint: 1})
			return response, err
		}},
	})
}
