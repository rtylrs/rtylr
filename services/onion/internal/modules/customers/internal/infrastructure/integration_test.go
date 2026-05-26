package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListCustomers_ReturnsJamieAndAlex(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	customers, total, err := repo.ListCustomers(context.Background(), domain.CustomerListFilter{
		OwnerID:  testutil.OwnerID,
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("ListCustomers error: %v", err)
	}
	if total == 0 {
		t.Fatal("expected non-zero total, got 0")
	}
	if len(customers) == 0 {
		t.Fatal("expected at least one customer row, got none")
	}

	names := make(map[string]bool)
	for _, c := range customers {
		names[c.FirstName] = true
	}

	if !names["Jamie"] {
		t.Errorf("expected to find customer 'Jamie', got names: %v", names)
	}
	if !names["Alex"] {
		t.Errorf("expected to find customer 'Alex', got names: %v", names)
	}
}

func TestListCustomers_AlexHasGoldTier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	customers, _, err := repo.ListCustomers(context.Background(), domain.CustomerListFilter{
		OwnerID:  testutil.OwnerID,
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("ListCustomers error: %v", err)
	}

	for _, c := range customers {
		if c.FirstName == "Alex" {
			if c.LoyaltyTier != "gold" {
				t.Errorf("expected Alex to have loyalty_tier='gold', got %q", c.LoyaltyTier)
			}
			return
		}
	}
	t.Fatal("customer 'Alex' not found in results")
}

func TestListCustomers_SearchGold_FindsAlex(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, total, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:  testutil.OwnerID,
		Search:   "Gold",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListCustomers search 'Gold': unexpected error: %v", err)
	}
	if total == 0 || len(customers) == 0 {
		t.Fatal("ListCustomers search 'Gold': expected at least one result")
	}

	found := false
	for _, c := range customers {
		if c.ID == 5302 {
			found = true
		}
	}
	if !found {
		t.Error("ListCustomers search 'Gold': expected customer id=5302 (Alex Gold) in results")
	}
}

func TestGetCustomerByID_AlexHasGoldTierAndHighSpend(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5302)
	if err != nil {
		t.Fatalf("GetCustomerByID(5302): unexpected error: %v", err)
	}

	if c.LoyaltyTier != "gold" {
		t.Errorf("GetCustomerByID(5302): expected loyalty_tier='gold', got %q", c.LoyaltyTier)
	}

	if c.TotalSpent < 2500 {
		t.Errorf("GetCustomerByID(5302): expected total_spent >= 2500, got %v", c.TotalSpent)
	}
}

func TestGetCustomerByID_JamieBronze(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerByID(5301): unexpected error: %v", err)
	}

	if c.LoyaltyTier != "bronze" {
		t.Errorf("GetCustomerByID(5301): expected loyalty_tier='bronze', got %q", c.LoyaltyTier)
	}

	if c.Email != "jamie@test.local" {
		t.Errorf("GetCustomerByID(5301): expected email='jamie@test.local', got %q", c.Email)
	}
}
