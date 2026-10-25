package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	meCacheTTL   = 5 * time.Minute
	teamCacheTTL = 5 * time.Minute
)

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

// invalidateSeatSubscriptionScope drops the business's cached subscription
// payload after a seat-affecting team change. Seat counts are embedded in that
// payload, so without this GET /hub/subscription/me keeps serving the
// pre-change seat usage until the 10-minute subscription cache TTL expires.
// The local replica is cleared synchronously; the published pattern clears the
// other processes' replicas and the Redis entry.
func invalidateSeatSubscriptionScope(ctx context.Context, businessID int64) {
	// The exact payload key carries no suffix, so pair it with the derived-key
	// wildcard instead of relying on one pattern to match both.
	local := cacher.GetInMemoryCache(ctx)
	for _, pattern := range cacheKeys.SubscriptionPatterns(businessID) {
		local.DeleteByPattern(ctx, pattern)
	}
	cacheInvalidate.InvalidateAsync(ctx, cacheKeys.InvalidateSubscription, strconv.FormatInt(businessID, 10))
}

type Service struct {
	repo                repository
	emails              teamEmailNotifications
	notifications       notificationPublisher
	members             memberProvisioner
	terminatedEmployees terminatedEmployeeSource
	audit               auditRecorder
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WithEmailNotifications(emails teamEmailNotifications) *Service {
	s.emails = emails
	return s
}

type notificationPublisher interface {
	CreateNotification(ctx context.Context, ownerID, businessID, userID int64, title, body, notificationType, link string) error
}

func (s *Service) WithNotificationPublisher(publisher notificationPublisher) *Service {
	s.notifications = publisher
	return s
}

type memberProvisioner interface {
	EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64, firstName, lastName, email string) error
	EnsureStaffForUser(ctx context.Context, ownerID, businessID, userID int64, role string) error
	DeleteProvisionedMember(ctx context.Context, ownerID, businessID, userID int64) error
}

func (s *Service) WithMemberProvisioner(provisioner memberProvisioner) *Service {
	s.members = provisioner
	return s
}

type TerminatedEmployee struct {
	BusinessID int64
	UserID     int64
}

type terminatedEmployeeSource interface {
	ListTerminatedEmployees(ctx context.Context) ([]TerminatedEmployee, error)
}

func (s *Service) WithTerminatedEmployeeSource(source terminatedEmployeeSource) *Service {
	s.terminatedEmployees = source
	return s
}

// EnqueueInviteEmail queues actorID's invitation to join businessID for the
// background worker to deliver, naming the inviter (or their email) and the
// business. Queue failures are returned to the invite workflow; a name that
// can't be read is left out rather than failing the invite.
func (s *Service) EnqueueInviteEmail(ctx context.Context, actorID, businessID int64, toEmail, inviteLink string) error {
	if s.emails == nil {
		return nil
	}
	invite := InviteEmail{
		To:        toEmail,
		Link:      inviteLink,
		ValidDays: int(domain.InviteTokenTTL / (24 * time.Hour)),
	}
	if email, fullName, _, err := s.repo.GetUserByID(ctx, actorID); err == nil {
		invite.InviterName = strings.TrimSpace(fullName)
		if invite.InviterName == "" {
			invite.InviterName = strings.TrimSpace(email)
		}
	}
	if name, _, err := s.repo.GetBusinessByID(ctx, businessID); err == nil {
		invite.BusinessName = strings.TrimSpace(name)
	}
	return s.emails.QueueInvite(ctx, invite)
}

// EnqueueAccessRequestEmail queues an app-access-request notification to the
// business owner. Queue failures are returned to the access-request workflow.
func (s *Service) EnqueueAccessRequestEmail(ctx context.Context, businessID int64, requesterName, requesterEmail, app string) error {
	if s.emails == nil {
		return nil
	}
	ownerEmail := s.GetBusinessOwnerEmail(ctx, businessID)
	if ownerEmail == "" {
		return nil
	}
	return s.emails.QueueAccessRequest(ctx, ownerEmail, requesterName, requesterEmail, app)
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

// GetEffectiveAccess is the middleware-facing normalized authorization boundary.
func (s *Service) GetEffectiveAccess(ctx context.Context, businessID, userID int64) (*authorization.EffectiveAccess, error) {
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

func (s *Service) GetProvisioningStatusForBusiness(ctx context.Context, businessID, userID int64) (*authorization.ProvisioningStatus, error) {
	status, found, err := s.repo.GetProvisioningStatusForBusiness(ctx, businessID, userID)
	if err != nil || !found {
		return nil, err
	}
	return status, nil
}

func (s *Service) IsMFAVerified(ctx context.Context, userID int64) (bool, error) {
	return s.repo.IsMFAVerified(ctx, userID)
}

func (s *Service) IsMFAVerifiedForBusiness(ctx context.Context, businessID, userID int64) (bool, error) {
	return s.repo.IsMFAVerifiedForBusiness(ctx, businessID, userID)
}

func (s *Service) MarkPasswordChanged(ctx context.Context, userID int64) error {
	if err := s.repo.MarkPasswordChanged(ctx, userID); err != nil {
		return err
	}
	invalidateUserScope(ctx, userID)
	return nil
}
