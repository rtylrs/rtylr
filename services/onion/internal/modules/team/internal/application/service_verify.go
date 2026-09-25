package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

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

	// Claim the token before writing anything: a concurrent accept loses the
	// atomic consume and never reaches the password write below.
	if err := s.repo.MarkInviteTokenUsed(ctx, token); err != nil {
		return 0, err
	}

	// The token is now claimed. A failure past this point must release the
	// claim: a consumed token plus a pending/email-unverified account is
	// unrecoverable for the invitee (sign-in is rejected as unverified and
	// re-accepting fails with ErrTokenAlreadyUsed), so the accept has to stay
	// retryable.
	failAccept := func(cause error) (int64, error) {
		if releaseErr := s.repo.ReleaseInviteToken(ctx, token); releaseErr != nil {
			logger.Ctx(ctx).Error("failed to release invite token after failed accept",
				logger.Int("user_id", int(userID)), logger.Error(releaseErr))
		}
		return 0, cause
	}

	if err := s.repo.SetPassword(ctx, userID, hash); err != nil {
		return failAccept(err)
	}

	// Resolve the business to activate the member in. The owner lookup is the
	// primary source; the member's own users.business_id is the fallback,
	// because it carries the same tenant and is set at invite time. Skipping
	// activation when the lookup fails used to surface as a successful accept
	// with the invitee stuck in pending_invitation.
	businessID, businessErr := s.repo.GetBusinessIDByOwnerID(ctx, ownerID)
	if businessErr != nil {
		if fallbackID, fallbackErr := s.repo.GetUserBusinessID(ctx, userID); fallbackErr == nil && fallbackID != 0 {
			businessID = fallbackID
		} else {
			logger.Ctx(ctx).Error("failed to resolve business for accepted invitation",
				logger.Int("user_id", int(userID)), logger.Int("owner_id", int(ownerID)), logger.Error(businessErr))
			return failAccept(fmt.Errorf("resolve business for accepted invitation: %w", businessErr))
		}
	}
	if stateErr := s.repo.UpsertProvisioningState(ctx, userID, businessID, "invitation", "active", true, false); stateErr != nil {
		return failAccept(stateErr)
	}
	if activateErr := s.repo.ActivateNormalizedAuthorization(ctx, businessID, userID); activateErr != nil {
		return failAccept(activateErr)
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
	_ = s.repo.BulkCreateHubNotification(ctx, ownerID, businessID, recipients, title, body, "info", link)
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
