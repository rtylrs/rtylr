package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpsertReward_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Free Dessert",
		Category:   "food",
		PointsCost: 300,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpsertReward: %v", err)
	}

	rewards, err := repo.ListRewards(ctx, testutil.BusinessID, false)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	for _, r := range rewards {
		if r.ID == id {
			if r.Name != "Free Dessert" {
				t.Errorf("Name = %q, want %q", r.Name, "Free Dessert")
			}
			return
		}
	}
	t.Fatalf("reward id=%d not found in list", id)
}

func TestUpsertReward_PointsCostIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Points Cost Reward",
		Category:   "general",
		PointsCost: 750,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpsertReward: %v", err)
	}

	rewards, err := repo.ListRewards(ctx, testutil.BusinessID, false)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	for _, r := range rewards {
		if r.ID == id {
			if r.PointsCost != 750 {
				t.Errorf("PointsCost = %d, want 750", r.PointsCost)
			}
			return
		}
	}
	t.Fatalf("reward id=%d not found", id)
}

func TestUpsertReward_InactiveNotReturnedByActiveOnly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Inactive Reward",
		Category:   "general",
		PointsCost: 100,
		Active:     false,
	})
	if err != nil {
		t.Fatalf("UpsertReward: %v", err)
	}

	active, err := repo.ListRewards(ctx, testutil.BusinessID, true)
	if err != nil {
		t.Fatalf("ListRewards activeOnly=true: %v", err)
	}
	for _, r := range active {
		if r.ID == id {
			t.Errorf("inactive reward id=%d should not appear in activeOnly list", id)
		}
	}
}

func TestUpsertReward_UpdateChangesPointsCost(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Changeable Cost Reward",
		Category:   "general",
		PointsCost: 100,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpsertReward create: %v", err)
	}

	_, err = repo.UpsertReward(ctx, domain.LoyaltyReward{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Changeable Cost Reward",
		Category:   "general",
		PointsCost: 500,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpsertReward update: %v", err)
	}

	rewards, err := repo.ListRewards(ctx, testutil.BusinessID, false)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	for _, r := range rewards {
		if r.ID == id {
			if r.PointsCost != 500 {
				t.Errorf("PointsCost after update = %d, want 500", r.PointsCost)
			}
			return
		}
	}
	t.Fatalf("reward id=%d not found", id)
}

func TestListRewards_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Owner Check Reward",
		Category:   "general",
		PointsCost: 50,
		Active:     true,
	})

	rewards, err := repo.ListRewards(ctx, testutil.BusinessID, false)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	for _, r := range rewards {
		if r.OwnerID != testutil.OwnerID {
			t.Errorf("reward %d has owner_id %d, want %d", r.ID, r.OwnerID, testutil.OwnerID)
		}
	}
}
