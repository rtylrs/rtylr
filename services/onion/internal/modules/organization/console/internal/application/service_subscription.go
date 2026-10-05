package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Service) GetSubscriptionByOwner(ctx context.Context, ownerID int64) (*SubscriptionDTO, error) {
	sub, err := s.repo.GetSubscriptionByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	dto := toSubDTO(*sub)
	return &dto, nil
}

type SubscriptionWriteInput struct {
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

func (s *Service) subscriptionWriteParams(ctx context.Context, input SubscriptionWriteInput) (SubscriptionWriteParams, error) {
	if input.OwnerID <= 0 || input.PlanID <= 0 {
		return SubscriptionWriteParams{}, fmt.Errorf("%w: owner_id and plan_id are required", ErrInvalidSubscription)
	}
	if _, err := s.repo.GetActiveBusinessByOwner(ctx, input.OwnerID); err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return SubscriptionWriteParams{}, ErrNotOrganizationOwner
		}
		return SubscriptionWriteParams{}, err
	}

	status, err := normalizeSubscriptionStatus(input.Status)
	if err != nil {
		return SubscriptionWriteParams{}, err
	}

	seatCount := input.SeatCount
	if seatCount <= 0 {
		seatCount, err = s.repo.CountActiveSeatsForOwner(ctx, input.OwnerID)
		if err != nil {
			return SubscriptionWriteParams{}, err
		}
		if seatCount < 1 {
			seatCount = 1
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
		OwnerID:              input.OwnerID,
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
	params, err := s.subscriptionWriteParams(ctx, input)
	if err != nil {
		return err
	}
	defer invalidateConsoleScope(ctx)
	defer invalidateOwnerSubscriptionScope(ctx, input.OwnerID)
	return s.repo.UpsertSubscription(ctx, params)
}

func (s *Service) UpdateSubscription(ctx context.Context, id int64, input SubscriptionWriteInput) error {
	defer invalidateConsoleScope(ctx)
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return err
	}
	input.OwnerID = sub.OwnerID
	defer invalidateOwnerSubscriptionScope(ctx, sub.OwnerID)
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

func (s *Service) activeSeatCountForOwner(ctx context.Context, ownerID int64) (int, error) {
	if _, err := s.repo.GetActiveBusinessByOwner(ctx, ownerID); err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return 0, ErrNotOrganizationOwner
		}
		return 0, err
	}
	seatCount, err := s.repo.CountActiveSeatsForOwner(ctx, ownerID)
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
