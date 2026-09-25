package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetDashboardStats_ActiveLoyaltyMembersIsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetDashboardStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.ActiveLoyaltyMembers < 0 {
		t.Errorf("expected non-negative ActiveLoyaltyMembers, got %d", stats.ActiveLoyaltyMembers)
	}
}

func TestGetDashboardStats_NewThisMonthIsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetDashboardStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.NewThisMonth < 0 {
		t.Errorf("expected non-negative NewThisMonth, got %d", stats.NewThisMonth)
	}
}

func TestInsertTimelineEntry_DescriptionIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "Customer placed order #42"
	if err := repo.InsertTimelineEntry(ctx, domain.TimelineEntry{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		CustomerID:  5301,
		Type:        "order",
		Description: want,
	}); err != nil {
		t.Fatalf("InsertTimelineEntry: %v", err)
	}

	entries, err := repo.GetCustomerTimeline(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerTimeline: %v", err)
	}

	found := false
	for _, e := range entries {
		if e.Description == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("inserted timeline entry with description %q not found", want)
	}
}

func TestGetCustomerTimeline_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.GetCustomerTimeline(ctx, 9999999, 5301)
	if err != nil {
		t.Fatalf("GetCustomerTimeline wrong owner: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for wrong owner, got %d", len(entries))
	}
}
