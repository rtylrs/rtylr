package application

import (
	"context"
	"errors"
	"time"
)

func toAddonGrantDTO(g AddonGrantRow) AddonGrantDTO {
	dto := AddonGrantDTO{
		ID:             g.ID,
		UserID:         g.UserID,
		BusinessID:     g.BusinessID,
		AddonID:        g.AddonID,
		AddonName:      g.AddonName,
		AddonPrice:     decimalStringToMoney(g.AddonPrice.String),
		GrantedBy:      g.GrantedBy,
		EnabledByGrant: g.EnabledByGrant,
	}
	if g.UserName.Valid {
		dto.UserName = g.UserName.String
	}
	if g.UserEmail.Valid {
		dto.UserEmail = g.UserEmail.String
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

func (s *Service) ListAddons(ctx context.Context) (*ListAddonsResult, error) {
	rows, err := s.repo.ListAddons(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]AddonDTO, 0, len(rows))
	for _, a := range rows {
		dtos = append(dtos, toAddonDTO(a))
	}
	return &ListAddonsResult{Addons: dtos}, nil
}

func (s *Service) ListAddonGrants(ctx context.Context, limit, offset int) (*ListAddonGrantsResult, error) {
	rows, total, err := s.repo.ListAddonGrants(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	dtos := make([]AddonGrantDTO, 0, len(rows))
	for _, g := range rows {
		dtos = append(dtos, toAddonGrantDTO(g))
	}
	return &ListAddonGrantsResult{Grants: dtos, Total: total}, nil
}

func (s *Service) resolveAddonGrantBusiness(ctx context.Context, userID, businessID int64) (*BusinessRow, error) {
	if businessID > 0 {
		business, err := s.repo.GetBusinessByID(ctx, businessID)
		if err != nil {
			return nil, err
		}
		if business.OwnerID != userID {
			return nil, ErrBusinessNotOwnedByUser
		}
		return business, nil
	}
	business, err := s.repo.GetActiveBusinessByOwner(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrNotOrganizationOwner
		}
		return nil, err
	}
	return business, nil
}

func (s *Service) CreateAddonGrant(ctx context.Context, userID, businessID int64, addonID string, grantedBy int64, note string, expiresAt *time.Time) (*AddonGrantDTO, error) {
	defer invalidateConsoleScope(ctx)
	defer invalidateOwnerSubscriptionScope(ctx, userID)
	exists, err := s.repo.AddonExists(ctx, addonID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAddonNotFound
	}

	business, err := s.resolveAddonGrantBusiness(ctx, userID, businessID)
	if err != nil {
		return nil, err
	}
	defer invalidateBusinessAddonScope(ctx, business.ID)

	wasEnabled, err := s.repo.IsBusinessAddonEnabled(ctx, business.ID, addonID)
	if err != nil {
		return nil, err
	}
	enabledByGrant := !wasEnabled

	id, err := s.repo.CreateAddonGrant(ctx, userID, business.ID, addonID, grantedBy, note, expiresAt, enabledByGrant)
	if err != nil {
		return nil, err
	}
	if err := s.repo.EnableBusinessAddon(ctx, business.ID, addonID); err != nil {
		return nil, err
	}

	grant, err := s.repo.GetAddonGrant(ctx, id)
	if err != nil {
		return &AddonGrantDTO{ID: id}, nil
	}
	dto := toAddonGrantDTO(*grant)
	return &dto, nil
}

func (s *Service) cleanupAddonGrantEnablement(ctx context.Context, scope *AddonGrantScopeRow) error {
	if scope == nil || !scope.EnabledByGrant {
		return nil
	}
	hasActiveGrant, err := s.repo.HasActiveAddonGrant(ctx, scope.BusinessID, scope.AddonID)
	if err != nil {
		return err
	}
	if hasActiveGrant {
		return nil
	}
	return s.repo.DisableBusinessAddon(ctx, scope.BusinessID, scope.AddonID)
}

// adoptAddonGrantEnablement hands the "this grant turned the add-on on" marker
// to a surviving active grant before the owning grant is revoked or deleted.
// Without it the last remaining grant could no longer turn the add-on off, and
// the enablement would outlive every grant.
func (s *Service) adoptAddonGrantEnablement(ctx context.Context, scope *AddonGrantScopeRow, removedID int64) error {
	if scope == nil || !scope.EnabledByGrant {
		return nil
	}
	hasActiveGrant, err := s.repo.HasActiveAddonGrant(ctx, scope.BusinessID, scope.AddonID)
	if err != nil {
		return err
	}
	if !hasActiveGrant {
		return nil
	}
	return s.repo.AdoptAddonGrantEnablement(ctx, scope.BusinessID, scope.AddonID, removedID)
}

func (s *Service) RevokeAddonGrant(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	scope, err := s.repo.GetAddonGrantScope(ctx, id)
	if err != nil {
		return err
	}
	defer invalidateOwnerSubscriptionScope(ctx, scope.UserID)
	defer invalidateBusinessAddonScope(ctx, scope.BusinessID)
	if err := s.repo.RevokeAddonGrant(ctx, id); err != nil {
		return err
	}
	if err := s.adoptAddonGrantEnablement(ctx, scope, id); err != nil {
		return err
	}
	return s.cleanupAddonGrantEnablement(ctx, scope)
}

func (s *Service) DeleteAddonGrant(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	scope, err := s.repo.GetAddonGrantScope(ctx, id)
	if err != nil {
		return err
	}
	defer invalidateOwnerSubscriptionScope(ctx, scope.UserID)
	defer invalidateBusinessAddonScope(ctx, scope.BusinessID)
	if err := s.repo.DeleteAddonGrant(ctx, id); err != nil {
		return err
	}
	if err := s.adoptAddonGrantEnablement(ctx, scope, id); err != nil {
		return err
	}
	return s.cleanupAddonGrantEnablement(ctx, scope)
}

// ── Stats ─────────────────────────────────────────────────────────────────────

type StatsDTO struct {
	TotalUsers          int64 `json:"total_users"`
	BannedUsers         int64 `json:"banned_users"`
	TotalBusinesses     int64 `json:"total_businesses"`
	ActiveSubs          int64 `json:"active_subs"`
	ActiveSubscriptions int64 `json:"active_subscriptions"`
	TrialSubs           int64 `json:"trial_subs"`
	ManualGrants        int64 `json:"manual_grants"`
	ManualAddonGrants   int64 `json:"manual_addon_grants"`
	TotalGrants         int64 `json:"total_grants"`
	ConsoleAdmins       int64 `json:"console_admins"`
}

func (s *Service) GetStats(ctx context.Context) (*StatsDTO, error) {
	// console_stats is maintained transactionally by the database. Reading its
	// single row is cheaper and fresher than caching a set of global counts.
	row, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &StatsDTO{
		TotalUsers:          row.TotalUsers,
		BannedUsers:         row.BannedUsers,
		TotalBusinesses:     row.TotalBusinesses,
		ActiveSubs:          row.ActiveSubs,
		ActiveSubscriptions: row.ActiveSubs,
		TrialSubs:           row.TrialSubs,
		ManualGrants:        row.ManualGrants,
		ManualAddonGrants:   row.ManualAddonGrants,
		TotalGrants:         row.ManualGrants,
		ConsoleAdmins:       row.ConsoleAdmins,
	}, nil
}
