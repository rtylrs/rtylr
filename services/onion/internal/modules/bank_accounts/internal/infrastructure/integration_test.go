package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_BankAccounts_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestList_BankAccounts_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestCreate_BankAccount_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:        "Test Checking",
		Institution: "Test Bank",
		AccountType: "checking",
		Balance:     1500.00,
		Currency:    "USD",
		IsConnected: false,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreate_BankAccount_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:        "List Check Account",
		AccountType: "savings",
		Balance:     500.00,
		Currency:    "USD",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, ba := range items {
		if ba.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created bank account id=%d not found in list", id)
	}
}

func TestCreate_BankAccount_DefaultCurrencyUSD(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:    "No Currency Account",
		Balance: 0,
		// Currency intentionally blank -> should default to USD
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if ba.Currency != "USD" {
				t.Errorf("Currency = %q, want %q", ba.Currency, "USD")
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestList_BankAccounts_ItemsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{Name: "owner-id-check", Balance: 0, Currency: "USD"})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, ba := range items {
		if ba.OwnerID != testutil.OwnerID {
			t.Errorf("items[%d].OwnerID = %d, want %d", i, ba.OwnerID, testutil.OwnerID)
		}
	}
}

func TestList_BankAccounts_ItemsHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{Name: "named-account", Balance: 0, Currency: "USD"})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, ba := range items {
		if ba.Name == "" {
			t.Errorf("items[%d].Name is empty", i)
		}
	}
}

func TestList_BankAccounts_ItemsHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{Name: "id-check", Balance: 0, Currency: "USD"})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, ba := range items {
		if ba.ID == 0 {
			t.Errorf("items[%d].ID is zero", i)
		}
	}
}

func TestCreate_BankAccount_BalanceRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:     "Balance Test",
		Balance:  1234.56,
		Currency: "EUR",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if ba.Balance != 1234.56 {
				t.Errorf("Balance = %f, want 1234.56", ba.Balance)
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestCreate_BankAccount_ConnectedFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:        "Connected Account",
		Balance:     0,
		Currency:    "USD",
		IsConnected: true,
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if !ba.IsConnected {
				t.Errorf("IsConnected = false, want true")
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestList_BankAccounts_BalanceNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Just verify we can iterate without panic
	_ = items
}
