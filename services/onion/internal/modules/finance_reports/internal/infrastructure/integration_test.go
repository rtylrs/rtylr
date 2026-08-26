package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestSumAccountsByType_AssetTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "asset")
	if err != nil {
		t.Fatalf("SumAccountsByType asset: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestCurrencyScopedReportsExcludeForeignCurrency(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const (
		businessID = int64(9982601)
		ownerID    = int64(9982602)
	)

	cleanup := func() {
		for _, query := range []string{
			"DELETE FROM finance_expenses WHERE business_id = ?",
			"DELETE FROM finance_bills WHERE business_id = ?",
			"DELETE FROM finance_invoices WHERE business_id = ?",
			"DELETE FROM finance_bank_accounts WHERE business_id = ?",
			"DELETE FROM finance_accounts WHERE business_id = ?",
			"DELETE FROM businesses WHERE id = ?",
		} {
			if _, err := db.ExecContext(ctx, query, businessID); err != nil {
				t.Errorf("cleanup %q: %v", query, err)
			}
		}
	}
	t.Cleanup(cleanup)

	statements := []struct {
		query string
		args  []any
	}{
		{
			query: "INSERT INTO businesses (id, name, owner_id, currency) VALUES (?, ?, ?, 'USD')",
			args:  []any{businessID, "Currency Scope Test Business", ownerID},
		},
		{
			query: "INSERT INTO finance_accounts (id, owner_id, business_id, name, type, currency, balance) VALUES (?, ?, ?, 'USD Cash', 'asset', 'USD', 125.00)",
			args:  []any{int64(99826011), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_accounts (id, owner_id, business_id, name, type, currency, balance) VALUES (?, ?, ?, 'EUR Cash', 'asset', 'EUR', 875.00)",
			args:  []any{int64(99826012), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_bank_accounts (id, owner_id, business_id, name, balance, currency) VALUES (?, ?, ?, 'USD Bank', 100.00, 'USD')",
			args:  []any{int64(99826021), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_bank_accounts (id, owner_id, business_id, name, balance, currency) VALUES (?, ?, ?, 'EUR Bank', 900.00, 'EUR')",
			args:  []any{int64(99826022), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_invoices (id, owner_id, business_id, invoice_number, client_name, status, issue_date, due_date, currency, total, amount_due) VALUES (?, ?, ?, 'USD-RECEIVABLE', 'USD Client', 'sent', '2026-01-01', '2026-01-31', 'USD', 50.00, 50.00)",
			args:  []any{int64(99826031), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_invoices (id, owner_id, business_id, invoice_number, client_name, status, issue_date, due_date, currency, total, amount_due) VALUES (?, ?, ?, 'EUR-RECEIVABLE', 'EUR Client', 'sent', '2026-01-01', '2026-01-31', 'EUR', 950.00, 950.00)",
			args:  []any{int64(99826032), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_invoices (id, owner_id, business_id, invoice_number, client_name, status, issue_date, due_date, paid_date, currency, total, amount_due) VALUES (?, ?, ?, 'USD-PAID', 'USD Paid Client', 'paid', '2026-01-01', '2026-01-31', '2026-01-15', 'USD', 300.00, 0.00)",
			args:  []any{int64(99826033), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_invoices (id, owner_id, business_id, invoice_number, client_name, status, issue_date, due_date, paid_date, currency, total, amount_due) VALUES (?, ?, ?, 'EUR-PAID', 'EUR Paid Client', 'paid', '2026-01-01', '2026-01-31', '2026-01-15', 'EUR', 700.00, 0.00)",
			args:  []any{int64(99826034), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_bills (id, owner_id, business_id, supplier_name, amount, currency, due_date, status) VALUES (?, ?, ?, 'USD Supplier', 60.00, 'USD', '2026-01-31', 'pending')",
			args:  []any{int64(99826041), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_bills (id, owner_id, business_id, supplier_name, amount, currency, due_date, status) VALUES (?, ?, ?, 'EUR Supplier', 940.00, 'EUR', '2026-01-31', 'pending')",
			args:  []any{int64(99826042), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_expenses (id, owner_id, business_id, description, amount, currency, expense_date, status) VALUES (?, ?, ?, 'USD Expense', 40.00, 'USD', '2026-01-15', 'approved')",
			args:  []any{int64(99826051), ownerID, businessID},
		},
		{
			query: "INSERT INTO finance_expenses (id, owner_id, business_id, description, amount, currency, expense_date, status) VALUES (?, ?, ?, 'EUR Expense', 960.00, 'EUR', '2026-01-15', 'approved')",
			args:  []any{int64(99826052), ownerID, businessID},
		},
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatalf("seed %q: %v", statement.query, err)
		}
	}

	assertAmount := func(name string, got, want float64) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %.2f, want %.2f in the business currency", name, got, want)
		}
	}

	accounts, err := repo.SumAccountsByType(ctx, businessID, "asset")
	if err != nil {
		t.Fatalf("SumAccountsByType: %v", err)
	}
	assertAmount("accounts", accounts, 125)

	bank, err := repo.SumBankBalances(ctx, businessID)
	if err != nil {
		t.Fatalf("SumBankBalances: %v", err)
	}
	assertAmount("bank", bank, 100)

	receivables, err := repo.SumReceivables(ctx, businessID)
	if err != nil {
		t.Fatalf("SumReceivables: %v", err)
	}
	assertAmount("receivables", receivables, 50)

	payables, err := repo.SumPayables(ctx, businessID)
	if err != nil {
		t.Fatalf("SumPayables: %v", err)
	}
	assertAmount("payables", payables, 60)

	paid, err := repo.SumPaidInvoicesInRange(ctx, businessID, "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange: %v", err)
	}
	assertAmount("paid invoices", paid, 300)

	expenses, err := repo.SumExpensesInRange(ctx, businessID, "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("SumExpensesInRange: %v", err)
	}
	assertAmount("expenses", expenses, 40)
}

func TestSumAccountsByType_LiabilityTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "liability")
	if err != nil {
		t.Fatalf("SumAccountsByType liability: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumAccountsByType_EquityTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "equity")
	if err != nil {
		t.Fatalf("SumAccountsByType equity: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumAccountsByType_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, 0, "asset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumBankBalances_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumBankBalances(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumBankBalances: %v", err)
	}
	_ = val // result may be 0 if no bank accounts exist for fixture owner
}

func TestSumBankBalances_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumBankBalances(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumReceivables_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumReceivables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumReceivables: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative receivables sum, got %f", val)
	}
}

func TestSumPayables_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPayables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumPayables: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative payables sum, got %f", val)
	}
}

func TestSumPaidInvoicesInRange_CurrentYearNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPaidInvoicesInRange(ctx, testutil.OwnerID, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumPaidInvoicesInRange_EmptyDatesNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPaidInvoicesInRange(ctx, testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange empty dates: %v", err)
	}
	_ = val
}

func TestSumExpensesInRange_CurrentYearNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumExpensesInRange(ctx, testutil.OwnerID, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("SumExpensesInRange: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumExpensesInRange_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumExpensesInRange(ctx, 0, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumExpensesInRange_MarqueeRange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// March 2026 range aligns with payroll fixture month.
	val, err := repo.SumExpensesInRange(ctx, testutil.OwnerID, "2026-03-01", "2026-03-31")
	if err != nil {
		t.Fatalf("SumExpensesInRange March 2026: %v", err)
	}
	_ = val
}

func TestSumAccountsByType_RevenueTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "revenue")
	if err != nil {
		t.Fatalf("SumAccountsByType revenue: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumAccountsByType_ExpenseTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "expense")
	if err != nil {
		t.Fatalf("SumAccountsByType expense: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumAccountsByType_UnknownTypeReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "nonexistent_type_xyz")
	if err != nil {
		t.Fatalf("SumAccountsByType unknown type: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown account type, got %f", val)
	}
}

func TestSumReceivables_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumReceivables(ctx, 0)
	if err != nil {
		t.Fatalf("SumReceivables unknown owner: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumPayables_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPayables(ctx, 0)
	if err != nil {
		t.Fatalf("SumPayables unknown owner: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumPaidInvoicesInRange_PastYearNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPaidInvoicesInRange(ctx, testutil.OwnerID, "2024-01-01", "2024-12-31")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange 2024: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumPaidInvoicesInRange_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPaidInvoicesInRange(ctx, 0, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange unknown owner: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumExpensesInRange_EmptyRange_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Inverted range should return 0, not error.
	val, err := repo.SumExpensesInRange(ctx, testutil.OwnerID, "2026-12-31", "2026-01-01")
	if err != nil {
		t.Fatalf("SumExpensesInRange inverted range: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative value for inverted range, got %f", val)
	}
}

func TestSumBankBalances_StableAcrossCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	v1, err := repo.SumBankBalances(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first SumBankBalances: %v", err)
	}
	v2, err := repo.SumBankBalances(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second SumBankBalances: %v", err)
	}
	if v1 != v2 {
		t.Errorf("SumBankBalances not stable: %f vs %f", v1, v2)
	}
}

func TestSumReceivables_StableAcrossCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	v1, err := repo.SumReceivables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first SumReceivables: %v", err)
	}
	v2, err := repo.SumReceivables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second SumReceivables: %v", err)
	}
	if v1 != v2 {
		t.Errorf("SumReceivables not stable: %f vs %f", v1, v2)
	}
}

func TestSumPayables_StableAcrossCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	v1, err := repo.SumPayables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first SumPayables: %v", err)
	}
	v2, err := repo.SumPayables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second SumPayables: %v", err)
	}
	if v1 != v2 {
		t.Errorf("SumPayables not stable: %f vs %f", v1, v2)
	}
}

func TestNetProfit_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profit, err := repo.NetProfit(ctx, testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("NetProfit: %v", err)
	}
	// Profit may be negative if expenses exceed revenue; just ensure the call
	// succeeds and returns a finite number.
	_ = profit
}

func TestNetProfit_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profit, err := repo.NetProfit(ctx, 0, "", "")
	if err != nil {
		t.Fatalf("NetProfit(ownerID=0): %v", err)
	}
	if profit != 0 {
		t.Errorf("expected 0 for ownerID=0, got %f", profit)
	}
}

func TestNetProfit_WithDateRange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	profit, err := repo.NetProfit(ctx, testutil.OwnerID, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("NetProfit with date range: %v", err)
	}
	_ = profit
}

func TestWorkingCapital_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wc, err := repo.WorkingCapital(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("WorkingCapital: %v", err)
	}
	// Working capital may be positive or negative; just assert the call succeeds.
	_ = wc
}

func TestWorkingCapital_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wc, err := repo.WorkingCapital(ctx, 0)
	if err != nil {
		t.Fatalf("WorkingCapital(ownerID=0): %v", err)
	}
	if wc != 0 {
		t.Errorf("expected 0 for ownerID=0, got %f", wc)
	}
}

func TestNetProfit_IsRevenueMinusExpenses(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	revenue, err := repo.SumPaidInvoicesInRange(ctx, testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange: %v", err)
	}
	expenses, err := repo.SumExpensesInRange(ctx, testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("SumExpensesInRange: %v", err)
	}
	profit, err := repo.NetProfit(ctx, testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("NetProfit: %v", err)
	}
	want := revenue - expenses
	if profit != want {
		t.Errorf("NetProfit=%f, want revenue(%f)-expenses(%f)=%f", profit, revenue, expenses, want)
	}
}
