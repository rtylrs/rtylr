package infrastructure

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure/persistence/sqlc"
)

type UserRow = application.UserRow
type NullString = application.NullString
type NullInt64 = application.NullInt64
type NullBool = application.NullBool
type NullTime = application.NullTime
type BusinessRow = application.BusinessRow
type CreateBusinessOwnerParams = application.CreateBusinessOwnerParams
type CreateBusinessOwnerRow = application.CreateBusinessOwnerRow
type SubscriptionRow = application.SubscriptionRow
type SubscriptionWriteParams = application.SubscriptionWriteParams
type PlanRow = application.PlanRow
type GrantRow = application.GrantRow
type AddonCatalogRow = application.AddonCatalogRow
type AddonGrantRow = application.AddonGrantRow
type AddonGrantScopeRow = application.AddonGrantScopeRow
type StatsRow = application.StatsRow
type ResourceTableRow = application.ResourceTableRow
type ResourceColumnRow = application.ResourceColumnRow
type ResourceRowsResult = application.ResourceRowsResult

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
