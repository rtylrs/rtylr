package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	consoleTTL            = 1 * time.Minute
	freeIncludedSeatLimit = 10
)

var (
	ErrNotOrganizationOwner   = errors.New("access can only be granted to organization owners")
	ErrBusinessNotOwnedByUser = errors.New("business does not belong to the requested owner")
	ErrAddonNotFound          = errors.New("addon not found")
	ErrInvalidSubscription    = errors.New("invalid subscription input")
)

// Console is a global admin surface — caches are keyed on the admin namespace
// rather than per-owner because they aggregate across all tenants.
func consoleKey(detail string) string {
	return "cache:console:" + detail
}

func invalidateConsoleScope(ctx context.Context) {
	cacher.InvalidateScope(ctx, "cache:console:*")
}

func invalidateOwnerSubscriptionScope(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx,
		cacheKeys.SubscriptionPrefix(ownerID),
		cacheKeys.UserPrefix(ownerID),
	)
}

func invalidateBusinessAddonScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx,
		cacheKeys.ScopedPrefix("addons", businessID),
		cacheKeys.BusinessPrefix(businessID),
	)
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
	defer invalidateConsoleScope(ctx)

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

type CreateBusinessOwnerInput struct {
	Email              string
	Password           string
	FullName           string
	BusinessName       string
	Industry           string
	Country            string
	Timezone           string
	Currency           string
	GrantPlan          bool
	PlanID             int64
	SubscriptionStatus string
	GrantDays          int
}

type CreateBusinessOwnerResult struct {
	Owner        UserDTO          `json:"owner"`
	Business     BusinessDTO      `json:"business"`
	Subscription *SubscriptionDTO `json:"subscription,omitempty"`
	Message      string           `json:"message"`
}

func (s *Service) CreateBusinessOwner(ctx context.Context, input CreateBusinessOwnerInput) (*CreateBusinessOwnerResult, error) {
	defer invalidateConsoleScope(ctx)

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return nil, errors.New("owner email is required")
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, errors.New("password is required")
	}
	if len(input.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	businessName := strings.TrimSpace(input.BusinessName)
	if businessName == "" {
		return nil, errors.New("business name is required")
	}

	timezone := strings.TrimSpace(input.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "USD"
	}

	status := strings.ToLower(strings.TrimSpace(input.SubscriptionStatus))
	if status == "" {
		if input.GrantDays > 0 {
			status = "trial"
		} else {
			status = "active"
		}
	}
	if status != "active" && status != "trial" {
		return nil, errors.New("subscription status must be active or trial")
	}
	if input.GrantDays < 0 {
		return nil, errors.New("grant days cannot be negative")
	}
	if input.GrantPlan && input.PlanID <= 0 {
		return nil, errors.New("plan is required when granting access")
	}

	if _, err := s.repo.GetUserByEmail(ctx, email); err == nil {
		return nil, errors.New("owner email already exists")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if _, total, err := s.repo.ListBusinesses(ctx, 1, 0, 0, businessName, 0); err != nil {
		return nil, err
	} else if total > 0 {
		return nil, errors.New("business name already exists")
	}

	hash, err := passwordutils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.CreateBusinessOwner(ctx, infrastructure.CreateBusinessOwnerParams{
		Email:              email,
		PasswordHash:       hash,
		FullName:           strings.TrimSpace(input.FullName),
		BusinessName:       businessName,
		Industry:           strings.TrimSpace(input.Industry),
		Country:            strings.TrimSpace(input.Country),
		Timezone:           timezone,
		Currency:           currency,
		GrantPlan:          input.GrantPlan,
		PlanID:             input.PlanID,
		SubscriptionStatus: status,
		GrantDays:          input.GrantDays,
	})
	if err != nil {
		return nil, err
	}

	owner, err := s.GetUser(ctx, created.OwnerID)
	if err != nil {
		return nil, err
	}
	business, err := s.GetBusiness(ctx, created.BusinessID)
	if err != nil {
		return nil, err
	}

	var subscription *SubscriptionDTO
	if created.SubscriptionID.Valid {
		sub, subErr := s.GetSubscriptionByOwner(ctx, created.OwnerID)
		if subErr != nil {
			return nil, subErr
		}
		subscription = sub
	}

	return &CreateBusinessOwnerResult{
		Owner:        *owner,
		Business:     *business,
		Subscription: subscription,
		Message:      "Owner and business created",
	}, nil
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

func nullableInputTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}

func nullableInputString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
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

func (s *Service) subscriptionWriteParams(ctx context.Context, input SubscriptionWriteInput) (infrastructure.SubscriptionWriteParams, error) {
	if input.OwnerID <= 0 || input.PlanID <= 0 {
		return infrastructure.SubscriptionWriteParams{}, fmt.Errorf("%w: owner_id and plan_id are required", ErrInvalidSubscription)
	}
	if _, err := s.repo.GetActiveBusinessByOwner(ctx, input.OwnerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return infrastructure.SubscriptionWriteParams{}, ErrNotOrganizationOwner
		}
		return infrastructure.SubscriptionWriteParams{}, err
	}

	status, err := normalizeSubscriptionStatus(input.Status)
	if err != nil {
		return infrastructure.SubscriptionWriteParams{}, err
	}

	seatCount := input.SeatCount
	if seatCount <= 0 {
		seatCount, err = s.repo.CountActiveSeatsForOwner(ctx, input.OwnerID)
		if err != nil {
			return infrastructure.SubscriptionWriteParams{}, err
		}
		if seatCount < 1 {
			seatCount = 1
		}
	}

	trialEndsAt := input.TrialEndsAt
	currentPeriodEndsAt := input.CurrentPeriodEndsAt
	if input.TotalDays != nil {
		if *input.TotalDays < 0 {
			return infrastructure.SubscriptionWriteParams{}, fmt.Errorf("%w: total_days cannot be negative", ErrInvalidSubscription)
		}
		endsAt := time.Now().UTC().AddDate(0, 0, *input.TotalDays)
		currentPeriodEndsAt = &endsAt
		if status == "trial" {
			trialEndsAt = &endsAt
		}
	}

	return infrastructure.SubscriptionWriteParams{
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
		if errors.Is(err, sql.ErrNoRows) {
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

	id, err := s.repo.CreateGrant(ctx, userID, planID, grantedBy, note, expiresAt)
	if err != nil {
		return nil, err
	}

	// Also activate the subscription for that owner.
	if subErr := s.repo.UpsertSubscription(ctx, infrastructure.SubscriptionWriteParams{
		OwnerID:             userID,
		PlanID:              planID,
		Status:              "active",
		SeatCount:           seatCount,
		CurrentPeriodEndsAt: nullableInputTime(expiresAt),
	}); subErr != nil {
		logger.Ctx(ctx).Warn("failed to upsert subscription for grant", logger.Error(subErr))
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
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Price       money.Amount `json:"price"`
	IsFree      bool    `json:"is_free"`
	SortOrder   int64   `json:"sort_order"`
}

type AddonGrantDTO struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	UserName       string  `json:"user_name"`
	UserEmail      string  `json:"user_email"`
	BusinessID     int64   `json:"business_id"`
	BusinessName   string  `json:"business_name"`
	AddonID        string  `json:"addon_id"`
	AddonName      string  `json:"addon_name"`
	AddonPrice     money.Amount `json:"addon_price"`
	GrantedBy      int64   `json:"granted_by"`
	GranterName    string  `json:"granter_name"`
	Note           string  `json:"note"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	RevokedAt      *string `json:"revoked_at,omitempty"`
	EnabledByGrant bool    `json:"enabled_by_grant"`
	CreatedAt      string  `json:"created_at"`
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

func toAddonGrantDTO(g infrastructure.AddonGrantRow) AddonGrantDTO {
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

func (s *Service) resolveAddonGrantBusiness(ctx context.Context, userID, businessID int64) (*infrastructure.BusinessRow, error) {
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
		if errors.Is(err, sql.ErrNoRows) {
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

func (s *Service) cleanupAddonGrantEnablement(ctx context.Context, scope *infrastructure.AddonGrantScopeRow) error {
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

func (s *Service) RevokeAddonGrant(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	scope, err := s.repo.GetAddonGrantScope(ctx, id)
	if err == nil {
		defer invalidateOwnerSubscriptionScope(ctx, scope.UserID)
		defer invalidateBusinessAddonScope(ctx, scope.BusinessID)
	}
	if err := s.repo.RevokeAddonGrant(ctx, id); err != nil {
		return err
	}
	return s.cleanupAddonGrantEnablement(ctx, scope)
}

func (s *Service) DeleteAddonGrant(ctx context.Context, id int64) error {
	defer invalidateConsoleScope(ctx)
	scope, err := s.repo.GetAddonGrantScope(ctx, id)
	if err == nil {
		defer invalidateOwnerSubscriptionScope(ctx, scope.UserID)
		defer invalidateBusinessAddonScope(ctx, scope.BusinessID)
	}
	if err := s.repo.DeleteAddonGrant(ctx, id); err != nil {
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
