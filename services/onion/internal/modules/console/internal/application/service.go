package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const consoleTTL = 1 * time.Minute

// Console is a global admin surface — caches are keyed on the admin namespace
// rather than per-owner because they aggregate across all tenants.
func consoleKey(detail string) string {
	return "cache:console:" + detail
}

func invalidateConsoleScope(ctx context.Context) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, "cache:console:*")
	}
}

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// ── Generic Resources ────────────────────────────────────────────────────────

type ResourceTableDTO struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

type ResourceColumnDTO struct {
	Name       string  `json:"name"`
	DataType   string  `json:"data_type"`
	IsNullable bool    `json:"is_nullable"`
	IsPrimary  bool    `json:"is_primary"`
	Default    *string `json:"default,omitempty"`
	Extra      string  `json:"extra"`
	Comment    string  `json:"comment"`
	Writable   bool    `json:"writable"`
}

type ListResourceTablesResult struct {
	Tables []ResourceTableDTO `json:"tables"`
}

type DescribeResourceResult struct {
	Table      string              `json:"table"`
	PrimaryKey string              `json:"primary_key"`
	Columns    []ResourceColumnDTO `json:"columns"`
}

type ListResourceRowsResult struct {
	Table      string              `json:"table"`
	PrimaryKey string              `json:"primary_key"`
	Columns    []ResourceColumnDTO `json:"columns"`
	Rows       []map[string]any    `json:"rows"`
	Total      int64               `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

type GetResourceRowResult struct {
	Table      string              `json:"table"`
	PrimaryKey string              `json:"primary_key"`
	Columns    []ResourceColumnDTO `json:"columns"`
	Row        map[string]any      `json:"row"`
}

func toResourceColumnDTO(c infrastructure.ResourceColumnRow) ResourceColumnDTO {
	dto := ResourceColumnDTO{
		Name:       c.Name,
		DataType:   c.DataType,
		IsNullable: c.IsNullable,
		IsPrimary:  c.IsPrimary,
		Extra:      c.Extra,
		Writable:   !c.IsPrimary && !strings.Contains(strings.ToLower(c.Extra), "auto_increment") && !strings.Contains(strings.ToLower(c.Extra), "generated"),
	}
	if c.Default.Valid {
		dto.Default = &c.Default.String
	}
	if c.Comment.Valid {
		dto.Comment = c.Comment.String
	}
	return dto
}

func toResourceColumnDTOs(columns []infrastructure.ResourceColumnRow) []ResourceColumnDTO {
	dtos := make([]ResourceColumnDTO, 0, len(columns))
	for _, col := range columns {
		dtos = append(dtos, toResourceColumnDTO(col))
	}
	return dtos
}

func (s *Service) ListResources(ctx context.Context) (*ListResourceTablesResult, error) {
	rows, err := s.repo.ListResourceTables(ctx)
	if err != nil {
		return nil, err
	}
	tables := make([]ResourceTableDTO, 0, len(rows))
	for _, row := range rows {
		dto := ResourceTableDTO{Name: row.Name}
		if row.Comment.Valid {
			dto.Comment = row.Comment.String
		}
		tables = append(tables, dto)
	}
	return &ListResourceTablesResult{Tables: tables}, nil
}

func (s *Service) DescribeResource(ctx context.Context, table string) (*DescribeResourceResult, error) {
	columns, primaryKey, err := s.repo.DescribeResourceTable(ctx, table)
	if err != nil {
		return nil, err
	}
	return &DescribeResourceResult{
		Table:      table,
		PrimaryKey: primaryKey,
		Columns:    toResourceColumnDTOs(columns),
	}, nil
}

func (s *Service) ListResourceRows(ctx context.Context, table string, limit, offset int) (*ListResourceRowsResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	result, err := s.repo.ListResourceRows(ctx, table, limit, offset)
	if err != nil {
		return nil, err
	}
	return &ListResourceRowsResult{
		Table:      table,
		PrimaryKey: result.PrimaryKey,
		Columns:    toResourceColumnDTOs(result.Columns),
		Rows:       result.Rows,
		Total:      result.Total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *Service) GetResourceRow(ctx context.Context, table, id string) (*GetResourceRowResult, error) {
	row, columns, primaryKey, err := s.repo.GetResourceRow(ctx, table, id)
	if err != nil {
		return nil, err
	}
	return &GetResourceRowResult{
		Table:      table,
		PrimaryKey: primaryKey,
		Columns:    toResourceColumnDTOs(columns),
		Row:        row,
	}, nil
}

func (s *Service) CreateResourceRow(ctx context.Context, table string, values map[string]any) (int64, error) {
	defer invalidateConsoleScope(ctx)
	return s.repo.CreateResourceRow(ctx, table, values)
}

func (s *Service) UpdateResourceRow(ctx context.Context, table, id string, values map[string]any) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.UpdateResourceRow(ctx, table, id, values)
}

func (s *Service) DeleteResourceRow(ctx context.Context, table, id string) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.DeleteResourceRow(ctx, table, id)
}

// ── Users ─────────────────────────────────────────────────────────────────────

type UserDTO struct {
	ID             int64   `json:"id"`
	Email          string  `json:"email"`
	FullName       string  `json:"full_name"`
	BusinessID     *int64  `json:"business_id,omitempty"`
	IsConsoleAdmin bool    `json:"is_console_admin"`
	IsBanned       bool    `json:"is_banned"`
	BanReason      *string `json:"ban_reason,omitempty"`
	BannedAt       *string `json:"banned_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

func toUserDTO(u infrastructure.UserRow) UserDTO {
	dto := UserDTO{
		ID:             u.ID,
		Email:          u.Email,
		IsConsoleAdmin: u.IsConsoleAdmin,
		IsBanned:       u.IsBanned,
	}
	if u.FullName.Valid {
		dto.FullName = u.FullName.String
	}
	if u.BusinessID.Valid {
		dto.BusinessID = &u.BusinessID.Int64
	}
	if u.BanReason.Valid {
		dto.BanReason = &u.BanReason.String
	}
	if u.BannedAt.Valid {
		t := u.BannedAt.Time.Format(time.RFC3339)
		dto.BannedAt = &t
	}
	if u.CreatedAt.Valid {
		dto.CreatedAt = u.CreatedAt.Time.Format(time.RFC3339)
	}
	return dto
}

type ListUsersResult struct {
	Users []UserDTO `json:"users"`
	Total int64     `json:"total"`
}

func (s *Service) ListUsers(
	ctx context.Context,
	limit, offset int,
	id int64,
	email, fullName string,
) (*ListUsersResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	fullName = strings.TrimSpace(fullName)
	key := consoleKey(fmt.Sprintf("users|l=%d|o=%d|id=%d|email=%s|name=%s", limit, offset, id, email, fullName))
	return cacher.Of[*ListUsersResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*ListUsersResult, error) {
			users, total, err := s.repo.ListUsers(ctx, limit, offset, id, email, fullName)
			if err != nil {
				return nil, err
			}
			dtos := make([]UserDTO, 0, len(users))
			for _, u := range users {
				dtos = append(dtos, toUserDTO(u))
			}
			return &ListUsersResult{Users: dtos, Total: total}, nil
		},
		cacher.WithExpiry(ctx, consoleTTL),
	)
}

func (s *Service) GetUser(ctx context.Context, id int64) (*UserDTO, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toUserDTO(*u)
	return &dto, nil
}

func (s *Service) CreateUser(ctx context.Context, email, password, fullName string, isConsoleAdmin bool) (*UserDTO, error) {
	hash, err := passwordutils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	id, err := s.repo.CreateUser(ctx, email, hash, fullName, isConsoleAdmin)
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, id int64, fullName, email string) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.UpdateUser(ctx, id, fullName, email)
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.DeleteUser(ctx, id)
}

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

func (s *Service) ListBusinesses(ctx context.Context, limit, offset int, search string) (*ListBusinessesResult, error) {
	key := consoleKey(fmt.Sprintf("biz|l=%d|o=%d|q=%s", limit, offset, search))
	return cacher.Of[*ListBusinessesResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*ListBusinessesResult, error) {
			rows, total, err := s.repo.ListBusinesses(ctx, limit, offset, search)
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
	ID                  int64   `json:"id"`
	OwnerID             int64   `json:"owner_id"`
	OwnerEmail          string  `json:"owner_email"`
	PlanID              int64   `json:"plan_id"`
	PlanName            string  `json:"plan_name"`
	Status              string  `json:"status"`
	SeatCount           int     `json:"seat_count"`
	TrialEndsAt         *string `json:"trial_ends_at,omitempty"`
	CurrentPeriodEndsAt *string `json:"current_period_ends_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
}

type ListSubscriptionsResult struct {
	Subscriptions []SubscriptionDTO `json:"subscriptions"`
	Total         int64             `json:"total"`
}

func toSubDTO(s infrastructure.SubscriptionRow) SubscriptionDTO {
	dto := SubscriptionDTO{
		ID: s.ID, OwnerID: s.OwnerID, PlanID: s.PlanID,
		PlanName: s.PlanName, Status: s.Status, SeatCount: s.SeatCount,
	}
	if s.OwnerEmail.Valid {
		dto.OwnerEmail = s.OwnerEmail.String
	}
	if s.TrialEndsAt.Valid {
		t := s.TrialEndsAt.Time.Format(time.RFC3339)
		dto.TrialEndsAt = &t
	}
	if s.CurrentPeriodEndsAt.Valid {
		t := s.CurrentPeriodEndsAt.Time.Format(time.RFC3339)
		dto.CurrentPeriodEndsAt = &t
	}
	if s.CreatedAt.Valid {
		dto.CreatedAt = s.CreatedAt.Time.Format(time.RFC3339)
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

func (s *Service) GetSubscriptionByOwner(ctx context.Context, ownerID int64) (*SubscriptionDTO, error) {
	sub, err := s.repo.GetSubscriptionByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	dto := toSubDTO(*sub)
	return &dto, nil
}

func (s *Service) UpsertSubscription(ctx context.Context, ownerID, planID int64, status string, seatCount int, expiresAt *time.Time) error {
	return s.repo.UpsertSubscription(ctx, ownerID, planID, status, seatCount, expiresAt)
}

func (s *Service) UpdateSubscription(ctx context.Context, id, planID int64, status string, seatCount int) error {
	return s.repo.UpdateSubscription(ctx, id, planID, status, seatCount)
}

func (s *Service) DeleteSubscription(ctx context.Context, id int64) error {
	return s.repo.DeleteSubscription(ctx, id)
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

func (s *Service) UpdatePlan(ctx context.Context, id int64, name string, pricePerSeatCents, annualPricePerSeatCents int, isActive bool) error {
	return s.repo.UpdatePlan(ctx, id, name, pricePerSeatCents, annualPricePerSeatCents, isActive)
}

// ── Grants ────────────────────────────────────────────────────────────────────

type GrantDTO struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	UserEmail   string  `json:"user_email"`
	PlanID      int64   `json:"plan_id"`
	PlanName    string  `json:"plan_name"`
	GrantedBy   int64   `json:"granted_by"`
	GranterName string  `json:"granter_name"`
	Note        string  `json:"note"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	RevokedAt   *string `json:"revoked_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type ListGrantsResult struct {
	Grants []GrantDTO `json:"grants"`
	Total  int64      `json:"total"`
}

func toGrantDTO(g infrastructure.GrantRow) GrantDTO {
	dto := GrantDTO{
		ID: g.ID, UserID: g.UserID, PlanID: g.PlanID, PlanName: g.PlanName, GrantedBy: g.GrantedBy,
	}
	if g.UserEmail.Valid {
		dto.UserEmail = g.UserEmail.String
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
	id, err := s.repo.CreateGrant(ctx, userID, planID, grantedBy, note, expiresAt)
	if err != nil {
		return nil, err
	}

	// Also activate the subscription for that user
	if subErr := s.repo.UpsertSubscription(ctx, userID, planID, "active", 1, expiresAt); subErr != nil {
		logger.Ctx(ctx).Warn("failed to upsert subscription for grant", logger.Error(subErr))
	}

	rows, _, err := s.repo.ListGrants(ctx, 1, 0)
	if err != nil || len(rows) == 0 {
		return &GrantDTO{ID: id}, nil
	}
	// Return the created grant by fetching from the list
	result, _, err2 := s.repo.ListGrants(ctx, 1000, 0)
	if err2 != nil {
		return &GrantDTO{ID: id}, nil
	}
	for _, g := range result {
		if g.ID == id {
			dto := toGrantDTO(g)
			return &dto, nil
		}
	}
	return &GrantDTO{ID: id}, nil
}

func (s *Service) RevokeGrant(ctx context.Context, id int64) error {
	return s.repo.RevokeGrant(ctx, id)
}

func (s *Service) DeleteGrant(ctx context.Context, id int64) error {
	return s.repo.DeleteGrant(ctx, id)
}

// ── Stats ─────────────────────────────────────────────────────────────────────

type StatsDTO struct {
	TotalUsers      int64 `json:"total_users"`
	BannedUsers     int64 `json:"banned_users"`
	TotalBusinesses int64 `json:"total_businesses"`
	ActiveSubs      int64 `json:"active_subs"`
	TrialSubs       int64 `json:"trial_subs"`
	ManualGrants    int64 `json:"manual_grants"`
}

func (s *Service) GetStats(ctx context.Context) (*StatsDTO, error) {
	row, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &StatsDTO{
		TotalUsers:      row.TotalUsers,
		BannedUsers:     row.BannedUsers,
		TotalBusinesses: row.TotalBusinesses,
		ActiveSubs:      row.ActiveSubs,
		TrialSubs:       row.TrialSubs,
		ManualGrants:    row.ManualGrants,
	}, nil
}
