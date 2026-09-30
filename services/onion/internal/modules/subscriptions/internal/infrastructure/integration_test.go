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

func TestSubscriptions_ReconcileSeatBilling_SplitsOnSeatChange(t *testing.T) {
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

	const testOwner int64 = 7000011
	plan := paidPlan(plans)
	monthStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := repo.ReconcileSeatBilling(ctx, testOwner, plan.ID, 0, 12, monthStart); err != nil {
		t.Fatalf("ReconcileSeatBilling open: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM seat_periods WHERE owner_id = ?", testOwner) })

	// A seat change on the 16th must leave the 1st-15th on the old count.
	if err := repo.ReconcileSeatBilling(ctx, testOwner, plan.ID, 0, 20, monthStart.AddDate(0, 0, 15)); err != nil {
		t.Fatalf("ReconcileSeatBilling change: %v", err)
	}

	periods, err := repo.GetPeriodsForMonth(ctx, testOwner, monthStart)
	if err != nil {
		t.Fatalf("GetPeriodsForMonth: %v", err)
	}
	if len(periods) != 2 {
		t.Fatalf("expected 2 seat periods in January, got %d: %+v", len(periods), periods)
	}
	if periods[0].SeatCount != 12 || periods[0].PeriodEnd == nil {
		t.Fatalf("first period should be 12 seats and closed: %+v", periods[0])
	}
	if !periods[0].PeriodEnd.Equal(monthStart.AddDate(0, 0, 14)) {
		t.Fatalf("first period ends %v, want %v (day before the change)",
			periods[0].PeriodEnd, monthStart.AddDate(0, 0, 14))
	}
	if periods[1].SeatCount != 20 || periods[1].PeriodEnd != nil {
		t.Fatalf("second period should be 20 seats and open: %+v", periods[1])
	}
	if !periods[1].PeriodStart.Equal(monthStart.AddDate(0, 0, 15)) {
		t.Fatalf("second period starts %v, want %v", periods[1].PeriodStart, monthStart.AddDate(0, 0, 15))
	}
}

// A subscription whose seats and price are unchanged must still get a period for
// the new billing month, otherwise the new month bills zero forever.
func TestSubscriptions_ReconcileSeatBilling_RollsOverMonth(t *testing.T) {
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

	const testOwner int64 = 7000012
	plan := paidPlan(plans)
	january := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := repo.ReconcileSeatBilling(ctx, testOwner, plan.ID, 0, 12, january); err != nil {
		t.Fatalf("ReconcileSeatBilling january: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM seat_periods WHERE owner_id = ?", testOwner) })

	if err := repo.ReconcileSeatBilling(ctx, testOwner, plan.ID, 0, 12, january.AddDate(0, 1, 0)); err != nil {
		t.Fatalf("ReconcileSeatBilling february: %v", err)
	}

	february, err := repo.GetPeriodsForMonth(ctx, testOwner, january.AddDate(0, 1, 0))
	if err != nil {
		t.Fatalf("GetPeriodsForMonth february: %v", err)
	}
	if len(february) != 1 {
		t.Fatalf("expected 1 seat period in February, got %d", len(february))
	}
	if !february[0].PeriodStart.Equal(january.AddDate(0, 1, 0)) {
		t.Fatalf("february period starts %v, want %v", february[0].PeriodStart, january.AddDate(0, 1, 0))
	}

	jan, err := repo.GetPeriodsForMonth(ctx, testOwner, january)
	if err != nil {
		t.Fatalf("GetPeriodsForMonth january: %v", err)
	}
	if len(jan) != 1 || jan[0].PeriodEnd == nil {
		t.Fatalf("january period should be closed by the rollover: %+v", jan)
	}
	if !jan[0].PeriodEnd.Equal(january.AddDate(0, 1, -1)) {
		t.Fatalf("january period ends %v, want %v (last day of the month)",
			jan[0].PeriodEnd, january.AddDate(0, 1, -1))
	}
}

// The reconciler is the only writer of seat_periods and has to converge on one
// open period; a second call with unchanged state must not add rows.
func TestSubscriptions_ReconcileSeatBilling_IsIdempotent(t *testing.T) {
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

	const testOwner int64 = 7000013
	plan := paidPlan(plans)
	asOf := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM seat_periods WHERE owner_id = ?", testOwner) })

	for i := 0; i < 3; i++ {
		if err := repo.ReconcileSeatBilling(ctx, testOwner, plan.ID, 0, 12, asOf); err != nil {
			t.Fatalf("ReconcileSeatBilling run %d: %v", i, err)
		}
	}

	var open int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM seat_periods WHERE owner_id = ? AND period_end IS NULL", testOwner).Scan(&open); err != nil {
		t.Fatalf("count open periods: %v", err)
	}
	if open != 1 {
		t.Fatalf("expected exactly 1 open seat period after 3 reconciles, got %d", open)
	}
}

func paidPlan(plans []domain.Plan) domain.Plan {
	for _, p := range plans {
		if p.PricePerSeatCents > 0 {
			return p
		}
	}
	return plans[0]
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

// The reconciler takes a row lock, so concurrent callers must converge on one
// open period. The old read path had no lock and could leave two, which are then
// both billed.
func TestSubscriptions_ReconcileSeatBilling_ConcurrentCallersLeaveOneOpenPeriod(t *testing.T) {
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

	const testOwner int64 = 7000014
	plan := paidPlan(plans)
	asOf := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM seat_periods WHERE owner_id = ?", testOwner) })

	const callers = 4
	errs := make(chan error, callers)
	for i := 0; i < callers; i++ {
		go func() {
			errs <- repo.ReconcileSeatBilling(ctx, testOwner, plan.ID, 0, 12, asOf)
		}()
	}
	for i := 0; i < callers; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent ReconcileSeatBilling: %v", err)
		}
	}

	var open int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM seat_periods WHERE owner_id = ? AND period_end IS NULL", testOwner).Scan(&open); err != nil {
		t.Fatalf("count open periods: %v", err)
	}
	if open != 1 {
		t.Fatalf("%d concurrent reconciles left %d open periods, want 1", callers, open)
	}
}

// An owner with no subscription row gets one on the default plan without
// disturbing an existing subscription's plan.
func TestSubscriptions_ReconcileSeatBilling_CreatesMissingRowOnly(t *testing.T) {
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

	paid := paidPlan(plans)
	free := plans[len(plans)-1]
	asOf := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM seat_periods WHERE owner_id = ?", testOwnerCreate) })
	if err := repo.ReconcileSeatBilling(ctx, testOwnerCreate, paid.ID, 0, 12, asOf); err != nil {
		t.Fatalf("ReconcileSeatBilling create: %v", err)
	}
	// Reconcile again with a different default plan: an existing owner must
	// keep the plan they are on.
	if err := repo.ReconcileSeatBilling(ctx, testOwnerCreate, free.ID, 0, 3, asOf); err != nil {
		t.Fatalf("ReconcileSeatBilling again: %v", err)
	}

	sub, err := repo.GetSubscription(ctx, testOwnerCreate)
	if err != nil {
		t.Fatalf("GetSubscription: %v", err)
	}
	if sub.PlanID != paid.ID {
		t.Errorf("reconcile changed the plan to %d, want %d", sub.PlanID, paid.ID)
	}
	if sub.SeatCount != 3 {
		t.Errorf("seat count = %d, want 3", sub.SeatCount)
	}
}

const testOwnerCreate int64 = 7000015
