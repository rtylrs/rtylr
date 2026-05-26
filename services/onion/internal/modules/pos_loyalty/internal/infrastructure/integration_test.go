package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// businessID=1 is the fixture business for owner_id=1001 (businesses: id=1, owner_id=1001).
const fixtureBusinessID int64 = 1

func TestLoyalty_GetConfig_NoErrorOrNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.GetConfig(context.Background(), fixtureBusinessID)
	if err != nil && !errors.Is(err, domain.ErrConfigNotFound) {
		t.Fatalf("GetConfig: %v", err)
	}
}

func TestLoyalty_GetConfig_WrongBusinessReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.GetConfig(context.Background(), 9999999)
	if !errors.Is(err, domain.ErrConfigNotFound) {
		t.Fatalf("expected ErrConfigNotFound for unknown business, got %v", err)
	}
}

func TestLoyalty_UpsertConfig_ThenGetReturnsValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       true,
		PointsPerDollar: 1.5,
		RedemptionRate:  0.01,
		MinRedeemPoints: 100,
	})
	if err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("GetConfig after upsert: %v", err)
	}
	if cfg.BusinessID != fixtureBusinessID {
		t.Errorf("BusinessID = %d, want %d", cfg.BusinessID, fixtureBusinessID)
	}
	if !cfg.IsEnabled {
		t.Error("expected IsEnabled=true after upsert")
	}
}

func TestLoyalty_UpsertConfig_MinRedeemPointsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       true,
		PointsPerDollar: 2.0,
		RedemptionRate:  0.05,
		MinRedeemPoints: 50,
	})
	if err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.MinRedeemPoints != 50 {
		t.Errorf("MinRedeemPoints = %d, want 50", cfg.MinRedeemPoints)
	}
}

func TestLoyalty_GetBalance_ZeroForUnknownCustomer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	balance, err := repo.GetBalance(context.Background(), fixtureBusinessID, 9999999)
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if balance != 0 {
		t.Errorf("expected balance=0 for unknown customer, got %d", balance)
	}
}

func TestLoyalty_ListTransactions_NoErrorForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	txns, err := repo.ListTransactions(context.Background(), fixtureBusinessID, 9999999)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if txns == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestLoyalty_CreateTransaction_ThenBalance(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8801
	note := "integration test earn"

	id, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 10, 10, &note)
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}
	if id == 0 {
		t.Fatal("CreateTransaction returned ID=0")
	}

	balance, err := repo.GetBalance(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if balance != 10 {
		t.Errorf("balance = %d, want 10", balance)
	}
}

func TestLoyalty_CreateTransaction_ListIncludesNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8802
	note := "integration test list"

	id, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 5, 5, &note)
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txns, err := repo.ListTransactions(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	found := false
	for _, tx := range txns {
		if tx.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("new transaction id=%d not found in list", id)
	}
}

func TestLoyalty_GetConfig_HasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Ensure config exists via upsert.
	_ = repo.UpsertConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       false,
		PointsPerDollar: 1.0,
		RedemptionRate:  0.01,
		MinRedeemPoints: 10,
	})

	cfg, err := repo.GetConfig(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.ID == 0 {
		t.Error("expected non-zero ID after upsert")
	}
}

func TestLoyalty_ListTransactions_TransactionsHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8803
	note := "id check"
	_, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 1, 1, &note)
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txns, err := repo.ListTransactions(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	for i, tx := range txns {
		if tx.ID == 0 {
			t.Errorf("txns[%d].ID is zero", i)
		}
	}
}

func TestLoyalty_ListTransactions_TransactionsHaveNonEmptyType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8804
	note := "type check"
	_, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 1, 1, &note)
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txns, err := repo.ListTransactions(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	for i, tx := range txns {
		if tx.Type == "" {
			t.Errorf("txns[%d].Type is empty", i)
		}
	}
}
