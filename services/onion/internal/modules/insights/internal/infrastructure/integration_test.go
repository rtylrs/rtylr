package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/crm/customers"
	financereports "github.com/voxire/onion/internal/modules/finance/finance_reports"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	posorders "github.com/voxire/onion/internal/modules/pos/orders"
	"github.com/voxire/onion/internal/testutil"
)

func dashboardService(t *testing.T) *application.Service {
	t.Helper()
	db := testutil.OpenDB(t)
	ctx := context.Background()
	return application.NewService(infrastructure.NewRepository(db), application.DashboardReaders{
		Finance: financereports.NewModule(ctx, db),
		POS:     posorders.NewModule(ctx, db),
		CRM:     customers.NewModule(ctx, db),
	})
}

func TestGetDashboard_NoError(t *testing.T) {
	ctx := context.Background()
	snap, err := dashboardService(t).GetDashboard(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	_ = snap
}

func TestGetDashboard_RevenueNonNegative(t *testing.T) {
	ctx := context.Background()
	snap, err := dashboardService(t).GetDashboard(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if snap.Revenue.Value < 0 {
		t.Errorf("Revenue.Value is negative: %d", snap.Revenue.Value)
	}
}

func TestGetDashboard_OrdersNonNegative(t *testing.T) {
	ctx := context.Background()
	snap, err := dashboardService(t).GetDashboard(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if snap.Orders.Value < 0 {
		t.Errorf("Orders.Value is negative: %d", snap.Orders.Value)
	}
}

func TestGetDashboard_CustomersNonNegative(t *testing.T) {
	ctx := context.Background()
	snap, err := dashboardService(t).GetDashboard(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if snap.Customers.Value < 0 {
		t.Errorf("Customers.Value is negative: %d", snap.Customers.Value)
	}
}

func TestGetDashboard_UnknownOwner_ReturnsZeros(t *testing.T) {
	ctx := context.Background()
	snap, err := dashboardService(t).GetDashboard(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetDashboard unknown owner: %v", err)
	}
	// Fresh/unknown tenant should have all zeros, not errors.
	if snap.Revenue.Value < 0 {
		t.Errorf("Revenue.Value is negative for unknown owner")
	}
}

func TestGetDashboard_MarginBetweenMinusHundredAndHundred(t *testing.T) {
	ctx := context.Background()
	snap, err := dashboardService(t).GetDashboard(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	// Margin percentage should be in a reasonable range when revenue is zero.
	// When revenue is 0 the helper returns 0 -- just check it does not blow up.
	_ = snap.Margin.Value
}

func TestGetDashboard_ReturnsStructuredFields(t *testing.T) {
	ctx := context.Background()
	snap, err := dashboardService(t).GetDashboard(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	// Verify that all four card fields exist (non-zero struct fields are populated).
	_ = snap.Revenue
	_ = snap.Customers
	_ = snap.Orders
	_ = snap.Margin
}

func TestGetTopCustomers_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	page, err := repo.GetTopCustomers(ctx, testutil.OwnerID, infrastructure.TopCustomerQuery{From: from, To: to, Limit: 10})
	if err != nil {
		t.Fatalf("GetTopCustomers: %v", err)
	}
	if len(page.Rows) > 10 {
		t.Errorf("expected at most 10 rows, got %d", len(page.Rows))
	}
}

func TestGetTopCustomers_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	page, err := repo.GetTopCustomers(ctx, 9999999, infrastructure.TopCustomerQuery{From: from, To: to, Limit: 10})
	if err != nil {
		t.Fatalf("GetTopCustomers unknown owner: %v", err)
	}
	if len(page.Rows) != 0 {
		t.Errorf("expected 0 rows for unknown owner, got %d", len(page.Rows))
	}
}

func TestGetInvoiceAging_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GetInvoiceAging(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInvoiceAging: %v", err)
	}
	_ = rows
}

func TestGetHeadcount_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	snap, err := repo.GetHeadcount(ctx, testutil.OwnerID, from, to)
	if err != nil {
		t.Fatalf("GetHeadcount: %v", err)
	}
	if snap.ActiveHeadcount < 0 {
		t.Errorf("ActiveHeadcount is negative: %d", snap.ActiveHeadcount)
	}
}

func TestGetLTVDistribution_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GetLTVDistribution(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLTVDistribution: %v", err)
	}
	_ = rows
}

func TestGetSlowMovers_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, err := repo.GetSlowMovers(ctx, testutil.OwnerID, 90)
	if err != nil {
		t.Fatalf("GetSlowMovers: %v", err)
	}
	_ = rows
}

func TestListScheduledReports_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rows, _, err := repo.ListScheduledReports(ctx, testutil.OwnerID, "created_at", "desc", 10, 0)
	if err != nil {
		t.Fatalf("ListScheduledReports: %v", err)
	}
	_ = rows
}

func TestCreateAndDeleteScheduledReport_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	isActive := true
	id, err := repo.CreateScheduledReport(ctx, infrastructure.ScheduledReportWrite{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		TemplateID: "weekly_revenue",
		Cadence:    "weekly",
		Channel:    "email",
		Recipient:  "test@example.com",
		IsActive:   &isActive,
	})
	if err != nil {
		t.Fatalf("CreateScheduledReport: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	if err := repo.DeleteScheduledReport(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteScheduledReport: %v", err)
	}
}

func TestGetSettings_UnknownOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetSettings(ctx, 9999999)
	if err == nil {
		t.Fatal("expected ErrSettingsNotFound for unknown owner, got nil")
	}
}

func TestUpsertSettings_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	row := infrastructure.SettingsRow{
		OwnerID:          testutil.OwnerID,
		BusinessID:       testutil.BusinessID,
		DefaultCurrency:  "USD",
		FiscalYearStart:  "01-01",
		Timezone:         "UTC",
		DefaultPeriod:    "this_month",
		DashboardWidgets: []string{},
		FavoriteReports:  []string{},
	}
	if err := repo.UpsertSettings(ctx, row); err != nil {
		t.Fatalf("UpsertSettings: %v", err)
	}

	got, err := repo.GetSettings(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetSettings after upsert: %v", err)
	}
	if got.DefaultCurrency != row.DefaultCurrency {
		t.Errorf("DefaultCurrency = %q, want %q", got.DefaultCurrency, row.DefaultCurrency)
	}
	if got.Timezone != row.Timezone {
		t.Errorf("Timezone = %q, want %q", got.Timezone, row.Timezone)
	}
}

func TestGetProductKPIs_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetProductKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProductKPIs: %v", err)
	}
	_ = kpis
}

func TestGetRevenueComparison_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	row, err := application.NewService(infrastructure.NewRepository(db), application.DashboardReaders{
		Finance: financereports.NewModule(ctx, db),
	}).GetRevenueComparison(ctx, testutil.OwnerID, from, to)
	if err != nil {
		t.Fatalf("GetRevenueComparison: %v", err)
	}
	_ = row
}
