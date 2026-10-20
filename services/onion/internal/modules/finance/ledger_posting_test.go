package finance_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/accounts"
	financereports "github.com/voxire/onion/internal/modules/finance/finance_reports"
	"github.com/voxire/onion/internal/modules/finance/journal"
	"github.com/voxire/onion/internal/modules/organization/business"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

const (
	ledgerBusinessID = int64(987654631)
	ledgerUserID     = int64(987654632)
)

// An approved expense and a recorded invoice payment used to stay out of the
// journal, so the balance sheet, journal and accounts were empty for a business
// that only used expenses and invoices. Reading the ledger now posts them,
// once each, against the standard accounts.
func TestLedger_PostsApprovedExpensesAndPaymentsOnRead(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	exec := func(query string, args ...any) int64 {
		t.Helper()
		res, err := db.ExecContext(ctx, query, args...)
		if err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE l FROM finance_journal_lines l JOIN finance_journal_entries e ON e.id = l.entry_id WHERE e.business_id = ?`, ledgerBusinessID)
		for _, table := range []string{"finance_journal_entries", "finance_accounts", "finance_expenses", "finance_payments", "finance_invoices"} {
			_, _ = db.ExecContext(ctx, "DELETE FROM "+table+" WHERE business_id = ?", ledgerBusinessID)
		}
		_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id = ?", ledgerBusinessID)
	}
	cleanup()
	t.Cleanup(cleanup)

	exec(`INSERT INTO businesses (id, name, owner_id, currency) VALUES (?, 'Ledger posting', ?, 'USD')`, ledgerBusinessID, ledgerUserID)
	exec(`INSERT INTO finance_expenses (owner_id, business_id, description, amount, currency, expense_date, payment_method, status)
		VALUES (?, ?, 'Phone number', 35, 'USD', '2026-10-19', 'card', 'approved')`, ledgerUserID, ledgerBusinessID)
	exec(`INSERT INTO finance_expenses (owner_id, business_id, description, amount, currency, expense_date, payment_method, status)
		VALUES (?, ?, 'Taxi', 12, 'USD', '2026-10-19', 'cash', 'pending')`, ledgerUserID, ledgerBusinessID)
	invoiceID := exec(`INSERT INTO finance_invoices (owner_id, business_id, invoice_number, client_name, issue_date, due_date, paid_date, status, total, amount_due)
		VALUES (?, ?, 'INV-1', 'Acme', '2026-10-01', '2026-10-15', '2026-10-10', 'paid', 500, 0)`, ledgerUserID, ledgerBusinessID)
	exec(`INSERT INTO finance_payments (owner_id, business_id, invoice_id, amount, currency, payment_date, method, reference)
		VALUES (?, ?, ?, 500, 'USD', '2026-10-10', 'bank_transfer', 'r-1')`, ledgerUserID, ledgerBusinessID, invoiceID)

	accountsModule := accounts.NewModule(ctx, db)
	journalModule := journal.NewModule(ctx, db)
	journalModule.SetLedgerAccountProvider(accountsModule.LedgerAccountProvider())
	accountsModule.SetLedgerSyncer(journalModule.LedgerSyncer())
	reports := financereports.NewModule(ctx, db).WithReportingCurrency(business.NewModule(ctx, db).ReportingCurrency)
	reports.SetLedgerSyncer(journalModule.LedgerSyncer())

	mux := http.NewServeMux()
	reports.RegisterAPI(mux)
	journalModule.RegisterAPI(mux)
	accountsModule.RegisterAPI(mux)
	device := &authPkg.Device{User: &authPkg.AuthUser{
		BusinessID: ledgerBusinessID,
		UserID:     ledgerUserID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: ledgerBusinessID,
			Apps: []authorization.EffectiveAppAccess{{
				AppID: authorization.AppFinance, Licensed: true, HasAccess: true,
				Permissions: []string{"finance:reports:view", "finance:journal:view", "finance:accounts:view"},
			}},
		},
	}}
	testutil.GrantBusinessAccess(t, db, ledgerBusinessID, ledgerUserID)
	authed := authPkg.WithAuthUser(device.Ctx(ctx), device.User)
	get := func(path string, out any) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil).WithContext(authed))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d %s", path, rec.Code, rec.Body.String())
		}
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("decode %s: %v", rec.Body.String(), err)
		}
	}

	var sheet struct {
		TotalAssets float64 `json:"total_assets"`
		TotalEquity float64 `json:"total_equity"`
		NetIncome   float64 `json:"net_income"`
		Lines       []struct {
			Code   string  `json:"code"`
			Name   string  `json:"name"`
			Amount float64 `json:"amount"`
		} `json:"lines"`
	}
	get("/finance/statement/v1/balance-sheet?as_of=2026-10-31", &sheet)
	// Bank received 500 and paid the 35 card expense; the pending taxi is not posted.
	if len(sheet.Lines) != 1 || sheet.Lines[0].Code != "1010" || sheet.Lines[0].Amount != 465 {
		t.Fatalf("lines = %+v, want Bank 465 only", sheet.Lines)
	}
	if sheet.NetIncome != 465 || sheet.TotalAssets != 465 || sheet.TotalEquity != 465 {
		t.Fatalf("assets %v equity %v net income %v, want 465 each", sheet.TotalAssets, sheet.TotalEquity, sheet.NetIncome)
	}

	var before struct {
		Lines []struct {
			Amount float64 `json:"amount"`
		} `json:"lines"`
	}
	get("/finance/statement/v1/balance-sheet?as_of=2026-10-09", &before)
	if len(before.Lines) != 0 {
		t.Fatalf("lines before anything was paid = %+v, want none", before.Lines)
	}

	var entries struct {
		Entries []struct {
			Description string `json:"description"`
		} `json:"entries"`
	}
	get("/finance/journal/v1/entries", &entries)
	get("/finance/journal/v1/entries", &entries)
	if len(entries.Entries) != 2 {
		t.Fatalf("journal = %+v, want the expense and the payment once each", entries.Entries)
	}

	var listed struct {
		Accounts []struct {
			Code           string  `json:"code"`
			CurrentBalance float64 `json:"current_balance"`
		} `json:"accounts"`
	}
	get("/finance/account/v1/accounts", &listed)
	balances := map[string]float64{}
	for _, a := range listed.Accounts {
		balances[a.Code] = a.CurrentBalance
	}
	if len(balances) != 4 || balances["1010"] != 465 || balances["4000"] != 500 || balances["6000"] != 35 || balances["1000"] != 0 {
		t.Fatalf("account balances = %v", balances)
	}
}
