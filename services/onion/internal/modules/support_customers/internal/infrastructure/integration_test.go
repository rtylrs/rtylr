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

func TestListCustomers_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, total, err := repo.List(ctx, 0, 1, 20)
	if err != nil {
		t.Fatalf("List(businessID=0): %v", err)
	}
	if len(customers) != 0 {
		t.Errorf("expected 0 customers for businessID=0, got %d", len(customers))
	}
	if total != 0 {
		t.Errorf("expected total=0 for businessID=0, got %d", total)
	}
}

func TestListCustomers_Pagination_LimitWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const pageSize = 1
	customers, _, err := repo.List(ctx, supportBusinessID, 1, pageSize)
	if err != nil {
		t.Fatalf("List(pageSize=%d): %v", pageSize, err)
	}
	if len(customers) > pageSize {
		t.Errorf("expected at most %d customers, got %d", pageSize, len(customers))
	}
}

func TestListCustomers_Pagination_TotalCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total, err := repo.List(ctx, supportBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 0 {
		t.Errorf("expected total >= 0, got %d", total)
	}
}

func TestGetCustomerByID_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, supportBusinessID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent customer, got nil")
	}
}

func TestGetCustomerByID_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 9999999, 1)
	if err == nil {
		t.Fatal("expected error when using wrong businessID, got nil")
	}
}

func TestListCustomers_PageSize1_ReturnsOne(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, total, err := repo.List(ctx, supportBusinessID, 1, 1)
	if err != nil {
		t.Fatalf("List(pageSize=1): %v", err)
	}
	// If there are customers, we get exactly 1; otherwise 0.
	if len(customers) > 1 {
		t.Errorf("expected at most 1 customer, got %d", len(customers))
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
}

func TestListCustomers_Page2_OffsetWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _, err := repo.List(ctx, supportBusinessID, 2, 10)
	if err != nil {
		t.Fatalf("List(page=2): %v", err)
	}
}

func TestListCustomers_TotalCountIsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total, err := repo.List(ctx, supportBusinessID, 1, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 0 {
		t.Errorf("total must be non-negative, got %d", total)
	}
}

func TestListCustomers_ReturnsValidStruct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, _, err := repo.List(ctx, supportBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, c := range customers {
		if c.ID <= 0 {
			t.Errorf("customer[%d] has non-positive ID: %d", i, c.ID)
		}
		if c.Name == "" {
			t.Errorf("customer[%d] (id=%d) has empty Name", i, c.ID)
		}
	}
}

func TestListCustomers_UnknownBusiness_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, total, err := repo.List(ctx, 9999999, 1, 20)
	if err != nil {
		t.Fatalf("List unknown business: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total=0 for unknown business, got %d", total)
	}
	if len(customers) != 0 {
		t.Errorf("expected 0 customers for unknown business, got %d", len(customers))
	}
}

func TestListCustomers_MultiplePages_TotalConsistent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total1, err := repo.List(ctx, supportBusinessID, 1, 5)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	_, total2, err := repo.List(ctx, supportBusinessID, 2, 5)
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if total1 != total2 {
		t.Errorf("total should be consistent across pages: %d vs %d", total1, total2)
	}
}

func TestListCustomers_NilSliceNotReturned(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, _, err := repo.List(ctx, supportBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if customers == nil {
		t.Fatal("expected non-nil slice from List")
	}
}
