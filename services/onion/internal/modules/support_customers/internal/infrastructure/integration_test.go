package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/support_customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// The support_customers repository uses businessID as its tenant key.
// Fixture data: businesses id=1 with owner_id=1001.
const supportBusinessID int64 = 1

func TestList_SupportCustomers_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, total, err := repo.List(ctx, supportBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("support_customers.List: unexpected error: %v", err)
	}

	if customers == nil {
		t.Fatal("support_customers.List: expected non-nil slice, got nil")
	}

	if total < 0 {
		t.Fatalf("support_customers.List: expected non-negative total, got %d", total)
	}
}

func TestGetByID_SupportCustomers_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, supportBusinessID, 0)
	if err == nil {
		t.Fatal("support_customers.GetByID(0): expected error for non-existent customer, got nil")
	}
}
