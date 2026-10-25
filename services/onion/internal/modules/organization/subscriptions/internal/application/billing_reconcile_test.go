package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/subscriptions/internal/domain"
)

// statefulRepo models the seat_periods ledger the way the MySQL schema does:
// GetPeriodsForMonth only returns rows whose billing_month matches, which is
// what makes a missing month rollover visible as an empty invoice. Writes are
// counted so a test can assert that a read stayed side-effect free.
type statefulRepo struct {
	mu sync.Mutex

	plan        *domain.Plan
	freePlan    *domain.Plan
	sub         *domain.Subscription
	seats       int
	periods     []domain.SeatPeriod
	businessIDs []int64
	nextID      int64
	writes      int
	upserts     int
	closings    int
	opens       int
	seatSets    int
	reconcile   int
}

func (r *statefulRepo) write() { r.writes++ }

func (r *statefulRepo) ListPlans(context.Context) ([]domain.Plan, error) {
	return []domain.Plan{*r.plan}, nil
}

func (r *statefulRepo) GetPlanBySlug(_ context.Context, slug string) (*domain.Plan, error) {
	for _, p := range []*domain.Plan{r.plan, r.freePlan} {
		if p != nil && p.Slug == slug {
			return p, nil
		}
	}
	return nil, domain.ErrPlanNotFound
}

func (r *statefulRepo) GetPlanByID(_ context.Context, _ int64) (*domain.Plan, error) {
	return r.plan, nil
}

func (r *statefulRepo) GetSubscription(_ context.Context, businessID int64) (*domain.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sub == nil || r.sub.BusinessID != businessID {
		return nil, domain.ErrSubscriptionNotFound
	}
	copied := *r.sub
	return &copied, nil
}

func (r *statefulRepo) UpsertSubscription(_ context.Context, businessID, planID int64, _ int, seatCount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.upserts++
	r.write()
	if r.sub == nil {
		r.sub = &domain.Subscription{
			ID:         1,
			BusinessID: businessID,
			PlanID:     planID,
		}
	}
	r.sub.PlanID = planID
	r.sub.PlanName = r.plan.Name
	r.sub.PlanSlug = r.plan.Slug
	r.sub.Apps = r.plan.Apps
	r.sub.PricePerSeatCents = r.plan.PricePerSeatCents
	r.sub.SeatCount = seatCount
	if r.sub.Status == "" {
		r.sub.Status = "active"
	}
	return nil
}

func (r *statefulRepo) GetBusinessApps(_ context.Context, businessID int64) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sub == nil || r.sub.BusinessID != businessID {
		return nil, domain.ErrSubscriptionNotFound
	}
	return r.sub.Apps, nil
}

func (r *statefulRepo) ExpireTrialSubscriptions(context.Context) error { return nil }

func (r *statefulRepo) CountActiveSeats(context.Context, int64) (int, error) {
	if r.seats < 1 {
		return 1, nil
	}
	return r.seats, nil
}

// ReconcileSeatBilling applies domain.PlanSeatLedger against the in-memory
// ledger, so the service tests exercise the real policy rather than a
// reimplementation of it.
func (r *statefulRepo) ReconcileSeatBilling(_ context.Context, businessID, defaultPlanID int64, defaultTrialDays, seatCount int, asOf time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if seatCount < 1 {
		seatCount = 1
	}
	today := domain.UTCDay(asOf)
	created := false
	if r.sub == nil || r.sub.BusinessID != businessID {
		def := r.plan
		for _, p := range []*domain.Plan{r.plan, r.freePlan} {
			if p != nil && p.ID == defaultPlanID {
				def = p
			}
		}
		status := "active"
		if defaultTrialDays > 0 {
			status = "trial"
		}
		r.sub = &domain.Subscription{
			ID: 1, BusinessID: businessID, PlanID: defaultPlanID,
			PlanName: def.Name, PlanSlug: def.Slug, Apps: def.Apps,
			PricePerSeatCents: def.PricePerSeatCents, Status: status,
		}
		created = true
	}

	var open []domain.SeatPeriod
	for _, p := range r.periods {
		if p.SubscriptionID == r.sub.ID && p.PeriodEnd == nil {
			open = append(open, p)
		}
	}

	plan := domain.PlanSeatLedger(open, today, seatCount, r.sub.PricePerSeatCents)
	if created {
		plan.CloseNoDay = nil
	}
	r.reconcile++

	for _, c := range plan.CloseNoDay {
		r.closePeriod(c.ID, c.End)
	}
	if plan.CloseActive {
		r.closePeriod(plan.RunningID, plan.CloseRunning)
	}
	if plan.OpenNew {
		r.nextID++
		r.opens++
		r.write()
		r.periods = append(r.periods, domain.SeatPeriod{
			ID: r.nextID, SubscriptionID: r.sub.ID, BusinessID: businessID,
			SeatCount: seatCount, PricePerSeatCents: r.sub.PricePerSeatCents,
			PeriodStart: plan.NewStart, BillingMonth: plan.NewBillingM,
		})
	}
	if r.sub.SeatCount != seatCount {
		r.seatSets++
		r.write()
		r.sub.SeatCount = seatCount
	}
	return nil
}

func (r *statefulRepo) closePeriod(periodID int64, end time.Time) {
	for i := range r.periods {
		if r.periods[i].ID == periodID && r.periods[i].PeriodEnd == nil {
			e := end
			r.periods[i].PeriodEnd = &e
			r.closings++
			r.write()
		}
	}
}

func (r *statefulRepo) ListBusinessIDsForReconciliation(_ context.Context, afterBusinessID int64, limit int) ([]int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []int64
	for _, id := range r.businessIDs {
		if id > afterBusinessID {
			out = append(out, id)
			if len(out) == limit {
				break
			}
		}
	}
	return out, nil
}

func (r *statefulRepo) GetPeriodsForMonth(_ context.Context, businessID int64, month time.Time) ([]domain.SeatPeriod, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	target := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	out := make([]domain.SeatPeriod, 0, len(r.periods))
	for _, p := range r.periods {
		if p.BusinessID == businessID && p.BillingMonth.Equal(target) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *statefulRepo) GetAddonPeriodsForMonth(_ context.Context, _ int64, _ time.Time) ([]domain.AddonBillingPeriod, error) {
	return nil, nil
}

func (r *statefulRepo) GetBillingHistory(context.Context, int64, int) ([]domain.BillingHistoryMonth, error) {
	return nil, nil
}

func (r *statefulRepo) UpdatePaddleSubscription(context.Context, int64, string, string, string) (bool, error) {
	return true, nil
}

func (r *statefulRepo) ApplyPaddleSubscriptionEvent(context.Context, string, string, time.Time, int64, string, string) (bool, int64, error) {
	return true, 0, nil
}

func (r *statefulRepo) RecordPaddleEvent(context.Context, string, string, time.Time) (bool, error) {
	return true, nil
}

func (r *statefulRepo) LastPaddleEventAt(context.Context, string) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

func (r *statefulRepo) GetBusinessIDByPaddleSubscription(context.Context, string) (int64, error) {
	return 0, domain.ErrSubscriptionNotFound
}

// ─── helpers ───────────────────────────────────────────────────────────────────

const testBusinessID int64 = 4242

// currentMonth is the UTC month the test clock sits in.
func currentMonth() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func paidPlan() *domain.Plan {
	return &domain.Plan{
		ID: 2, Name: "Pro", Slug: "pro", Apps: []string{"dash"},
		PricePerSeatCents: 1000, TrialDays: 0,
	}
}

// newRepo models the seeded shape: a free default plan plus a paid one, with the
// owner already on the paid plan.
func newRepo(seats int) *statefulRepo {
	return &statefulRepo{
		plan:     paidPlan(),
		freePlan: &domain.Plan{ID: 1, Name: "Free", Slug: "free", Apps: []string{"dash", "pos"}},
		seats:    seats,
		sub: &domain.Subscription{
			ID: 1, BusinessID: testBusinessID, PlanID: 2,
			PlanName: "Pro", PlanSlug: "pro", Apps: []string{"dash"},
			PricePerSeatCents: 1000, Status: "active", SeatCount: seats,
		},
	}
}

// openPeriod seeds a single open seat period starting on the 1st of `month`.
func (r *statefulRepo) openPeriod(month time.Time, seats, price int) {
	r.nextID++
	r.periods = append(r.periods, domain.SeatPeriod{
		ID: r.nextID, SubscriptionID: 1, BusinessID: testBusinessID,
		SeatCount: seats, PricePerSeatCents: price,
		PeriodStart: month, BillingMonth: month,
	})
}

// ─── Finding 1: the seat ledger never rolled over a billing month ──────────────

// The regression: a subscription whose seats and price were unchanged from last
// month kept one period with billing_month = last month, so the current month
// had no rows and GetBillingDetail reported zero line items and a zero total.
func TestReconcile_MonthRolloverBillsTheNewMonth(t *testing.T) {
	prevMonth := currentMonth().AddDate(0, -1, 0)
	repo := newRepo(15)
	repo.openPeriod(prevMonth, 15, 1000)
	s := &Service{repo: repo}

	// Before any reconciliation the new month has nothing to bill.
	before, err := s.GetBillingDetail(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBillingDetail: %v", err)
	}
	if len(before.LineItems) != 0 {
		t.Fatalf("a read must not create billing periods, got %d line items", len(before.LineItems))
	}

	if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
		t.Fatalf("ReconcileBusinessBilling: %v", err)
	}

	after, err := s.GetBillingDetail(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBillingDetail: %v", err)
	}
	if len(after.LineItems) != 1 {
		t.Fatalf("expected 1 seat line for the new billing month, got %+v", after.LineItems)
	}
	if after.TotalCents <= 0 {
		t.Fatalf("new billing month bills %d cents, want > 0", after.TotalCents)
	}
	if !after.BillingMonth.Equal(currentMonth()) {
		t.Fatalf("billing month %v, want %v", after.BillingMonth, currentMonth())
	}
}

// Seats added mid-cycle are prorated: the two periods must tile the elapsed
// days of the month with no gap and no overlap, so the merchant is neither
// under- nor double-charged.
func TestReconcile_SeatsAddedMidCycleIsProrated(t *testing.T) {
	month := currentMonth()
	repo := newRepo(20)
	repo.openPeriod(month, 3, 1000)
	s := &Service{repo: repo}

	if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
		t.Fatalf("ReconcileBusinessBilling: %v", err)
	}

	detail, err := s.GetBillingDetail(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBillingDetail: %v", err)
	}
	if len(detail.LineItems) != 2 {
		t.Fatalf("expected the old and new seat periods, got %d: %+v", len(detail.LineItems), detail.LineItems)
	}

	old, new := detail.LineItems[0], detail.LineItems[1]
	if old.SeatCount != 3 {
		t.Errorf("first period should bill 3 seats, got %d", old.SeatCount)
	}
	if new.SeatCount != 20 {
		t.Errorf("second period should bill 20 seats, got %d", new.SeatCount)
	}
	if old.PeriodEnd == nil {
		t.Fatal("first period should be closed")
	}
	if !old.PeriodEnd.Equal(new.PeriodStart.AddDate(0, 0, -1)) {
		t.Errorf("periods must be adjacent: old ends %v, new starts %v", old.PeriodEnd, new.PeriodStart)
	}

	// Days billed across both periods must equal the days of the month elapsed.
	elapsed := int(time.Now().UTC().Sub(month).Hours()/24) + 1
	if got := daysInItem(old) + daysInItem(new); got != elapsed {
		t.Fatalf("periods bill %d days, want the %d days elapsed this month (overlap or gap)", got, elapsed)
	}
	if detail.CurrentSeatCount != 20 {
		t.Errorf("subscription reports %d seats, want the current 20", detail.CurrentSeatCount)
	}
}

// Seats removed mid-cycle follow the same rule, with no credit for the days
// already billed.
func TestReconcile_SeatsRemovedMidCycleIsProrated(t *testing.T) {
	month := currentMonth()
	repo := newRepo(3)
	repo.openPeriod(month, 20, 1000)
	s := &Service{repo: repo}

	if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
		t.Fatalf("ReconcileBusinessBilling: %v", err)
	}

	detail, err := s.GetBillingDetail(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBillingDetail: %v", err)
	}
	if len(detail.LineItems) != 2 {
		t.Fatalf("expected 2 periods, got %+v", detail.LineItems)
	}
	if detail.LineItems[0].SeatCount != 20 || detail.LineItems[1].SeatCount != 3 {
		t.Fatalf("expected 20 then 3 seats, got %d then %d",
			detail.LineItems[0].SeatCount, detail.LineItems[1].SeatCount)
	}
	elapsed := int(time.Now().UTC().Sub(month).Hours()/24) + 1
	if got := daysInItem(detail.LineItems[0]) + daysInItem(detail.LineItems[1]); got != elapsed {
		t.Fatalf("periods bill %d days, want %d", got, elapsed)
	}
}

// The renewal must re-derive the amount from current state, never reuse the
// amount computed when the subscription was first created.
func TestReconcile_PriceChangeIsNotReusedFromTheFirstCycle(t *testing.T) {
	month := currentMonth()
	repo := newRepo(15)
	repo.openPeriod(month, 15, 1000)
	s := &Service{repo: repo}

	if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
		t.Fatalf("ReconcileBusinessBilling: %v", err)
	}
	first, err := s.GetBillingDetail(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBillingDetail: %v", err)
	}

	// The plan now charges half as much per seat.
	repo.plan = &domain.Plan{ID: 2, Name: "Pro", Slug: "pro", Apps: []string{"dash"}, PricePerSeatCents: 500}
	repo.sub.PricePerSeatCents = 500
	if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
		t.Fatalf("ReconcileBusinessBilling after price change: %v", err)
	}

	second, err := s.GetBillingDetail(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBillingDetail: %v", err)
	}
	if len(second.LineItems) != 2 {
		t.Fatalf("a price change must split the period, got %+v", second.LineItems)
	}
	if second.LineItems[1].PricePerSeatCents != 500 {
		t.Errorf("newest period still bills %d cents/seat, want 500", second.LineItems[1].PricePerSeatCents)
	}
	if second.TotalCents >= first.TotalCents && first.TotalCents > 0 {
		t.Errorf("total went from %d to %d cents: the new price was not applied to the new period",
			first.TotalCents, second.TotalCents)
	}
}

// Two open periods is what the old unlocked read path left behind. The
// reconciler must collapse them so no day is billed twice.
func TestReconcile_CollapsesDuplicateOpenPeriods(t *testing.T) {
	month := currentMonth()
	repo := newRepo(15)
	repo.openPeriod(month, 15, 1000)
	repo.openPeriod(month, 15, 1000)
	s := &Service{repo: repo}

	if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
		t.Fatalf("ReconcileBusinessBilling: %v", err)
	}

	open := 0
	for _, p := range repo.periods {
		if p.PeriodEnd == nil {
			open++
		}
	}
	if open != 1 {
		t.Fatalf("expected exactly 1 open seat period, got %d", open)
	}

	detail, err := s.GetBillingDetail(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBillingDetail: %v", err)
	}
	elapsed := int(time.Now().UTC().Sub(month).Hours()/24) + 1
	billed := 0
	for _, li := range detail.LineItems {
		billed += daysInItem(li)
	}
	if billed != elapsed {
		t.Fatalf("duplicate periods billed %d days, want %d: a day is charged twice", billed, elapsed)
	}
}

// Reconciling an unchanged subscription repeatedly must not churn the ledger:
// the job runs every fifteen minutes.
func TestReconcile_RepeatedRunsDoNotChurnPeriods(t *testing.T) {
	repo := newRepo(15)
	s := &Service{repo: repo}
	for i := 0; i < 5; i++ {
		if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
			t.Fatalf("ReconcileBusinessBilling run %d: %v", i, err)
		}
	}
	if len(repo.periods) != 1 {
		t.Fatalf("5 reconciles of an unchanged subscription produced %d periods, want 1", len(repo.periods))
	}
}

// An owner with no row yet gets one on the default plan, and it is created by
// the reconciliation path rather than by a read.
func TestReconcile_CreatesMissingSubscription(t *testing.T) {
	repo := newRepo(1)
	repo.sub = nil
	s := &Service{repo: repo}

	if _, err := s.GetMySubscription(context.Background(), testBusinessID); err != nil {
		t.Fatalf("GetMySubscription: %v", err)
	}
	if repo.sub != nil {
		t.Fatal("a read created the subscription row")
	}

	if err := s.ReconcileBusinessBilling(context.Background(), testBusinessID); err != nil {
		t.Fatalf("ReconcileBusinessBilling: %v", err)
	}
	if repo.sub == nil {
		t.Fatal("ReconcileBusinessBilling did not create the subscription row")
	}
	if repo.sub.PlanSlug != defaultPlanSlug {
		t.Fatalf("expected the default plan %q, got %q", defaultPlanSlug, repo.sub.PlanSlug)
	}
	if len(repo.periods) != 1 {
		t.Fatalf("expected 1 seat period after creating the subscription, got %d", len(repo.periods))
	}
}

// ReconcileAllSubscriptions walks every candidate owner and keeps going when
// one fails, so a single bad tenant cannot stop billing for everyone.
func TestReconcileAllSubscriptions_PagesThroughBusinesses(t *testing.T) {
	repo := newRepo(15)
	repo.businessIDs = []int64{1, 2, 3, 4, 5, 6, 7}
	s := &Service{repo: repo}
	if err := s.ReconcileAllSubscriptions(context.Background()); err != nil {
		t.Fatalf("ReconcileAllSubscriptions: %v", err)
	}
	if repo.reconcile != len(repo.businessIDs) {
		t.Fatalf("reconciled %d businesses, want %d", repo.reconcile, len(repo.businessIDs))
	}
}

// ─── Finding 2: the GET path performed unlocked writes ────────────────────────

// Every read entry point must be side-effect free. A GET that writes breaks
// HTTP/edge caching, makes retries non-idempotent, races with itself under
// concurrent reads, and can be triggered by a prefetch.
func TestReads_PerformNoWrites(t *testing.T) {
	cases := map[string]func(t *testing.T, s *Service, businessID int64){
		"GetMySubscription": func(t *testing.T, s *Service, businessID int64) {
			if _, err := s.GetMySubscription(context.Background(), businessID); err != nil {
				t.Fatalf("GetMySubscription: %v", err)
			}
		},
		"GetBillingDetail": func(t *testing.T, s *Service, businessID int64) {
			if _, err := s.GetBillingDetail(context.Background(), businessID); err != nil {
				t.Fatalf("GetBillingDetail: %v", err)
			}
		},
		"GetBillingHistory": func(t *testing.T, s *Service, businessID int64) {
			if _, err := s.GetBillingHistory(context.Background(), businessID, 12); err != nil {
				t.Fatalf("GetBillingHistory: %v", err)
			}
		},
		"GetBusinessApps": func(t *testing.T, s *Service, businessID int64) {
			if _, err := s.GetBusinessApps(context.Background(), businessID); err != nil {
				t.Fatalf("GetBusinessApps: %v", err)
			}
		},
	}

	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			// With an existing subscription.
			repo := newRepo(15)
			call(t, &Service{repo: repo}, testBusinessID)
			assertNoWrites(t, repo, name)

			// And with no subscription row, which is where the old path
			// inserted one.
			fresh := newRepo(1)
			fresh.sub = nil
			call(t, &Service{repo: fresh}, testBusinessID)
			assertNoWrites(t, fresh, name)
		})
	}
}

func assertNoWrites(t *testing.T, repo *statefulRepo, name string) {
	t.Helper()
	if repo.writes != 0 || repo.reconcile != 0 || repo.upserts != 0 {
		t.Fatalf("%s performed %d writes and %d reconciles (upserts=%d closes=%d opens=%d seatSets=%d); a read must be side-effect-free",
			name, repo.writes, repo.reconcile, repo.upserts, repo.closings, repo.opens, repo.seatSets)
	}
}

// A read for an owner with no row still reports the default plan, so the client
// sees the same response the old write-then-read produced.
func TestReads_ProjectDefaultPlanWithoutWriting(t *testing.T) {
	repo := newRepo(4)
	repo.sub = nil
	s := &Service{repo: repo}

	sub, err := s.GetMySubscription(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetMySubscription: %v", err)
	}
	if sub.PlanSlug != "free" || sub.Status != "active" || sub.SeatCount != 4 {
		t.Fatalf("unexpected projected subscription: %+v", sub)
	}
	apps, err := s.GetBusinessApps(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("GetBusinessApps: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("GetBusinessApps = %v, want the default plan's apps", apps)
	}
	assertNoWrites(t, repo, "GetMySubscription")
}

func daysInItem(li domain.BillingLineItem) int {
	end := time.Now().UTC()
	if li.PeriodEnd != nil {
		end = *li.PeriodEnd
	}
	if end.Before(li.PeriodStart) {
		return 0
	}
	return int(end.Sub(li.PeriodStart).Hours()/24) + 1
}
