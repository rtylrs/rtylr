package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const idempotencyBusinessID = 901

// These tests need a real MySQL: they assert the unique key added by
// schemas/hr/zz_hr_payroll_records_uniqueness.sql actually holds, and that the
// repository's replace-in-transaction is safe against it. They are skipped
// when DB_HOST is unset (see testutil.OpenDB).

type periodSpec struct {
	label  string
	start  time.Time
	end    time.Time
	staff  []int64
	netPay map[int64]float64
}

func aprilSpec(staff ...int64) periodSpec {
	return periodSpec{
		label: "2026-04",
		start: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		staff: staff,
		netPay: map[int64]float64{
			4101: 1600.00,
			4102: 2400.00,
		},
	}
}

func (p periodSpec) records() []domain.PayrollRecordInput {
	out := make([]domain.PayrollRecordInput, 0, len(p.staff))
	for _, staffID := range p.staff {
		net := p.netPay[staffID]
		out = append(out, domain.PayrollRecordInput{
			StaffID:     staffID,
			PeriodStart: p.start,
			PeriodEnd:   p.end,
			PeriodLabel: p.label,
			BasePay:     net,
			NetPay:      net,
		})
	}
	return out
}

// netPaySum is computed from the rows the application would read back through
// ListPayrollRecords, not from what the run was asked to write, so a duplicate
// row shows up in the total the same way a payroll export would see it.
func netPaySum(t *testing.T, repo *infrastructure.Repository, label string) float64 {
	t.Helper()
	records, err := repo.ListPayrollRecords(context.Background(), idempotencyBusinessID, label)
	if err != nil {
		t.Fatalf("ListPayrollRecords(%q): %v", label, err)
	}
	var total float64
	for _, r := range records {
		total += r.NetPay
	}
	return total
}

func countRows(t *testing.T, db *sql.DB, label string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM hr_payroll_records WHERE business_id = ? AND period_label = ?",
		idempotencyBusinessID, label).Scan(&count); err != nil {
		t.Fatalf("count rows for %q: %v", label, err)
	}
	return count
}

func TestRunPayrollReRunSamePeriodDoesNotDuplicate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	period := aprilSpec(4101, 4102)
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE business_id = ? AND period_label = ?", idempotencyBusinessID, period.label)
	})

	firstCount, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, period.records())
	if err != nil {
		t.Fatalf("first CreatePayrollRecords: %v", err)
	}
	if firstCount != 2 {
		t.Fatalf("first run wrote %d records, want 2", firstCount)
	}
	firstRows := countRows(t, db, period.label)
	firstNet := netPaySum(t, repo, period.label)
	if firstRows != 2 || firstNet != 4000.00 {
		t.Fatalf("first run: rows=%d net=%v, want rows=2 net=4000", firstRows, firstNet)
	}

	secondCount, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, period.records())
	if err != nil {
		t.Fatalf("re-run CreatePayrollRecords: %v", err)
	}
	if secondCount != 2 {
		t.Errorf("re-run reported %d records, want 2", secondCount)
	}
	if rows := countRows(t, db, period.label); rows != firstRows {
		t.Errorf("row count changed on re-run: %d -> %d", firstRows, rows)
	}
	if net := netPaySum(t, repo, period.label); net != firstNet {
		t.Errorf("net_pay total changed on re-run: %v -> %v", firstNet, net)
	}
}

func TestRunPayrollDifferentPeriodIsUnaffected(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	april := aprilSpec(4101, 4102)
	may := periodSpec{
		label:  "2026-05",
		start:  time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		end:    time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
		staff:  []int64{4101},
		netPay: map[int64]float64{4101: 1600.00},
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE business_id = ? AND period_label IN (?, ?)", idempotencyBusinessID, april.label, may.label)
	})

	if _, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, may.records()); err != nil {
		t.Fatalf("May CreatePayrollRecords: %v", err)
	}
	mayRows := countRows(t, db, may.label)
	mayNet := netPaySum(t, repo, may.label)

	if _, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, april.records()); err != nil {
		t.Fatalf("April CreatePayrollRecords: %v", err)
	}
	// Re-running April twice must not touch May.
	if _, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, april.records()); err != nil {
		t.Fatalf("April re-run: %v", err)
	}

	if rows := countRows(t, db, may.label); rows != mayRows {
		t.Errorf("May row count changed: %d -> %d", mayRows, rows)
	}
	if net := netPaySum(t, repo, may.label); net != mayNet {
		t.Errorf("May net_pay total changed: %v -> %v", mayNet, net)
	}
}

func TestRunPayrollAfterPartialPaymentIsRefused(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	period := aprilSpec(4101, 4102)
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE business_id = ? AND period_label = ?", idempotencyBusinessID, period.label)
	})

	if _, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, period.records()); err != nil {
		t.Fatalf("CreatePayrollRecords: %v", err)
	}

	// Settle one employee, leaving the period partially paid.
	if _, err := db.ExecContext(ctx,
		"UPDATE hr_payroll_records SET status = 'paid', paid_at = NOW() WHERE business_id = ? AND period_label = ? AND staff_id = ?",
		idempotencyBusinessID, period.label, int64(4101)); err != nil {
		t.Fatalf("mark one record paid: %v", err)
	}
	rowsBefore := countRows(t, db, period.label)

	// A recompute of a partially paid period would restate figures for money
	// that has already gone out, so it is refused and nothing is written.
	if _, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, period.records()); !errors.Is(err, domain.ErrPayrollPeriodAlreadyPaid) {
		t.Fatalf("re-run of partially paid period: err = %v, want ErrPayrollPeriodAlreadyPaid", err)
	}
	if rows := countRows(t, db, period.label); rows != rowsBefore {
		t.Errorf("refused re-run still changed rows: %d -> %d", rowsBefore, rows)
	}
	if net := netPaySum(t, repo, period.label); net != 4000.00 {
		t.Errorf("refused re-run changed net_pay total: got %v, want 4000", net)
	}
}

func TestMarkPayrollPaidRefusesDuplicateSibling(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	period := aprilSpec(4101)

	// A legacy duplicate: two unpaid rows for the same employee and period,
	// which is the state a repeated run used to leave behind.
	rows := period.records()
	rows = append(rows, rows...)
	for _, record := range rows {
		if _, err := db.ExecContext(ctx,
			"INSERT INTO hr_payroll_records (owner_id, business_id, staff_id, period_start, period_end, period_label, net_pay) VALUES (?, ?, ?, ?, ?, ?, ?)",
			idempotencyBusinessID, idempotencyBusinessID, record.StaffID, record.PeriodStart, record.PeriodEnd, record.PeriodLabel,
			fmt.Sprintf("%.2f", record.NetPay)); err != nil {
			t.Fatalf("seed duplicate row: %v", err)
		}
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE business_id = ? AND period_label = ?", idempotencyBusinessID, period.label)
	})

	var ids []int64
	rowsQuery, err := db.QueryContext(ctx,
		"SELECT id FROM hr_payroll_records WHERE business_id = ? AND period_label = ? ORDER BY id",
		idempotencyBusinessID, period.label)
	if err != nil {
		t.Fatalf("read duplicate ids: %v", err)
	}
	for rowsQuery.Next() {
		var id int64
		if err := rowsQuery.Scan(&id); err != nil {
			t.Fatalf("scan duplicate id: %v", err)
		}
		ids = append(ids, id)
	}
	rowsQuery.Close()
	if len(ids) != 2 {
		t.Skipf("unique key already applied to this database (%d duplicate rows), nothing to prove here", len(ids))
	}

	if err := repo.MarkPayrollPaid(ctx, idempotencyBusinessID, ids[0]); err != nil {
		t.Fatalf("paying the first duplicate: %v", err)
	}
	if err := repo.MarkPayrollPaid(ctx, idempotencyBusinessID, ids[1]); !errors.Is(err, domain.ErrPayrollRecordHasPaidSibling) {
		t.Fatalf("paying the duplicate sibling: err = %v, want ErrPayrollRecordHasPaidSibling", err)
	}

	var paid int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hr_payroll_records WHERE business_id = ? AND period_label = ? AND status = 'paid'",
		idempotencyBusinessID, period.label).Scan(&paid); err != nil {
		t.Fatalf("count paid rows: %v", err)
	}
	if paid != 1 {
		t.Errorf("%d records settled for one employee in one period, want 1", paid)
	}
}

// TestMarkPayrollPaidAllowsDifferentStaffInSamePeriod guards the sibling check
// from over-reaching: paying a period out employee by employee is the normal
// path, so a paid record for a different employee must not block anyone.
func TestMarkPayrollPaidAllowsDifferentStaffInSamePeriod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	period := aprilSpec(4101, 4102)
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE business_id = ? AND period_label = ?", idempotencyBusinessID, period.label)
	})

	if _, err := repo.CreatePayrollRecords(ctx, idempotencyBusinessID, period.records()); err != nil {
		t.Fatalf("CreatePayrollRecords: %v", err)
	}

	ids := map[int64]int64{}
	rowsQuery, err := db.QueryContext(ctx,
		"SELECT id, staff_id FROM hr_payroll_records WHERE business_id = ? AND period_label = ?", idempotencyBusinessID, period.label)
	if err != nil {
		t.Fatalf("read ids: %v", err)
	}
	for rowsQuery.Next() {
		var id, staffID int64
		if err := rowsQuery.Scan(&id, &staffID); err != nil {
			t.Fatalf("scan row: %v", err)
		}
		ids[staffID] = id
	}
	rowsQuery.Close()

	for _, staffID := range []int64{4101, 4102} {
		if err := repo.MarkPayrollPaid(ctx, idempotencyBusinessID, ids[staffID]); err != nil {
			t.Fatalf("paying staff %d after another was paid: %v", staffID, err)
		}
	}
}
