package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestRound3BankAccountRoundTripsMaskAndSeparatesPrimaryFromConnection(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(987654301)

	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM finance_bank_accounts WHERE business_id = ?`, businessID)
	}
	t.Cleanup(cleanup)

	id, err := repo.Create(ctx, businessID, domain.BankAccountInput{
		Name:        "round3-mask-account",
		LastFour:    "0123",
		Currency:    "USD",
		Balance:     money.Amount(10000),
		IsPrimary:   true,
		IsConnected: false,
	})
	if err != nil {
		t.Fatalf("create bank account: %v", err)
	}

	account, err := repo.GetByID(ctx, businessID, id)
	if err != nil {
		t.Fatalf("get bank account: %v", err)
	}
	if account.LastFour != "0123" {
		t.Fatalf("LastFour = %q, want 0123", account.LastFour)
	}
	if !account.IsPrimary {
		t.Fatal("IsPrimary = false, want true")
	}
	if account.IsConnected {
		t.Fatal("IsConnected = true, want the independent connection state false")
	}
}

func TestRound3BankAccountListPageHonorsPageAndSort(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(987654302)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM finance_bank_accounts WHERE business_id = ?`, businessID)
	})

	for i, balance := range []money.Amount{30000, 10000, 20000} {
		_, err := repo.Create(ctx, businessID, domain.BankAccountInput{
			Name:     []string{"round3-c", "round3-a", "round3-b"}[i],
			Currency: "USD",
			Balance:  balance,
		})
		if err != nil {
			t.Fatalf("create account %d: %v", i, err)
		}
	}

	accounts, total, err := repo.ListPage(ctx, businessID, 2, 1, "balance", "desc")
	if err != nil {
		t.Fatalf("list page: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(accounts) != 1 {
		t.Fatalf("returned %d accounts, want 1", len(accounts))
	}
	if accounts[0].Name != "round3-b" {
		t.Fatalf("page 2 account = %q, want round3-b", accounts[0].Name)
	}
}

func TestRound3BankAccountTotalUsesBusinessCurrency(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(987654303)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO businesses (id, name, owner_id, currency)
		VALUES (?, ?, ?, 'EUR')`, businessID, "round3-bank-currency", testutil.OwnerID); err != nil {
		t.Fatalf("seed business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM finance_bank_accounts WHERE business_id = ?`, businessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM businesses WHERE id = ?`, businessID)
	})

	for _, input := range []domain.BankAccountInput{
		{Name: "round3-usd", Currency: "USD", Balance: money.Amount(10000)},
		{Name: "round3-eur", Currency: "EUR", Balance: money.Amount(25000)},
	} {
		if _, err := repo.Create(ctx, businessID, input); err != nil {
			t.Fatalf("create %s: %v", input.Currency, err)
		}
	}

	got, err := repo.TotalBalance(ctx, businessID)
	if err != nil {
		t.Fatalf("total balance: %v", err)
	}
	if got != money.Amount(25000) {
		t.Fatalf("total balance = %s, want 250.00 EUR", got.String())
	}
}
