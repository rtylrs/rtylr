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
	email, fullName, isConsoleAdmin, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.MyProfile{}, err
	}
	businessName, businessOwnerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return domain.MyProfile{}, err
	}
	businessApps, err := s.GetOwnerApps(ctx, businessOwnerID)
	if err != nil {
		return domain.MyProfile{}, err
	}
	planSlug, subscriptionStatus, err := s.GetOwnerSubscriptionInfo(ctx, businessOwnerID)
	if err != nil {
		return domain.MyProfile{}, err
	}

	snapshot, found, err := s.repo.GetAuthorizationSnapshot(ctx, businessID, userID, businessApps)
	if err != nil {
		return domain.MyProfile{}, err
	}
	if !found {
		return domain.MyProfile{}, domain.ErrForbidden
	}
	effectiveAccess := authorization.Resolve(snapshot)
	provisioningStatus, found, err := s.repo.GetProvisioningStatusForBusiness(ctx, businessID, userID)
	if err != nil {
		return domain.MyProfile{}, err
	}
	if !found || provisioningStatus.State != authorization.ProvisioningActive || !provisioningStatus.EmailVerified {
		return domain.MyProfile{}, domain.ErrForbidden
	}

	return domain.MyProfile{
		ID:                 userID,
		Email:              email,
		FullName:           fullName,
		BusinessID:         businessID,
		BusinessName:       businessName,
		IsConsoleAdmin:     isConsoleAdmin,
		Onboarded:          s.IsOnboarded(ctx, userID),
		PlanSlug:           planSlug,
		SubscriptionStatus: subscriptionStatus,
		EffectiveAccess:    &effectiveAccess,
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
	memberType := input.MemberType
	if memberType == "" {
		memberType = "member"
	}
	if memberType != "member" && memberType != "admin" {
		return 0, "", errors.New("invalid member type")
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

	actorID := ownerID
	ownerID, err := s.requireTeamManager(ctx, actorID, businessID)
	if err != nil {
		return 0, "", err
	}
	if err := s.enforceAuthorizationDelegation(ctx, actorID, ownerID, businessID, domain.UpdatePermissionsInput{
		MemberType:                   memberType,
		LicensedAppIDs:               &input.LicensedAppIDs,
		PermissionProfileAssignments: &input.PermissionProfileAssignments,
		GroupIDs:                     &input.GroupIDs,
	}); err != nil {
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

	if err := s.repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, actorID, userID, input.LicensedAppIDs, memberType, "pending", input.PermissionProfileAssignments, input.GroupIDs); err != nil {
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
	} else if err := s.repo.EnsureStaffForUser(ctx, ownerID, businessID, userID, "member"); err != nil {
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
