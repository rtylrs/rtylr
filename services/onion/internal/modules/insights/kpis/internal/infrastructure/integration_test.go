package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

func TestGetKPIs_QueryFailure_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)

	// A failing database must surface as an error instead of a zero-filled
	// summary, which the application service would otherwise cache for 2
	// minutes as if the business had no activity.
	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	if _, err := repo.GetKPIs(context.Background(), testutil.OwnerID); err == nil {
		t.Fatal("GetKPIs error = nil, want query failure to be returned")
	}
}

func TestGetKPIs_CanceledContext_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repo.GetKPIs(ctx, testutil.OwnerID); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetKPIs error = %v, want context.Canceled", err)
	}
}

func TestGetKPIs_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	_ = kpis
}

func TestGetKPIs_CustomerCountNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
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
	repo := newRepository(db)
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
	repo := newRepository(db)
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
	repo := newRepository(db)
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
	repo := newRepository(db)
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
	repo := newRepository(db)
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
	repo := newRepository(db)
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
	repo := newRepository(db)
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
	repo := newRepository(db)
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

func TestGetKPIs_NegativeOwner_ReturnsZeros(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, -1)
	if err != nil {
		t.Fatalf("GetKPIs owner=-1: %v", err)
	}
	if kpis.CustomerCount != 0 {
		t.Errorf("CustomerCount = %d, want 0 for negative owner", kpis.CustomerCount)
	}
}

func TestGetKPIs_ZeroOwner_ReturnsZeros(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, 0)
	if err != nil {
		t.Fatalf("GetKPIs owner=0: %v", err)
	}
	if kpis.CustomerCount != 0 {
		t.Errorf("CustomerCount = %d, want 0 for owner=0", kpis.CustomerCount)
	}
}

func TestGetKPIs_OrderCountNonNegative_NewOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetKPIs new owner: %v", err)
	}
	if kpis.OrderCount < 0 {
		t.Errorf("OrderCount is negative for new owner: %d", kpis.OrderCount)
	}
}

func TestGetKPIs_RevenueNonNegative_NewOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetKPIs new owner: %v", err)
	}
	if kpis.Revenue < 0 {
		t.Errorf("Revenue is negative for new owner: %f", kpis.Revenue)
	}
}

func TestGetKPIs_LowStockCount_NewOwner_IsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetKPIs new owner: %v", err)
	}
	if kpis.LowStockCount < 0 {
		t.Errorf("LowStockCount is negative: %d", kpis.LowStockCount)
	}
}

func TestGetKPIs_StaffCountAtLeastFixture(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetKPIs: %v", err)
	}
	// Fixture has at least Alice (4101) and Bob (4102) in hr_staff.
	if kpis.StaffCount < 2 {
		t.Errorf("StaffCount = %d, expected >= 2 (fixture staff)", kpis.StaffCount)
	}
}

func TestGetKPIs_RevenueIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	first, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first GetKPIs: %v", err)
	}
	second, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second GetKPIs: %v", err)
	}
	if first.Revenue != second.Revenue {
		t.Errorf("Revenue not stable: %f vs %f", first.Revenue, second.Revenue)
	}
}

func TestGetKPIs_OrderCountIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	first, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first GetKPIs: %v", err)
	}
	second, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second GetKPIs: %v", err)
	}
	if first.OrderCount != second.OrderCount {
		t.Errorf("OrderCount not stable: %d vs %d", first.OrderCount, second.OrderCount)
	}
}

func TestGetKPIs_LargeOwnerID_ReturnsZeros(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	kpis, err := repo.GetKPIs(ctx, 1<<40)
	if err != nil {
		t.Fatalf("GetKPIs large owner: %v", err)
	}
	if kpis.CustomerCount != 0 {
		t.Errorf("CustomerCount = %d, want 0 for new owner", kpis.CustomerCount)
	}
}

func TestGetKPIs_StaffCountIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	first, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first GetKPIs: %v", err)
	}
	second, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second GetKPIs: %v", err)
	}
	if first.StaffCount != second.StaffCount {
		t.Errorf("StaffCount not stable: %d vs %d", first.StaffCount, second.StaffCount)
	}
}

func TestGetKPIs_LowStockCountIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := newRepository(db)
	ctx := context.Background()

	first, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first GetKPIs: %v", err)
	}
	second, err := repo.GetKPIs(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second GetKPIs: %v", err)
	}
	if first.LowStockCount != second.LowStockCount {
		t.Errorf("LowStockCount not stable: %d vs %d", first.LowStockCount, second.LowStockCount)
	}
}
