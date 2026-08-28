package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "finance reports view", App: authorization.AppFinance, Permission: "finance:reports:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewStatementAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetBalanceSheet(ctx, nil)
			return response, err
		}},
		{Name: "finance reports export", App: authorization.AppFinance, Permission: "finance:reports:export", Call: func(ctx context.Context) (any, error) {
			response, err := NewStatementAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ExportCashFlow(ctx, nil)
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
