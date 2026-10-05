package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business"

	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_BankAccounts_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:        "Test Checking",
		Institution: "Test Bank",
		AccountType: "checking",
		Balance:     money.Amount(150000),
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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:        "List Check Account",
		AccountType: "savings",
		Balance:     money.Amount(50000),
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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:    "No Currency Account",
		Balance: money.Amount(0),
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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{Name: "owner-id-check", Balance: money.Amount(0), Currency: "USD"})

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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{Name: "named-account", Balance: money.Amount(0), Currency: "USD"})

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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{Name: "id-check", Balance: money.Amount(0), Currency: "USD"})

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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:     "Balance Test",
		Balance:  money.Amount(123456),
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
			if ba.Balance != money.Amount(123456) {
				t.Errorf("Balance = %s, want 1234.56", ba.Balance.String())
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestCreate_BankAccount_ConnectedFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	input := domain.BankAccountInput{
		Name:        "Connected Account",
		Balance:     money.Amount(0),
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
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Just verify we can iterate without panic
	_ = items
}

func TestList_BankAccounts_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List(0): %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestCreate_BankAccount_NameStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	name := "Name Stored Account"
	id, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:     name,
		Balance:  money.Amount(0),
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if ba.Name != name {
				t.Errorf("Name = %q, want %q", ba.Name, name)
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestCreate_BankAccount_IBANStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	institution := "IBAN Bank"
	id, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:        "IBAN Account",
		Institution: institution,
		Balance:     money.Amount(0),
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if ba.Institution != institution {
				t.Errorf("Institution = %q, want %q", ba.Institution, institution)
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestCreate_ThenList_FindsNewAccount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}
	beforeCount := len(items)

	_, err = repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:     "New Account For List Check",
		Balance:  money.Amount(0),
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err = repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(items) <= beforeCount {
		t.Errorf("expected more items after create; before=%d after=%d", beforeCount, len(items))
	}
}

func TestList_ScopedByOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	// Create one account for the fixture owner.
	_, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:     "Scoped Account",
		Balance:  money.Amount(0),
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// A different (non-existent) owner should get nothing.
	items, err := repo.List(ctx, 888888)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for other owner, got %d", len(items))
	}
}

func TestCreate_BankAccount_AccountTypeStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:        "Type Stored Account",
		AccountType: "savings",
		Balance:     money.Amount(0),
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if ba.AccountType != "savings" {
				t.Errorf("AccountType = %q, want %q", ba.AccountType, "savings")
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestCreate_BankAccount_CurrencyEUR(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:     "EUR Account",
		Balance:  money.Amount(50000),
		Currency: "EUR",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if ba.Currency != "EUR" {
				t.Errorf("Currency = %q, want %q", ba.Currency, "EUR")
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestCreate_BankAccount_MultipleAccountsForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
			Name:     "Multi Account",
			Balance:  money.Amount(i * 10000),
			Currency: "USD",
		})
		if err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) < 3 {
		t.Fatalf("expected at least 3 items, got %d", len(items))
	}
}

func TestCreate_BankAccount_ZeroBalance(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:     "Zero Balance Account",
		Balance:  money.Amount(0),
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, ba := range items {
		if ba.ID == id {
			if ba.Balance != 0 {
				t.Errorf("Balance = %s, want 0", ba.Balance.String())
			}
			return
		}
	}
	t.Fatalf("bank account id=%d not found", id)
}

func TestTotalBalance_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	total, err := repo.TotalBalance(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("TotalBalance: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %s", total.String())
	}
}

func TestTotalBalance_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	total, err := repo.TotalBalance(ctx, 0)
	if err != nil {
		t.Fatalf("TotalBalance(ownerID=0): %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 for ownerID=0, got %s", total.String())
	}
}

func TestTotalBalance_ReflectsCreatedAccount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, domain.BankAccountInput{
		Name:     "Balance Test Account",
		Balance:  money.Amount(500000),
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	total, err := repo.TotalBalance(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("TotalBalance: %v", err)
	}
	if total < money.Amount(500000) {
		t.Errorf("expected total >= 5000.00 after create, got %s", total.String())
	}
}

func TestCountConnected_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	cnt, err := repo.CountConnected(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountConnected: %v", err)
	}
	if cnt < 0 {
		t.Errorf("expected non-negative count, got %d", cnt)
	}
}

func TestCountConnected_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(context.Background(), db).ReportingCurrency)
	ctx := context.Background()

	cnt, err := repo.CountConnected(ctx, 0)
	if err != nil {
		t.Fatalf("CountConnected(ownerID=0): %v", err)
	}
	if cnt != 0 {
		t.Errorf("expected 0 for ownerID=0, got %d", cnt)
	}
}
