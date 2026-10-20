package api

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/testutil"
)

const reportDetailsBusinessID = int64(987654521)

// An expense that is still waiting for approval is left out of the P&L and
// the dashboard totals, which used to make it look as if it never reached the
// books. Both reports now say how much is waiting, split approved spend by
// category, and the dashboard carries its chart, activity and overdue lists.
func TestReports_CarryPendingExpensesAndBreakdowns(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	cleanup := func() {
		for _, table := range []string{"finance_expenses", "finance_expense_categories", "finance_invoices"} {
			_, _ = db.ExecContext(ctx, "DELETE FROM "+table+" WHERE business_id = ?", reportDetailsBusinessID)
		}
	}
	cleanup()
	t.Cleanup(cleanup)

	now := time.Now().UTC()
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	day := thisMonth.Format("2006-01-02")
	exec := func(query string, args ...any) int64 {
		t.Helper()
		res, err := db.ExecContext(ctx, query, args...)
		if err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	phone := exec(`INSERT INTO finance_expense_categories (owner_id, business_id, name, color) VALUES (1, ?, 'Communications', '#14b8a6')`, reportDetailsBusinessID)
	rent := exec(`INSERT INTO finance_expense_categories (owner_id, business_id, name, color) VALUES (1, ?, 'Rent', '#8b5cf6')`, reportDetailsBusinessID)
	addExpense := func(category any, amount float64, status string) {
		exec(`INSERT INTO finance_expenses (owner_id, business_id, category_id, description, amount, expense_date, status) VALUES (1, ?, ?, 'x', ?, ?, ?)`,
			reportDetailsBusinessID, category, amount, day, status)
	}
	addExpense(rent, 900, "approved")
	addExpense(phone, 40, "approved")
	addExpense(nil, 10, "approved")
	addExpense(phone, 35, "pending")
	addExpense(rent, 500, "rejected")
	exec(`INSERT INTO finance_invoices (owner_id, business_id, invoice_number, client_name, issue_date, due_date, paid_date, status, total, amount_due)
		VALUES (1, ?, 'INV-1', 'Acme', ?, ?, ?, 'paid', 2000, 0)`, reportDetailsBusinessID, day, day, day)
	overdueDue := thisMonth.AddDate(0, -1, 0).Format("2006-01-02")
	exec(`INSERT INTO finance_invoices (owner_id, business_id, invoice_number, client_name, issue_date, due_date, status, total, amount_due)
		VALUES (1, ?, 'INV-2', 'Initech', ?, ?, 'sent', 300, 300)`, reportDetailsBusinessID, overdueDue, overdueDue)

	service := newInvoicesServiceForTest(db)

	pl, err := service.GetPLReport(ctx, reportDetailsBusinessID, "month", day)
	if err != nil {
		t.Fatalf("P&L: %v", err)
	}
	if pl.OperatingExpenses != 950 {
		t.Fatalf("operating expenses = %v, want the 950 approved", pl.OperatingExpenses)
	}
	if pl.PendingExpenses != 35 || pl.PendingExpenseCount != 1 {
		t.Fatalf("pending = %v over %d, want 35 over 1", pl.PendingExpenses, pl.PendingExpenseCount)
	}
	if len(pl.ExpensesByCategory) != 3 {
		t.Fatalf("categories = %+v, want rent, communications and uncategorized", pl.ExpensesByCategory)
	}
	first := pl.ExpensesByCategory[0]
	if first.Name != "Rent" || first.Amount != 900 || first.Color != "#8b5cf6" || first.CategoryID == nil || *first.CategoryID != rent {
		t.Fatalf("largest category = %+v, want Rent 900", first)
	}
	if last := pl.ExpensesByCategory[2]; last.CategoryID != nil || last.Amount != 10 {
		t.Fatalf("uncategorized = %+v, want nil category for 10", last)
	}

	dash, err := service.GetDashboard(ctx, reportDetailsBusinessID)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if dash.PendingExpenses != 35 || dash.PendingExpenseCount != 1 {
		t.Fatalf("dashboard pending = %v over %d", dash.PendingExpenses, dash.PendingExpenseCount)
	}
	if len(dash.Monthly) != 6 {
		t.Fatalf("monthly = %d months, want 6", len(dash.Monthly))
	}
	current := dash.Monthly[5]
	if current.Month != thisMonth.Format("2006-01") || current.Revenue != 2000 || current.Expenses != 950 {
		t.Fatalf("current month = %+v, want 2000 revenue and 950 expenses", current)
	}
	if dash.Monthly[0].Revenue != 0 || dash.Monthly[0].Expenses != 0 {
		t.Fatalf("oldest month = %+v, want zeros", dash.Monthly[0])
	}
	kinds := map[string]int{}
	for _, a := range dash.RecentActivity {
		kinds[a.Kind]++
		if a.Status == "rejected" {
			t.Fatalf("rejected expense in activity: %+v", a)
		}
	}
	if kinds["expense"] != 4 || kinds["payment"] != 1 {
		t.Fatalf("activity kinds = %v, want 4 expenses and 1 payment", kinds)
	}
	if dash.OverdueInvoiceCount != 1 || len(dash.OverdueInvoices) != 1 || dash.OverdueInvoices[0].InvoiceNumber != "INV-2" || dash.OverdueInvoices[0].DaysOverdue < 28 {
		t.Fatalf("overdue = %d %+v", dash.OverdueInvoiceCount, dash.OverdueInvoices)
	}
}
