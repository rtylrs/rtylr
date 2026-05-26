package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestSubscriptions_ListPlans_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	plans, err := repo.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	_ = plans
}

func TestSubscriptions_ListPlans_AllActive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	plans, err := repo.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	for _, p := range plans {
		if !p.IsActive {
			t.Errorf("plan %s should be active but is not", p.Slug)
		}
	}
}

func TestSubscriptions_GetPlanBySlug_KnownPlan(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	plans, err := repo.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}
	slug := plans[0].Slug
	p, err := repo.GetPlanBySlug(context.Background(), slug)
	if err != nil {
		t.Fatalf("GetPlanBySlug(%s): %v", slug, err)
	}
	if p.Slug != slug {
		t.Errorf("expected slug %s, got %s", slug, p.Slug)
	}
}

func TestSubscriptions_GetPlanBySlug_Unknown_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetPlanBySlug(context.Background(), "nonexistent-plan-slug")
	if err == nil {
		t.Fatal("expected error for unknown plan slug, got nil")
	}
	if err != domain.ErrPlanNotFound {
		t.Errorf("expected ErrPlanNotFound, got %v", err)
	}
}

func TestSubscriptions_GetPlanByID_FirstPlan(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	plans, err := repo.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}
	id := plans[0].ID
	p, err := repo.GetPlanByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetPlanByID(%d): %v", id, err)
	}
	if p.ID != id {
		t.Errorf("expected id %d, got %d", id, p.ID)
	}
}

func TestSubscriptions_GetPlanByID_Unknown_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetPlanByID(context.Background(), 9999999)
	if err == nil {
		t.Fatal("expected error for unknown plan ID")
	}
}

func TestSubscriptions_GetSubscription_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	sub, err := repo.GetSubscription(context.Background(), testutil.OwnerID)
	if err != nil {
		if err == domain.ErrSubscriptionNotFound {
			t.Skip("fixture owner has no subscription")
		}
		t.Fatalf("GetSubscription: %v", err)
	}
	if sub.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, sub.OwnerID)
	}
}

func TestSubscriptions_GetSubscription_Unknown_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetSubscription(context.Background(), 9999999)
	if err == nil {
		t.Fatal("expected ErrSubscriptionNotFound for unknown owner")
	}
}

func TestSubscriptions_GetOwnerApps_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	apps, err := repo.GetOwnerApps(context.Background(), testutil.OwnerID)
	if err != nil {
		if err == domain.ErrSubscriptionNotFound {
			t.Skip("fixture owner has no subscription")
		}
		t.Fatalf("GetOwnerApps: %v", err)
	}
	_ = apps
}

func TestSubscriptions_GetPeriodsForMonth_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	periods, err := repo.GetPeriodsForMonth(context.Background(), testutil.OwnerID, time.Now())
	if err != nil {
		t.Fatalf("GetPeriodsForMonth: %v", err)
	}
	_ = periods
}

func TestSubscriptions_GetBillingHistory_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	history, err := repo.GetBillingHistory(context.Background(), testutil.OwnerID, 12)
	if err != nil {
		t.Fatalf("GetBillingHistory: %v", err)
	}
	_ = history
}

func TestSubscriptions_GetOwnerIDByPaddleSubscription_Unknown(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetOwnerIDByPaddleSubscription(context.Background(), "nonexistent-paddle-sub-id")
	if err == nil {
		t.Fatal("expected error for unknown paddle subscription ID")
	}
}
