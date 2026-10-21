package infrastructure_test

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

var regressionBusinessBase int64 = time.Now().UnixNano()

func nextRegressionBusinessID() int64 {
	return atomic.AddInt64(&regressionBusinessBase, 1)
}

func TestCustomerMetricsAndTopCustomersHonorPeriodSortAndPage(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := nextRegressionBusinessID()
	repo := newRepository(db)

	insertCustomer := func(name string, createdAt time.Time) int64 {
		t.Helper()
		result, err := db.ExecContext(ctx,
			`INSERT INTO crm_customers (owner_id, business_id, first_name, last_name, created_at) VALUES (?, ?, ?, ?, ?)`,
			businessID, businessID, name, "test", createdAt)
		if err != nil {
			t.Fatalf("insert customer %s: %v", name, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("customer id %s: %v", name, err)
		}
		return id
	}
	insertOrder := func(customerID int64, suffix string, at time.Time, total float64) {
		t.Helper()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO pos_orders (owner_id, business_id, order_number, customer_id, total, payment_status, created_at) VALUES (?, ?, ?, ?, ?, 'paid', ?)`,
			businessID, businessID, suffix, customerID, total, at); err != nil {
			t.Fatalf("insert order %s: %v", suffix, err)
		}
	}

	aliceID := insertCustomer("Alice", time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC))
	bobID := insertCustomer("Bob", time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))
	carolID := insertCustomer("Carol", time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC))
	insertOrder(aliceID, "customer-period-old", time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC), 1)
	insertOrder(aliceID, "customer-period-a1", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), 10)
	insertOrder(aliceID, "customer-period-a2", time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC), 20)
	insertOrder(aliceID, "customer-period-feb", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), 100)
	insertOrder(bobID, "customer-period-b", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), 5)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_orders WHERE business_id = ?`, businessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_customers WHERE business_id = ?`, businessID)
	})

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	kpis, err := repo.GetCustomerKPIs(ctx, businessID, from, to)
	if err != nil {
		t.Fatalf("GetCustomerKPIs: %v", err)
	}
	if kpis.Total != 3 || kpis.NewThisPeriod != 2 || kpis.Active != 2 || kpis.Churned != 0 {
		t.Fatalf("January KPIs = %+v, want total=3 new=2 active=2 churned=0", kpis)
	}

	churn, err := repo.GetChurnRates(ctx, businessID,
		time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetChurnRates: %v", err)
	}
	if len(churn) != 1 || churn[0].ActiveCount != 0 || churn[0].ChurnedCount != 2 {
		t.Fatalf("March churn = %+v, want both purchasing customers churned", churn)
	}

	firstSpendPage, err := repo.GetTopCustomers(ctx, businessID, infrastructure.TopCustomerQuery{
		From: from, To: to, Limit: 1, Offset: 0, Descending: true,
	})
	if err != nil {
		t.Fatalf("top customers spend page 1: %v", err)
	}
	if firstSpendPage.Total != 2 || len(firstSpendPage.Rows) != 1 || firstSpendPage.Rows[0].CustomerID != aliceID {
		t.Fatalf("spend page 1 = %+v, want Alice and total 2", firstSpendPage)
	}
	secondSpendPage, err := repo.GetTopCustomers(ctx, businessID, infrastructure.TopCustomerQuery{
		From: from, To: to, Limit: 1, Offset: 1, Descending: true,
	})
	if err != nil {
		t.Fatalf("top customers spend page 2: %v", err)
	}
	if len(secondSpendPage.Rows) != 1 || secondSpendPage.Rows[0].CustomerID != bobID {
		t.Fatalf("spend page 2 = %+v, want Bob", secondSpendPage)
	}
	orderCountPage, err := repo.GetTopCustomers(ctx, businessID, infrastructure.TopCustomerQuery{
		From: from, To: to, Limit: 1, Offset: 0, ByOrderCount: true, Descending: false,
	})
	if err != nil {
		t.Fatalf("top customers order-count page: %v", err)
	}
	if len(orderCountPage.Rows) != 1 || orderCountPage.Rows[0].CustomerID != bobID {
		t.Fatalf("ascending order-count page = %+v, want Bob", orderCountPage)
	}
	_ = carolID
}

func TestHeadcountTimeOffAndUtilizationHonorPeriodAndTerminationHistory(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := nextRegressionBusinessID()
	repo := newRepository(db)

	insertEmployee := func(name, status string, start, end time.Time) int64 {
		t.Helper()
		var endValue any
		if !end.IsZero() {
			endValue = end
		}
		result, err := db.ExecContext(ctx,
			`INSERT INTO hr_employees (owner_id, business_id, first_name, last_name, email, status, start_date, end_date) VALUES (?, ?, ?, 'Test', ?, ?, ?, ?)`,
			businessID, businessID, name, name+"@example.com", status, start, endValue)
		if err != nil {
			t.Fatalf("insert employee %s: %v", name, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("employee id %s: %v", name, err)
		}
		return id
	}

	departedID := insertEmployee("Departed", "terminated", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))
	existingID := insertEmployee("Existing", "active", time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC), time.Time{})
	hiredID := insertEmployee("Hired", "active", time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), time.Time{})
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM hr_leave_requests WHERE business_id = ?`, businessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM hr_employees WHERE business_id = ?`, businessID)
	})

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	headcount, err := repo.GetHeadcount(ctx, businessID, from, to)
	if err != nil {
		t.Fatalf("GetHeadcount: %v", err)
	}
	if headcount.HeadcountAtStart != 2 || headcount.ActiveHeadcount != 2 || headcount.HiredThisPeriod != 1 || headcount.DepartedThisPeriod != 1 {
		t.Fatalf("headcount = %+v, want start=2 end=2 hired=1 departed=1", headcount)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_leave_requests (owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days, status) VALUES (?, ?, ?, 1, '2026-01-12', '2026-01-12', 1, 'pending'), (?, ?, ?, 1, '2026-01-13', '2026-01-13', 1, 'approved'), (?, ?, ?, 1, '2026-02-02', '2026-02-02', 1, 'pending')`,
		businessID, businessID, existingID, businessID, businessID, hiredID, businessID, businessID, departedID); err != nil {
		t.Fatalf("insert leave requests: %v", err)
	}
	timeOff, err := repo.GetTimeOffSummary(ctx, businessID, from, to)
	if err != nil {
		t.Fatalf("GetTimeOffSummary: %v", err)
	}
	if timeOff.Pending != 1 || timeOff.Approved != 1 {
		t.Fatalf("January time off = %+v, want pending=1 approved=1", timeOff)
	}

	shiftResult, err := db.ExecContext(ctx,
		`INSERT INTO hr_shifts (owner_id, business_id, start_time, end_time, status) VALUES (?, ?, '09:00:00', '17:00:00', 'scheduled')`,
		businessID, businessID)
	if err != nil {
		t.Fatalf("insert shift: %v", err)
	}
	shiftID, _ := shiftResult.LastInsertId()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_shift_assignments (owner_id, business_id, employee_id, shift_id, date, status) VALUES (?, ?, ?, ?, '2026-01-05', 'scheduled')`,
		businessID, businessID, hiredID, shiftID); err != nil {
		t.Fatalf("insert shift assignment: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM hr_shift_assignments WHERE business_id = ?`, businessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM hr_shifts WHERE business_id = ?`, businessID)
	})
	utilization, err := repo.GetUtilization(ctx, businessID, from, to)
	if err != nil {
		t.Fatalf("GetUtilization: %v", err)
	}
	if utilization.ScheduledHours != 8 {
		t.Fatalf("scheduled hours = %v, want 8", utilization.ScheduledHours)
	}
	if utilization.AvailableHours != 352 { // two employees at period end × 22 weekdays × 8 hours
		t.Fatalf("available hours = %v, want 352", utilization.AvailableHours)
	}
}

func TestScheduledReportPaginationAndActivePresenceArePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := nextRegressionBusinessID()
	repo := newRepository(db)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM insights_scheduled_reports WHERE business_id = ?`, businessID)
	})

	create := func(recipient string, active *bool) int64 {
		t.Helper()
		id, err := repo.CreateScheduledReport(ctx, infrastructure.ScheduledReportWrite{
			OwnerID: businessID, BusinessID: businessID, TemplateID: "revenue-monthly",
			Cadence: "monthly", Channel: "email", Recipient: recipient, IsActive: active,
		})
		if err != nil {
			t.Fatalf("create report %s: %v", recipient, err)
		}
		return id
	}

	cID := create("c@example.com", nil)
	aID := create("a@example.com", nil)
	bID := create("b@example.com", nil)

	cRow, err := repo.GetScheduledReport(ctx, businessID, cID)
	if err != nil {
		t.Fatalf("get report c: %v", err)
	}
	if !cRow.IsActive {
		t.Fatal("omitted create is_active should use the active default")
	}
	if err := repo.UpdateScheduledReport(ctx, infrastructure.ScheduledReportWrite{
		ID: cID, BusinessID: businessID, TemplateID: cRow.TemplateID, Cadence: cRow.Cadence,
		Channel: cRow.Channel, Recipient: cRow.Recipient, IsActive: nil,
	}); err != nil {
		t.Fatalf("update report without is_active: %v", err)
	}
	cRow, _ = repo.GetScheduledReport(ctx, businessID, cID)
	if !cRow.IsActive {
		t.Fatal("omitted update is_active should preserve the active value")
	}

	paused := false
	if err := repo.UpdateScheduledReport(ctx, infrastructure.ScheduledReportWrite{
		ID: cID, BusinessID: businessID, TemplateID: cRow.TemplateID, Cadence: cRow.Cadence,
		Channel: cRow.Channel, Recipient: cRow.Recipient, IsActive: &paused,
	}); err != nil {
		t.Fatalf("pause report: %v", err)
	}
	if err := repo.UpdateScheduledReport(ctx, infrastructure.ScheduledReportWrite{
		ID: cID, BusinessID: businessID, TemplateID: cRow.TemplateID, Cadence: cRow.Cadence,
		Channel: cRow.Channel, Recipient: cRow.Recipient, IsActive: nil,
	}); err != nil {
		t.Fatalf("update paused report without is_active: %v", err)
	}
	cRow, _ = repo.GetScheduledReport(ctx, businessID, cID)
	if cRow.IsActive {
		t.Fatal("omitted update is_active should preserve an explicitly paused value")
	}

	page1, total, err := repo.ListScheduledReports(ctx, businessID, "recipient", "asc", 2, 0)
	if err != nil {
		t.Fatalf("list reports page 1: %v", err)
	}
	if total != 3 || len(page1) != 2 || page1[0].Recipient != "a@example.com" || page1[1].Recipient != "b@example.com" {
		t.Fatalf("page 1 = %+v, total=%d", page1, total)
	}
	page2, _, err := repo.ListScheduledReports(ctx, businessID, "recipient", "asc", 2, 2)
	if err != nil {
		t.Fatalf("list reports page 2: %v", err)
	}
	if len(page2) != 1 || page2[0].ID != cID {
		t.Fatalf("page 2 = %+v, want report %d", page2, cID)
	}
	if aID == 0 || bID == 0 {
		t.Fatal("expected created report ids")
	}
}

func TestInsightsSettingsRoundTripAllContractFields(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := nextRegressionBusinessID()
	repo := newRepository(db)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM insights_settings WHERE business_id = ?`, businessID)
	})

	want := infrastructure.SettingsRow{
		OwnerID: businessID, BusinessID: businessID, DefaultCurrency: "EUR",
		FiscalYearStart: "04-01", Timezone: "Europe/Paris", DefaultPeriod: "this_quarter",
		ComparePrevious: true, DashboardWidgets: []string{"revenue", "customers"},
		FavoriteReports: []string{"weekly-revenue"},
	}
	if err := repo.UpsertSettings(ctx, want); err != nil {
		t.Fatalf("UpsertSettings: %v", err)
	}
	got, err := repo.GetSettings(ctx, businessID)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if got.DefaultPeriod != want.DefaultPeriod || got.ComparePrevious != want.ComparePrevious ||
		!reflect.DeepEqual(got.DashboardWidgets, want.DashboardWidgets) || !reflect.DeepEqual(got.FavoriteReports, want.FavoriteReports) {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}
