package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	financecontract "github.com/voxire/onion/internal/contracts/finance"
	"github.com/voxire/onion/internal/modules/finance/finance_reports/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// TestInsightsProjectInvoices uses businesses no fixture or other test touches,
// so every total it checks is exact.
func TestInsightsProjectInvoices(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const (
		biz, other         int64 = 9982701, 9982702
		project, otherProj int64 = 9982711, 9982712
	)
	remove := func() {
		if _, err := db.ExecContext(ctx, `DELETE FROM finance_invoices WHERE business_id IN (?, ?)`, biz, other); err != nil {
			t.Fatalf("clear: %v", err)
		}
	}
	remove()
	t.Cleanup(remove)

	today := time.Now()
	day := func(offset int) string { return today.AddDate(0, 0, offset).Format(time.DateOnly) }
	n := 0
	invoice := func(businessID, projectID int64, status, due string, total, paid float64) {
		t.Helper()
		n++
		if _, err := db.ExecContext(ctx, `INSERT INTO finance_invoices
			(owner_id, business_id, invoice_number, client_name, project_id, status, issue_date, due_date, currency, total, amount_paid, amount_due)
			VALUES (?, ?, ?, 'Client', NULLIF(?, 0), ?, ?, ?, 'USD', ?, ?, ?)`,
			businessID, businessID, fmt.Sprintf("PRJ-%02d", n), projectID, status, day(-60), due, total, paid, total-paid); err != nil {
			t.Fatalf("insert invoice %d: %v", n, err)
		}
	}

	invoice(biz, project, "paid", day(-30), 1000, 1000)
	invoice(biz, project, "partial", day(-5), 500, 200)   // past due: overdue 300
	invoice(biz, project, "sent", day(10), 400, 0)        // not yet due
	invoice(biz, project, "viewed", day(0), 50.25, 0)     // due today: not overdue
	invoice(biz, project, "overdue", day(3), 125.5, 25.5) // marked overdue: overdue 100
	invoice(biz, project, "draft", day(-90), 9000, 0)     // not billed
	invoice(biz, project, "cancelled", day(-90), 8000, 0) // never billed
	invoice(biz, project, "void", day(-90), 7000, 0)      // never billed
	invoice(biz, otherProj, "sent", day(-90), 6000, 0)    // another project
	invoice(biz, 0, "sent", day(-90), 5000, 0)            // untagged
	invoice(other, project, "sent", day(-90), 4000, 0)    // another business

	got, err := repo.InsightsProjectInvoices(ctx, biz, project)
	if err != nil {
		t.Fatalf("InsightsProjectInvoices: %v", err)
	}
	want := financecontract.ProjectInvoiceTotals{
		InvoiceCount: 5,
		Invoiced:     207575, // 1000 + 500 + 400 + 50.25 + 125.50
		Paid:         122550, // 1000 + 200 + 25.50
		Outstanding:  85025,  // 300 + 400 + 50.25 + 100
		Overdue:      40000,  // 300 + 100
		OverdueCount: 2,
	}
	if got != want {
		t.Errorf("totals = %+v, want %+v", got, want)
	}

	empty, err := repo.InsightsProjectInvoices(ctx, other, otherProj)
	if err != nil {
		t.Fatalf("InsightsProjectInvoices(empty): %v", err)
	}
	if empty != (financecontract.ProjectInvoiceTotals{}) {
		t.Errorf("project without invoices = %+v, want zeros", empty)
	}
}
