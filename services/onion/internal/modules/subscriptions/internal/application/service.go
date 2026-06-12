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
	paddlePkg "github.com/voxire/onion/internal/platform/paddle"
)

const subscriptionCacheTTL = 10 * time.Minute
const defaultPlanSlug = "free"

// invalidateSubscriptionScope clears every subscription + user-scoped cache
// entry for the owner. Called by every mutation that changes plan/seat/state.
func invalidateSubscriptionScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.SubscriptionPrefix(ownerID))
		cache.DeleteByPattern(ctx, cacheKeys.UserPrefix(ownerID))
	}
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
		return sub, nil
	}
	if !errors.Is(err, domain.ErrSubscriptionNotFound) {
		return nil, err
	}
	if upsertErr := s.UpsertSubscription(ctx, ownerID, defaultPlanSlug); upsertErr != nil {
		return nil, upsertErr
	}
	return s.repo.GetSubscription(ctx, ownerID)
}

func (s *Service) UpsertSubscription(ctx context.Context, ownerID int64, planSlug string) error {
	plan, err := s.repo.GetPlanBySlug(ctx, planSlug)
	if err != nil {
		return err
	}

	if upsertErr := s.repo.UpsertSubscription(ctx, ownerID, plan.ID, plan.TrialDays); upsertErr != nil {
		return upsertErr
	}
	defer invalidateSubscriptionScope(ctx, ownerID)

	// Ensure a seat period exists for the current month.
	sub, err := s.repo.GetSubscription(ctx, ownerID)
	if err != nil {
		return err
	}

	activePeriod, err := s.repo.GetActivePeriod(ctx, sub.ID)
	if err != nil {
		return err
	}

	if activePeriod == nil {
		now := time.Now()
		billingMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		if openErr := s.repo.OpenPeriod(ctx, sub.ID, ownerID, sub.SeatCount, plan.PricePerSeatCents, now, billingMonth); openErr != nil {
			return openErr
		}
	}

	return nil
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

// GetBillingDetail returns the current month's usage line items and totals.
func (s *Service) GetBillingDetail(ctx context.Context, ownerID, businessID int64) (*domain.BillingDetail, error) {
	sub, err := s.getSubscriptionOrCreateDefault(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periods, err := s.repo.GetPeriodsForMonth(ctx, ownerID, currentMonth)
	if err != nil {
		return nil, err
	}

	lineItems := make([]domain.BillingLineItem, 0, len(periods))
	totalCents := 0
	for _, p := range periods {
		item := computeLineItem(p.SeatCount, p.PricePerSeatCents, p.PeriodStart, p.PeriodEnd, currentMonth)
		lineItems = append(lineItems, item)
		totalCents += item.AmountCents
	}
	addonPeriods, err := s.repo.GetAddonPeriodsForMonth(ctx, businessID, currentMonth)
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
		Status:               sub.Status,
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
		addonPeriods, err := s.repo.GetAddonPeriodsForMonth(ctx, businessID, months[i].Month)
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
	sub, err := s.repo.GetSubscription(ctx, ownerID)
	if err != nil {
		return err
	}

	now := time.Now()
	billingMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Close the active period as of yesterday (the last full day at old count).
	yesterday := now.AddDate(0, 0, -1)
	if closeErr := s.repo.CloseActivePeriod(ctx, sub.ID, yesterday); closeErr != nil {
		return closeErr
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

// computeLineItem mirrors the infrastructure helper for use in the service layer.
func computeLineItem(seatCount, pricePerSeatCents int, start time.Time, end *time.Time, billingMonth time.Time) domain.BillingLineItem {
	effectiveEnd := time.Now()
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

	amount := (seatCount * daysUsed * pricePerSeatCents) / daysInMonth

	return domain.BillingLineItem{
		PeriodStart:       start,
		PeriodEnd:         end,
		SeatCount:         seatCount,
		PricePerSeatCents: pricePerSeatCents,
		AmountCents:       amount,
		Type:              "seat",
		Description:       "Seat usage",
		Quantity:          seatCount,
		UnitPriceCents:    pricePerSeatCents,
	}
}

func computeAddonLineItem(period domain.AddonBillingPeriod) domain.BillingLineItem {
	item := computeLineItem(1, period.PriceCents, period.PeriodStart, period.PeriodEnd, period.BillingMonth)
	item.Type = "addon"
	item.Description = period.Name
	item.AddonID = period.AddonID
	item.Quantity = 1
	item.UnitPriceCents = period.PriceCents
	item.SeatCount = 0
	item.PricePerSeatCents = 0
	return item
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
// It updates the subscription record based on the event type.
func (s *Service) HandlePaddleWebhook(ctx context.Context, _ paddlePkg.WebhookEvent, data paddlePkg.SubscriptionData) error {
	// Map Paddle subscription ID → owner ID
	ownerID, err := s.repo.GetOwnerIDByPaddleSubscription(ctx, data.ID)
	if err != nil {
		// Subscription might not exist yet (first webhook before our DB is updated)
		ownerIDRaw := data.CustomData["owner_id"]
		if ownerIDRaw == "" {
			return fmt.Errorf("cannot resolve owner from paddle subscription %s", data.ID)
		}
		_, scanErr := fmt.Sscanf(ownerIDRaw, "%d", &ownerID)
		if scanErr != nil || ownerID == 0 {
			return fmt.Errorf("invalid owner_id in custom_data: %s", ownerIDRaw)
		}
	}

	// Map Paddle status to our status
	status := "active"
	switch data.Status {
	case "canceled":
		status = "cancelled"
	case "paused":
		status = "cancelled"
	case "trialing":
		status = "trial"
	}

	if err := s.repo.UpdatePaddleSubscription(ctx, ownerID, data.ID, data.CustomerID, status); err != nil {
		return err
	}
	invalidateSubscriptionScope(ctx, ownerID)
	return nil
}
