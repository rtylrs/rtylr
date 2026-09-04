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
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/config"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
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

	var effectiveAccess *authorization.EffectiveAccess
	if snapshot, found, snapshotErr := s.repo.GetAuthorizationSnapshot(ctx, businessID, userID, apps); snapshotErr == nil && found {
		resolved := authorization.Resolve(snapshot)
		effectiveAccess = &resolved
		if resolved.DashRole != "" {
			role = domain.Role(resolved.DashRole)
		}
		isOwner = isOwner || resolved.IsOwner
		if len(resolved.SubscribedApps) > 0 {
			apps = make([]string, 0, len(resolved.SubscribedApps)+1)
			apps = append(apps, "dash")
			for _, appID := range resolved.SubscribedApps {
				apps = append(apps, string(appID))
			}
		}
	}
	var provisioningStatus *authorization.ProvisioningStatus
	if status, found, statusErr := s.repo.GetProvisioningStatus(ctx, userID); statusErr == nil && found {
		provisioningStatus = status
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
		EffectiveAccess:    effectiveAccess,
		ProvisioningStatus: provisioningStatus,
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
	provisioningMode := input.ProvisioningMode
	if provisioningMode == "" {
		provisioningMode = "invitation"
		if input.Password != "" {
			provisioningMode = "direct_credentials"
		}
	}
	if provisioningMode != "invitation" && provisioningMode != "direct_credentials" {
		return 0, "", errors.New("invalid provisioning mode")
	}
	username := strings.ToLower(strings.TrimSpace(input.Username))
	if provisioningMode == "direct_credentials" {
		if input.Password == "" {
			return 0, "", errors.New("password is required for direct credentials")
		}
		if username == "" {
			return 0, "", errors.New("username is required for direct credentials")
		}
		if len(username) < 3 || len(username) > 64 {
			return 0, "", errors.New("username must be between 3 and 64 characters")
		}
		for _, character := range username {
			if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '.' && character != '_' && character != '-' {
				return 0, "", errors.New("username contains invalid characters")
			}
		}
	}
	if input.Password != "" {
		if err := passwordutils.ValidateStrength(input.Password); err != nil {
			return 0, "", err
		}
	}

	ownerID, err := s.requireTeamManager(ctx, ownerID, businessID)
	if err != nil {
		return 0, "", err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))

	existingID, _, _, err := s.repo.GetUserByEmail(ctx, normalizedEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, "", err
	}

	userID := existingID
	createdUser := false
	cleanupCreatedUser := func(cause error) (int64, string, error) {
		if !createdUser {
			return 0, "", cause
		}
		if cleanupErr := s.repo.DeleteInvitedUser(ctx, businessID, ownerID, userID); cleanupErr != nil {
			return 0, "", fmt.Errorf("%w; failed to clean up newly created user: %w", cause, cleanupErr)
		}
		return 0, "", cause
	}

	if userID == 0 {
		// Hash password if provided (manual auth), otherwise empty (invite link)
		passwordHash := ""
		if input.Password != "" {
			hash, hashErr := passwordutils.HashPassword(input.Password)
			if hashErr != nil {
				return 0, "", hashErr
			}
			passwordHash = hash
		}

		userID, err = s.repo.CreateUser(ctx, normalizedEmail, username, passwordHash, strings.TrimSpace(input.FullName))
		if err != nil {
			return 0, "", err
		}
		createdUser = true
	}
	if userID != 0 && existingID != 0 && input.Password != "" {
		return 0, "", errors.New("direct credentials require a new user")
	}

	// Attach the member to the business before any authorization row is written.
	// Nothing else ever set users.business_id for an invitee, so their JWT was
	// issued with no business: /users/me reported business_id "0", onion logged
	// "failed to resolve business for personal workspace", every Flow endpoint
	// 403d, and the auth app dropped the session back to the sign-in screen with
	// nothing to explain why.
	//
	// A failure here is logged rather than fatal: it leaves the member in the
	// state they were already in before this existed, and failing the invite
	// outright would be a worse outcome than an invite that needs the business
	// attaching by hand.
	if attached, attachErr := s.repo.AttachUserToBusiness(ctx, userID, businessID); attachErr != nil {
		logger.Ctx(ctx).Error("failed to attach invited member to business",
			logger.Int("user_id", int(userID)), logger.Int("business_id", int(businessID)), logger.Error(attachErr))
	} else if !attached {
		// Already had one. Expected when inviting somebody who belongs to
		// another business; recorded because it also means this invite did not
		// change which business their session resolves to.
		logger.Ctx(ctx).Info("invited member already belongs to a business",
			logger.Int("user_id", int(userID)), logger.Int("business_id", int(businessID)))
	}

	apps := input.Apps
	if apps == nil {
		apps = []string{}
	}

	if err := s.repo.UpsertPermission(ctx, businessID, ownerID, userID, apps, role); err != nil {
		return cleanupCreatedUser(err)
	}
	if err := s.repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, userID, apps, role, input.DashRole, input.AppAssignments, input.GroupIDs); err != nil {
		return cleanupCreatedUser(err)
	}
	if err := s.repo.SetNormalizedAuthorizationPending(ctx, businessID, userID); err != nil {
		return cleanupCreatedUser(err)
	}
	provisioningState := "pending_invitation"
	passwordChangeRequired := false
	if provisioningMode == "direct_credentials" {
		provisioningState = "pending_activation"
		passwordChangeRequired = true
	}
	if err := s.repo.UpsertProvisioningState(ctx, userID, businessID, provisioningMode, provisioningState, false, passwordChangeRequired); err != nil {
		return cleanupCreatedUser(err)
	}

	// businessID, not the invitee's own users.business_id, which is still NULL
	// at this point and made every invited member's employee record land with
	// business_id = 0.
	if err := s.repo.EnsureEmployeeForUser(ctx, ownerID, businessID, userID); err != nil {
		logger.Ctx(ctx).Error("failed to create employee record for invited user",
			logger.Int("owner_id", int(ownerID)), logger.Int("user_id", int(userID)), logger.Error(err))
	} else if err := s.repo.EnsureStaffForUser(ctx, ownerID, businessID, userID, string(role)); err != nil {
		logger.Ctx(ctx).Error("failed to create staff directory record for invited user",
			logger.Int("owner_id", int(ownerID)), logger.Int("business_id", int(businessID)),
			logger.Int("user_id", int(userID)), logger.Error(err))
	}

	// New team member changes the business's team list cache + the new user's
	// /me payload (apps/permissions just got attached).
	invalidateUserScope(ctx, userID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}

	if provisioningMode == "direct_credentials" {
		return userID, "", nil
	}

	// Generate invite token (32 bytes = 64 hex chars)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return cleanupCreatedUser(err)
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	if err := s.repo.CreateInviteToken(ctx, userID, ownerID, token, expiresAt); err != nil {
		return cleanupCreatedUser(err)
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

	userID, ownerID, _, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, err
	}

	hash, err := passwordutils.HashPassword(password)
	if err != nil {
		return 0, err
	}

	if err := s.repo.SetPassword(ctx, userID, hash); err != nil {
		return 0, err
	}

	if err := s.repo.MarkInviteTokenUsed(ctx, token); err != nil {
		return 0, err
	}
	if businessID, businessErr := s.repo.GetBusinessIDByOwnerID(ctx, ownerID); businessErr == nil {
		if stateErr := s.repo.UpsertProvisioningState(ctx, userID, businessID, "invitation", "active", true, false); stateErr != nil {
			return 0, stateErr
		}
		if activateErr := s.repo.ActivateNormalizedAuthorization(ctx, businessID, userID); activateErr != nil {
			return 0, activateErr
		}
	}
	invalidateUserScope(ctx, userID)

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
	if _, ok := domain.ValidApps[app]; !ok {
		return domain.ErrPermissionNotFound
	}
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
	if err := s.repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, targetUserID, apps, existing.Role, "", nil, nil); err != nil {
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
	if err := s.enforceAuthorizationDelegation(ctx, actorID, ownerID, businessID, input); err != nil {
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
	if len(apps) == 0 && len(input.AppAssignments) > 0 {
		for _, assignment := range input.AppAssignments {
			if assignment.AppID != "" {
				apps = appendUniqueString(apps, assignment.AppID)
			}
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
	if err := s.repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, input.UserID, apps, role, input.DashRole, input.AppAssignments, input.GroupIDs); err != nil {
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
	if businessID, businessErr := s.repo.GetBusinessIDByOwnerID(ctx, ownerID); businessErr == nil && businessID > 0 {
		apps, _ := s.GetOwnerApps(ctx, ownerID)
		if snapshot, found, snapshotErr := s.repo.GetAuthorizationSnapshot(ctx, businessID, userID, apps); snapshotErr == nil && found {
			return authorization.Resolve(snapshot).Apps != nil && snapshot.MembershipActive
		}
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
	if apps, appsErr := s.GetOwnerApps(ctx, ownerID); appsErr == nil {
		if snapshot, found, snapshotErr := s.repo.GetAuthorizationSnapshot(ctx, businessID, actorID, apps); snapshotErr == nil && found {
			resolved := authorization.Resolve(snapshot)
			if resolved.IsOwner || resolved.IsAdmin || containsString(resolved.DashPermissions, "hub:team:edit") {
				return ownerID, nil
			}
			return 0, domain.ErrForbidden
		}
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

func (s *Service) enforceAuthorizationDelegation(ctx context.Context, actorID, ownerID, businessID int64, input domain.UpdatePermissionsInput) error {
	if actorID == ownerID {
		return nil
	}
	if businessID <= 0 {
		return domain.ErrForbidden
	}

	apps, err := s.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return domain.ErrForbidden
	}
	snapshot, found, err := s.repo.GetAuthorizationSnapshot(ctx, businessID, actorID, apps)
	if err != nil || !found {
		// Do not allow a legacy fallback actor to submit normalized role,
		// license, group, or Dash-role fields during the migration window.
		if input.DashRole != "" || input.Apps != nil || input.Role != "" || input.AppAssignments != nil || input.GroupIDs != nil {
			return domain.ErrDelegationCeiling
		}
		return nil
	}
	access := authorization.Resolve(snapshot)
	if access.IsOwner || access.IsAdmin {
		return nil
	}

	hasRoleEdit := containsString(access.DashPermissions, "hub:roles:edit")
	hasTeamEdit := containsString(access.DashPermissions, "hub:team:edit")
	hasLicenseManage := containsString(access.DashPermissions, "hub:licenses:manage")
	if input.DashRole != "" {
		if !hasRoleEdit || !authorization.CanAssignDashboardRole(&access, input.DashRole) {
			return domain.ErrDelegationCeiling
		}
	}
	if input.Role != "" && !hasRoleEdit {
		return domain.ErrDelegationCeiling
	}
	if input.Apps != nil && !hasLicenseManage {
		return domain.ErrDelegationCeiling
	}
	if input.GroupIDs != nil && !hasTeamEdit {
		return domain.ErrDelegationCeiling
	}
	if len(input.AppAssignments) == 0 {
		return nil
	}
	if !hasRoleEdit {
		return domain.ErrDelegationCeiling
	}
	for _, assignment := range input.AppAssignments {
		if assignment.Source == string(authorization.SourceDashDefault) || assignment.Source == string(authorization.SourceSystem) {
			return domain.ErrDelegationCeiling
		}
		if len(assignment.RoleIDs) == 0 {
			return domain.ErrDelegationCeiling
		}
		for _, roleID := range assignment.RoleIDs {
			grant, grantErr := s.repo.GetAuthorizationRoleGrant(ctx, businessID, roleID)
			if grantErr != nil || grant.AppID != assignment.AppID || !authorization.CanDelegatePermissions(&access, authorization.AppID(grant.AppID), grant.Permissions) {
				return domain.ErrDelegationCeiling
			}
		}
	}
	return nil
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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
			Email:          email,
			Apps:           input.Apps,
			Role:           input.Role,
			DashRole:       input.DashRole,
			AppAssignments: input.AppAssignments,
			GroupIDs:       append(append([]int64{}, input.GroupIDs...), input.GroupID),
		})
		if err != nil {
			message := err.Error()
			if len(positiveInt64s(append(append([]int64{}, input.GroupIDs...), input.GroupID))) > 0 && strings.Contains(strings.ToLower(message), "group") {
				message = fmt.Sprintf("group assignment failed: %v", err)
			}
			results = append(results, domain.BulkInviteResult{
				Email:   email,
				Status:  "error",
				Message: message,
			})
			continue
		}
		groupAssignmentFailed := false
		for _, groupID := range positiveInt64s(append(append([]int64{}, input.GroupIDs...), input.GroupID)) {
			if groupErr := s.AddGroupMembers(ctx, ownerID, groupID, ownerID, []int64{userID}); groupErr != nil {
				results = append(results, domain.BulkInviteResult{
					Email:   email,
					UserID:  userID,
					Token:   token,
					Status:  "error",
					Message: fmt.Sprintf("group assignment failed: %v", groupErr),
				})
				groupAssignmentFailed = true
				break
			}
		}
		if groupAssignmentFailed {
			continue
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

func positiveInt64s(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
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
