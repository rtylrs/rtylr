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
		OwnerID:     testutil.OwnerID,
		TaxName:     "VAT",
		TaxRate:     20.0,
		IsInclusive: false,
		IsActive:    true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}
}

func TestTax_UpsertConfig_ThenGet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID:  testutil.OwnerID,
		TaxName:  "GST",
		TaxRate:  10.0,
		IsActive: true,
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
		OwnerID: testutil.OwnerID, TaxRate: 15.0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig 15%%: %v", err)
	}
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRate: 18.0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig 18%%: %v", err)
	}

	cfg, err := repo.GetConfig(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRate != 18.0 {
		t.Errorf("expected tax rate 18.0, got %f", cfg.TaxRate)
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
		OwnerID: testutil.OwnerID, TaxName: "Sales Tax", TaxRate: 7.5, IsActive: true,
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
		OwnerID: testutil.OwnerID, TaxRate: 0.0, IsActive: false,
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
	if report.GrossProfit != report.TotalRevenue-report.TotalExpenses {
		t.Errorf("GrossProfit mismatch: %.2f != %.2f - %.2f", report.GrossProfit, report.TotalRevenue, report.TotalExpenses)
	}
}

func TestTax_GetPLReport_NetEqualsGross(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	report, err := repo.GetPLReport(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPLReport: %v", err)
	}
	if report.NetProfit != report.GrossProfit {
		t.Errorf("expected NetProfit == GrossProfit, got %.2f vs %.2f", report.NetProfit, report.GrossProfit)
	}
}

func TestTax_UpsertConfig_WithIsInclusive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	if err := repo.UpsertConfig(context.Background(), domain.TaxConfigInput{
		OwnerID:     testutil.OwnerID,
		TaxName:     "Inclusive Tax",
		TaxRate:     10.0,
		IsInclusive: true,
		IsActive:    true,
	}); err != nil {
		t.Fatalf("UpsertConfig inclusive: %v", err)
	}
}

func TestTax_GetConfig_ReturnsOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRate: 5.0, IsActive: true,
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
		OwnerID: testutil.OwnerID, TaxRate: 5.0, IsActive: true,
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
		OwnerID: testutil.OwnerID, TaxName: "PersistName", TaxRate: 3.0, IsActive: true,
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
		OwnerID:     testutil.OwnerID,
		TaxRate:     8.0,
		IsInclusive: true,
		IsActive:    true,
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
		OwnerID: testutil.OwnerID, TaxRate: 0.0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig zero rate: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRate < 0 {
		t.Errorf("TaxRate should be non-negative, got %f", cfg.TaxRate)
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
	if report.TotalRevenue < 0 {
		t.Errorf("TotalRevenue should be non-negative, got %f", report.TotalRevenue)
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
	if report.TotalExpenses < 0 {
		t.Errorf("TotalExpenses should be non-negative, got %f", report.TotalExpenses)
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
	if r1.TotalRevenue != r2.TotalRevenue {
		t.Errorf("idempotent calls differ: %.2f vs %.2f", r1.TotalRevenue, r2.TotalRevenue)
	}
}

func TestTax_UpsertConfig_HighRate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxName: "High Rate", TaxRate: 99.99, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig high rate: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRate != 99.99 {
		t.Errorf("expected TaxRate=99.99, got %f", cfg.TaxRate)
	}
}

func TestTax_UpsertConfig_EmptyTaxName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxName: "", TaxRate: 5.0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig empty name: %v", err)
	}
}

func TestTax_GetConfig_HasCreatedAt(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRate: 5.0, IsActive: true,
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
		OwnerID: testutil.OwnerID, TaxRate: 12.0, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig active: %v", err)
	}
	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRate: 12.0, IsActive: false,
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
	expected := report.TotalRevenue - report.TotalExpenses
	if report.GrossProfit != expected {
		t.Errorf("GrossProfit=%.2f, want TotalRevenue-TotalExpenses=%.2f", report.GrossProfit, expected)
	}
}

func TestTax_UpsertConfig_FractionalRate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.TaxConfigInput{
		OwnerID: testutil.OwnerID, TaxRate: 7.25, IsActive: true,
	}); err != nil {
		t.Fatalf("UpsertConfig fractional: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.TaxRate != 7.25 {
		t.Errorf("expected TaxRate=7.25, got %f", cfg.TaxRate)
	}
}
