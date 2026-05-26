package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListInvoices_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	invoices, err := repo.ListInvoices(context.Background(), testutil.OwnerID, "", "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListInvoices error: %v", err)
	}
	// Fixture may have no invoices; just assert the call succeeds.
	_ = invoices
}

func TestListFinanceExpenses_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	expenses, err := repo.ListExpenses(context.Background(), testutil.OwnerID, 0, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListExpenses error: %v", err)
	}
	// Fixture may have no expenses; just assert the call succeeds.
	_ = expenses
}

func TestListExpenseCategories_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	cats, err := repo.ListExpenseCategories(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListExpenseCategories error: %v", err)
	}
	_ = cats
}
