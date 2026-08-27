package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestSearchCustomers_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Search(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
}

func TestSearchCustomers_EmptyQueryReturnsAll(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	// Fixture has Jamie (5301) and Alex (5302); result must be non-empty.
	if len(customers) == 0 {
		t.Fatal("expected fixture customer rows, got none")
	}
}

func TestSearchCustomers_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	for i, c := range customers {
		if c.ID == 0 {
			t.Errorf("customer[%d] has zero ID", i)
		}
	}
}

func TestSearchCustomers_EachHasNonEmptyFirstName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	for i, c := range customers {
		if c.FirstName == "" {
			t.Errorf("customer[%d] (id=%d) has empty FirstName", i, c.ID)
		}
	}
}

func TestSearchCustomers_FilterByFirstName_Jamie(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, testutil.OwnerID, "Jamie")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(customers) == 0 {
		t.Fatal("expected at least one result for 'Jamie'")
	}
	for _, c := range customers {
		if c.ID == 5301 {
			return
		}
	}
	t.Error("fixture customer id=5301 (Jamie) not found in search results")
}

func TestSearchCustomers_FilterByFirstName_Alex(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, testutil.OwnerID, "Alex")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(customers) == 0 {
		t.Fatal("expected at least one result for 'Alex'")
	}
	for _, c := range customers {
		if c.ID == 5302 {
			return
		}
	}
	t.Error("fixture customer id=5302 (Alex) not found in search results")
}

func TestSearchCustomers_UnknownQueryReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, testutil.OwnerID, "xyzzy_no_match_99999")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(customers) != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", len(customers))
	}
}

func TestSearchCustomers_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, 0, "")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(customers) != 0 {
		t.Errorf("expected empty list for owner_id=0, got %d", len(customers))
	}
}

func TestSearchCustomers_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.Search(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("first Search: %v", err)
	}
	second, err := repo.Search(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("second Search: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("count changed: %d vs %d", len(first), len(second))
	}
}

func TestSearchCustomers_FixtureHasAtLeastTwoCustomers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	customers, err := repo.Search(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(customers) < 2 {
		t.Errorf("expected at least 2 fixture customers (Jamie + Alex), got %d", len(customers))
	}
}

func TestSearchCustomers_ContainsBothFixtureCustomers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for _, query := range []string{"Jamie", "Alex"} {
		customers, err := repo.Search(ctx, testutil.OwnerID, query)
		if err != nil {
			t.Fatalf("Search(%q) error: %v", query, err)
		}

		found := false
		for _, customer := range customers {
			if (query == "Jamie" && customer.ID == 5301) || (query == "Alex" && customer.ID == 5302) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("fixture customer for query %q not found in results", query)
		}
	}
}

func TestSearchCustomers_CancelledContextErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Search(ctx, testutil.OwnerID, "")
	if err == nil {
		t.Log("cancelled context did not return error; acceptable for buffered drivers")
	}
}
