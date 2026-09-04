package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateLoyaltyTier_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Diamond",
		MinPoints:   5000,
		DiscountPct: 20.0,
		Perks:       []byte(`[]`),
		SortOrder:   5,
	})
	if err != nil {
		t.Fatalf("CreateLoyaltyTier: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateLoyaltyTier_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Ruby",
		MinPoints:   2000,
		DiscountPct: 10.0,
		Perks:       []byte(`[]`),
		SortOrder:   3,
	})
	if err != nil {
		t.Fatalf("CreateLoyaltyTier: %v", err)
	}

	tiers, err := repo.ListLoyaltyTiers(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers: %v", err)
	}

	found := false
	for _, tier := range tiers {
		if tier.ID == id && tier.Name == "Ruby" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("tier id=%d named 'Ruby' not found in list", id)
	}
}

func TestListLoyaltyTiers_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tiers, err := repo.ListLoyaltyTiers(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers: %v", err)
	}
	if tiers == nil {
		t.Fatal("expected non-nil tiers slice")
	}
}

func TestListLoyaltyTiers_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tiers, err := repo.ListLoyaltyTiers(ctx, 999999)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers wrong owner: %v", err)
	}
	if len(tiers) != 0 {
		t.Fatalf("expected 0 tiers for unknown owner, got %d", len(tiers))
	}
}

func TestUpdateLoyaltyTier_ChangesDiscountPct(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "DiscountUpdate",
		MinPoints:   1000,
		DiscountPct: 5.0,
		Perks:       []byte(`[]`),
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateLoyaltyTier: %v", err)
	}

	if err := repo.UpdateLoyaltyTier(ctx, domain.LoyaltyTierUpdateInput{
		ID:          id,
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "DiscountUpdate",
		MinPoints:   1000,
		DiscountPct: 15.0,
		Perks:       []byte(`[]`),
		SortOrder:   1,
	}); err != nil {
		t.Fatalf("UpdateLoyaltyTier: %v", err)
	}

	tiers, err := repo.ListLoyaltyTiers(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers: %v", err)
	}
	for _, tier := range tiers {
		if tier.ID == id {
			if tier.DiscountPct != 15.0 {
				t.Errorf("DiscountPct = %f, want 15.0", tier.DiscountPct)
			}
			return
		}
	}
	t.Fatalf("tier id=%d not found", id)
}
