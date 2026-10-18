package scopes

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/utils/passwordutils"

	"github.com/voxire/onion/internal/modules/organization/scopes/internal/persistence/sqlc"
)

// Reader resolves current identity and membership state without caching grants.
type Reader struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewReader(db *sql.DB) *Reader { return &Reader{db: db, queries: sqlc.New(db)} }

// ActiveBusinessOwner returns sql.ErrNoRows for unknown or inactive businesses.
func (r *Reader) ActiveBusinessOwner(ctx context.Context, businessID int64) (int64, error) {
	if r.db == nil {
		return 0, sql.ErrNoRows
	}
	return r.queries.GetActiveBusinessOwnerID(ctx, businessID)
}

// UserBusiness verifies ownership or an active membership. Zero selects the user's default business.
func (r *Reader) UserBusiness(ctx context.Context, businessID, userID int64) (int64, error) {
	if r.db == nil || userID <= 0 {
		return 0, sql.ErrNoRows
	}
	if businessID <= 0 {
		resolved, err := r.queries.GetUserDefaultBusinessID(ctx, userID)
		if err != nil {
			return 0, err
		}
		if !resolved.Valid || resolved.Int64 <= 0 {
			return 0, sql.ErrNoRows
		}
		businessID = resolved.Int64
	}
	return r.queries.GetAuthorizedBusinessID(ctx, sqlc.GetAuthorizedBusinessIDParams{ID: businessID, UserID: userID})
}

// SubscriptionAccess reads the owner's current billing status and unrevoked grant. Missing subscriptions return sql.ErrNoRows.
func (r *Reader) SubscriptionAccess(ctx context.Context, ownerID int64) (string, bool, time.Time, error) {
	row, err := r.queries.SubscriptionAccess(ctx, ownerID)
	return string(row.Status.SubscriptionsStatus), row.Granted, row.TrialEndsAt.Time, err
}

// UploadKeyScope resolves an active, unexpired upload credential; invalid keys return sql.ErrNoRows.
func (r *Reader) UploadKeyScope(ctx context.Context, hash string) (int64, int64, error) {
	row, err := r.queries.UploadKeyScope(ctx, hash)
	return row.OwnerID, row.BusinessID, err
}

// VerifyAccountPassword grants credential verification without exposing stored hashes.
func (r *Reader) VerifyAccountPassword(ctx context.Context, userID int64, password string) (bool, error) {
	hash, err := r.queries.AccountPasswordHash(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !strings.HasPrefix(hash, "$2")) {
		passwordutils.CheckDummyHash(password)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return passwordutils.CheckPasswordHash(password, hash), nil
}

// BusinessUserName resolves display-only historical membership; it grants no access.
func (r *Reader) BusinessUserName(ctx context.Context, businessID, userID int64) (string, error) {
	value, err := r.queries.BusinessUserName(ctx, sqlc.BusinessUserNameParams{UserID: userID, BusinessID: businessID})
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value.String, err
}

// IsBusinessUser checks current ownership or an active grant, without caching authorization.
func (r *Reader) IsBusinessUser(ctx context.Context, businessID, userID int64) (bool, error) {
	_, err := r.UserBusiness(ctx, businessID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
func (r *Reader) AuthorizationVersion(ctx context.Context, businessID, userID int64) (int64, error) {
	return r.queries.AuthorizationVersion(ctx, sqlc.AuthorizationVersionParams{UserID: userID, BusinessID: businessID})
}

// AccountEmail resolves an authenticated account, including an invitee without membership.
// Missing accounts return sql.ErrNoRows.
func (r *Reader) AccountEmail(ctx context.Context, userID int64) (string, error) {
	return r.queries.AccountEmail(ctx, userID)
}

// LocationLimit is the committed plan limit at operation start. Missing subscriptions or NULL limits are unlimited.
func (r *Reader) LocationLimit(ctx context.Context, businessID int64) (int64, bool, error) {
	value, err := r.queries.LocationLimit(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return int64(value.Int32), value.Valid, err
}

// BusinessUser is display-only identity data, including historical members.
type BusinessUser struct {
	ID          int64
	Name, Email string
}

// BusinessUsers returns only requested identities associated with the business.
// Missing identities are omitted; database failures are returned. This grants no access.
func (r *Reader) BusinessUsers(ctx context.Context, businessID int64, userIDs []int64) ([]BusinessUser, error) {
	result := make([]BusinessUser, 0, len(userIDs))
	for start := 0; start < len(userIDs); start += 256 {
		end := min(start+256, len(userIDs))
		rows, err := r.queries.BusinessUsers(ctx, sqlc.BusinessUsersParams{BusinessID: businessID, UserIds: userIDs[start:end]})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			result = append(result, BusinessUser{row.ID, row.Name, row.Email})
		}
	}
	return result, nil
}

// ActiveBusinessUsers lists the owner and active members of the business,
// for pickers such as a deal's assigned rep. Unlike BusinessUsers it leaves
// out former members, who can no longer be assigned work. This grants no access.
func (r *Reader) ActiveBusinessUsers(ctx context.Context, businessID int64) ([]BusinessUser, error) {
	rows, err := r.queries.ActiveBusinessUsers(ctx, businessID)
	if err != nil {
		return nil, err
	}
	result := make([]BusinessUser, 0, len(rows))
	for _, row := range rows {
		result = append(result, BusinessUser{row.ID, row.Name, row.Email})
	}
	return result, nil
}

// AccountByEmail resolves an invitation target. Missing accounts return sql.ErrNoRows.
func (r *Reader) AccountByEmail(ctx context.Context, email string) (int64, error) {
	return r.queries.AccountByEmail(ctx, email)
}

// BusinessUserByUsername resolves mentions against current membership, not a cached/default business.
func (r *Reader) BusinessUserByUsername(ctx context.Context, businessID int64, username string) (int64, error) {
	return r.queries.BusinessUserByUsername(ctx, sqlc.BusinessUserByUsernameParams{BusinessID: businessID, Username: sql.NullString{String: username, Valid: true}})
}

// DefaultBusiness reads the account's committed default. No default returns sql.ErrNoRows.
func (r *Reader) DefaultBusiness(ctx context.Context, userID int64) (int64, error) {
	value, err := r.queries.GetUserDefaultBusinessID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !value.Valid || value.Int64 <= 0 {
		return 0, sql.ErrNoRows
	}
	return value.Int64, nil
}

// AssistantIdentity resolves only the active business the person can access.
// Missing/revoked identities return sql.ErrNoRows; strings are display context, not grants.
func (r *Reader) AssistantIdentity(ctx context.Context, businessID, userID int64) (name, business, timezone, currency string, err error) {
	value, err := r.queries.AssistantIdentity(ctx, sqlc.AssistantIdentityParams{BusinessID: businessID, UserID: userID})
	return value.Name, value.Business, value.Timezone, value.Currency, err
}
