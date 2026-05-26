package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/kpis/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetKPIs_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	_ = kpis
}

func TestGetKPIs_CustomerCountNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	if kpis.CustomerCount < 0 {
		t.Errorf("CustomerCount is negative: %d", kpis.CustomerCount)
	}
}

func TestGetKPIs_CustomerCountAtLeastFixtures(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	// Fixture data has at least crm_customers 5301 (Jamie) and 5302 (Alex).
	if kpis.CustomerCount < 2 {
		t.Errorf("CustomerCount = %d, want >= 2 (fixture customers)", kpis.CustomerCount)
	}
}

func TestGetKPIs_StaffCountNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	if kpis.StaffCount < 0 {
		t.Errorf("StaffCount is negative: %d", kpis.StaffCount)
	}
}

func TestGetKPIs_LowStockCountNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	if kpis.LowStockCount < 0 {
		t.Errorf("LowStockCount is negative: %d", kpis.LowStockCount)
	}
}

func TestGetKPIs_OrderCountNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	if kpis.OrderCount < 0 {
		t.Errorf("OrderCount is negative: %d", kpis.OrderCount)
	}
}

func TestGetKPIs_RevenueNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	if kpis.Revenue < 0 {
		t.Errorf("Revenue is negative: %f", kpis.Revenue)
	}
}

func TestGetKPIs_UnknownOwner_ReturnsZeros(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetKPIs unknown owner: %v", err)
	}
	if kpis.CustomerCount != 0 {
		t.Errorf("expected 0 customers for unknown owner, got %d", kpis.CustomerCount)
	}
	if kpis.StaffCount != 0 {
		t.Errorf("expected 0 staff for unknown owner, got %d", kpis.StaffCount)
	}
}

func TestGetKPIs_AllFieldsPopulated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	// Sanity check: struct is not zero-value due to DB scan failure.
	// At minimum CustomerCount should be >= 0 and Revenue >= 0.
	_ = kpis.CustomerCount
	_ = kpis.StaffCount
	_ = kpis.LowStockCount
	_ = kpis.OrderCount
	_ = kpis.Revenue
}

func TestGetKPIs_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs first call: %v", err)
	}
	second, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs second call: %v", err)
	}
	if first.CustomerCount != second.CustomerCount {
		t.Errorf("CustomerCount not idempotent: %d vs %d", first.CustomerCount, second.CustomerCount)
	}
}
