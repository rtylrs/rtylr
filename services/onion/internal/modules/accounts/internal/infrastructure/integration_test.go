package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListAccounts_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	accounts, err := repo.List(context.Background(), testutil.OwnerID)
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
