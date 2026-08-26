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
)

// invalidateSubscriptionScope clears every subscription + user-scoped cache
// entry for the owner. Called by every mutation that changes plan/seat/state.
func invalidateSubscriptionScope(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx,
		cacheKeys.SubscriptionPrefix(ownerID),
		cacheKeys.UserPrefix(ownerID),
	)
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

func (s *Service) GetMySubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	return cacher.Of[*domain.Subscription]().GetOrComputeSingle(
		ctx,
		cacheKeys.Subscription(ownerID),
		func(ctx context.Context) (*domain.Subscription, error) {
			return s.getSubscriptionOrCreateDefault(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, subscriptionCacheTTL),
	)
}

func (s *Service) getSubscriptionOrCreateDefault(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, ownerID)
	if err == nil {
		return s.syncSubscriptionBillingState(ctx, sub)
	}
	if !errors.Is(err, domain.ErrSubscriptionNotFound) {
		return nil, err
	}
	if upsertErr := s.UpsertSubscription(ctx, ownerID, defaultPlanSlug); upsertErr != nil {
		return nil, upsertErr
	}
	sub, err = s.repo.GetSubscription(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return s.withSeatAllowance(sub), nil
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

	sub, err := s.repo.GetSubscription(ctx, ownerID)
	if err != nil {
		return err
	}
	_, err = s.syncSubscriptionBillingState(ctx, sub)
	return err
}

func (s *Service) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	apps, err := s.repo.GetOwnerApps(ctx, ownerID)
	if err == nil {
		return apps, nil
	}
	if !errors.Is(err, domain.ErrSubscriptionNotFound) {
		return nil, err
	}
	if _, createErr := s.getSubscriptionOrCreateDefault(ctx, ownerID); createErr != nil {
		return nil, createErr
	}
	return s.repo.GetOwnerApps(ctx, ownerID)
}

// ExpireTrialSubscriptions transitions trials whose promised end date has
// passed. The request middleware also checks the timestamp, so the nightly job
// is durable cleanup rather than the only enforcement point.
func (s *Service) ExpireTrialSubscriptions(ctx context.Context) error {
	return s.repo.ExpireTrialSubscriptions(ctx)
}

// GetBillingDetail returns the current month's usage line items and totals.
func (s *Service) GetBillingDetail(ctx context.Context, ownerID, businessID int64) (*domain.BillingDetail, error) {
	sub, err := s.getSubscriptionOrCreateDefault(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
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

// RecordSeatChange is called when a team member is added or removed.
// It closes the current seat period and opens a new one reflecting the new count.
func (s *Service) RecordSeatChange(ctx context.Context, ownerID int64, newSeatCount int) error {
	if newSeatCount < 1 {
		newSeatCount = 1
	}
	sub, err := s.repo.GetSubscription(ctx, ownerID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	billingMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	activePeriod, err := s.repo.GetActivePeriod(ctx, sub.ID)
	if err != nil {
		return err
	}
	if activePeriod != nil &&
		activePeriod.SeatCount == newSeatCount &&
		activePeriod.PricePerSeatCents == sub.PricePerSeatCents {
		if sub.SeatCount != newSeatCount {
			if err := s.repo.UpdateSeatCount(ctx, ownerID, newSeatCount); err != nil {
				return err
			}
			invalidateSubscriptionScope(ctx, ownerID)
		}
		return nil
	}

	if activePeriod != nil {
		// Close the active period as of yesterday (the last full day at old count).
		yesterday := now.AddDate(0, 0, -1)
		if closeErr := s.repo.CloseActivePeriod(ctx, sub.ID, yesterday); closeErr != nil {
			return closeErr
		}
	}

	// Open a new period starting today with the new seat count.
	if openErr := s.repo.OpenPeriod(ctx, sub.ID, ownerID, newSeatCount, sub.PricePerSeatCents, now, billingMonth); openErr != nil {
		return openErr
	}

	// Update seat_count on the subscription.
	if err := s.repo.UpdateSeatCount(ctx, ownerID, newSeatCount); err != nil {
		return err
	}
	invalidateSubscriptionScope(ctx, ownerID)
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

func (s *Service) syncSubscriptionBillingState(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	seatCount, err := s.currentActiveSeatCount(ctx, sub.OwnerID)
	if err != nil {
		return nil, err
	}

	activePeriod, err := s.repo.GetActivePeriod(ctx, sub.ID)
	if err != nil {
		return nil, err
	}
	needsPeriodSync := activePeriod == nil ||
		activePeriod.SeatCount != seatCount ||
		activePeriod.PricePerSeatCents != sub.PricePerSeatCents

	if sub.SeatCount != seatCount || needsPeriodSync {
		if err := s.RecordSeatChange(ctx, sub.OwnerID, seatCount); err != nil {
			return nil, err
		}
		sub, err = s.repo.GetSubscription(ctx, sub.OwnerID)
		if err != nil {
			return nil, err
		}
	}

	return s.withSeatAllowance(sub), nil
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

	firstOfNextMonth := time.Date(billingMonth.Year(), billingMonth.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := int(firstOfNextMonth.Sub(billingMonth).Hours() / 24)

	daysUsed := int(effectiveEnd.Sub(start).Hours()/24) + 1
	if daysUsed < 1 {
		daysUsed = 1
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
