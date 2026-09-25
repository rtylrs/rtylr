package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	"github.com/voxire/onion/internal/modules/tax/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestTax_GetConfig_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetConfig(context.Background(), testutil.OwnerID)
	if err != nil && !errors.Is(err, domain.ErrTaxConfigNotFound) {
		t.Fatalf("GetConfig: %v", err)
	}
}

func TestTax_UpsertConfig_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID:            testutil.OwnerID,
		TaxName:            "VAT",
		TaxRateBasisPoints: 2000,
		IsInclusive:        false,
		IsActive:           true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}
}

func TestTax_UpsertConfig_ThenGet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID:            testutil.OwnerID,
		TaxName:            "GST",
		TaxRateBasisPoints: 1000,
		IsActive:           true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig after Upsert: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, cfg.OwnerID)
	}
}

func TestTax_UpsertConfig_UpdatesRate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 1500, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig 15%%: %v", err)
	}
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 1800, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig 18%%: %v", err)
	}

	cfg, err := repo.GetConfig(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRateBasisPoints != 1800 {
		t.Errorf("expected tax rate 1800 basis points, got %d", cfg.TaxRateBasisPoints)
	}
}

func TestTax_GetConfig_UnknownOwner_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetConfig(context.Background(), 9999999)
	if err == nil {
		t.Fatal("expected ErrTaxConfigNotFound for unknown owner")
	}
	if !errors.Is(err, domain.ErrTaxConfigNotFound) {
		t.Errorf("expected ErrTaxConfigNotFound, got %v", err)
	}
}

func TestTax_UpsertConfig_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxName: "Sales Tax", TaxRateBasisPoints: 750, IsActive: true,
	}
	for i := 0; i < 3; i++ {
		if err := repo.UpsertConfig(context.Background(), input); err != nil {
			t.Fatalf("UpsertConfig iteration %d: %v", i, err)
		}
	}
}

func TestTax_UpsertConfig_InactiveFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 0, IsActive: false,
	}); err != nil {
		t.Fatalf("UpsertConfig inactive: %v", err)
	}

	cfg, err := repo.GetConfig(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.IsActive {
		t.Error("expected IsActive=false after upsert with inactive")
	}
}

func TestTax_GetPLReport_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	report, err := repo.GetPLReport(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport: %v", err)
	}
	if report.GrossProfitMinorUnits != report.TotalRevenueMinorUnits-report.TotalExpensesMinorUnits {
		t.Errorf("GrossProfit mismatch: %d != %d - %d", report.GrossProfitMinorUnits, report.TotalRevenueMinorUnits, report.TotalExpensesMinorUnits)
	}
}

func TestTax_GetPLReport_NetEqualsGross(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	report, err := repo.GetPLReport(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport: %v", err)
	}
	if report.NetProfitMinorUnits != report.GrossProfitMinorUnits {
		t.Errorf("expected NetProfit == GrossProfit, got %d vs %d", report.NetProfitMinorUnits, report.GrossProfitMinorUnits)
	}
}

func TestTax_UpsertConfig_WithIsInclusive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID:            testutil.OwnerID,
		TaxName:            "Inclusive Tax",
		TaxRateBasisPoints: 1000,
		IsInclusive:        true,
		IsActive:           true,
	}); err != nil {
		t.Fatalf("UpsertConfig inclusive: %v", err)
	}
}

func TestTax_GetConfig_ReturnsOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 500, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("expected OwnerID=%d, got %d", testutil.OwnerID, cfg.OwnerID)
	}
}

func TestTax_GetConfig_ReturnsNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 500, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.ID == 0 {
		t.Error("expected non-zero config ID")
	}
}

func TestTax_UpsertConfig_TaxNamePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxName: "PersistName", TaxRateBasisPoints: 300, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxName != "PersistName" {
		t.Errorf("expected TaxName=%q, got %q", "PersistName", cfg.TaxName)
	}
}

func TestTax_UpsertConfig_IsInclusivePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID:            testutil.OwnerID,
		TaxRateBasisPoints: 800,
		IsInclusive:        true,
		IsActive:           true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !cfg.IsInclusive {
		t.Error("expected IsInclusive=true")
	}
}

func TestTax_UpsertConfig_ZeroRateAllowed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig zero rate: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRateBasisPoints < 0 {
		t.Errorf("TaxRateBasisPoints should be non-negative, got %d", cfg.TaxRateBasisPoints)
	}
}

func TestTax_GetPLReport_UnknownOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetPLReport(ctx, 8888888)
	if err != nil {
		t.Fatalf("GetPLReport unknown owner: %v", err)
	}
}

func TestTax_GetPLReport_TotalRevenueNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	report, err := repo.GetPLReport(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport: %v", err)
	}
	if report.TotalRevenueMinorUnits < 0 {
		t.Errorf("TotalRevenue should be non-negative, got %d", report.TotalRevenueMinorUnits)
	}
}

func TestTax_GetPLReport_TotalExpensesNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	report, err := repo.GetPLReport(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport: %v", err)
	}
	if report.TotalExpensesMinorUnits < 0 {
		t.Errorf("TotalExpenses should be non-negative, got %d", report.TotalExpensesMinorUnits)
	}
}

func TestTax_GetPLReport_IdempotentCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	r1, err := repo.GetPLReport(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport first: %v", err)
	}
	r2, err := repo.GetPLReport(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport second: %v", err)
	}
	if r1.TotalRevenueMinorUnits != r2.TotalRevenueMinorUnits {
		t.Errorf("idempotent calls differ: %d vs %d", r1.TotalRevenueMinorUnits, r2.TotalRevenueMinorUnits)
	}
}

func TestTax_UpsertConfig_HighRate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxName: "High Rate", TaxRateBasisPoints: 9999, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig high rate: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRateBasisPoints != 9999 {
		t.Errorf("expected TaxRateBasisPoints=9999, got %d", cfg.TaxRateBasisPoints)
	}
}

func TestTax_UpsertConfig_EmptyTaxName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxName: "", TaxRateBasisPoints: 500, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig empty name: %v", err)
	}
}

func TestTax_GetConfig_HasCreatedAt(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 500, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestTax_UpsertConfig_ActiveThenInactive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 1200, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig active: %v", err)
	}
	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 1200, IsActive: false,
	}); err != nil {
		t.Fatalf("UpsertConfig inactive: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.IsActive {
		t.Error("expected IsActive=false after toggling off")
	}
}

func TestTax_GetPLReport_GrossProfitFormula(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	report, err := repo.GetPLReport(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport: %v", err)
	}
	expected := report.TotalRevenueMinorUnits - report.TotalExpensesMinorUnits
	if report.GrossProfitMinorUnits != expected {
		t.Errorf("GrossProfit=%d, want TotalRevenue-TotalExpenses=%d", report.GrossProfitMinorUnits, expected)
	}
}

func TestTax_UpsertConfig_FractionalRate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRateBasisPoints: 725, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig fractional: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRateBasisPoints != 725 {
		t.Errorf("expected TaxRateBasisPoints=725, got %d", cfg.TaxRateBasisPoints)
	}
}
