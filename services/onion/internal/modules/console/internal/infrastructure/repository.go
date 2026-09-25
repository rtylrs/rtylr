package infrastructure

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure/persistence/sqlc"
)

type UserRow struct {
	ID             int64
	Email          string
	FullName       sql.NullString
	BusinessID     sql.NullInt64
	IsConsoleAdmin bool
	IsBanned       bool
	BanReason      sql.NullString
	BannedAt       sql.NullTime
	CreatedAt      sql.NullTime
	PasswordHash   string
}

type BusinessRow struct {
	ID        int64
	Name      string
	OwnerID   int64
	CreatedAt sql.NullTime
}

type CreateBusinessOwnerParams struct {
	Email              string
	PasswordHash       string
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

type CreateBusinessOwnerRow struct {
	OwnerID        int64
	BusinessID     int64
	SubscriptionID sql.NullInt64
}

type SubscriptionRow struct {
	ID                   int64
	OwnerID              int64
	OwnerName            sql.NullString
	PlanID               int64
	PlanName             string
	Status               string
	SeatCount            int
	ActiveSeatCount      int
	IncludedSeatCount    int
	BillableSeatCount    int
	TrialEndsAt          sql.NullTime
	NextBillingDate      sql.NullTime
	CurrentPeriodEndsAt  sql.NullTime
	PaddleSubscriptionID sql.NullString
	PaddleCustomerID     sql.NullString
	CreatedAt            sql.NullTime
	UpdatedAt            sql.NullTime
	OwnerEmail           sql.NullString
	BusinessID           sql.NullInt64
	BusinessName         sql.NullString
}

type SubscriptionWriteParams struct {
	OwnerID              int64
	PlanID               int64
	Status               string
	SeatCount            int
	TrialEndsAt          sql.NullTime
	NextBillingDate      sql.NullTime
	CurrentPeriodEndsAt  sql.NullTime
	PaddleSubscriptionID sql.NullString
	PaddleCustomerID     sql.NullString
}

type PlanRow struct {
	ID                      int64
	Name                    string
	Slug                    string
	PricePerSeatCents       int
	AnnualPricePerSeatCents int
	TrialDays               int
	MaxSeats                sql.NullInt64
	MaxLocations            sql.NullInt64
	IsActive                bool
	IsSolo                  bool
	IsFeatured              bool
	TierOrder               int
	Tagline                 string
	CreatedAt               sql.NullTime
}

type GrantRow struct {
	ID              int64
	UserID          int64
	UserName        sql.NullString
	PlanID          int64
	PlanName        string
	GrantedBy       int64
	GranterName     sql.NullString
	Note            sql.NullString
	ExpiresAt       sql.NullTime
	RevokedAt       sql.NullTime
	CreatedAt       sql.NullTime
	UserEmail       sql.NullString
	BusinessID      sql.NullInt64
	BusinessName    sql.NullString
	ActiveSeatCount int
}

type AddonCatalogRow struct {
	ID          string
	Name        string
	Description sql.NullString
	Icon        sql.NullString
	Price       sql.NullString
	IsFree      sql.NullBool
	SortOrder   sql.NullInt64
}

type AddonGrantRow struct {
	ID             int64
	UserID         int64
	UserName       sql.NullString
	UserEmail      sql.NullString
	BusinessID     int64
	BusinessName   sql.NullString
	AddonID        string
	AddonName      string
	AddonPrice     sql.NullString
	GrantedBy      int64
	GranterName    sql.NullString
	Note           sql.NullString
	ExpiresAt      sql.NullTime
	RevokedAt      sql.NullTime
	EnabledByGrant bool
	CreatedAt      sql.NullTime
}

type AddonGrantScopeRow struct {
	UserID         int64
	BusinessID     int64
	AddonID        string
	EnabledByGrant bool
}

type StatsRow struct {
	TotalUsers        int64
	BannedUsers       int64
	TotalBusinesses   int64
	ActiveSubs        int64
	TrialSubs         int64
	ManualGrants      int64
	ManualAddonGrants int64
	ConsoleAdmins     int64
}

type ResourceTableRow struct {
	Name    string
	Comment sql.NullString
}

type ResourceColumnRow struct {
	Name       string
	DataType   string
	IsNullable bool
	IsPrimary  bool
	Default    sql.NullString
	Extra      string
	Comment    sql.NullString
}

type ResourceRowsResult struct {
	Columns    []ResourceColumnRow
	PrimaryKey string
	Rows       []map[string]any
	Total      int64
}

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

var ErrResourceNotAllowed = errors.New("console resource is not allowlisted")

// Application-owned schema prefixes are intentionally narrow; system tables
// and unprefixed tables must be added explicitly to the exact-name set below.
var allowedResourceTablePrefixes = []string{
	"booking_", "careers_", "crm_", "erp_", "finance_", "flow_", "hr_",
	"oauth_", "pos_", "store_", "support_", "voxire_", "rtylr_",
}

var allowedResourceTables = map[string]struct{}{
	"access_requests": {}, "activity_log": {}, "addons": {}, "businesses": {},
	"chat_sessions": {}, "clients": {}, "config": {}, "conversation_messages": {},
	"conversation_summaries": {}, "custom_domains": {}, "email_jobs": {},
	"expenses": {}, "faq": {}, "feature_flags": {}, "users": {},
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
