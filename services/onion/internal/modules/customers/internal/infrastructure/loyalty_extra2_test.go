package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLoyaltyTransactions_PaginationPage1(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	txs, total, err := repo.ListLoyaltyTransactions(ctx, testutil.OwnerID, 0, 1, 5)
	if err != nil {
		t.Fatalf("ListLoyaltyTransactions(page=1, pageSize=5): %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	if len(txs) > 5 {
		t.Errorf("expected at most 5 results per page, got %d", len(txs))
	}
}

func TestUpsertLoyaltyConfig_PointsPerCurrencyUnit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertLoyaltyConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:               testutil.OwnerID,
		IsEnabled:             true,
		PointsPerCurrencyUnit: 3.5,
	}); err != nil {
		t.Fatalf("UpsertLoyaltyConfig: %v", err)
	}

	cfg, err := repo.GetLoyaltyConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLoyaltyConfig: %v", err)
	}
	if cfg.PointsPerCurrencyUnit != 3.5 {
		t.Errorf("PointsPerCurrencyUnit = %f, want 3.5", cfg.PointsPerCurrencyUnit)
	}
}

func TestGetLoyaltyConfig_IsEnabledIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertLoyaltyConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:   testutil.OwnerID,
		IsEnabled: true,
	}); err != nil {
		t.Fatalf("UpsertLoyaltyConfig: %v", err)
	}

	cfg, err := repo.GetLoyaltyConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLoyaltyConfig: %v", err)
	}
	if !cfg.IsEnabled {
		t.Error("expected IsEnabled=true after upsert")
	}
}

func TestGetLoyaltyConfig_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertLoyaltyConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:   testutil.OwnerID,
		IsEnabled: false,
	}); err != nil {
		t.Fatalf("UpsertLoyaltyConfig: %v", err)
	}

	cfg, err := repo.GetLoyaltyConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLoyaltyConfig: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", cfg.OwnerID, testutil.OwnerID)
	}
}
