package application

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/config"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

const (
	meCacheTTL   = 5 * time.Minute
	teamCacheTTL = 5 * time.Minute
)

var defaultFreePlanApps = []string{"dash", "pos", "erp", "hr", "crm", "finance", "flow", "insights", "menu", "recruit"}

func invalidateUserScope(ctx context.Context, userID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.UserPrefix(userID))
	}
}

func invalidateBusinessTeamScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.BusinessPrefix(businessID))
	}
}

// invalidateSeatSubscriptionScope drops the owner's cached subscription
// payload after a seat-affecting team change. Seat counts are embedded in that
// payload, so without this GET /hub/subscription/me keeps serving the
// pre-change seat usage until the 10-minute subscription cache TTL expires.
// The local replica is cleared synchronously; the published pattern clears the
// other processes' replicas and the Redis entry.
func invalidateSeatSubscriptionScope(ctx context.Context, ownerID int64) {
	// The exact payload key carries no suffix, so pair it with the derived-key
	// wildcard instead of relying on one pattern to match both.
	local := cacher.GetInMemoryCache(ctx)
	for _, pattern := range []string{cacheKeys.Subscription(ownerID), cacheKeys.SubscriptionPrefix(ownerID)} {
		local.DeleteByPattern(ctx, pattern)
	}
	cacheInvalidate.InvalidateAsync(ctx, cacheKeys.InvalidateSubscription, strconv.FormatInt(ownerID, 10))
}

type Service struct {
	repo *infrastructure.Repository
	db   *sql.DB
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
		db:   db,
	}
}

// EnqueueInviteEmail queues a team invitation email for the background worker to
// deliver. Queue failures are returned to the invite workflow.
func (s *Service) EnqueueInviteEmail(ctx context.Context, toEmail, inviteLink string) error {
	if s.db == nil {
		return nil
	}
	subject := "Your team invited you to rtylr on rtylr"
	return emailPkg.EnqueueTemplate(ctx, s.db, emailPkg.PriorityHigh, toEmail, subject, emailPkg.TemplateInvitation, emailPkg.InvitationData{
		Subject:      subject,
		InviterName:  "Your team",
		BusinessName: "rtylr",
		InviteLink:   inviteLink,
	})
}

// EnqueueAccessRequestEmail queues an app-access-request notification to the
// business owner. Queue failures are returned to the access-request workflow.
func (s *Service) EnqueueAccessRequestEmail(ctx context.Context, businessID int64, requesterName, requesterEmail, app string) error {
	if s.db == nil {
		return nil
	}
	ownerEmail := s.GetBusinessOwnerEmail(ctx, businessID)
	if ownerEmail == "" {
		return nil
	}
	appUpper := strings.ToUpper(app)
	subject := requesterName + " requested access to " + appUpper
	return emailPkg.EnqueueTemplate(ctx, s.db, emailPkg.PriorityNormal, ownerEmail, subject, emailPkg.TemplateAccessRequest, emailPkg.AccessRequestData{
		Subject:        subject,
		RequesterName:  requesterName,
		RequesterEmail: requesterEmail,
		App:            appUpper,
		TeamLink:       config.HubAppURL + "/team",
	})
}

func (s *Service) ListStaffUsers(ctx context.Context, ownerID, page, pageSize int64) ([]domain.StaffUser, bool, error) {
	limit, offset := paginateTeam(page, pageSize)
	rows, err := s.repo.ListStaffUsers(ctx, ownerID, int32(limit+1), int32(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

// ListStaffUsersByBusiness resolves the business owner and lists all staff
// under that owner. Lets non-owner members see the full team, not just the
// rows they personally own (which would be empty). Result is cached per
// businessID and invalidated by any team mutation under that business.
func (s *Service) ListStaffUsersByBusiness(ctx context.Context, businessID, page, pageSize int64) ([]domain.StaffUser, bool, error) {
	limit, offset := paginateTeam(page, pageSize)
	key := fmt.Sprintf("%s|p=%d|ps=%d", cacheKeys.BusinessTeam(businessID), page, pageSize)
	rows, err := cacher.Of[[]domain.StaffUser]().GetOrComputeSingle(
		ctx,
		key,
		func(ctx context.Context) ([]domain.StaffUser, error) {
			return s.repo.ListStaffUsersByBusinessID(ctx, businessID, int32(limit+1), int32(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		},
		cacher.WithExpiry(ctx, teamCacheTTL),
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func (s *Service) CountStaffUsersByBusiness(ctx context.Context, businessID int64) (int64, error) {
	return s.repo.CountStaffUsersByBusinessID(ctx, businessID)
}

func paginateTeam(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return int(pageSize), int((page - 1) * pageSize)
}

// GetMyProfile builds the composite /user/me payload (identity + business +
// subscription + effective resources). Cached per-user; any team or
// subscription mutation under the user invalidates `cache:user:{id}:*`.
func (s *Service) GetMyProfile(ctx context.Context, userID, businessID int64) (domain.MyProfile, error) {
	resolvedBusinessID, err := s.resolveProfileBusinessID(ctx, userID, businessID)
	if err != nil {
		return domain.MyProfile{}, err
	}

	return cacher.Of[domain.MyProfile]().GetOrComputeSingle(
		ctx,
		cacheKeys.UserMeForBusiness(userID, resolvedBusinessID),
		func(ctx context.Context) (domain.MyProfile, error) {
			return s.computeMyProfile(ctx, userID, resolvedBusinessID)
		},
		cacher.WithExpiry(ctx, meCacheTTL),
	)
}

// GetEffectiveAccess is the middleware-facing authorization boundary. A nil
// result means the normalized member row is not available yet, so callers may
// remain on the legacy read path during the additive migration.
func (s *Service) GetEffectiveAccess(ctx context.Context, businessID, userID int64) (*authorization.EffectiveAccess, error) {
	if authPkg.LegacyAuthorizationMode() {
		_, _, _, apps, role, err := s.GetMeByBusiness(ctx, businessID, userID)
		if err != nil {
			return nil, err
		}
		legacy := authorization.LegacyAccess(businessID, apps, string(role))
		return &legacy, nil
	}
	profile, err := s.GetMyProfile(ctx, userID, businessID)
	if err != nil {
		return nil, err
	}
	return profile.EffectiveAccess, nil
}

func (s *Service) GetProvisioningStatus(ctx context.Context, userID int64) (*authorization.ProvisioningStatus, error) {
	status, found, err := s.repo.GetProvisioningStatus(ctx, userID)
	if err != nil || !found {
		return nil, err
	}
	return status, nil
}

func (s *Service) IsMFAVerified(ctx context.Context, userID int64) (bool, error) {
	return s.repo.IsMFAVerified(ctx, userID)
}

func (s *Service) MarkPasswordChanged(ctx context.Context, userID int64) error {
	if err := s.repo.MarkPasswordChanged(ctx, userID); err != nil {
		return err
	}
	invalidateUserScope(ctx, userID)
	return nil
}
