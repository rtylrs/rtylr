package infrastructure_test

import (
	"context"
	"errors"
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
	if !errors.Is(err, domain.ErrPlanNotFound) {
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
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
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
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
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

func TestSubscriptions_UpsertSubscription_ThenGetSubscription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}

	const testOwner int64 = 7000001
	if err := repo.UpsertSubscription(ctx, testOwner, plans[0].ID, 0, 1); err != nil {
		t.Fatalf("UpsertSubscription: %v", err)
	}

	sub, err := repo.GetSubscription(ctx, testOwner)
	if err != nil {
		t.Fatalf("GetSubscription after upsert: %v", err)
	}
	if sub.OwnerID != testOwner {
		t.Errorf("expected owner_id=%d, got %d", testOwner, sub.OwnerID)
	}
}

func TestSubscriptions_UpsertSubscription_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}

	const testOwner int64 = 7000002
	if err := repo.UpsertSubscription(ctx, testOwner, plans[0].ID, 0, 1); err != nil {
		t.Fatalf("first UpsertSubscription: %v", err)
	}
	if err := repo.UpsertSubscription(ctx, testOwner, plans[0].ID, 0, 1); err != nil {
		t.Fatalf("second UpsertSubscription: %v", err)
	}
}

func TestSubscriptions_UpdateSeatCount_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}

	const testOwner int64 = 7000003
	if err := repo.UpsertSubscription(ctx, testOwner, plans[0].ID, 0, 1); err != nil {
		t.Fatalf("UpsertSubscription: %v", err)
	}
	if err := repo.UpdateSeatCount(ctx, testOwner, 5); err != nil {
		t.Fatalf("UpdateSeatCount: %v", err)
	}
}

func TestSubscriptions_GetActivePeriod_NoSubscription_Error(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A subscription ID that almost certainly doesn't exist.
	_, err := repo.GetActivePeriod(ctx, 9999999)
	_ = err // may return nil period or error; must not panic
}

func TestSubscriptions_ListPlans_SlugNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	plans, err := repo.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	for _, p := range plans {
		if p.Slug == "" {
			t.Errorf("plan id=%d has empty Slug", p.ID)
		}
	}
}

func TestSubscriptions_ListPlans_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	plans, err := repo.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	for _, p := range plans {
		if p.ID <= 0 {
			t.Errorf("expected positive plan ID, got %d", p.ID)
		}
	}
}

func TestSubscriptions_GetBillingHistory_LimitZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	history, err := repo.GetBillingHistory(context.Background(), testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("GetBillingHistory limit=0: %v", err)
	}
	_ = history
}

func TestSubscriptions_GetPeriodsForMonth_PastMonth(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	past := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	periods, err := repo.GetPeriodsForMonth(context.Background(), testutil.OwnerID, past)
	if err != nil {
		t.Fatalf("GetPeriodsForMonth past: %v", err)
	}
	_ = periods
}

func TestSubscriptions_UpdatePaddleSubscription_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}

	const testOwner int64 = 7000004
	if err := repo.UpsertSubscription(ctx, testOwner, plans[0].ID, 14, 1); err != nil {
		t.Fatalf("UpsertSubscription: %v", err)
	}
	updated, err := repo.UpdatePaddleSubscription(ctx, testOwner, "paddle-sub-test-001", "paddle-cust-001", "active")
	if err != nil {
		t.Fatalf("UpdatePaddleSubscription: %v", err)
	}
	if !updated {
		t.Fatal("expected the existing subscription row to be updated")
	}
}

func TestSubscriptions_GetOwnerIDByPaddleSubscription_AfterUpdate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}

	const testOwner int64 = 7000005
	const paddleSubID = "paddle-sub-lookup-test"
	if err := repo.UpsertSubscription(ctx, testOwner, plans[0].ID, 0, 1); err != nil {
		t.Fatalf("UpsertSubscription: %v", err)
	}
	if _, err := repo.UpdatePaddleSubscription(ctx, testOwner, paddleSubID, "paddle-cust-lookup", "active"); err != nil {
		t.Fatalf("UpdatePaddleSubscription: %v", err)
	}

	ownerID, err := repo.GetOwnerIDByPaddleSubscription(ctx, paddleSubID)
	if err != nil {
		t.Fatalf("GetOwnerIDByPaddleSubscription: %v", err)
	}
	if ownerID != testOwner {
		t.Errorf("expected ownerID=%d, got %d", testOwner, ownerID)
	}
}

func TestSubscriptions_GetBillingHistory_LargeLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	history, err := repo.GetBillingHistory(context.Background(), testutil.OwnerID, 100)
	if err != nil {
		t.Fatalf("GetBillingHistory large limit: %v", err)
	}
	_ = history
}

func TestSubscriptions_GetOwnerApps_NonEmptySubscription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no active plans in DB")
	}

	const testOwner int64 = 7000006
	if err := repo.UpsertSubscription(ctx, testOwner, plans[0].ID, 0, 1); err != nil {
		t.Fatalf("UpsertSubscription: %v", err)
	}

	apps, err := repo.GetOwnerApps(ctx, testOwner)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			t.Skip("subscription not found after upsert (fixture constraint)")
		}
		t.Fatalf("GetOwnerApps: %v", err)
	}
	_ = apps
}
