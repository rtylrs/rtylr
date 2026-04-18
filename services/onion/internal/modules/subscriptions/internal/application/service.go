package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure"
)

type Service struct {
	repo domain.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *Service) GetMySubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
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
	return s.repo.GetOwnerApps(ctx, ownerID)
}

// GetBillingDetail returns the current month's usage line items and totals.
func (s *Service) GetBillingDetail(ctx context.Context, ownerID int64) (*domain.BillingDetail, error) {
	sub, err := s.repo.GetSubscription(ctx, ownerID)
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
func (s *Service) GetBillingHistory(ctx context.Context, ownerID int64, limit int) ([]domain.BillingHistoryMonth, error) {
	return s.repo.GetBillingHistory(ctx, ownerID, limit)
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
	return s.repo.UpdateSeatCount(ctx, ownerID, newSeatCount)
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
	}
}
