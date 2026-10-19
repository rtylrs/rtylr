package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/expenses/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

const (
	expenseCategoriesBusinessID int64 = 907201
	expenseCategoriesUserID     int64 = 907202
)

func expenseCategoriesAPI(t *testing.T, db *sql.DB) (*ExpensesAPI, context.Context) {
	t.Helper()
	testutil.GrantBusinessAccess(t, db, expenseCategoriesBusinessID, expenseCategoriesUserID)
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_expenses WHERE business_id = ?", expenseCategoriesBusinessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_expense_categories WHERE business_id = ?", expenseCategoriesBusinessID)
	}
	cleanup()
	t.Cleanup(cleanup)

	user := &authPkg.AuthUser{
		BusinessID: expenseCategoriesBusinessID,
		UserID:     expenseCategoriesUserID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: expenseCategoriesBusinessID,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"finance:expenses:view", "finance:expenses:create"},
			}},
		},
	}
	ctx := authPkg.WithAuthUser(context.Background(), user)
	return NewExpensesAPI(application.NewService(ctx, moduleinfra.NewRepository(db))), ctx
}

func categoryNames(t *testing.T, api *ExpensesAPI, ctx context.Context) []string {
	t.Helper()
	res, err := api.ListExpenseCategories(ctx, &financeexpensesv1.ListExpenseCategoriesRequest{})
	if err != nil {
		t.Fatalf("ListExpenseCategories: %v", err)
	}
	names := make([]string, 0, len(res.GetCategories()))
	for _, c := range res.GetCategories() {
		names = append(names, c.GetName())
	}
	return names
}

func wantStatus(t *testing.T, err error, status int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected HTTP %d, got success", status)
	}
	var coded interface{ HTTPStatusCode() int }
	if !errors.As(err, &coded) || coded.HTTPStatusCode() != status {
		t.Fatalf("expected HTTP %d, got %v", status, err)
	}
}

// Regression: the portal created categories through finance.report.v1 and
// listed them through finance.expenses.v1, whose cache the create never
// cleared. A created category must be listed by the very next read, and
// pressing Create again must not add a second row.
func TestExpenseCategories_CreateIsListedAtOnceAndNeverDuplicates(t *testing.T) {
	db := testutil.OpenDB(t)
	api, ctx := expenseCategoriesAPI(t, db)

	if names := categoryNames(t, api, ctx); len(names) != 0 {
		t.Fatalf("fresh business lists %v, want none", names)
	}

	first, err := api.CreateExpenseCategory(ctx, &financeexpensesv1.CreateExpenseCategoryRequest{Name: "  Communications ", Color: "#14b8a6"})
	if err != nil {
		t.Fatalf("CreateExpenseCategory: %v", err)
	}
	if !first.GetCreated() || first.GetCategory().GetId() == 0 || first.GetCategory().GetName() != "Communications" {
		t.Fatalf("first create = %+v, want a new trimmed row", first)
	}

	if names := categoryNames(t, api, ctx); len(names) != 1 || names[0] != "Communications" {
		t.Fatalf("after create lists %v, want [Communications]", names)
	}

	again, err := api.CreateExpenseCategory(ctx, &financeexpensesv1.CreateExpenseCategoryRequest{Name: "communications", Color: "#000000"})
	if err != nil {
		t.Fatalf("second CreateExpenseCategory: %v", err)
	}
	if again.GetCreated() || again.GetCategory().GetId() != first.GetCategory().GetId() {
		t.Fatalf("second create = %+v, want the existing row %d back", again, first.GetCategory().GetId())
	}
	if names := categoryNames(t, api, ctx); len(names) != 1 {
		t.Fatalf("after duplicate create lists %v, want one row", names)
	}

	_, err = api.CreateExpenseCategory(ctx, &financeexpensesv1.CreateExpenseCategoryRequest{Name: "   "})
	wantStatus(t, err, http.StatusBadRequest)
}

func TestExpenseCategories_UpdateRenamesAndRefusesTakenNames(t *testing.T) {
	db := testutil.OpenDB(t)
	api, ctx := expenseCategoriesAPI(t, db)

	travel, err := api.CreateExpenseCategory(ctx, &financeexpensesv1.CreateExpenseCategoryRequest{Name: "Travel", Color: "#111111"})
	if err != nil {
		t.Fatalf("create Travel: %v", err)
	}
	if _, err = api.CreateExpenseCategory(ctx, &financeexpensesv1.CreateExpenseCategoryRequest{Name: "Rent"}); err != nil {
		t.Fatalf("create Rent: %v", err)
	}

	updated, err := api.UpdateExpenseCategory(ctx, &financeexpensesv1.UpdateExpenseCategoryRequest{
		Id:   travel.GetCategory().GetId(),
		Body: &financeexpensesv1.UpdateExpenseCategoryBody{Name: "Travel & Transport", Color: "#222222"},
	})
	if err != nil {
		t.Fatalf("UpdateExpenseCategory: %v", err)
	}
	if updated.GetCategory().GetName() != "Travel & Transport" || updated.GetCategory().GetColor() != "#222222" {
		t.Fatalf("update returned %+v", updated.GetCategory())
	}
	if names := categoryNames(t, api, ctx); len(names) != 2 || names[0] != "Rent" || names[1] != "Travel & Transport" {
		t.Fatalf("after rename lists %v", names)
	}

	// Changing only the colour keeps the category's own name.
	if _, err = api.UpdateExpenseCategory(ctx, &financeexpensesv1.UpdateExpenseCategoryRequest{
		Id:   travel.GetCategory().GetId(),
		Body: &financeexpensesv1.UpdateExpenseCategoryBody{Name: "travel & transport", Color: "#333333"},
	}); err != nil {
		t.Fatalf("recolour: %v", err)
	}

	_, err = api.UpdateExpenseCategory(ctx, &financeexpensesv1.UpdateExpenseCategoryRequest{
		Id:   travel.GetCategory().GetId(),
		Body: &financeexpensesv1.UpdateExpenseCategoryBody{Name: "RENT"},
	})
	wantStatus(t, err, http.StatusConflict)

	_, err = api.UpdateExpenseCategory(ctx, &financeexpensesv1.UpdateExpenseCategoryRequest{
		Id:   travel.GetCategory().GetId() + 1_000_000,
		Body: &financeexpensesv1.UpdateExpenseCategoryBody{Name: "Ghost"},
	})
	wantStatus(t, err, http.StatusNotFound)
}

func TestExpenseCategories_DeleteOnlyWhenNoExpenseIsFiledUnderIt(t *testing.T) {
	db := testutil.OpenDB(t)
	api, ctx := expenseCategoriesAPI(t, db)

	used, err := api.CreateExpenseCategory(ctx, &financeexpensesv1.CreateExpenseCategoryRequest{Name: "Software"})
	if err != nil {
		t.Fatalf("create Software: %v", err)
	}
	spare, err := api.CreateExpenseCategory(ctx, &financeexpensesv1.CreateExpenseCategoryRequest{Name: "Spare"})
	if err != nil {
		t.Fatalf("create Spare: %v", err)
	}

	usedID := used.GetCategory().GetId()
	if _, err = api.CreateExpense(ctx, &financeexpensesv1.CreateExpenseRequest{
		CategoryId: &usedID, Amount: 12, Currency: "USD", Description: "Licence", ExpenseDate: "2026-10-01",
	}); err != nil {
		t.Fatalf("CreateExpense: %v", err)
	}

	_, err = api.DeleteExpenseCategory(ctx, &financeexpensesv1.DeleteExpenseCategoryRequest{Id: usedID})
	wantStatus(t, err, http.StatusConflict)

	if _, err = api.DeleteExpenseCategory(ctx, &financeexpensesv1.DeleteExpenseCategoryRequest{Id: spare.GetCategory().GetId()}); err != nil {
		t.Fatalf("delete Spare: %v", err)
	}
	if names := categoryNames(t, api, ctx); len(names) != 1 || names[0] != "Software" {
		t.Fatalf("after delete lists %v, want [Software]", names)
	}

	_, err = api.DeleteExpenseCategory(ctx, &financeexpensesv1.DeleteExpenseCategoryRequest{Id: spare.GetCategory().GetId()})
	wantStatus(t, err, http.StatusNotFound)
}

// An expense may only be filed under one of its own business's categories.
func TestExpenseCategories_ExpenseRejectsAnotherBusinessCategory(t *testing.T) {
	db := testutil.OpenDB(t)
	api, ctx := expenseCategoriesAPI(t, db)

	res, err := db.ExecContext(context.Background(),
		"INSERT INTO finance_expense_categories (owner_id, business_id, name) VALUES (?, ?, ?)",
		expenseCategoriesUserID, expenseCategoriesBusinessID+1, "Foreign")
	if err != nil {
		t.Fatalf("seed foreign category: %v", err)
	}
	foreignID, _ := res.LastInsertId()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_expense_categories WHERE id = ?", foreignID)
	})

	_, err = api.CreateExpense(ctx, &financeexpensesv1.CreateExpenseRequest{
		CategoryId: &foreignID, Amount: 5, Currency: "USD", Description: "Probe", ExpenseDate: "2026-10-01",
	})
	wantStatus(t, err, http.StatusBadRequest)
}
