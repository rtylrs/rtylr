package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	paddlePkg "github.com/voxire/onion/internal/platform/paddle"
)

const (
	subscriptionCacheTTL = 10 * time.Minute
	defaultPlanSlug      = "free"
	reconcileBatchSize   = 200
)

// invalidateSubscriptionScope clears every subscription + user-scoped cache
// entry for the owner. Called by every mutation that changes plan/seat/state.
func invalidateSubscriptionScope(ctx context.Context, ownerID int64) {
	patterns := append(cacheKeys.SubscriptionPatterns(ownerID), cacheKeys.UserPrefix(ownerID))
	cacher.InvalidateScope(ctx, patterns...)
}

type Service struct {
	repo domain.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.ListPlans(ctx)
}

// GetMySubscription is a pure read. Everything that changes subscription or
// seat state lives behind ReconcileOwnerBilling, which the background job calls;
// a GET used to do it inline, so every cached read (and every prefetch or
// crawler hit that missed the cache) mutated rows without a lock.
func (s *Service) GetMySubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	return cacher.Of[*domain.Subscription]().GetOrComputeSingle(
		ctx,
		cacheKeys.Subscription(ownerID),
		func(ctx context.Context) (*domain.Subscription, error) {
			return s.readSubscription(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, subscriptionCacheTTL),
	)
}

func (s *Service) readSubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, ownerID)
	if errors.Is(err, domain.ErrSubscriptionNotFound) {
		return s.defaultSubscription(ctx, ownerID)
	}
	if err != nil {
		return nil, err
	}
	return s.withSeatAllowance(sub), nil
}

// defaultSubscription projects the default plan for an owner that has no
// subscription row yet. The background job persists the row; a read only
// reports what it would have contained, so the response is unchanged without
// the write.
func (s *Service) defaultSubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	plan, err := s.repo.GetPlanBySlug(ctx, defaultPlanSlug)
	if err != nil {
		return nil, err
	}
	seatCount, err := s.currentActiveSeatCount(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return s.withSeatAllowance(&domain.Subscription{
		OwnerID:           ownerID,
		PlanName:          plan.Name,
		PlanSlug:          plan.Slug,
		Apps:              plan.Apps,
		PricePerSeatCents: plan.PricePerSeatCents,
		Status:            "active",
		SeatCount:         seatCount,
	}), nil
}

func (s *Service) UpsertSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	plan, err := s.repo.GetPlanBySlug(ctx, planSlug)
	if err != nil {
		return err
	}

	seatCount, err := s.currentActiveSeatCount(ctx, ownerID)
	if err != nil {
		return err
	}

	if upsertErr := s.repo.UpsertSubscription(ctx, ownerID, plan.ID, plan.TrialDays, seatCount); upsertErr != nil {
		return upsertErr
	}
	defer invalidateSubscriptionScope(ctx, ownerID)
	// A plan change moves the per-seat price, so the running seat period has to
	// be split rather than left at the old price for the rest of the month.
	return s.ReconcileOwnerBilling(ctx, ownerID)
}

func (s *Service) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	apps, err := s.repo.GetOwnerApps(ctx, ownerID)
	if err == nil {
		return apps, nil
	}
	if !errors.Is(err, domain.ErrSubscriptionNotFound) {
		return nil, err
	}
	plan, planErr := s.repo.GetPlanBySlug(ctx, defaultPlanSlug)
	if planErr != nil {
		return nil, planErr
	}
	return plan.Apps, nil
}

// ExpireTrialSubscriptions transitions trials whose promised end date has
// passed. The request middleware also checks the timestamp, so the nightly job
// is durable cleanup rather than the only enforcement point.
func (s *Service) ExpireTrialSubscriptions(ctx context.Context) error {
	return s.repo.ExpireTrialSubscriptions(ctx)
}

// GetBillingDetail returns the current month's usage line items and totals.
func (s *Service) GetBillingDetail(ctx context.Context, ownerID, businessID int64) (*domain.BillingDetail, error) {
	sub, err := s.readSubscription(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	currentMonth := domain.FirstOfUTCMonth(time.Now().UTC())
	periods, err := s.repo.GetPeriodsForMonth(ctx, ownerID, currentMonth)
	if err != nil {
		return nil, err
	}

	lineItems := make([]domain.BillingLineItem, 0, len(periods))
	totalCents := 0
	seatDiscountCents := 0
	for _, p := range periods {
		item := computeLineItem(p.SeatCount, p.PricePerSeatCents, p.PeriodStart, p.PeriodEnd, currentMonth)
		if sub.IsGranted {
			item = applyGrantToSeatLineItem(item)
		}
		lineItems = append(lineItems, item)
		totalCents += item.AmountCents
		seatDiscountCents += item.IncludedCents
	}
	addonPeriods, err := s.repo.GetAddonPeriodsForMonth(ctx, ownerID, businessID, currentMonth)
	if err != nil {
		return nil, err
	}
	for _, p := range addonPeriods {
		item := computeAddonLineItem(p)
		lineItems = append(lineItems, item)
		totalCents += item.AmountCents
	}

	detail := &domain.BillingDetail{
		BillingMonth:         currentMonth,
		NextBillingDate:      sub.NextBillingDate,
		LineItems:            lineItems,
		TotalCents:           totalCents,
		PlanName:             sub.PlanName,
		PricePerSeatCents:    sub.PricePerSeatCents,
		CurrentSeatCount:     sub.SeatCount,
		IncludedSeatLimit:    sub.IncludedSeatLimit,
		IncludedSeatCount:    sub.IncludedSeatCount,
		BillableSeatCount:    sub.BillableSeatCount,
		SeatDiscountCents:    seatDiscountCents,
		Status:               sub.Status,
		IsGranted:            sub.IsGranted,
		GrantExpiresAt:       sub.GrantExpiresAt,
		TrialEndsAt:          sub.TrialEndsAt,
		PaddleSubscriptionID: sub.PaddleSubscriptionID,
	}

	return detail, nil
}

// GetBillingHistory returns the last N billing months.
func (s *Service) GetBillingHistory(ctx context.Context, ownerID, businessID int64, limit int) ([]domain.BillingHistoryMonth, error) {
	months, err := s.repo.GetBillingHistory(ctx, ownerID, limit)
	if err != nil {
		return nil, err
	}
	for i := range months {
		addonPeriods, err := s.repo.GetAddonPeriodsForMonth(ctx, ownerID, businessID, months[i].Month)
		if err != nil {
			return nil, err
		}
		for _, p := range addonPeriods {
			item := computeAddonLineItem(p)
			months[i].LineItems = append(months[i].LineItems, item)
			months[i].TotalCents += item.AmountCents
		}
	}
	return months, nil
}

// ReconcileOwnerBilling re-derives one owner's seat ledger from current state:
// the subscription row (created on the default plan if absent) and the seat
// period for the current billing month, plus subscriptions.seat_count.
//
// This is the only path that writes billing state. It is driven by the
// subscriptions.reconcile job and by plan changes, never by a GET. The
// repository call is transactional and takes a row lock, so calling it
// concurrently for the same owner converges on a single open period instead of
// two that are both billed.
func (s *Service) ReconcileOwnerBilling(ctx context.Context, ownerID int64) error {
	// The default plan is only materialised when the owner has no row; passing
	// it unconditionally lets the repository keep the existing plan on the
	// create-if-absent insert.
	plan, err := s.repo.GetPlanBySlug(ctx, defaultPlanSlug)
	if err != nil {
		return err
	}
	seatCount, err := s.currentActiveSeatCount(ctx, ownerID)
	if err != nil {
		return err
	}
	if err := s.repo.ReconcileSeatBilling(ctx, ownerID, plan.ID, plan.TrialDays, seatCount, time.Now().UTC()); err != nil {
		return err
	}
	invalidateSubscriptionScope(ctx, ownerID)
	return nil
}

// ReconcileAllSubscriptions is the subscriptions.reconcile job: it walks every
// owner that could owe for the current billing month and rolls their ledger
// forward. Without it nothing advances a seat period past the month it was
// opened in, so every later month bills zero.
func (s *Service) ReconcileAllSubscriptions(ctx context.Context) error {
	after := int64(0)
	reconciled, failed := 0, 0
	for {
		owners, err := s.repo.ListOwnerIDsForReconciliation(ctx, after, reconcileBatchSize)
		if err != nil {
			return err
		}
		if len(owners) == 0 {
			break
		}
		for _, ownerID := range owners {
			after = ownerID
			if err := s.ReconcileOwnerBilling(ctx, ownerID); err != nil {
				failed++
				logger.Ctx(ctx).Error("seat billing reconciliation failed",
					logger.Any("owner_id", ownerID),
					logger.Error(err),
				)
				continue
			}
			reconciled++
		}
		if len(owners) < reconcileBatchSize {
			break
		}
	}
	logger.Ctx(ctx).Info("seat billing reconciliation finished",
		logger.Int("reconciled", reconciled),
		logger.Int("failed", failed),
	)
	return nil
}

func (s *Service) currentActiveSeatCount(ctx context.Context, ownerID int64) (int, error) {
	seatCount, err := s.repo.CountActiveSeats(ctx, ownerID)
	if err != nil {
		return 0, err
	}
	if seatCount < 1 {
		return 1, nil
	}
	return seatCount, nil
}

func (s *Service) withSeatAllowance(sub *domain.Subscription) *domain.Subscription {
	includedLimit := domain.FreeIncludedSeatLimit
	includedSeats, billableSeats := splitSeats(sub.SeatCount, includedLimit)
	if sub.IsGranted {
		includedSeats = sub.SeatCount
		billableSeats = 0
		if includedLimit < sub.SeatCount {
			includedLimit = sub.SeatCount
		}
	}
	sub.IncludedSeatLimit = includedLimit
	sub.IncludedSeatCount = includedSeats
	sub.BillableSeatCount = billableSeats
	return sub
}

func applyGrantToSeatLineItem(item domain.BillingLineItem) domain.BillingLineItem {
	item.IncludedCents += item.AmountCents
	item.AmountCents = 0
	item.BillableSeatCount = 0
	item.IncludedSeatCount = item.SeatCount
	if item.IncludedSeatLimit < item.SeatCount {
		item.IncludedSeatLimit = item.SeatCount
	}
	item.Description = "Granted plan access"
	return item
}

// computeLineItem mirrors the infrastructure helper for use in the service layer.
func computeLineItem(seatCount, pricePerSeatCents int, start time.Time, end *time.Time, billingMonth time.Time) domain.BillingLineItem {
	return computeLineItemWithAllowance(seatCount, pricePerSeatCents, start, end, billingMonth, domain.FreeIncludedSeatLimit)
}

func computeLineItemWithAllowance(seatCount, pricePerSeatCents int, start time.Time, end *time.Time, billingMonth time.Time, includedSeatLimit int) domain.BillingLineItem {
	effectiveEnd := time.Now().UTC()
	if end != nil {
		effectiveEnd = *end
	}

	daysInMonth := domain.DaysInUTCMonth(billingMonth)

	// A period closed before it started never covered a day, so it bills
	// nothing. The reconciler produces that shape when it collapses a
	// duplicate open period, and charging it a day would double-bill the
	// day the replacement period already covers.
	daysUsed := 0
	if !effectiveEnd.Before(start) {
		daysUsed = int(effectiveEnd.Sub(start).Hours()/24) + 1
	}
	if daysUsed > daysInMonth {
		daysUsed = daysInMonth
	}

	includedSeats, billableSeats := splitSeats(seatCount, includedSeatLimit)
	amount := (billableSeats * daysUsed * pricePerSeatCents) / daysInMonth
	includedAmount := (includedSeats * daysUsed * pricePerSeatCents) / daysInMonth

	return domain.BillingLineItem{
		PeriodStart:       start,
		PeriodEnd:         end,
		SeatCount:         seatCount,
		IncludedSeatLimit: includedSeatLimit,
		IncludedSeatCount: includedSeats,
		BillableSeatCount: billableSeats,
		IncludedCents:     includedAmount,
		PricePerSeatCents: pricePerSeatCents,
		AmountCents:       amount,
		Type:              "seat",
		Description:       "Seat usage",
		Quantity:          seatCount,
		UnitPriceCents:    pricePerSeatCents,
	}
}

func computeAddonLineItem(period domain.AddonBillingPeriod) domain.BillingLineItem {
	item := computeLineItemWithAllowance(1, period.PriceCents, period.PeriodStart, period.PeriodEnd, period.BillingMonth, 0)
	item.Type = "addon"
	item.Description = period.Name
	item.AddonID = period.AddonID
	item.Quantity = 1
	item.UnitPriceCents = period.PriceCents
	item.SeatCount = 0
	item.PricePerSeatCents = 0
	item.IsGranted = period.IsGranted
	item.GrantExpiresAt = period.GrantExpiresAt
	if period.IsGranted {
		item.IncludedCents += item.AmountCents
		item.AmountCents = 0
	}
	return item
}

func splitSeats(seatCount, includedLimit int) (int, int) {
	if seatCount < 0 {
		seatCount = 0
	}
	if includedLimit < 0 {
		includedLimit = 0
	}
	included := seatCount
	if included > includedLimit {
		included = includedLimit
	}
	return included, seatCount - included
}

// ── Paddle ─────────────────────────────────────────────────────────────────────

// CreateCheckout creates a Paddle transaction for the given plan and returns the checkout URL.
func (s *Service) CreateCheckout(ctx context.Context, ownerID int64, planSlug, ownerEmail string) (string, error) {
	plan, err := s.repo.GetPlanBySlug(ctx, planSlug)
	if err != nil {
		return "", err
	}

	if plan.PaddlePriceID == "" {
		return "", fmt.Errorf("plan %s has no Paddle price configured", planSlug)
	}

	return paddlePkg.CreateTransaction(ctx, ownerID, plan.PaddlePriceID, ownerEmail)
}

// HandlePaddleWebhook processes a verified Paddle webhook event.
//
// Delivery is at-least-once and possibly out of order, so the handler first
// dedupes on event_id and rejects events older than the newest already
// applied for the subscription; a stale retried "active" must never undo a
// later cancellation (#755). Unknown statuses leave the row untouched rather
// than silently re-activating it.
func (s *Service) HandlePaddleWebhook(ctx context.Context, event paddlePkg.WebhookEvent, data paddlePkg.SubscriptionData) error {
	if event.EventID == "" {
		return errors.New("paddle webhook event_id required")
	}
	if data.ID == "" {
		return errors.New("paddle subscription id required")
	}

	status, ok := mapPaddleStatus(data.Status)
	if !ok {
		logger.Ctx(ctx).Warn("unknown paddle subscription status ignored",
			logger.String("status", data.Status),
			logger.String("event_id", event.EventID))
		return nil
	}

	var fallbackOwnerID int64
	if ownerIDRaw := data.CustomData["owner_id"]; ownerIDRaw != "" {
		if _, err := fmt.Sscanf(ownerIDRaw, "%d", &fallbackOwnerID); err != nil || fallbackOwnerID == 0 {
			return fmt.Errorf("invalid owner_id in custom_data: %s", ownerIDRaw)
		}
	}

	applied, ownerID, err := s.repo.ApplyPaddleSubscriptionEvent(ctx, event.EventID, data.ID, event.OccurredAt, fallbackOwnerID, data.CustomerID, status)
	if err != nil {
		return err
	}
	if !applied {
		logger.Ctx(ctx).Info("duplicate or out-of-order paddle webhook ignored",
			logger.String("event_id", event.EventID),
			logger.String("event_type", event.EventType))
		return nil
	}
	invalidateSubscriptionScope(ctx, ownerID)
	return nil
}

// mapPaddleStatus translates Paddle subscription states. The second return is
// false for unrecognized statuses, which callers must not map to a default.
func mapPaddleStatus(paddleStatus string) (string, bool) {
	switch paddleStatus {
	case "active":
		return "active", true
	case "trialing":
		return "trial", true
	case "past_due":
		// Dunning: keep the state visible instead of granting full access.
		return "past_due", true
	case "canceled", "paused":
		return "cancelled", true
	case "expired":
		return "expired", true
	default:
		return "", false
	}
}
