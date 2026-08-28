package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/expenses/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "finance expenses view", App: authorization.AppFinance, Permission: "finance:expenses:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewExpensesAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListExpenses(ctx, nil)
			return response, err
		}},
		{Name: "finance expenses create", App: authorization.AppFinance, Permission: "finance:expenses:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewExpensesAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CreateExpense(ctx, &financeexpensesv1.CreateExpenseRequest{Amount: 10, Currency: "USD", Description: "Authorization Probe", ExpenseDate: "2026-08-27"})
			return response, err
		}},
		{Name: "finance expenses approve", App: authorization.AppFinance, Permission: "finance:expenses:approve", Call: func(ctx context.Context) (any, error) {
			response, err := NewExpensesAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).UpdateExpenseStatus(ctx, &financeexpensesv1.UpdateExpenseStatusRequest{Id: 1, Body: &financeexpensesv1.UpdateExpenseStatusBody{Status: "approved"}})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
