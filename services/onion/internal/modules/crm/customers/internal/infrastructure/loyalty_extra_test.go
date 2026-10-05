package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLoyaltyTransactions_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	txs, total, err := repo.ListLoyaltyTransactions(ctx, testutil.OwnerID, 5301, 1, 20)
	if err != nil {
		t.Fatalf("ListLoyaltyTransactions: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = txs
}

func TestListLoyaltyTransactions_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	txs, total, err := repo.ListLoyaltyTransactions(ctx, 0, 5301, 1, 20)
	if err != nil {
		t.Fatalf("ListLoyaltyTransactions(ownerID=0): %v", err)
	}
	if total != 0 || len(txs) != 0 {
		t.Fatalf("expected 0 transactions for ownerID=0, got total=%d len=%d", total, len(txs))
	}
}

func TestUpsertLoyaltyConfig_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	err := repo.UpsertLoyaltyConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:               testutil.OwnerID,
		BusinessID:            testutil.BusinessID,
		IsEnabled:             true,
		PointsPerCurrencyUnit: 1.0,
		CurrencyPerPoint:      0.01,
		BronzeMin:             0,
		SilverMin:             500,
		GoldMin:               1500,
		PlatinumMin:           3000,
	})
	if err != nil {
		t.Fatalf("UpsertLoyaltyConfig: %v", err)
	}
}

func TestGetLoyaltyConfig_AfterUpsert_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	if err := repo.UpsertLoyaltyConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		IsEnabled:  true,
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

func TestRecordVisit_UpdatesTotalSpent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	before, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5302)
	if err != nil {
		t.Fatalf("GetCustomerByID before visit: %v", err)
	}

	if err := repo.RecordVisit(ctx, testutil.OwnerID, 5302, 100.00); err != nil {
		t.Fatalf("RecordVisit: %v", err)
	}

	after, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5302)
	if err != nil {
		t.Fatalf("GetCustomerByID after visit: %v", err)
	}

	if after.TotalSpent <= before.TotalSpent {
		t.Errorf("expected TotalSpent to increase: before=%f after=%f", before.TotalSpent, after.TotalSpent)
	}
}
