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
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

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
	// Resolve the business's plan entitlement for every member, not just the
	// owner: a dash admin is a tenant administrator and inherits the whole plan
	// by default. Which apps a member actually receives is decided in the
	// repository, where the normalized role is known.
	billingID := businessOwnerID
	if isPrimaryOwner {
		billingID = userID
	}
	businessApps, appsErr := s.GetOwnerApps(ctx, billingID)
	if appsErr != nil || len(businessApps) == 0 {
		businessApps = append([]string(nil), defaultFreePlanApps...)
	}
	if isOwner {
		role = domain.RoleOwner
		apps = businessApps
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
	// The legacy apps column is deliberately not passed through: it is written
	// by a separate, partially migrated path, so a stale row there would grant
	// an app that no assignment backs. Owners and admins inherit the plan in the
	// repository; everyone else is bounded by their own role assignments.
	if snapshot, found, snapshotErr := s.repo.GetAuthorizationSnapshot(ctx, businessID, userID, businessApps); snapshotErr == nil && found {
		resolved := authorization.Resolve(snapshot)
		effectiveAccess = &resolved
		if resolved.DashRole != "" {
			role = domain.Role(resolved.DashRole)
		}
		isOwner = isOwner || resolved.IsOwner
		// Always restate the response from the resolver, so the sidebar lists
		// exactly the apps the member can open rather than whatever the legacy
		// column happens to hold.
		apps = make([]string, 0, len(resolved.SubscribedApps)+1)
		apps = append(apps, "dash")
		for _, appID := range resolved.SubscribedApps {
			apps = append(apps, string(appID))
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
	if err := s.repo.CreateInviteToken(ctx, userID, ownerID, businessID, token, expiresAt); err != nil {
		return cleanupCreatedUser(err)
	}

	return userID, token, nil
}
