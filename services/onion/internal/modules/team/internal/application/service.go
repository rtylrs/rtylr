package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/config"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	meCacheTTL   = 5 * time.Minute
	teamCacheTTL = 5 * time.Minute
)

var defaultFreePlanApps = []string{"dash", "pos", "erp", "hr", "crm", "finance", "flow", "insights"}

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
	rows, err := s.repo.ListStaffUsers(ctx, ownerID, int32(limit+1), int32(offset))
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
			return s.repo.ListStaffUsersByBusinessID(ctx, businessID, int32(limit+1), int32(offset))
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

func (s *Service) resolveProfileBusinessID(ctx context.Context, userID, businessID int64) (int64, error) {
	if businessID > 0 {
		return businessID, nil
	}

	resolvedBusinessID, err := s.repo.GetUserBusinessID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return resolvedBusinessID, nil
}

func (s *Service) computeMyProfile(ctx context.Context, userID, businessID int64) (domain.MyProfile, error) {
	email, fullName, isConsoleAdmin, apps, role, err := s.GetMeByBusiness(ctx, businessID, userID)
	if err != nil {
		return domain.MyProfile{}, err
	}
	if apps == nil {
		apps = []string{}
	}

	businessName, businessOwnerID, isPrimaryOwner := s.GetBusinessInfo(ctx, businessID, userID)
	// A co-owner has role=owner in user_app_permissions but is not the primary owner.
	isCoOwner := !isPrimaryOwner && role == domain.RoleOwner
	isOwner := isPrimaryOwner || isCoOwner

	var planSlug, subscriptionStatus string
	if isOwner {
		role = domain.RoleOwner
		// Always use the primary owner's subscription for entitlements.
		billingID := businessOwnerID
		if isPrimaryOwner {
			billingID = userID
		}
		ownerApps, appsErr := s.GetOwnerApps(ctx, billingID)
		if appsErr != nil || len(ownerApps) == 0 {
			apps = append([]string(nil), defaultFreePlanApps...)
		} else {
			apps = ownerApps
		}
		if slug, status, subErr := s.GetOwnerSubscriptionInfo(ctx, billingID); subErr == nil {
			planSlug = slug
			subscriptionStatus = status
		} else {
			planSlug = "free"
			subscriptionStatus = "active"
		}
	}

	effectiveResources := map[string]string{}
	if !isOwner {
		if resPerms, resErr := s.GetUserEffectiveResources(ctx, businessOwnerID, userID); resErr == nil {
			for k, v := range resPerms {
				effectiveResources[k] = string(v)
			}
		}
	}

	return domain.MyProfile{
		ID:                 userID,
		Email:              email,
		FullName:           fullName,
		Apps:               apps,
		Role:               string(role),
		BusinessID:         businessID,
		BusinessName:       businessName,
		IsOwner:            isOwner,
		IsConsoleAdmin:     isConsoleAdmin,
		Onboarded:          s.IsOnboarded(ctx, userID),
		PlanSlug:           planSlug,
		SubscriptionStatus: subscriptionStatus,
		EffectiveResources: effectiveResources,
	}, nil
}

func (s *Service) GetStaffUserByID(ctx context.Context, ownerID, businessID, userID int64) (*domain.StaffUser, error) {
	if userID == 0 {
		return nil, domain.ErrUserNotFound
	}

	resolvedOwnerID, err := s.resolveBusinessOwnerID(ctx, ownerID, businessID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetStaffUserByID(ctx, resolvedOwnerID, businessID, userID)
}

// InviteUser creates a user account (no password) and generates an invite token.
// Returns userID and invite token string.
//
// businessID is passed in (rather than looked up) so the service can invalidate
// the business-scoped team list cache without an extra query. Tests that don't
// care about the cache key may pass 0.
func (s *Service) InviteUser(ctx context.Context, ownerID, businessID int64, input domain.InviteUserInput) (int64, string, error) {
	if input.Email == "" {
		return 0, "", errors.New("email is required")
	}
	role := input.Role
	if role == "" {
		role = domain.RoleStaff
	}
	if !isAssignableRole(role) {
		return 0, "", fmt.Errorf("invalid role: %s", role)
	}
	if role == domain.RoleOwner {
		return 0, "", domain.ErrForbidden
	}

	ownerID, err := s.requireTeamManager(ctx, ownerID, businessID)
	if err != nil {
		return 0, "", err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))

	// Hash password if provided (manual auth), otherwise empty (invite link).
	passwordHash := ""
	if input.Password != "" {
		hash, hashErr := passwordutils.HashPassword(input.Password)
		if hashErr != nil {
			return 0, "", hashErr
		}
		passwordHash = hash
	}

	apps := input.Apps
	if apps == nil {
		apps = []string{}
	}

	// Generate invite token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return 0, "", err
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	userID, err := s.repo.CreateInviteTransaction(ctx, businessID, ownerID, normalizedEmail, passwordHash, strings.TrimSpace(input.FullName), apps, role, token, expiresAt)
	if err != nil {
		return 0, "", err
	}

	// New team member changes the business's team list cache + the new user's
	// /me payload (apps/permissions just got attached).
	invalidateUserScope(ctx, userID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}

	return userID, token, nil
}

// VerifyInviteToken checks if the token is valid and returns the user info along with their app access list.
func (s *Service) VerifyInviteToken(ctx context.Context, token string) (int64, string, string, []string, error) {
	userID, ownerID, _, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, "", "", nil, err
	}

	email, fullName, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return 0, "", "", nil, err
	}

	// Fetch the user's app permissions so the frontend can redirect to the right app after accept.
	apps := []string{}
	if perm, permErr := s.repo.GetPermission(ctx, ownerID, userID); permErr == nil && perm != nil {
		apps = perm.Apps
	}

	return userID, email, fullName, apps, nil
}

// AcceptInvite sets the user's password and marks the token as used.
func (s *Service) AcceptInvite(ctx context.Context, token, password string) (int64, error) {
	if err := passwordutils.ValidateStrength(password); err != nil {
		return 0, err
	}

	userID, _, _, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, err
	}

	hash, err := passwordutils.HashPassword(password)
	if err != nil {
		return 0, err
	}

	if err := s.repo.SetPasswordAndConsumeInviteToken(ctx, userID, token, hash); err != nil {
		return 0, err
	}

	return userID, nil
}

// GetMeByBusiness returns user info with permissions looked up by business_id.
func (s *Service) GetMeByBusiness(ctx context.Context, businessID, userID int64) (string, string, bool, []string, domain.Role, error) {
	email, fullName, isConsoleAdmin, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", false, nil, "", err
	}

	perm, err := s.repo.GetPermissionByBusiness(ctx, businessID, userID)
	if err != nil {
		return email, fullName, isConsoleAdmin, []string{}, "", nil
	}

	if perm == nil {
		return email, fullName, isConsoleAdmin, []string{}, "", nil
	}

	return email, fullName, isConsoleAdmin, perm.Apps, perm.Role, nil
}

// RequestAccess stores a pending access request from a staff user.
func (s *Service) RequestAccess(ctx context.Context, businessID, userID int64, app string) error {
	return s.repo.CreateAccessRequest(ctx, businessID, userID, app)
}

// NotifyAdminsOfAccessRequest writes an in-app notification to hub_notifications
// for every admin and owner in the business so the bell icon lights up immediately.
func (s *Service) NotifyAdminsOfAccessRequest(ctx context.Context, businessID int64, requesterName, app string) {
	recipients, err := s.repo.GetAdminsAndOwnerUserIDs(ctx, businessID)
	if err != nil || len(recipients) == 0 {
		return
	}
	_, ownerID, _ := s.repo.GetBusinessByID(ctx, businessID)
	title := requesterName + " wants access to " + strings.ToUpper(app)
	body := "Go to Team to review and approve."
	link := config.HubAppURL + "/team"
	_ = s.repo.BulkCreateHubNotification(ctx, ownerID, recipients, title, body, "info", link)
}

// ListAccessRequests returns pending access requests for the given business.
// Only admins and owners may call this.
func (s *Service) ListAccessRequests(ctx context.Context, actorID, businessID int64) ([]domain.AccessRequest, error) {
	if _, err := s.requireTeamManager(ctx, actorID, businessID); err != nil {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListAccessRequests(ctx, businessID)
}

// ApproveAccessRequest grants the requested app to the user and marks the request resolved.
func (s *Service) ApproveAccessRequest(ctx context.Context, actorID, businessID, requestID, targetUserID int64, app string) error {
	ownerID, err := s.requireTeamManager(ctx, actorID, businessID)
	if err != nil {
		return domain.ErrForbidden
	}

	existing, err := s.repo.GetStaffUserByID(ctx, ownerID, businessID, targetUserID)
	if err != nil {
		return err
	}
	apps := existing.Apps
	for _, a := range apps {
		if a == app {
			// already has access — just resolve the request
			return s.repo.ResolveAccessRequest(ctx, requestID, businessID, "approved")
		}
	}
	apps = append(apps, app)

	if err := s.repo.UpsertPermission(ctx, businessID, ownerID, targetUserID, apps, existing.Role); err != nil {
		return err
	}
	invalidateUserScope(ctx, targetUserID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	return s.repo.ResolveAccessRequest(ctx, requestID, businessID, "approved")
}

// GetBusinessOwnerEmail returns the email of the business owner.
func (s *Service) GetBusinessOwnerEmail(ctx context.Context, businessID int64) string {
	_, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return ""
	}
	email, _, _, err := s.repo.GetUserByID(ctx, ownerID)
	if err != nil {
		return ""
	}
	return email
}

// GetUserInfo returns email and full name for a user.
func (s *Service) GetUserInfo(ctx context.Context, userID int64) (string, string, error) {
	email, fullName, _, err := s.repo.GetUserByID(ctx, userID)
	return email, fullName, err
}

func (s *Service) IsOnboarded(ctx context.Context, userID int64) bool {
	return s.repo.IsOnboarded(ctx, userID)
}

func (s *Service) MarkOnboarded(ctx context.Context, userID int64) error {
	if err := s.repo.MarkOnboarded(ctx, userID); err != nil {
		return err
	}
	invalidateUserScope(ctx, userID)
	return nil
}

// GetBusinessInfo returns the business name, owner ID, and whether the user is the owner.
func (s *Service) GetBusinessInfo(ctx context.Context, businessID, userID int64) (string, int64, bool) {
	if businessID == 0 {
		return "", 0, false
	}

	name, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return "", 0, false
	}

	return name, ownerID, ownerID == userID
}

func (s *Service) UpdatePermissions(ctx context.Context, actorID, businessID int64, input domain.UpdatePermissionsInput) error {
	if input.UserID == 0 {
		return errors.New("user id is required")
	}

	ownerID, err := s.requireTeamManager(ctx, actorID, businessID)
	if err != nil {
		return err
	}

	role := input.Role
	apps := []string{}
	if input.Apps != nil {
		apps = *input.Apps
	}
	if input.Apps == nil || role == "" {
		existing, err := s.repo.GetStaffUserByID(ctx, ownerID, businessID, input.UserID)
		if err != nil {
			return err
		}
		if input.Apps == nil {
			apps = existing.Apps
		}
		if role == "" {
			role = existing.Role
		}
	}
	if role == "" {
		role = domain.RoleStaff
	}
	if !isAssignableRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}
	if role == domain.RoleOwner && input.UserID != ownerID {
		return domain.ErrForbidden
	}

	if input.FullName != nil {
		if err := s.repo.UpdateUserProfile(ctx, input.UserID, *input.FullName); err != nil {
			return err
		}
	}

	if err := s.repo.UpsertPermission(ctx, businessID, ownerID, input.UserID, apps, role); err != nil {
		return err
	}

	invalidateUserScope(ctx, input.UserID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	return nil
}

func (s *Service) DeactivateUser(ctx context.Context, actorID, businessID, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	ownerID, err := s.requireTeamManager(ctx, actorID, businessID)
	if err != nil {
		return err
	}
	if userID == ownerID {
		return domain.ErrForbidden
	}

	if err := s.repo.DeactivateUser(ctx, ownerID, userID); err != nil {
		return err
	}

	invalidateUserScope(ctx, userID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	return nil
}

func (s *Service) resolveBusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	if businessID == 0 {
		return fallbackOwnerID, nil
	}

	_, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return 0, err
	}

	return ownerID, nil
}

func (s *Service) BusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	return s.resolveBusinessOwnerID(ctx, fallbackOwnerID, businessID)
}

func (s *Service) CanAccessOwner(ctx context.Context, userID, ownerID int64) bool {
	if userID == 0 || ownerID == 0 {
		return false
	}
	if userID == ownerID {
		return true
	}
	perm, err := s.repo.GetPermission(ctx, ownerID, userID)
	if err != nil {
		return false
	}
	return perm != nil && perm.IsActive
}

func (s *Service) TeamManagerOwnerID(ctx context.Context, actorID, businessID int64) (int64, error) {
	return s.requireTeamManager(ctx, actorID, businessID)
}

func (s *Service) requireTeamManager(ctx context.Context, actorID, businessID int64) (int64, error) {
	ownerID, err := s.resolveBusinessOwnerID(ctx, actorID, businessID)
	if err != nil {
		return 0, err
	}
	if actorID == ownerID {
		return ownerID, nil
	}
	if businessID == 0 {
		return 0, domain.ErrForbidden
	}

	perm, err := s.repo.GetPermissionByBusiness(ctx, businessID, actorID)
	if err != nil {
		return 0, domain.ErrForbidden
	}
	if perm != nil && perm.IsActive && (perm.Role == domain.RoleAdmin || perm.Role == domain.RoleOwner) {
		return ownerID, nil
	}

	return 0, domain.ErrForbidden
}

func isAssignableRole(role domain.Role) bool {
	switch role {
	case domain.RoleOwner, domain.RoleAdmin, domain.RoleOperator, domain.RoleBilling, domain.RoleMenuManager, domain.RoleStoreManager, domain.RoleStaff, domain.RoleViewer:
		return true
	default:
		return false
	}
}

// GetOwnerApps returns the list of apps the owner is subscribed to.
func (s *Service) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	return s.repo.GetOwnerApps(ctx, ownerID)
}

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (s *Service) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	return s.repo.GetOwnerSubscriptionInfo(ctx, ownerID)
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// It is invoked asynchronously so it never blocks team operations, but it returns its error so the
// caller can log billing failures instead of silently discarding them (which would corrupt billing).
func (s *Service) RecordSeatChange(ctx context.Context, ownerID int64) error {
	newCount, err := s.repo.CountActiveMembers(ctx, ownerID)
	if err != nil {
		return err
	}

	return s.repo.RecordSeatChange(ctx, ownerID, newCount)
}

// ─── Groups ──────────────────────────────────────────────────────────────────

func (s *Service) CreateGroup(ctx context.Context, ownerID int64, input domain.CreateGroupInput) (int64, error) {
	if input.Name == "" {
		return 0, errors.New("name is required")
	}
	if input.Color == "" {
		input.Color = "#6366f1"
	}
	return s.repo.CreateGroup(ctx, ownerID, input)
}

func (s *Service) ListGroups(ctx context.Context, ownerID int64) ([]domain.Group, error) {
	return s.repo.ListGroups(ctx, ownerID)
}

func (s *Service) GetGroup(ctx context.Context, ownerID, groupID int64) (*domain.Group, error) {
	return s.repo.GetGroup(ctx, ownerID, groupID)
}

func (s *Service) GetGroupDetail(ctx context.Context, ownerID, groupID int64) (*domain.GroupDetail, error) {
	return s.repo.GetGroupDetail(ctx, ownerID, groupID)
}

func (s *Service) UpdateGroup(ctx context.Context, ownerID, groupID int64, input domain.UpdateGroupInput) error {
	if input.Name == "" {
		return errors.New("name is required")
	}
	if err := s.repo.UpdateGroup(ctx, ownerID, groupID, input); err != nil {
		return err
	}
	s.invalidateGroupMemberScopes(ctx, ownerID, groupID)
	return nil
}

func (s *Service) DeleteGroup(ctx context.Context, ownerID, groupID int64) error {
	memberIDs := s.groupMemberIDs(ctx, ownerID, groupID)
	if err := s.repo.DeleteGroup(ctx, ownerID, groupID); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, ownerID, memberIDs)
	return nil
}

func (s *Service) AddGroupMembers(ctx context.Context, ownerID, groupID, callerID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return errors.New("no user IDs provided")
	}
	// Verify the group belongs to this owner before adding members — otherwise
	// a caller could target a group id owned by another business.
	if _, err := s.repo.GetGroup(ctx, ownerID, groupID); err != nil {
		return err
	}
	if err := s.repo.AddGroupMembers(ctx, ownerID, groupID, callerID, userIDs); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, ownerID, userIDs)
	return nil
}

func (s *Service) RemoveGroupMember(ctx context.Context, ownerID, groupID, userID int64) error {
	if _, err := s.repo.GetGroup(ctx, ownerID, groupID); err != nil {
		return domain.ErrPermissionNotFound
	}
	if err := s.repo.RemoveGroupMember(ctx, groupID, userID); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, ownerID, []int64{userID})
	return nil
}

func (s *Service) GetUserGroups(ctx context.Context, ownerID, userID int64) ([]domain.GroupSummary, error) {
	return s.repo.GetUserGroups(ctx, ownerID, userID)
}

func (s *Service) GetUserEffectiveResources(ctx context.Context, ownerID, userID int64) (map[string]domain.PermissionLevel, error) {
	return s.repo.GetUserEffectiveResources(ctx, ownerID, userID)
}

func (s *Service) invalidateGroupMemberScopes(ctx context.Context, ownerID, groupID int64) {
	s.invalidateMemberScopes(ctx, ownerID, s.groupMemberIDs(ctx, ownerID, groupID))
}

func (s *Service) groupMemberIDs(ctx context.Context, ownerID, groupID int64) []int64 {
	detail, err := s.repo.GetGroupDetail(ctx, ownerID, groupID)
	if err != nil || detail == nil {
		return nil
	}
	userIDs := make([]int64, 0, len(detail.Members))
	for _, member := range detail.Members {
		userIDs = append(userIDs, member.UserID)
	}
	return userIDs
}

func (s *Service) invalidateMemberScopes(ctx context.Context, ownerID int64, userIDs []int64) {
	for _, userID := range userIDs {
		if userID != 0 {
			invalidateUserScope(ctx, userID)
		}
	}
	if businessID, err := s.repo.GetBusinessIDByOwnerID(ctx, ownerID); err == nil && businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
}

// BulkInvite invites multiple users with the same role/apps/group in one call.
func (s *Service) BulkInvite(ctx context.Context, ownerID, businessID int64, input domain.BulkInviteInput) ([]domain.BulkInviteResult, error) {
	results := make([]domain.BulkInviteResult, 0, len(input.Emails))
	for _, email := range input.Emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}
		userID, token, err := s.InviteUser(ctx, ownerID, businessID, domain.InviteUserInput{
			Email: email,
			Apps:  input.Apps,
			Role:  input.Role,
		})
		if err != nil {
			results = append(results, domain.BulkInviteResult{
				Email:   email,
				Status:  "error",
				Message: err.Error(),
			})
			continue
		}
		if input.GroupID > 0 {
			if groupErr := s.AddGroupMembers(ctx, ownerID, input.GroupID, ownerID, []int64{userID}); groupErr != nil {
				results = append(results, domain.BulkInviteResult{
					Email:   email,
					UserID:  userID,
					Token:   token,
					Status:  "error",
					Message: fmt.Sprintf("group assignment failed: %v", groupErr),
				})
				continue
			}
		}
		status := "invited"
		if token == "" {
			status = "existing"
		}
		results = append(results, domain.BulkInviteResult{
			Email:   email,
			UserID:  userID,
			Token:   token,
			Status:  status,
			Message: "ok",
		})
	}
	return results, nil
}

func (s *Service) GetMe(ctx context.Context, ownerID, userID int64) (string, string, []string, domain.Role, error) {
	email, fullName, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", nil, "", err
	}

	perm, err := s.repo.GetPermission(ctx, ownerID, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return email, fullName, []string{}, "", nil
	}

	if perm == nil {
		return email, fullName, []string{}, "", nil
	}

	return email, fullName, perm.Apps, perm.Role, nil
}
