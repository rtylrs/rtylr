package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func TestLoyalty_UpsertConfig_DisableThenEnable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       false,
		PointsPerDollar: 1.0,
		RedemptionRate:  0.01,
		MinRedeemPoints: 10,
	})
	if err != nil {
		t.Fatalf("UpsertConfig disable: %v", err)
	}

	err = repo.UpsertConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       true,
		PointsPerDollar: 1.0,
		RedemptionRate:  0.01,
		MinRedeemPoints: 10,
	})
	if err != nil {
		t.Fatalf("UpsertConfig enable: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !cfg.IsEnabled {
		t.Error("expected IsEnabled=true after re-enable")
	}
}

func TestLoyalty_CreateTransaction_WithOrderID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8810
	orderID := int64(5001)
	note := "order earn"

	id, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, &orderID, "earn", 20, 20, &note)
	if err != nil {
		t.Fatalf("CreateTransaction with orderID: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero transaction ID")
	}

	txns, err := repo.ListTransactions(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	found := false
	for _, tx := range txns {
		if tx.ID == id {
			found = true
			if tx.OrderID == nil || *tx.OrderID != orderID {
				t.Errorf("expected OrderID=%d, got %v", orderID, tx.OrderID)
			}
			break
		}
	}
	if !found {
		t.Errorf("transaction id=%d not in list", id)
	}
}

func TestLoyalty_CreateTransaction_WithNilNote(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8811

	id, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 5, 5, nil)
	if err != nil {
		t.Fatalf("CreateTransaction with nil note: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero transaction ID")
	}
}

func TestLoyalty_GetBalance_AfterMultipleEarns(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8812
	note := "multi earn"

	_, _ = repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 10, 10, &note)
	_, _ = repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 5, 15, &note)

	balance, err := repo.GetBalance(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if balance <= 0 {
		t.Errorf("expected positive balance, got %d", balance)
	}
}

func TestLoyalty_UpsertConfig_PointsPerDollarUpdated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       true,
		PointsPerDollar: 3.0,
		RedemptionRate:  0.01,
		MinRedeemPoints: 10,
	})
	if err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.PointsPerDollar != 3.0 {
		t.Errorf("PointsPerDollar = %f, want 3.0", cfg.PointsPerDollar)
	}
}

func TestLoyalty_ListTransactions_BusinessIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8813
	note := "biz id check"
	_, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 7, 7, &note)
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txns, err := repo.ListTransactions(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	for i, tx := range txns {
		if tx.BusinessID != fixtureBusinessID {
			t.Errorf("txns[%d].BusinessID = %d, want %d", i, tx.BusinessID, fixtureBusinessID)
		}
	}
}

func TestLoyalty_ListTransactions_CustomerIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8814
	note := "cust id check"
	_, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 3, 3, &note)
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txns, err := repo.ListTransactions(ctx, fixtureBusinessID, customerID)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	for i, tx := range txns {
		if tx.CustomerID != customerID {
			t.Errorf("txns[%d].CustomerID = %d, want %d", i, tx.CustomerID, customerID)
		}
	}
}

func TestLoyalty_CreateTransaction_RedeemType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8815
	earnNote := "redeem test earn"
	_, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "earn", 50, 50, &earnNote)
	if err != nil {
		t.Fatalf("CreateTransaction earn: %v", err)
	}

	redeemNote := "redeem test redeem"
	id, err := repo.CreateTransaction(ctx, fixtureBusinessID, customerID, nil, "redeem", -30, 20, &redeemNote)
	if err != nil {
		t.Fatalf("CreateTransaction redeem: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero transaction ID for redeem")
	}
}

func TestLoyalty_GetConfig_RedemptionRateField(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       true,
		PointsPerDollar: 1.0,
		RedemptionRate:  0.02,
		MinRedeemPoints: 10,
	})
	if err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.RedemptionRate != 0.02 {
		t.Errorf("RedemptionRate = %f, want 0.02", cfg.RedemptionRate)
	}
}

func TestLoyalty_UpsertConfig_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.UpdateLoyaltyConfigInput{
		BusinessID:      fixtureBusinessID,
		IsEnabled:       true,
		PointsPerDollar: 1.5,
		RedemptionRate:  0.01,
		MinRedeemPoints: 25,
	}
	if err := repo.UpsertConfig(ctx, input); err != nil {
		t.Fatalf("UpsertConfig first: %v", err)
	}
	if err := repo.UpsertConfig(ctx, input); err != nil {
		t.Fatalf("UpsertConfig second: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.MinRedeemPoints != 25 {
		t.Errorf("MinRedeemPoints = %d, want 25", cfg.MinRedeemPoints)
	}
}

// Regression: the deferred RELEASE_LOCK ran after tx.Commit(), where
// database/sql returns ErrTxDone without sending the statement, so the MySQL
// named lock stayed on the pooled connection and a concurrent earn/redeem for
// the same customer could fail on a 5s GET_LOCK timeout.
func TestLoyalty_CreateBalanceTransaction_ReleasesAdvisoryLock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const customerID int64 = 8899
	orderID := int64(77001)
	if _, err := repo.CreateBalanceTransaction(ctx, fixtureBusinessID, customerID, &orderID, "earn", 10, "lock release test"); err != nil {
		t.Fatalf("CreateBalanceTransaction: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_loyalty_transactions WHERE business_id = ? AND customer_id = ?`, fixtureBusinessID, customerID)
	})

	assertAdvisoryLockReleased(t, db, fmt.Sprintf("pos_loyalty:%d:%d", fixtureBusinessID, customerID))
}

func assertAdvisoryLockReleased(t *testing.T, db *sql.DB, lockName string) {
	t.Helper()
	var holder sql.NullInt64
	if err := db.QueryRowContext(context.Background(), "SELECT IS_USED_LOCK(?)", lockName).Scan(&holder); err != nil {
		t.Fatalf("check advisory lock %q: %v", lockName, err)
	}
	if holder.Valid {
		t.Fatalf("advisory lock %q is still held by connection %d", lockName, holder.Int64)
	}
}
