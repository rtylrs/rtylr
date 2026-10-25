package application

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *Service) GetSubscriptionByBusiness(ctx context.Context, businessID int64) (*SubscriptionDTO, error) {
	sub, err := s.repo.GetSubscriptionByBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	dto := toSubDTO(*sub)
	return &dto, nil
}

// GetSubscriptionForUser returns the subscription of the business the user
// owns, preferring the one their session opens in. A user who owns no active
// business has none: ErrRecordNotFound.
func (s *Service) GetSubscriptionForUser(ctx context.Context, userID int64) (*SubscriptionDTO, error) {
	businesses, err := s.repo.ListActiveBusinessesByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(businesses) == 0 {
		return nil, ErrRecordNotFound
	}
	return s.GetSubscriptionByBusiness(ctx, businesses[0].ID)
}

// resolveBillingBusiness picks the business a console billing write applies
// to. Partials may still name only the owner; that is accepted while the owner
// has exactly one active business, so an owner with several must say which.
func (s *Service) resolveBillingBusiness(ctx context.Context, ownerID, businessID int64) (*BusinessRow, error) {
	if businessID > 0 {
		business, err := s.repo.GetBusinessByID(ctx, businessID)
		if err != nil {
			return nil, err
		}
		if ownerID > 0 && business.OwnerID != ownerID {
			return nil, ErrBusinessNotOwnedByUser
		}
		return business, nil
	}
	if ownerID <= 0 {
		return nil, ErrBusinessRequired
	}
	businesses, err := s.repo.ListActiveBusinessesByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	switch len(businesses) {
	case 0:
		return nil, ErrNotOrganizationOwner
	case 1:
		return &businesses[0], nil
	default:
		return nil, ErrOwnerHasSeveralBusinesses
	}
}

type SubscriptionWriteInput struct {
	BusinessID int64
	// OwnerID names the business through its owner when BusinessID is unset.
	OwnerID              int64
	PlanID               int64
	Status               string
	SeatCount            int
	TrialEndsAt          *time.Time
	NextBillingDate      *time.Time
	CurrentPeriodEndsAt  *time.Time
	TotalDays            *int
	PaddleSubscriptionID string
	PaddleCustomerID     string
}

func nullableInputTime(value *time.Time) NullTime {
	if value == nil {
		return NullTime{}
	}
	return NullTime{Time: value.UTC(), Valid: true}
}

func nullableInputString(value string) NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return NullString{}
	}
	return NullString{String: value, Valid: true}
}

func normalizeSubscriptionStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "active", nil
	}
	switch status {
	case "active", "trial", "cancelled", "expired":
		return status, nil
	default:
		return "", fmt.Errorf("%w: status must be active, trial, cancelled, or expired", ErrInvalidSubscription)
	}
}

// subscriptionWriteParams expects input.BusinessID to be resolved already.
func (s *Service) subscriptionWriteParams(ctx context.Context, input SubscriptionWriteInput) (SubscriptionWriteParams, error) {
	if input.BusinessID <= 0 || input.PlanID <= 0 {
		return SubscriptionWriteParams{}, fmt.Errorf("%w: business_id and plan_id are required", ErrInvalidSubscription)
	}

	status, err := normalizeSubscriptionStatus(input.Status)
	if err != nil {
		return SubscriptionWriteParams{}, err
	}

	seatCount := input.SeatCount
	if seatCount <= 0 {
		seatCount, err = s.activeSeatCountForBusiness(ctx, input.BusinessID)
		if err != nil {
			return SubscriptionWriteParams{}, err
		}
	}

	trialEndsAt := input.TrialEndsAt
	currentPeriodEndsAt := input.CurrentPeriodEndsAt
	if input.TotalDays != nil {
		if *input.TotalDays < 0 {
			return SubscriptionWriteParams{}, fmt.Errorf("%w: total_days cannot be negative", ErrInvalidSubscription)
		}
		endsAt := time.Now().UTC().AddDate(0, 0, *input.TotalDays)
		currentPeriodEndsAt = &endsAt
		if status == "trial" {
			trialEndsAt = &endsAt
		}
	}

	return SubscriptionWriteParams{
		BusinessID:           input.BusinessID,
		PlanID:               input.PlanID,
		Status:               status,
		SeatCount:            seatCount,
		TrialEndsAt:          nullableInputTime(trialEndsAt),
		NextBillingDate:      nullableInputTime(input.NextBillingDate),
		CurrentPeriodEndsAt:  nullableInputTime(currentPeriodEndsAt),
		PaddleSubscriptionID: nullableInputString(input.PaddleSubscriptionID),
		PaddleCustomerID:     nullableInputString(input.PaddleCustomerID),
	}, nil
}

func (s *Service) UpsertSubscription(ctx context.Context, input SubscriptionWriteInput) error {
	business, err := s.resolveBillingBusiness(ctx, input.OwnerID, input.BusinessID)
	if err != nil {
		return err
	}
	input.BusinessID = business.ID
	params, err := s.subscriptionWriteParams(ctx, input)
	if err != nil {
		return err
	}
	defer invalidateConsoleScope(ctx)
	defer invalidateBusinessSubscriptionScope(ctx, business.ID)
	return s.repo.UpsertSubscription(ctx, params)
}

func (s *Service) UpdateSubscription(ctx context.Context, id int64, input SubscriptionWriteInput) error {
	defer invalidateConsoleScope(ctx)
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return err
	}
	input.BusinessID = sub.BusinessID
	defer invalidateBusinessSubscriptionScope(ctx, sub.BusinessID)
	params, err := s.subscriptionWriteParams(ctx, input)
	if err != nil {
		return err
	}
	return s.repo.UpdateSubscription(ctx, id, params)
}

func (s *Service) DeleteSubscription(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.DeleteSubscription(ctx, id)
}

func (s *Service) activeSeatCountForBusiness(ctx context.Context, businessID int64) (int, error) {
	seatCount, err := s.repo.CountActiveSeatsForBusiness(ctx, businessID)
	if err != nil {
		return 0, err
	}
	if seatCount < 1 {
		return 1, nil
	}
	return seatCount, nil
}

// ── Plans ─────────────────────────────────────────────────────────────────────

type PlanDTO struct {
	ID                      int64  `json:"id"`
	Name                    string `json:"name"`
	Slug                    string `json:"slug"`
	PricePerSeatCents       int    `json:"price_per_seat_cents"`
	AnnualPricePerSeatCents int    `json:"annual_price_per_seat_cents"`
	TrialDays               int    `json:"trial_days"`
	MaxSeats                *int64 `json:"max_seats,omitempty"`
	MaxLocations            *int64 `json:"max_locations,omitempty"`
	IsActive                bool   `json:"is_active"`
	IsSolo                  bool   `json:"is_solo"`
	IsFeatured              bool   `json:"is_featured"`
	TierOrder               int    `json:"tier_order"`
	Tagline                 string `json:"tagline"`
	CreatedAt               string `json:"created_at"`
}

func (s *Service) ListPlans(ctx context.Context) ([]PlanDTO, error) {
	rows, err := s.repo.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]PlanDTO, 0, len(rows))
	for _, p := range rows {
		dto := PlanDTO{
			ID: p.ID, Name: p.Name, Slug: p.Slug,
			PricePerSeatCents: p.PricePerSeatCents, AnnualPricePerSeatCents: p.AnnualPricePerSeatCents,
			TrialDays: p.TrialDays, IsActive: p.IsActive, IsSolo: p.IsSolo,
			IsFeatured: p.IsFeatured, TierOrder: p.TierOrder, Tagline: p.Tagline,
		}
		if p.MaxSeats.Valid {
			dto.MaxSeats = &p.MaxSeats.Int64
		}
		if p.MaxLocations.Valid {
			dto.MaxLocations = &p.MaxLocations.Int64
		}
		if p.CreatedAt.Valid {
			dto.CreatedAt = p.CreatedAt.Time.Format(time.RFC3339)
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}
