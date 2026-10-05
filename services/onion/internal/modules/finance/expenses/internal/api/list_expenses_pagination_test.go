package api

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/expenses/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

// A dedicated business keeps the pagination fixture away from the shared
// fixture tenant and from the seeded businesses other packages own.
const (
	expensesPaginationBusinessID int64 = 907101
	expensesPaginationUserID     int64 = 907001
	expensesPaginationRows       int   = 505
)

func expensesPaginationUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: expensesPaginationBusinessID,
		UserID:     expensesPaginationUserID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: expensesPaginationBusinessID,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

// seedPaginationExpenses inserts 505 expenses with distinct, strictly
// decreasing expense dates so the page boundaries are deterministic.
func seedPaginationExpenses(t *testing.T, db *sql.DB) {
	t.Helper()

	testutil.GrantBusinessAccess(t, db, expensesPaginationBusinessID, expensesPaginationUserID)

	placeholders := make([]string, 0, expensesPaginationRows)
	args := make([]any, 0, expensesPaginationRows*5)
	base := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	for i := 0; i < expensesPaginationRows; i++ {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
		args = append(args,
			expensesPaginationUserID,
			expensesPaginationBusinessID,
			fmt.Sprintf("pagination expense %03d", i),
			float64(1+i),
			base.AddDate(0, 0, -i),
		)
	}
	if _, err := db.ExecContext(context.Background(),
		"INSERT INTO finance_expenses (owner_id, business_id, description, amount, expense_date) VALUES "+
			strings.Join(placeholders, ", "), args...); err != nil {
		t.Fatalf("seed pagination expenses: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			"DELETE FROM finance_expenses WHERE business_id = ?", expensesPaginationBusinessID)
	})
}

func paginationExpensesAPI(t *testing.T, db *sql.DB) (*ExpensesAPI, context.Context) {
	t.Helper()
	ctx := authPkg.WithAuthUser(context.Background(), expensesPaginationUser("finance:expenses:view"))
	return NewExpensesAPI(application.NewService(ctx, moduleinfra.NewRepository(db))), ctx
}

func expenseIDs(expenses []*financeexpensesv1.Expense) []int64 {
	ids := make([]int64, 0, len(expenses))
	for _, expense := range expenses {
		ids = append(ids, expense.GetId())
	}
	return ids
}

// Regression: ListExpenses ignored page/page_size and replayed the first 500
// rows on every request. Each page must now return a distinct window and echo
// the pagination it applied.
func TestListExpenses_HonoursPageAndPageSize(t *testing.T) {
	db := testutil.OpenDB(t)
	seedPaginationExpenses(t, db)
	api, ctx := paginationExpensesAPI(t, db)

	first, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListExpenses page 1: %v", err)
	}
	if first.GetPage() != 1 || first.GetPageSize() != 50 {
		t.Fatalf("page 1 echoed page=%d page_size=%d, want 1/50", first.GetPage(), first.GetPageSize())
	}
	if len(first.GetExpenses()) != 50 {
		t.Fatalf("page 1 returned %d rows, want 50", len(first.GetExpenses()))
	}
	if !first.GetHasMore() {
		t.Fatal("page 1 has_more = false, want true for a full page of 505 rows")
	}

	second, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{Page: 2, PageSize: 50})
	if err != nil {
		t.Fatalf("ListExpenses page 2: %v", err)
	}
	if second.GetPage() != 2 || second.GetPageSize() != 50 {
		t.Fatalf("page 2 echoed page=%d page_size=%d, want 2/50", second.GetPage(), second.GetPageSize())
	}
	if len(second.GetExpenses()) != 50 {
		t.Fatalf("page 2 returned %d rows, want 50", len(second.GetExpenses()))
	}
	seen := make(map[int64]bool, len(first.GetExpenses()))
	for _, id := range expenseIDs(first.GetExpenses()) {
		seen[id] = true
	}
	for _, id := range expenseIDs(second.GetExpenses()) {
		if seen[id] {
			t.Fatalf("expense id=%d appears on both page 1 and page 2", id)
		}
	}

	last, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{Page: 11, PageSize: 50})
	if err != nil {
		t.Fatalf("ListExpenses page 11: %v", err)
	}
	if len(last.GetExpenses()) != 5 {
		t.Fatalf("page 11 returned %d rows, want the trailing 5", len(last.GetExpenses()))
	}
	if last.GetHasMore() {
		t.Fatal("page 11 has_more = true, want false on the trailing short page")
	}

	empty, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{Page: 12, PageSize: 50})
	if err != nil {
		t.Fatalf("ListExpenses page 12: %v", err)
	}
	if len(empty.GetExpenses()) != 0 {
		t.Fatalf("page 12 returned %d rows, want 0", len(empty.GetExpenses()))
	}
	if empty.GetHasMore() {
		t.Fatal("page 12 has_more = true, want false past the end")
	}
}

func TestListExpenses_DefaultsAndCapsPageSize(t *testing.T) {
	db := testutil.OpenDB(t)
	seedPaginationExpenses(t, db)
	api, ctx := paginationExpensesAPI(t, db)

	defaulted, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{})
	if err != nil {
		t.Fatalf("ListExpenses with no pagination: %v", err)
	}
	if defaulted.GetPage() != 1 || defaulted.GetPageSize() != 50 {
		t.Fatalf("unpaged request echoed page=%d page_size=%d, want 1/50", defaulted.GetPage(), defaulted.GetPageSize())
	}
	if len(defaulted.GetExpenses()) != 50 {
		t.Fatalf("unpaged request returned %d rows, want 50", len(defaulted.GetExpenses()))
	}

	// page 0 is clamped to the first page instead of silently replaying from 0.
	zeroPage, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{Page: 0, PageSize: 50})
	if err != nil {
		t.Fatalf("ListExpenses page 0: %v", err)
	}
	if zeroPage.GetPage() != 1 {
		t.Fatalf("page 0 echoed page=%d, want 1", zeroPage.GetPage())
	}

	capped, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{Page: 1, PageSize: 1000})
	if err != nil {
		t.Fatalf("ListExpenses oversized page: %v", err)
	}
	if capped.GetPageSize() != 500 {
		t.Fatalf("page_size 1000 echoed %d, want the 500 cap", capped.GetPageSize())
	}
	if len(capped.GetExpenses()) != 500 {
		t.Fatalf("capped page returned %d rows, want 500", len(capped.GetExpenses()))
	}
}

// Regression: rows past the first 500 were unreachable because the endpoint
// always answered with has_more=false. A second maximum-size page must return
// the remaining rows.
func TestListExpenses_ReturnsRowsPastFiveHundred(t *testing.T) {
	db := testutil.OpenDB(t)
	seedPaginationExpenses(t, db)
	api, ctx := paginationExpensesAPI(t, db)

	overflow, err := api.ListExpenses(ctx, &financeexpensesv1.ListExpensesRequest{Page: 2, PageSize: 500})
	if err != nil {
		t.Fatalf("ListExpenses page 2 page_size 500: %v", err)
	}
	if len(overflow.GetExpenses()) != expensesPaginationRows-500 {
		t.Fatalf("second maximum page returned %d rows, want %d", len(overflow.GetExpenses()), expensesPaginationRows-500)
	}
	if overflow.GetPage() != 2 || overflow.GetPageSize() != 500 {
		t.Fatalf("second maximum page echoed page=%d page_size=%d, want 2/500", overflow.GetPage(), overflow.GetPageSize())
	}
	if overflow.GetHasMore() {
		t.Fatal("second maximum page has_more = true, want false")
	}
}
