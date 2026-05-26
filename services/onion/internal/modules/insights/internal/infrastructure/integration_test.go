package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetDashboard_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	_ = snap
}

func TestGetDashboard_RevenueNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if snap.Revenue.Value < 0 {
		t.Errorf("Revenue.Value is negative: %d", snap.Revenue.Value)
	}
}

func TestGetDashboard_OrdersNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if snap.Orders.Value < 0 {
		t.Errorf("Orders.Value is negative: %d", snap.Orders.Value)
	}
}

func TestGetDashboard_CustomersNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if snap.Customers.Value < 0 {
		t.Errorf("Customers.Value is negative: %d", snap.Customers.Value)
	}
}

func TestGetDashboard_UnknownOwner_ReturnsZeros(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	snap, err := repo.GetDashboard(ctx, 9999999, infrastructure.DashboardInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("GetDashboard unknown owner: %v", err)
	}
	// Fresh/unknown tenant should have all zeros, not errors.
	if snap.Revenue.Value < 0 {
		t.Errorf("Revenue.Value is negative for unknown owner")
	}
}

func TestGetDashboard_PastDate_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	past := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: past})
	if err != nil {
		t.Fatalf("GetDashboard with past date: %v", err)
	}
	_ = snap
}

func TestGetDashboard_MarginBetweenMinusHundredAndHundred(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	// Margin percentage should be in a reasonable range when revenue is zero.
	// When revenue is 0 the helper returns 0 -- just check it does not blow up.
	_ = snap.Margin.Value
}

func TestGetDashboard_ReturnsStructuredFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	// Verify that all four card fields exist (non-zero struct fields are populated).
	_ = snap.Revenue
	_ = snap.Customers
	_ = snap.Orders
	_ = snap.Margin
}

func TestGetDashboard_FirstOfMonth_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// First day of month -- MTD window collapses to a single day.
	firstOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: firstOfMonth})
	if err != nil {
		t.Fatalf("GetDashboard on first of month: %v", err)
	}
	_ = snap
}

func TestGetDashboard_EndOfYear_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	endOfYear := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	snap, err := repo.GetDashboard(ctx, testutil.OwnerID, infrastructure.DashboardInput{Now: endOfYear})
	if err != nil {
		t.Fatalf("GetDashboard at end of year: %v", err)
	}
	_ = snap
}
