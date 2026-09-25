package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateTransaction_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.CreateTransaction(ctx, domain.AwardInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		CustomerID:  5301,
		Points:      75,
		Description: "type check",
	}, "award"); err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txs, _, err := repo.ListTransactions(ctx, testutil.BusinessID, 5301, 1, 50)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}

	found := false
	for _, tx := range txs {
		if tx.Type == "award" && tx.Points == 75 {
			found = true
			break
		}
	}
	if !found {
		t.Error("created transaction not found with matching type and points")
	}
}

func TestCreateTransaction_PointsAreStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.CreateTransaction(ctx, domain.AwardInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5302,
		Points:     333,
	}, "award"); err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txs, _, err := repo.ListTransactions(ctx, testutil.BusinessID, 5302, 1, 100)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}

	found := false
	for _, tx := range txs {
		if tx.Points == 333 {
			found = true
			break
		}
	}
	if !found {
		t.Error("transaction with 333 points not found")
	}
}

func TestCreateRedeemTransaction_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.CreateRedeemTransaction(ctx, domain.RedeemInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		CustomerID:  5301,
		Points:      50,
		Description: "redeem test",
	})
	if err != nil {
		t.Fatalf("CreateRedeemTransaction: %v", err)
	}
}

func TestListTransactions_Pagination_LimitWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Seed a few transactions
	for i := 0; i < 3; i++ {
		_ = repo.CreateTransaction(ctx, domain.AwardInput{
			OwnerID:    testutil.OwnerID,
			BusinessID: testutil.BusinessID,
			CustomerID: 5301,
			Points:     10,
		}, "award")
	}

	txs, _, err := repo.ListTransactions(ctx, testutil.BusinessID, 5301, 1, 2)
	if err != nil {
		t.Fatalf("ListTransactions pageSize=2: %v", err)
	}
	if len(txs) > 2 {
		t.Errorf("expected at most 2 transactions, got %d", len(txs))
	}
}

func TestListTransactions_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_ = repo.CreateTransaction(ctx, domain.AwardInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Points:     20,
	}, "award")

	txs, _, err := repo.ListTransactions(ctx, testutil.BusinessID, 0, 1, 50)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	for _, tx := range txs {
		if tx.OwnerID != testutil.OwnerID {
			t.Errorf("transaction %d has owner_id %d, want %d", tx.ID, tx.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListTransactions_FilterByCustomer_OnlyThatCustomer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_ = repo.CreateTransaction(ctx, domain.AwardInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Points:     10,
	}, "award")

	txs, _, err := repo.ListTransactions(ctx, testutil.BusinessID, 5301, 1, 50)
	if err != nil {
		t.Fatalf("ListTransactions for customer: %v", err)
	}
	for _, tx := range txs {
		if tx.CustomerID != 5301 {
			t.Errorf("transaction %d has customer_id %d, want 5301", tx.ID, tx.CustomerID)
		}
	}
}
