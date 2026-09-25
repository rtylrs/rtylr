package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetDaySummary_TodayNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Empty date string triggers the "today" branch.
	_, err := repo.GetDaySummary(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetDaySummary (today) error: %v", err)
	}
}

func TestGetDaySummary_ByDateNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetDaySummary(ctx, testutil.OwnerID, "2026-01-01")
	if err != nil {
		t.Fatalf("GetDaySummary (by date) error: %v", err)
	}
}

func TestGetDaySummary_TotalOrdersNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	if summary.TotalOrders < 0 {
		t.Errorf("TotalOrders is negative: %d", summary.TotalOrders)
	}
}

func TestGetDaySummary_TotalSalesNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	if summary.TotalSales < 0 {
		t.Errorf("TotalSales is negative: %g", summary.TotalSales)
	}
}

func TestGetDaySummary_TotalRefundsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	if summary.TotalRefunds < 0 {
		t.Errorf("TotalRefunds is negative: %g", summary.TotalRefunds)
	}
}

func TestGetDaySummary_PaymentBreakdownIsMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	if summary.PaymentBreakdown == nil {
		t.Error("expected PaymentBreakdown map to be initialized, got nil")
	}
}

func TestGetDaySummary_UnknownOwnerReturnsZeroCounts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, 0, "")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	if summary.TotalOrders != 0 {
		t.Errorf("expected TotalOrders=0 for unknown owner, got %d", summary.TotalOrders)
	}
	if summary.TotalSales != 0 {
		t.Errorf("expected TotalSales=0 for unknown owner, got %g", summary.TotalSales)
	}
}

func TestGetDaySummary_InvalidDateReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetDaySummary(ctx, testutil.OwnerID, "not-a-date")
	if err == nil {
		t.Fatal("expected parse error for invalid date, got nil")
	}
}

func TestGetDaySummary_ByDatePaymentBreakdownInitialized(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, testutil.OwnerID, "2026-01-15")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	if summary.PaymentBreakdown == nil {
		t.Error("PaymentBreakdown should be initialized even with no data")
	}
}

func TestGetDaySummary_TwiceReturnsSameValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetDaySummary(ctx, testutil.OwnerID, "2026-01-01")
	if err != nil {
		t.Fatalf("first GetDaySummary: %v", err)
	}
	second, err := repo.GetDaySummary(ctx, testutil.OwnerID, "2026-01-01")
	if err != nil {
		t.Fatalf("second GetDaySummary: %v", err)
	}
	if first.TotalOrders != second.TotalOrders {
		t.Errorf("TotalOrders changed: %d vs %d", first.TotalOrders, second.TotalOrders)
	}
	if first.TotalSales != second.TotalSales {
		t.Errorf("TotalSales changed: %g vs %g", first.TotalSales, second.TotalSales)
	}
}

func TestGetDaySummary_PaymentBreakdownValuesNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	for method, amount := range summary.PaymentBreakdown {
		if amount < 0 {
			t.Errorf("PaymentBreakdown[%q]=%g is negative", method, amount)
		}
	}
}

func TestGetDaySummary_TotalTaxNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetDaySummary(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetDaySummary error: %v", err)
	}
	if summary.TotalTax < 0 {
		t.Errorf("TotalTax is negative: %g", summary.TotalTax)
	}
}
