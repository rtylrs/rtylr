package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateCustomer_PhoneIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "Phone",
		LastName:   "Test",
		Phone:      "+15551234567",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.Phone != "+15551234567" {
		t.Errorf("Phone = %q, want %q", c.Phone, "+15551234567")
	}
}

func TestCreateCustomer_FirstNameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "Cassidy",
		LastName:   "Rivers",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.FirstName != "Cassidy" {
		t.Errorf("FirstName = %q, want %q", c.FirstName, "Cassidy")
	}
}

func TestCreateCustomer_OwnerIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "Owner",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", c.OwnerID, testutil.OwnerID)
	}
}

func TestCreateCustomer_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.CreateCustomer(ctx, domain.CustomerInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, FirstName: "UniqueA"})
	if err != nil {
		t.Fatalf("CreateCustomer A: %v", err)
	}
	id2, err := repo.CreateCustomer(ctx, domain.CustomerInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, FirstName: "UniqueB"})
	if err != nil {
		t.Fatalf("CreateCustomer B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestUpdateCustomer_ChangesFirstName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "BeforeUpdate",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	if err := repo.UpdateCustomer(ctx, domain.CustomerUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "AfterUpdate",
	}); err != nil {
		t.Fatalf("UpdateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.FirstName != "AfterUpdate" {
		t.Errorf("FirstName = %q, want %q", c.FirstName, "AfterUpdate")
	}
}

func TestUpdateCustomer_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "WrongOwner",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	err = repo.UpdateCustomer(ctx, domain.CustomerUpdateInput{
		ID:        id,
		OwnerID:   999999,
		FirstName: "Sneaky",
	})
	if err == nil {
		t.Fatal("expected error updating customer with wrong ownerID, got nil")
	}
}

func TestDeleteCustomer_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "DeleteWrongOwner",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	err = repo.DeleteCustomer(ctx, 999999, id)
	if err == nil {
		t.Fatal("expected error deleting customer with wrong ownerID, got nil")
	}
}

func TestListCustomers_Pagination_Page2(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	all, _, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Page:       1,
		PageSize:   100,
	})
	if err != nil {
		t.Fatalf("ListCustomers page1: %v", err)
	}

	if len(all) < 2 {
		t.Skip("fewer than 2 customers; skipping pagination test")
	}

	page2, _, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Page:       2,
		PageSize:   1,
	})
	if err != nil {
		t.Fatalf("ListCustomers page2: %v", err)
	}
	if len(page2) > 1 {
		t.Errorf("expected at most 1 customer on page2, got %d", len(page2))
	}
}

func TestListCustomers_FilterByGoldTier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, _, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Tier:       "gold",
		Page:       1,
		PageSize:   50,
	})
	if err != nil {
		t.Fatalf("ListCustomers tier=gold: %v", err)
	}
	for _, c := range customers {
		if c.LoyaltyTier != "gold" {
			t.Errorf("customer %d has tier %q, want gold", c.ID, c.LoyaltyTier)
		}
	}
}
