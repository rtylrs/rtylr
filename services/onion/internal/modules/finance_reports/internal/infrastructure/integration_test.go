package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestSumAccountsByType_AssetTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "asset")
	if err != nil {
		t.Fatalf("SumAccountsByType asset: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumAccountsByType_LiabilityTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "liability")
	if err != nil {
		t.Fatalf("SumAccountsByType liability: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumAccountsByType_EquityTypeNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, testutil.OwnerID, "equity")
	if err != nil {
		t.Fatalf("SumAccountsByType equity: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumAccountsByType_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumAccountsByType(ctx, 0, "asset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumBankBalances_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumBankBalances(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumBankBalances: %v", err)
	}
	_ = val // result may be 0 if no bank accounts exist for fixture owner
}

func TestSumBankBalances_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumBankBalances(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumReceivables_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumReceivables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumReceivables: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative receivables sum, got %f", val)
	}
}

func TestSumPayables_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPayables(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumPayables: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative payables sum, got %f", val)
	}
}

func TestSumPaidInvoicesInRange_CurrentYearNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPaidInvoicesInRange(ctx, testutil.OwnerID, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumPaidInvoicesInRange_EmptyDatesNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumPaidInvoicesInRange(ctx, testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("SumPaidInvoicesInRange empty dates: %v", err)
	}
	_ = val
}

func TestSumExpensesInRange_CurrentYearNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumExpensesInRange(ctx, testutil.OwnerID, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("SumExpensesInRange: %v", err)
	}
	if val < 0 {
		t.Errorf("expected non-negative sum, got %f", val)
	}
}

func TestSumExpensesInRange_UnknownOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	val, err := repo.SumExpensesInRange(ctx, 0, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for unknown owner, got %f", val)
	}
}

func TestSumExpensesInRange_MarqueeRange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// March 2026 range aligns with payroll fixture month.
	val, err := repo.SumExpensesInRange(ctx, testutil.OwnerID, "2026-03-01", "2026-03-31")
	if err != nil {
		t.Fatalf("SumExpensesInRange March 2026: %v", err)
	}
	_ = val
}
