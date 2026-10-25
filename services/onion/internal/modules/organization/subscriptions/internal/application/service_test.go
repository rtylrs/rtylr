package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/subscriptions/internal/domain"
	paddlePkg "github.com/voxire/onion/internal/platform/paddle"
)

// ─── mock repository ──────────────────────────────────────────────────────────

type mockRepo struct {
	eventNew            bool
	paddleUpdateMissing bool
	plans               []domain.Plan
	plan                *domain.Plan
	planSlug            string
	subscription        *domain.Subscription
	getSubCalls         int
	ownerApps           []string
	billingHistory      []domain.BillingHistoryMonth
	addonPeriods        []domain.AddonBillingPeriod
	businessID          int64
	upsertBusinessID    int64
	upsertPlanID        int64
	upsertTrialDays     int
	upsertSeatCount     int
	activeSeats         int
	listPlansErr        error
	planBySlugErr       error
	getSubErr           error
	upsertSubErr        error
	countSeatsErr       error
	getAppsErr          error
	getPeriodsErr       error
	getAddonPeriodsErr  error
	billingHistErr      error
	paddleUpdateErr     error
	paddleOwnerErr      error
	reconcileCalls      int
	reconcileErr        error
	businessIDs         []int64
	listOwnersErr       error
}

func (m *mockRepo) ListPlans(_ context.Context) ([]domain.Plan, error) {
	return m.plans, m.listPlansErr
}

func (m *mockRepo) GetPlanBySlug(_ context.Context, slug string) (*domain.Plan, error) {
	m.planSlug = slug
	return m.plan, m.planBySlugErr
}

func (m *mockRepo) GetPlanByID(_ context.Context, _ int64) (*domain.Plan, error) {
	return m.plan, m.planBySlugErr
}

func (m *mockRepo) GetSubscription(_ context.Context, _ int64) (*domain.Subscription, error) {
	m.getSubCalls++
	if m.getSubErr != nil {
		if m.getSubCalls == 1 {
			return nil, m.getSubErr
		}
		if errors.Is(m.getSubErr, domain.ErrSubscriptionNotFound) {
			return m.subscription, nil
		}
		return nil, m.getSubErr
	}
	return m.subscription, nil
}

func (m *mockRepo) UpsertSubscription(_ context.Context, businessID, planID int64, trialDays, seatCount int) error {
	m.upsertBusinessID = businessID
	m.upsertPlanID = planID
	m.upsertTrialDays = trialDays
	m.upsertSeatCount = seatCount
	return m.upsertSubErr
}

func (m *mockRepo) GetBusinessApps(_ context.Context, _ int64) ([]string, error) {
	return m.ownerApps, m.getAppsErr
}

func (m *mockRepo) ExpireTrialSubscriptions(_ context.Context) error { return nil }

func (m *mockRepo) CountActiveSeats(_ context.Context, _ int64) (int, error) {
	if m.countSeatsErr != nil {
		return 0, m.countSeatsErr
	}
	if m.activeSeats > 0 {
		return m.activeSeats, nil
	}
	return 1, nil
}

func (m *mockRepo) GetPeriodsForMonth(_ context.Context, _ int64, _ time.Time) ([]domain.SeatPeriod, error) {
	return nil, m.getPeriodsErr
}

func (m *mockRepo) GetAddonPeriodsForMonth(_ context.Context, _ int64, _ time.Time) ([]domain.AddonBillingPeriod, error) {
	return m.addonPeriods, m.getAddonPeriodsErr
}

func (m *mockRepo) ReconcileSeatBilling(_ context.Context, _, _ int64, _, _ int, _ time.Time) error {
	m.reconcileCalls++
	return m.reconcileErr
}

func (m *mockRepo) ListBusinessIDsForReconciliation(_ context.Context, _ int64, _ int) ([]int64, error) {
	return m.businessIDs, m.listOwnersErr
}

func (m *mockRepo) GetBillingHistory(_ context.Context, _ int64, _ int) ([]domain.BillingHistoryMonth, error) {
	return m.billingHistory, m.billingHistErr
}

func (m *mockRepo) UpdatePaddleSubscription(_ context.Context, _ int64, _, _, _ string) (bool, error) {
	return !m.paddleUpdateMissing, m.paddleUpdateErr
}

func (m *mockRepo) ApplyPaddleSubscriptionEvent(_ context.Context, _, _ string, _ time.Time, fallbackBusinessID int64, _, _ string) (bool, int64, error) {
	if !m.eventNew {
		return false, 0, nil
	}
	if m.paddleUpdateErr != nil {
		return false, 0, m.paddleUpdateErr
	}
	if m.paddleUpdateMissing {
		return false, 0, errors.New("no subscription row")
	}
	if m.paddleOwnerErr != nil && fallbackBusinessID == 0 {
		return false, 0, m.paddleOwnerErr
	}
	return m.eventNew, fallbackBusinessID, nil
}

func (m *mockRepo) RecordPaddleEvent(_ context.Context, _, _ string, _ time.Time) (bool, error) {
	return m.eventNew, nil
}

func (m *mockRepo) LastPaddleEventAt(_ context.Context, _ string) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

func (m *mockRepo) GetBusinessIDByPaddleSubscription(_ context.Context, _ string) (int64, error) {
	return m.businessID, m.paddleOwnerErr
}

// ─── tests ─────────────────────────────────────────────────────────────────────

func TestComputeLineItem_FullMonth(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)
	li := computeLineItem(12, 1000, billing, &end, billing)

	// April has 30 days. First 10 seats are included; 2 seats x 30 days x $10 / 30 = $20.
	if li.AmountCents != 2000 {
		t.Errorf("expected 2000 cents, got %d", li.AmountCents)
	}
	if li.SeatCount != 12 {
		t.Errorf("seat count round trip: %d", li.SeatCount)
	}
	if li.IncludedSeatCount != 10 || li.BillableSeatCount != 2 {
		t.Errorf("unexpected seat split: %+v", li)
	}
}

func TestComputeLineItem_PartialMonth(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	li := computeLineItem(12, 3000, start, &end, billing)
	// First 10 seats are included; 2 seats x 15 days x $30 / 30 days = $30.
	if li.AmountCents != 3000 {
		t.Errorf("expected 3000 cents for half-month, got %d", li.AmountCents)
	}
}

func TestComputeAddonLineItem_FullMonth(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	li := computeAddonLineItem(domain.AddonBillingPeriod{
		AddonID:      "email_marketing",
		Name:         "Email Marketing",
		PriceCents:   1500,
		PeriodStart:  billing,
		PeriodEnd:    &end,
		BillingMonth: billing,
	})

	if li.Type != "addon" || li.AddonID != "email_marketing" || li.Description != "Email Marketing" {
		t.Fatalf("unexpected addon line metadata: %+v", li)
	}
	if li.AmountCents != 1500 {
		t.Fatalf("expected full-month addon charge of 1500 cents, got %d", li.AmountCents)
	}
	if li.SeatCount != 0 || li.PricePerSeatCents != 0 {
		t.Fatalf("addon line should not masquerade as seat usage: %+v", li)
	}
}

func TestComputeLineItem_OpenEndedUsesNow(t *testing.T) {
	now := time.Now().UTC()
	billing := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	li := computeLineItem(11, 1000, billing, nil, billing)
	if li.PeriodEnd != nil {
		t.Errorf("nil end should be preserved on the line item, got %v", li.PeriodEnd)
	}
	if li.AmountCents <= 0 {
		t.Errorf("open-ended period should still bill, got %d", li.AmountCents)
	}
}

// A period whose end precedes its start never covered a day, so it must bill
// zero rather than the old one-day floor. The floor charged a phantom day for
// every duplicate open period the reconciler collapses, which double-billed the
// day its replacement already covers.
func TestComputeLineItem_InvertedPeriodBillsNothing(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	// Start AFTER end: no day was ever covered.
	start := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)

	li := computeLineItem(12, 3000, start, &end, billing)
	if li.AmountCents != 0 || li.IncludedCents != 0 {
		t.Fatalf("inverted period billed amount=%d included=%d, want 0", li.AmountCents, li.IncludedCents)
	}
}

// A single-day period is legitimate: start == end covers exactly one day.
func TestComputeLineItem_SingleDayPeriodBillsOneDay(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	day := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	li := computeLineItem(12, 3000, day, &day, billing)
	// 2 billable seats x 1 day x 3000 / 30 = 200.
	if li.AmountCents != 200 {
		t.Fatalf("single-day period billed %d cents, want 200", li.AmountCents)
	}
}

func TestListPlans_Delegates(t *testing.T) {
	plans := []domain.Plan{{ID: 1, Name: "Starter"}, {ID: 2, Name: "Pro"}}
	s := &Service{repo: &mockRepo{plans: plans}}
	got, err := s.ListPlans(context.Background())
	if err != nil || len(got) != 2 {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestListPlans_PropagatesError(t *testing.T) {
	want := errors.New("db fail")
	s := &Service{repo: &mockRepo{listPlansErr: want}}
	if _, err := s.ListPlans(context.Background()); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetMySubscription_PropagatesError(t *testing.T) {
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{getSubErr: want}}
	if _, err := s.GetMySubscription(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetMySubscription_ReturnsSubscription(t *testing.T) {
	sub := &domain.Subscription{ID: 1, PlanName: "Pro", Status: "active"}
	s := &Service{repo: &mockRepo{subscription: sub}}
	got, err := s.GetMySubscription(context.Background(), 1)
	if err != nil || got == nil || got.PlanName != "Pro" {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestGetAIQuotaEntitlementReturnsSubscriptionFacts(t *testing.T) {
	repo := &mockRepo{subscription: &domain.Subscription{
		PlanSlug: "pro", AICreditsPerMonth: 1200, SeatCount: 6, Status: "active",
	}}
	s := &Service{repo: repo}

	slug, credits, seats, status, err := s.GetAIQuotaEntitlement(context.Background(), 123)
	if err != nil {
		t.Fatalf("GetAIQuotaEntitlement: %v", err)
	}
	if slug != "pro" || credits != 1200 || seats != 6 || status != "active" {
		t.Fatalf("entitlement = (%q, %d, %d, %q), want (pro, 1200, 6, active)", slug, credits, seats, status)
	}
}

func TestGetAIQuotaEntitlementUsesDefaultPlan(t *testing.T) {
	repo := &mockRepo{
		getSubErr:   domain.ErrSubscriptionNotFound,
		plan:        &domain.Plan{Slug: "free", Name: "Free", AICreditsPerMonth: 30, Apps: []string{"flow"}},
		activeSeats: 4,
	}
	s := &Service{repo: repo}

	slug, credits, seats, status, err := s.GetAIQuotaEntitlement(context.Background(), 456)
	if err != nil {
		t.Fatalf("GetAIQuotaEntitlement: %v", err)
	}
	if slug != "free" || credits != 30 || seats != 4 || status != "active" {
		t.Fatalf("default entitlement = (%q, %d, %d, %q), want (free, 30, 4, active)", slug, credits, seats, status)
	}
}

func TestGetFreePlanAICredits(t *testing.T) {
	repo := &mockRepo{plan: &domain.Plan{Slug: "free", AICreditsPerMonth: 25}}
	s := &Service{repo: repo}

	credits, err := s.GetFreePlanAICredits(context.Background())
	if err != nil {
		t.Fatalf("GetFreePlanAICredits: %v", err)
	}
	if credits != 25 || repo.planSlug != defaultPlanSlug {
		t.Fatalf("free plan credits = %d (slug %q), want 25 (%q)", credits, repo.planSlug, defaultPlanSlug)
	}
}

// A GET must not create the subscription row. It projects the default plan so
// the response is unchanged, and the background job is what persists it.
func TestGetMySubscription_ProjectsDefaultWithoutWriting(t *testing.T) {
	repo := &mockRepo{
		plan:         &domain.Plan{ID: 42, Slug: "free", TrialDays: 0},
		activeSeats:  4,
		getSubErr:    domain.ErrSubscriptionNotFound,
		reconcileErr: errors.New("a read must never reach the reconciler"),
	}
	s := &Service{repo: repo}

	got, err := s.GetMySubscription(context.Background(), 98765)
	if err != nil {
		t.Fatalf("GetMySubscription: %v", err)
	}
	if got == nil || got.PlanSlug != "free" || got.Status != "active" {
		t.Fatalf("expected active free subscription, got %+v", got)
	}
	if got.SeatCount != 4 {
		t.Fatalf("expected the current active seat count, got %d", got.SeatCount)
	}
	if repo.planSlug != defaultPlanSlug {
		t.Fatalf("expected default plan slug %q, got %q", defaultPlanSlug, repo.planSlug)
	}
	if repo.upsertBusinessID != 0 || repo.reconcileCalls != 0 {
		t.Fatalf("read wrote: upsertOwner=%d reconcileCalls=%d", repo.upsertBusinessID, repo.reconcileCalls)
	}
}

func TestGetBusinessApps_Delegates(t *testing.T) {
	apps := []string{"pos", "erp"}
	s := &Service{repo: &mockRepo{ownerApps: apps}}
	got, err := s.GetBusinessApps(context.Background(), 1)
	if err != nil || len(got) != 2 {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestGetBusinessApps_PropagatesError(t *testing.T) {
	want := errors.New("fail")
	s := &Service{repo: &mockRepo{getAppsErr: want}}
	if _, err := s.GetBusinessApps(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetBillingHistory_Delegates(t *testing.T) {
	history := []domain.BillingHistoryMonth{{TotalCents: 1000}}
	s := &Service{repo: &mockRepo{billingHistory: history}}
	got, err := s.GetBillingHistory(context.Background(), 2, 6)
	if err != nil || len(got) != 1 {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestGetBillingHistory_IncludesAddonCharges(t *testing.T) {
	month := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	s := &Service{repo: &mockRepo{
		billingHistory: []domain.BillingHistoryMonth{{Month: month, TotalCents: 1000}},
		addonPeriods: []domain.AddonBillingPeriod{{
			AddonID:      "custom_domain",
			Name:         "Custom Domain",
			PriceCents:   900,
			PeriodStart:  month,
			PeriodEnd:    &end,
			BillingMonth: month,
		}},
	}}

	got, err := s.GetBillingHistory(context.Background(), 2, 6)
	if err != nil {
		t.Fatalf("GetBillingHistory: %v", err)
	}
	if got[0].TotalCents != 1900 {
		t.Fatalf("expected seat + addon total of 1900 cents, got %d", got[0].TotalCents)
	}
	if len(got[0].LineItems) != 1 || got[0].LineItems[0].Type != "addon" {
		t.Fatalf("expected addon line item in history, got %+v", got[0].LineItems)
	}
}

func TestGetBillingHistory_PropagatesError(t *testing.T) {
	want := errors.New("db fail")
	s := &Service{repo: &mockRepo{billingHistErr: want}}
	if _, err := s.GetBillingHistory(context.Background(), 2, 6); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestHandlePaddleWebhook_ResolvesBusinessFromCustomData(t *testing.T) {
	// When GetBusinessIDByPaddleSubscription fails but custom_data has business_id
	s := &Service{repo: &mockRepo{paddleOwnerErr: errors.New("not found")}}
	data := paddlePkg.SubscriptionData{
		ID:         "sub_123",
		Status:     "active",
		CustomData: map[string]string{"business_id": "42"},
	}
	err := s.HandlePaddleWebhook(context.Background(), paddlePkg.WebhookEvent{EventID: "evt_owner"}, data)
	// The transaction mock treats this as a duplicate after the fallback
	// business was parsed, so the handler acknowledges it without a second write.
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandlePaddleWebhook_FailsWithNoBusinessInfo(t *testing.T) {
	s := &Service{repo: &mockRepo{paddleOwnerErr: errors.New("not found"), eventNew: true}}
	data := paddlePkg.SubscriptionData{
		ID:         "sub_xyz",
		Status:     "active",
		CustomData: map[string]string{},
	}
	if err := s.HandlePaddleWebhook(context.Background(), paddlePkg.WebhookEvent{}, data); err == nil {
		t.Fatal("expected error when business cannot be resolved")
	}
}

func TestComputeLineItem_SingleSeat(t *testing.T) {
	billing := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	// March has 31 days. 1 billable seat x 31 days x 500 / 31 = 500 cents (no free allowance in this unit test).
	li := computeLineItemWithAllowance(1, 500, billing, &end, billing, 0)
	if li.AmountCents != 500 {
		t.Errorf("expected 500 cents, got %d", li.AmountCents)
	}
}

func TestComputeLineItem_PeriodEndPreservedOnResult(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	li := computeLineItem(1, 100, billing, &end, billing)
	if li.PeriodEnd == nil {
		t.Error("expected PeriodEnd to be preserved")
	}
}

func TestHandlePaddleWebhook_DuplicateEventAckedWithoutProcessing(t *testing.T) {
	// A redelivered event_id must short-circuit before any business resolution or
	// status write, so Paddle retries cannot re-apply stale transitions.
	s := &Service{repo: &mockRepo{paddleOwnerErr: errors.New("must not be reached")}}
	data := paddlePkg.SubscriptionData{
		ID:         "sub_dup",
		Status:     "active",
		CustomData: map[string]string{},
	}
	event := paddlePkg.WebhookEvent{EventID: "evt_1", EventType: "subscription.updated"}
	if err := s.HandlePaddleWebhook(context.Background(), event, data); err != nil {
		t.Fatalf("duplicate event should ack cleanly, got %v", err)
	}
}

func TestHandlePaddleWebhook_MissingRowReturnsErrorForRetry(t *testing.T) {
	// RowsAffected==0 must surface as an error so the handler answers 500 and
	// Paddle retries after the lazily-created subscription row exists (#755).
	s := &Service{repo: &mockRepo{
		eventNew:            true,
		paddleUpdateMissing: true,
	}}
	data := paddlePkg.SubscriptionData{
		ID:         "sub_missing",
		Status:     "active",
		CustomData: map[string]string{"business_id": "42"},
	}
	if err := s.HandlePaddleWebhook(context.Background(), paddlePkg.WebhookEvent{EventID: "evt_2"}, data); err == nil {
		t.Fatal("expected error when no subscription row matched")
	}
}

func TestMapPaddleStatus(t *testing.T) {
	cases := map[string]string{
		"active":    "active",
		"trialing":  "trial",
		"past_due":  "past_due",
		"canceled":  "cancelled",
		"paused":    "cancelled",
		"expired":   "expired",
		"scheduled": "",
	}
	for in, want := range cases {
		got, ok := mapPaddleStatus(in)
		if want == "" {
			if ok {
				t.Errorf("mapPaddleStatus(%q) accepted unknown status", in)
			}
			continue
		}
		if !ok || got != want {
			t.Errorf("mapPaddleStatus(%q) = %q,%v; want %q", in, got, ok, want)
		}
	}
}
