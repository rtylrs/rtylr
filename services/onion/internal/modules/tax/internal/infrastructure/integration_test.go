package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	"github.com/voxire/onion/internal/modules/tax/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestTax_GetConfig_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetConfig(context.Background(), testutil.OwnerID)
	if err != nil && err != domain.ErrTaxConfigNotFound {
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
	if err != domain.ErrTaxConfigNotFound {
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
