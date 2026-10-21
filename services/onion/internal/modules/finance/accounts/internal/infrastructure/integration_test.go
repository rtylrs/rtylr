package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListAccounts_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	accounts, err := repo.List(context.Background(), testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List accounts error: %v", err)
	}
	// May be empty for a fresh fixture, just assert the call succeeds.
	_ = accounts
}

func TestListAccountsForTree_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	accounts, err := repo.ListForTree(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListForTree accounts error: %v", err)
	}
	_ = accounts
}

func TestListAccounts_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	accounts, err := repo.List(ctx, 0, 100, 0)
	if err != nil {
		t.Fatalf("List(ownerID=0): %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("expected 0 accounts for ownerID=0, got %d", len(accounts))
	}
}

func TestListAccounts_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	accounts, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if accounts == nil {
		t.Fatal("expected non-nil slice from List")
	}
}

func TestListForTree_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	accounts, err := repo.ListForTree(ctx, 0)
	if err != nil {
		t.Fatalf("ListForTree(ownerID=0): %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("expected 0 accounts for ownerID=0, got %d", len(accounts))
	}
}

func TestListForTree_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	accounts, err := repo.ListForTree(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListForTree: %v", err)
	}
	if accounts == nil {
		t.Fatal("expected non-nil slice from ListForTree")
	}
}

func TestCreateAccount_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "Test Account",
		Type:     "asset",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateAccount_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "Named Account",
		Type:     "asset",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	accounts, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			if a.Name != "Named Account" {
				t.Errorf("Name = %q, want %q", a.Name, "Named Account")
			}
			return
		}
	}
	t.Fatalf("account id=%d not found", id)
}

func TestCreateAccount_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "Liability Account",
		Type:     "liability",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	accounts, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			if a.Type != "liability" {
				t.Errorf("Type = %q, want %q", a.Type, "liability")
			}
			return
		}
	}
	t.Fatalf("account id=%d not found", id)
}

func TestCreateAccount_CodeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "Code Account",
		Type:     "asset",
		Currency: "USD",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	accounts, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			// Currency round-trip check
			if a.Currency != "USD" {
				t.Errorf("Currency = %q, want %q", a.Currency, "USD")
			}
			return
		}
	}
	t.Fatalf("account id=%d not found", id)
}

func TestUpdateAccount_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "Before Update Account",
		Type:     "asset",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.AccountInput{
		Name:     "After Update Account",
		Type:     "asset",
		IsActive: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	accounts, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List after update: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			if a.Name != "After Update Account" {
				t.Errorf("Name = %q, want %q", a.Name, "After Update Account")
			}
			return
		}
	}
	t.Fatalf("account id=%d not found after update", id)
}

func TestDeleteAccount_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "Delete Me Account",
		Type:     "asset",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	accounts, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			t.Errorf("deleted account id=%d still present in list", id)
		}
	}
}

func TestCreate_ThenList_FindsAccount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "FindMe Account",
		Type:     "expense",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	accounts, err := repo.List(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			return
		}
	}
	t.Fatalf("account id=%d not found in list", id)
}

func TestListForTree_AfterCreate_ContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.AccountInput{
		Name:     "Tree Account",
		Type:     "asset",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	accounts, err := repo.ListForTree(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListForTree: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			return
		}
	}
	t.Fatalf("new account id=%d not found in tree list", id)
}
