package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

func (s *Service) BanUser(ctx context.Context, id int64, reason string) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.BanUser(ctx, id, reason)
}

func (s *Service) UnbanUser(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.UnbanUser(ctx, id)
}

func (s *Service) SetConsoleAdmin(ctx context.Context, id int64, isAdmin bool) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.SetConsoleAdmin(ctx, id, isAdmin)
}

// ── Businesses ────────────────────────────────────────────────────────────────

type BusinessDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	OwnerID   int64  `json:"owner_id"`
	CreatedAt string `json:"created_at"`
}

type ListBusinessesResult struct {
	Businesses []BusinessDTO `json:"businesses"`
	Total      int64         `json:"total"`
}

func (s *Service) ListBusinesses(
	ctx context.Context,
	limit, offset int,
	id int64,
	name string,
	ownerID int64,
) (*ListBusinessesResult, error) {
	name = strings.TrimSpace(name)
	key := consoleKey(fmt.Sprintf("biz|l=%d|o=%d|id=%d|name=%s|owner=%d", limit, offset, id, name, ownerID))
	return cacher.Of[*ListBusinessesResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*ListBusinessesResult, error) {
			rows, total, err := s.repo.ListBusinesses(ctx, limit, offset, id, name, ownerID)
			if err != nil {
				return nil, err
			}
			dtos := make([]BusinessDTO, 0, len(rows))
			for _, b := range rows {
				dto := BusinessDTO{ID: b.ID, Name: b.Name, OwnerID: b.OwnerID}
				if b.CreatedAt.Valid {
					dto.CreatedAt = b.CreatedAt.Time.Format(time.RFC3339)
				}
				dtos = append(dtos, dto)
			}
			return &ListBusinessesResult{Businesses: dtos, Total: total}, nil
		},
		cacher.WithExpiry(ctx, consoleTTL),
	)
}

func (s *Service) GetBusiness(ctx context.Context, id int64) (*BusinessDTO, error) {
	b, err := s.repo.GetBusinessByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := BusinessDTO{ID: b.ID, Name: b.Name, OwnerID: b.OwnerID}
	if b.CreatedAt.Valid {
		dto.CreatedAt = b.CreatedAt.Time.Format(time.RFC3339)
	}
	return &dto, nil
}

func (s *Service) UpdateBusiness(ctx context.Context, id int64, name string) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.UpdateBusiness(ctx, id, name)
}

func (s *Service) DeleteBusiness(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.DeleteBusiness(ctx, id)
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

type SubscriptionDTO struct {
	ID                   int64   `json:"id"`
	OwnerID              int64   `json:"owner_id"`
	OwnerName            string  `json:"owner_name"`
	OwnerEmail           string  `json:"owner_email"`
	BusinessID           *int64  `json:"business_id,omitempty"`
	BusinessName         string  `json:"business_name"`
	PlanID               int64   `json:"plan_id"`
	PlanName             string  `json:"plan_name"`
	Status               string  `json:"status"`
	SeatCount            int     `json:"seat_count"`
	ActiveSeatCount      int     `json:"active_seat_count"`
	IncludedSeatLimit    int     `json:"included_seat_limit"`
	IncludedSeatCount    int     `json:"included_seat_count"`
	BillableSeatCount    int     `json:"billable_seat_count"`
	TrialEndsAt          *string `json:"trial_ends_at,omitempty"`
	NextBillingDate      *string `json:"next_billing_date,omitempty"`
	CurrentPeriodEndsAt  *string `json:"current_period_ends_at,omitempty"`
	PaddleSubscriptionID *string `json:"paddle_subscription_id,omitempty"`
	PaddleCustomerID     *string `json:"paddle_customer_id,omitempty"`
	DaysRemaining        *int    `json:"days_remaining,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

type ListSubscriptionsResult struct {
	Subscriptions []SubscriptionDTO `json:"subscriptions"`
	Total         int64             `json:"total"`
}

func toSubDTO(s infrastructure.SubscriptionRow) SubscriptionDTO {
	dto := SubscriptionDTO{
		ID: s.ID, OwnerID: s.OwnerID, PlanID: s.PlanID,
		PlanName: s.PlanName, Status: s.Status, SeatCount: s.SeatCount,
		ActiveSeatCount: s.ActiveSeatCount, IncludedSeatLimit: freeIncludedSeatLimit,
		IncludedSeatCount: s.IncludedSeatCount, BillableSeatCount: s.BillableSeatCount,
	}
	if s.OwnerName.Valid {
		dto.OwnerName = s.OwnerName.String
	}
	if s.OwnerEmail.Valid {
		dto.OwnerEmail = s.OwnerEmail.String
	}
	if s.BusinessID.Valid {
		id := s.BusinessID.Int64
		dto.BusinessID = &id
	}
	if s.BusinessName.Valid {
		dto.BusinessName = s.BusinessName.String
	}
	if s.TrialEndsAt.Valid {
		t := s.TrialEndsAt.Time.Format(time.RFC3339)
		dto.TrialEndsAt = &t
	}
	if s.NextBillingDate.Valid {
		t := s.NextBillingDate.Time.Format("2006-01-02")
		dto.NextBillingDate = &t
	}
	if s.CurrentPeriodEndsAt.Valid {
		t := s.CurrentPeriodEndsAt.Time.Format(time.RFC3339)
		dto.CurrentPeriodEndsAt = &t
		days := int(time.Until(s.CurrentPeriodEndsAt.Time).Hours() / 24)
		if days < 0 {
			days = 0
		}
		dto.DaysRemaining = &days
	}
	if s.PaddleSubscriptionID.Valid {
		dto.PaddleSubscriptionID = &s.PaddleSubscriptionID.String
	}
	if s.PaddleCustomerID.Valid {
		dto.PaddleCustomerID = &s.PaddleCustomerID.String
	}
	if s.CreatedAt.Valid {
		dto.CreatedAt = s.CreatedAt.Time.Format(time.RFC3339)
	}
	if s.UpdatedAt.Valid {
		dto.UpdatedAt = s.UpdatedAt.Time.Format(time.RFC3339)
	}
	return dto
}

func (s *Service) ListSubscriptions(ctx context.Context, limit, offset int, status string) (*ListSubscriptionsResult, error) {
	key := consoleKey(fmt.Sprintf("subs|l=%d|o=%d|s=%s", limit, offset, status))
	return cacher.Of[*ListSubscriptionsResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*ListSubscriptionsResult, error) {
			return s.listSubscriptionsImpl(ctx, limit, offset, status)
		},
		cacher.WithExpiry(ctx, consoleTTL),
	)
}

func (s *Service) listSubscriptionsImpl(ctx context.Context, limit, offset int, status string) (*ListSubscriptionsResult, error) {
	rows, total, err := s.repo.ListSubscriptions(ctx, limit, offset, status)
	if err != nil {
		return nil, err
	}
	dtos := make([]SubscriptionDTO, 0, len(rows))
	for _, sub := range rows {
		dtos = append(dtos, toSubDTO(sub))
	}
	return &ListSubscriptionsResult{Subscriptions: dtos, Total: total}, nil
}

func (s *Service) GetSubscription(ctx context.Context, id int64) (*SubscriptionDTO, error) {
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toSubDTO(*sub)
	return &dto, nil
}
