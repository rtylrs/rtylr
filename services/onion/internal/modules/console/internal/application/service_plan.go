package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
)

func (s *Service) UpdatePlan(ctx context.Context, id int64, name string, pricePerSeatCents, annualPricePerSeatCents int, isActive bool) error {
	return s.repo.UpdatePlan(ctx, id, name, pricePerSeatCents, annualPricePerSeatCents, isActive)
}

// ── Grants ────────────────────────────────────────────────────────────────────

type GrantDTO struct {
	ID                int64   `json:"id"`
	UserID            int64   `json:"user_id"`
	UserName          string  `json:"user_name"`
	UserEmail         string  `json:"user_email"`
	BusinessID        *int64  `json:"business_id,omitempty"`
	BusinessName      string  `json:"business_name"`
	ActiveSeatCount   int     `json:"active_seat_count"`
	IncludedSeatLimit int     `json:"included_seat_limit"`
	PlanID            int64   `json:"plan_id"`
	PlanName          string  `json:"plan_name"`
	GrantedBy         int64   `json:"granted_by"`
	GranterName       string  `json:"granter_name"`
	Note              string  `json:"note"`
	ExpiresAt         *string `json:"expires_at,omitempty"`
	RevokedAt         *string `json:"revoked_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

type ListGrantsResult struct {
	Grants []GrantDTO `json:"grants"`
	Total  int64      `json:"total"`
}

func toGrantDTO(g infrastructure.GrantRow) GrantDTO {
	dto := GrantDTO{
		ID: g.ID, UserID: g.UserID, PlanID: g.PlanID, PlanName: g.PlanName, GrantedBy: g.GrantedBy,
		ActiveSeatCount: g.ActiveSeatCount, IncludedSeatLimit: freeIncludedSeatLimit,
	}
	if g.UserName.Valid {
		dto.UserName = g.UserName.String
	}
	if g.UserEmail.Valid {
		dto.UserEmail = g.UserEmail.String
	}
	if g.BusinessID.Valid {
		id := g.BusinessID.Int64
		dto.BusinessID = &id
	}
	if g.BusinessName.Valid {
		dto.BusinessName = g.BusinessName.String
	}
	if g.GranterName.Valid {
		dto.GranterName = g.GranterName.String
	}
	if g.Note.Valid {
		dto.Note = g.Note.String
	}
	if g.ExpiresAt.Valid {
		t := g.ExpiresAt.Time.Format(time.RFC3339)
		dto.ExpiresAt = &t
	}
	if g.RevokedAt.Valid {
		t := g.RevokedAt.Time.Format(time.RFC3339)
		dto.RevokedAt = &t
	}
	if g.CreatedAt.Valid {
		dto.CreatedAt = g.CreatedAt.Time.Format(time.RFC3339)
	}
	return dto
}

func (s *Service) ListGrants(ctx context.Context, limit, offset int) (*ListGrantsResult, error) {
	rows, total, err := s.repo.ListGrants(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	dtos := make([]GrantDTO, 0, len(rows))
	for _, g := range rows {
		dtos = append(dtos, toGrantDTO(g))
	}
	return &ListGrantsResult{Grants: dtos, Total: total}, nil
}

func (s *Service) CreateGrant(ctx context.Context, userID, planID, grantedBy int64, note string, expiresAt *time.Time) (*GrantDTO, error) {
	defer invalidateConsoleScope(ctx)
	defer invalidateOwnerSubscriptionScope(ctx, userID)
	seatCount, err := s.activeSeatCountForOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.CreateGrantAndActivateSubscription(ctx, userID, planID, grantedBy, note, expiresAt, infrastructure.SubscriptionWriteParams{
		OwnerID:             userID,
		PlanID:              planID,
		Status:              "active",
		SeatCount:           seatCount,
		CurrentPeriodEndsAt: nullableInputTime(expiresAt),
	})
	if err != nil {
		return nil, err
	}

	grant, err := s.repo.GetGrant(ctx, id)
	if err != nil {
		return &GrantDTO{ID: id}, nil
	}
	dto := toGrantDTO(*grant)
	return &dto, nil
}

func (s *Service) RevokeGrant(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	if ownerID, err := s.repo.GetGrantUserID(ctx, id); err == nil {
		defer invalidateOwnerSubscriptionScope(ctx, ownerID)
	}
	return s.repo.RevokeGrant(ctx, id)
}

func (s *Service) DeleteGrant(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	if ownerID, err := s.repo.GetGrantUserID(ctx, id); err == nil {
		defer invalidateOwnerSubscriptionScope(ctx, ownerID)
	}
	return s.repo.DeleteGrant(ctx, id)
}

// ── Addon Grants ──────────────────────────────────────────────────────────────

type AddonDTO struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Price       money.Amount `json:"price"`
	IsFree      bool         `json:"is_free"`
	SortOrder   int64        `json:"sort_order"`
}

type AddonGrantDTO struct {
	ID             int64        `json:"id"`
	UserID         int64        `json:"user_id"`
	UserName       string       `json:"user_name"`
	UserEmail      string       `json:"user_email"`
	BusinessID     int64        `json:"business_id"`
	BusinessName   string       `json:"business_name"`
	AddonID        string       `json:"addon_id"`
	AddonName      string       `json:"addon_name"`
	AddonPrice     money.Amount `json:"addon_price"`
	GrantedBy      int64        `json:"granted_by"`
	GranterName    string       `json:"granter_name"`
	Note           string       `json:"note"`
	ExpiresAt      *string      `json:"expires_at,omitempty"`
	RevokedAt      *string      `json:"revoked_at,omitempty"`
	EnabledByGrant bool         `json:"enabled_by_grant"`
	CreatedAt      string       `json:"created_at"`
}

type ListAddonsResult struct {
	Addons []AddonDTO `json:"addons"`
}

type ListAddonGrantsResult struct {
	Grants []AddonGrantDTO `json:"grants"`
	Total  int64           `json:"total"`
}

func decimalStringToMoney(value string) money.Amount {
	if value == "" {
		return 0
	}
	parsed, err := money.FromString(value)
	if err != nil {
		return 0
	}
	return parsed
}

func toAddonDTO(a infrastructure.AddonCatalogRow) AddonDTO {
	dto := AddonDTO{
		ID:    a.ID,
		Name:  a.Name,
		Price: decimalStringToMoney(a.Price.String),
	}
	if a.Description.Valid {
		dto.Description = a.Description.String
	}
	if a.Icon.Valid {
		dto.Icon = a.Icon.String
	}
	if a.IsFree.Valid {
		dto.IsFree = a.IsFree.Bool
	}
	if a.SortOrder.Valid {
		dto.SortOrder = a.SortOrder.Int64
	}
	return dto
}
