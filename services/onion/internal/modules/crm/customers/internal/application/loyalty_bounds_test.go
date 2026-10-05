package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

func TestAwardAndRedeemRejectPointsPastStorageBoundBeforeRepositoryAccess(t *testing.T) {
	service := &Service{}
	ctx := context.Background()
	points := int(domain.MaxLoyaltyPoints + 1)

	awardErr := service.AwardPoints(ctx, domain.AwardPointsInput{
		OwnerID:    1,
		BusinessID: 2,
		CustomerID: 3,
		Points:     points,
	})
	if !errors.Is(awardErr, domain.ErrLoyaltyPointsOutOfRange) {
		t.Fatalf("AwardPoints error = %v, want ErrLoyaltyPointsOutOfRange", awardErr)
	}

	redeemErr := service.RedeemPoints(ctx, domain.RedeemPointsInput{
		OwnerID:    1,
		BusinessID: 2,
		CustomerID: 3,
		Points:     points,
	})
	if !errors.Is(redeemErr, domain.ErrLoyaltyPointsOutOfRange) {
		t.Fatalf("RedeemPoints error = %v, want ErrLoyaltyPointsOutOfRange", redeemErr)
	}
}

func TestCustomerDerivedCacheInvalidationIncludesDashboard(t *testing.T) {
	ctx := context.Background()
	if err := cacher.InitializeLocalCache(ctx); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}
	businessID := int64(987654321)
	cache := cacher.Of[domain.DashboardStats]()
	key := "crm:dashboard:987654321|domain=v2"
	cache.Put(ctx, key, domain.DashboardStats{TotalCustomers: 1}, cacher.MemoryOnly())

	if _, ok := cache.Get(ctx, key, cacher.MemoryOnly()); !ok {
		t.Fatal("test setup did not populate dashboard cache")
	}

	invalidateCustomerScope(ctx, businessID)

	if _, ok := cache.Get(ctx, key, cacher.MemoryOnly()); ok {
		t.Fatal("dashboard cache survived customer-derived invalidation")
	}
}

func TestNormalizeLoyaltyTierPageRejectsUnrepresentableOffset(t *testing.T) {
	if _, _, err := normalizeLoyaltyTierPage(1<<31, 20); !errors.Is(err, domain.ErrLoyaltyTierPageOutOfRange) {
		t.Fatalf("error = %v, want ErrLoyaltyTierPageOutOfRange", err)
	}
}
