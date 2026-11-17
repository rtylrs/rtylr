package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

// TestInsightsDealTotals writes to two businesses no fixture or other test
// uses, so the totals it checks are exact.
func TestInsightsDealTotals(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := configuredRepository(db)
	const biz, other int64 = 75640001, 75640002
	removeDeals := func() {
		if _, err := db.ExecContext(ctx, `DELETE FROM crm_deals WHERE business_id IN (?, ?)`, biz, other); err != nil {
			t.Fatalf("clear deals: %v", err)
		}
	}
	removeDeals()
	t.Cleanup(removeDeals)

	insert := func(businessID int64, value, billing, status string, closedAt time.Time) {
		t.Helper()
		closed := sql.NullTime{Time: closedAt, Valid: !closedAt.IsZero()}
		if _, err := db.ExecContext(ctx, `INSERT INTO crm_deals (business_id, owner_id, pipeline_id, stage_id, title, value, billing, status, closed_at)
			VALUES (?, ?, 1, 1, ?, ?, ?, ?, ?)`, businessID, businessID, unique(t, "Deal"), value, billing, status, closed); err != nil {
			t.Fatalf("insert deal: %v", err)
		}
	}
	at := func(month time.Month, day, hour, minute, second int) time.Time {
		return time.Date(2026, month, day, hour, minute, second, 0, time.UTC)
	}
	from, to := at(time.October, 1, 0, 0, 0), at(time.October, 15, 12, 0, 0)

	// Open: 1,000 once + 250.50 a month (3,006 first year) + 99.99 a year.
	insert(biz, "1000.00", "one_time", "open", time.Time{})
	insert(biz, "250.50", "monthly", "open", time.Time{})
	insert(biz, "99.99", "yearly", "open", time.Time{})
	// Won in the window, one of them on its closing edge.
	insert(biz, "500.00", "one_time", "won", at(time.October, 5, 10, 0, 0))
	insert(biz, "100.00", "monthly", "won", to)
	// Won outside the window, and a loss inside it.
	insert(biz, "7000.00", "one_time", "won", to.Add(time.Second))
	insert(biz, "8000.00", "one_time", "won", from.Add(-time.Second))
	insert(biz, "9000.00", "one_time", "lost", at(time.October, 10, 9, 0, 0))
	// Another business's deals stay out of biz's totals.
	insert(other, "5000.00", "one_time", "open", time.Time{})
	insert(other, "6000.00", "one_time", "won", at(time.October, 5, 10, 0, 0))

	open := func(businessID int64) (int64, money.Amount) {
		t.Helper()
		deals, value, err := repo.OpenDealPipeline(ctx, businessID)
		if err != nil {
			t.Fatalf("OpenDealPipeline: %v", err)
		}
		return deals, value
	}
	won := func(businessID int64) (int64, money.Amount) {
		t.Helper()
		deals, value, err := repo.WonDealsBetween(ctx, businessID, from, to)
		if err != nil {
			t.Fatalf("WonDealsBetween: %v", err)
		}
		return deals, value
	}

	if deals, value := open(biz); deals != 3 || value != 410599 {
		t.Errorf("open pipeline = %d deals, %d minor; want 3 deals, 410599", deals, value)
	}
	if deals, value := won(biz); deals != 2 || value != 170000 {
		t.Errorf("won in window = %d deals, %d minor; want 2 deals, 170000", deals, value)
	}
	if deals, value := open(other); deals != 1 || value != 500000 {
		t.Errorf("other business open pipeline = %d deals, %d minor; want 1 deal, 500000", deals, value)
	}
	if deals, value := won(other); deals != 1 || value != 600000 {
		t.Errorf("other business won = %d deals, %d minor; want 1 deal, 600000", deals, value)
	}
	if deals, value := open(75640003); deals != 0 || value != 0 {
		t.Errorf("business without deals = %d deals, %d minor; want zeros", deals, value)
	}
}

// TestInsightsRecurringRevenue uses its own businesses, so the MRR it checks
// is exact.
func TestInsightsRecurringRevenue(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := configuredRepository(db)
	const biz, other, empty int64 = 75650001, 75650002, 75650003
	removeContracts := func() {
		if _, err := db.ExecContext(ctx, `DELETE FROM crm_contracts WHERE business_id IN (?, ?)`, biz, other); err != nil {
			t.Fatalf("clear contracts: %v", err)
		}
	}
	removeContracts()
	t.Cleanup(removeContracts)

	date := func(y int, m time.Month, d int) sql.NullTime {
		return sql.NullTime{Time: time.Date(y, m, d, 0, 0, 0, 0, time.UTC), Valid: true}
	}
	none := sql.NullTime{}
	insert := func(businessID int64, mrr, value, billing string, signed, start, end sql.NullTime) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `INSERT INTO crm_contracts (business_id, owner_id, mrr, value, billing, signed_at, start_date, end_date)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, businessID, businessID, mrr, value, billing, signed, start, end); err != nil {
			t.Fatalf("insert contract: %v", err)
		}
	}
	monthStart := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)

	// Active since September: an explicit MRR of 100.00, and a yearly 1,200.00
	// with no MRR entered (100.00 a month).
	insert(biz, "100.00", "0", "monthly", date(2026, 9, 10), date(2026, 9, 10), none)
	insert(biz, "0", "1200.00", "yearly", date(2026, 9, 12), none, date(2027, 9, 11))
	// Signed this month: 50.50 a month, counted now but not at month start.
	insert(biz, "50.50", "0", "monthly", date(2026, 10, 3), date(2026, 10, 3), none)
	// Not counted: unsigned, not started yet, ended before today, one-time.
	insert(biz, "700.00", "0", "monthly", none, none, none)
	insert(biz, "300.00", "0", "monthly", date(2026, 10, 1), date(2026, 11, 1), none)
	insert(biz, "400.00", "0", "monthly", date(2026, 8, 1), date(2026, 8, 1), date(2026, 10, 10))
	insert(biz, "0", "9000.00", "one_time", date(2026, 9, 1), none, none)
	// Another business's contract stays out of biz's MRR.
	insert(other, "999.00", "0", "monthly", date(2026, 9, 1), none, none)

	read := func(businessID int64) (int64, money.Amount, money.Amount) {
		t.Helper()
		contracts, mrr, atMonthStart, err := repo.RecurringRevenue(ctx, businessID, monthStart, now)
		if err != nil {
			t.Fatalf("RecurringRevenue: %v", err)
		}
		return contracts, mrr, atMonthStart
	}

	// Now: 100.00 + 100.00 + 50.50 = 250.50. Month start: 100.00 + 100.00, plus
	// the contract that ended on 10 October, which was still running then.
	if contracts, mrr, atMonthStart := read(biz); contracts != 3 || mrr != 25050 || atMonthStart != 60000 {
		t.Errorf("recurring revenue = %d contracts, MRR %d, %d at month start; want 3, 25050, 60000", contracts, mrr, atMonthStart)
	}
	if contracts, mrr, atMonthStart := read(other); contracts != 1 || mrr != 99900 || atMonthStart != 99900 {
		t.Errorf("other business = %d contracts, MRR %d, %d at month start; want 1, 99900, 99900", contracts, mrr, atMonthStart)
	}
	if contracts, mrr, atMonthStart := read(empty); contracts != 0 || mrr != 0 || atMonthStart != 0 {
		t.Errorf("business without contracts = %d contracts, MRR %d, %d at month start; want zeros", contracts, mrr, atMonthStart)
	}
}
