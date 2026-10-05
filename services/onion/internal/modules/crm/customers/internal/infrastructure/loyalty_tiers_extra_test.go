package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateLoyaltyTier_MinPointsIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "MinPoints Tier",
		MinPoints:  1750,
		Perks:      []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("CreateLoyaltyTier: %v", err)
	}

	tiers, err := repo.ListLoyaltyTiers(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers: %v", err)
	}
	for _, tier := range tiers {
		if tier.ID == id {
			if tier.MinPoints != 1750 {
				t.Errorf("MinPoints = %d, want 1750", tier.MinPoints)
			}
			return
		}
	}
	t.Fatalf("tier id=%d not found", id)
}

func TestCreateLoyaltyTier_DiscountPctIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "DiscountPct Tier",
		DiscountPct: 15.5,
		Perks:       []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("CreateLoyaltyTier: %v", err)
	}

	tiers, err := repo.ListLoyaltyTiers(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers: %v", err)
	}
	for _, tier := range tiers {
		if tier.ID == id {
			if tier.DiscountPct != 15.5 {
				t.Errorf("DiscountPct = %f, want 15.5", tier.DiscountPct)
			}
			return
		}
	}
	t.Fatalf("tier id=%d not found", id)
}

func TestListLoyaltyTiers_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	_, _ = repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Owner Check Tier",
		Perks:      []byte(`[]`),
	})

	tiers, err := repo.ListLoyaltyTiers(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers: %v", err)
	}
	for _, tier := range tiers {
		if tier.OwnerID != testutil.OwnerID {
			t.Errorf("tier %d has owner_id %d, want %d", tier.ID, tier.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListLoyaltyTiers_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	tiers, err := repo.ListLoyaltyTiers(ctx, 0)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers(ownerID=0): %v", err)
	}
	if len(tiers) != 0 {
		t.Errorf("expected 0 tiers for ownerID=0, got %d", len(tiers))
	}
}

func TestUpdateLoyaltyTier_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Before Tier Name",
		Perks:      []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("CreateLoyaltyTier: %v", err)
	}

	if err := repo.UpdateLoyaltyTier(ctx, domain.LoyaltyTierUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "After Tier Name",
		Perks:      []byte(`[]`),
	}); err != nil {
		t.Fatalf("UpdateLoyaltyTier: %v", err)
	}

	tiers, err := repo.ListLoyaltyTiers(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLoyaltyTiers: %v", err)
	}
	for _, tier := range tiers {
		if tier.ID == id {
			if tier.Name != "After Tier Name" {
				t.Errorf("Name = %q, want %q", tier.Name, "After Tier Name")
			}
			return
		}
	}
	t.Fatalf("tier id=%d not found after update", id)
}
