package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/expenses/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListExpenses_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	expenses, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List expenses error: %v", err)
	}
	// May be empty for fixture data; assert no error and valid return.
	_ = expenses
}

func TestListExpenseCategories_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	cats, err := repo.ListCategories(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories error: %v", err)
	}
	// May be empty for fixture data; assert no error and valid return.
	_ = cats
}
