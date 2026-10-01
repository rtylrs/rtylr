package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onkithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

func expenseUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func expensePermissionStatus(err error) int {
	var oneKitErr *onkithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestExpensesAPIRequiresPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), expenseUser())
	api := &ExpensesAPI{}

	cases := []struct {
		name string
		call func() error
	}{
		{
			name: "list",
			call: func() error {
				_, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{})
				return err
			},
		},
		{
			name: "create",
			call: func() error {
				_, err := api.CreateExpense(ctx, &financeexpensesv1.CreateExpenseRequest{})
				return err
			},
		},
		{
			name: "update",
			call: func() error {
				_, err := api.UpdateExpense(ctx, &financeexpensesv1.UpdateExpenseRequest{})
				return err
			},
		},
		{
			name: "delete",
			call: func() error {
				_, err := api.DeleteExpense(ctx, &financeexpensesv1.DeleteExpenseRequest{})
				return err
			},
		},
		{
			name: "get",
			call: func() error {
				_, err := api.GetExpense(ctx, &financeexpensesv1.GetExpenseRequest{})
				return err
			},
		},
		{
			name: "status",
			call: func() error {
				_, err := api.UpdateExpenseStatus(ctx, &financeexpensesv1.UpdateExpenseStatusRequest{})
				return err
			},
		},
		{
			name: "categories",
			call: func() error {
				_, err := api.ListExpenseCategories(ctx, &financeexpensesv1.ListExpenseCategoriesRequest{})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := expensePermissionStatus(tc.call()); got != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
			}
		})
	}
}

func TestRequireExpensePermissionChecksFinanceAppAndResolvedAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), expenseUser("finance:expenses:view"))
	if err := requireExpensePermission(ctx, "finance:expenses:view"); err != nil {
		t.Fatalf("granted Finance permission rejected: %v", err)
	}

	wrongApp := expenseUser("finance:expenses:view")
	wrongApp.EffectiveAccess.Apps[0].AppID = authorization.AppCRM
	if err := requireExpensePermission(authPkg.WithAuthUser(context.Background(), wrongApp), "finance:expenses:view"); expensePermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("wrong app permission status = %d, want %d", expensePermissionStatus(err), http.StatusForbidden)
	}

	noAccess := expenseUser("finance:expenses:view")
	noAccess.EffectiveAccess.Apps[0].HasAccess = false
	if err := requireExpensePermission(authPkg.WithAuthUser(context.Background(), noAccess), "finance:expenses:view"); expensePermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("app without access permission status = %d, want %d", expensePermissionStatus(err), http.StatusForbidden)
	}

	// A permission assignment cannot replace the app's active license.
	unlicensed := expenseUser("finance:expenses:view")
	unlicensed.EffectiveAccess.Apps[0].Licensed = false
	if err := requireExpensePermission(authPkg.WithAuthUser(context.Background(), unlicensed), "finance:expenses:view"); expensePermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("an unlicensed app returned %v, want 403", err)
	}
}
