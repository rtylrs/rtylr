package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateCustomer_IsActiveByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "NewActive",
		LastName:   "User",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if !c.IsActive {
		t.Error("expected new customer to be active by default")
	}
}

func TestCreateCustomer_NotesIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "VIP customer, handle with care"
	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "NotesCustomer",
		Notes:      want,
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.Notes != want {
		t.Errorf("Notes = %q, want %q", c.Notes, want)
	}
}

func TestCreateCustomer_CityIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "CityCustomer",
		City:       "New York",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.City != "New York" {
		t.Errorf("City = %q, want %q", c.City, "New York")
	}
}

func TestCreateCustomer_CountryIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "CountryCustomer",
		Country:    "US",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.Country != "US" {
		t.Errorf("Country = %q, want %q", c.Country, "US")
	}
}

func TestListCustomers_SearchByEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "SearchEmail",
		Email:      "searchme@unique-test.example",
	})

	results, total, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Search:     "searchme@unique-test.example",
		Page:       1,
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("ListCustomers with search: %v", err)
	}
	if total == 0 || len(results) == 0 {
		t.Error("expected at least one result for email search")
	}
}

func TestGetCustomerByID_LoyaltyPointsIsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "LoyaltyCheck",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID: %v", err)
	}
	if c.LoyaltyPoints < 0 {
		t.Errorf("LoyaltyPoints = %d, expected non-negative", c.LoyaltyPoints)
	}
}
