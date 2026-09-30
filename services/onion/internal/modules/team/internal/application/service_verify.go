package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

// VerifyInviteToken checks if the token is valid and returns the user info along with their app access list.
func (s *Service) VerifyInviteToken(ctx context.Context, token string) (int64, string, string, []string, error) {
	invite, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, "", "", nil, err
	}

	email, fullName, _, err := s.repo.GetUserByID(ctx, invite.UserID)
	if err != nil {
		return 0, "", "", nil, err
	}

	// Fetch the user's app permissions so the frontend can redirect to the right app after accept.
	apps := []string{}
	if perm, permErr := s.repo.GetPermission(ctx, invite.OwnerID, invite.UserID); permErr == nil && perm != nil {
		apps = perm.Apps
	}

	return invite.UserID, email, fullName, apps, nil
}

// resolveInviteBusinessID returns the business an accepted invitation belongs
// to, verified as the invitee's.
//
// The invite row is the authoritative binding: InviteUser records the business
// it granted access in, so that is what the accept activates. users.business_id
// is deliberately not consulted. It is the member's *own* default business, not
// the invitation's, so it names a different tenant whenever somebody who
// already belongs to one is invited to another — the previous code substituted
// it whenever the owner lookup failed and then activated the member and granted
// its authorization in that unrelated business.
//
// Invitations issued before the business was recorded on the row fall back to
// the inviter's business and are verified the same way, so no write ever lands
// under an id that has not been positively checked for this member.
func (s *Service) resolveInviteBusinessID(ctx context.Context, invite domain.InviteToken) (int64, error) {
	candidate := invite.BusinessID
	if candidate <= 0 {
		ownerBusinessID, err := s.repo.GetBusinessIDByOwnerID(ctx, invite.OwnerID)
		if err != nil {
			return 0, err
		}
		candidate = ownerBusinessID
	}
	if candidate <= 0 {
		return 0, tenant.ErrNoBusinessScope
	}
	return tenant.ResolveBusinessID(ctx, candidate, invite.UserID)
}

// AcceptInvite sets the user's password and marks the token as used.
func (s *Service) AcceptInvite(ctx context.Context, token, password string) (int64, error) {
	if err := passwordutils.ValidateStrength(password); err != nil {
		return 0, err
	}

	invite, err := s.repo.GetInviteToken(ctx, token)
	if err != nil {
		return 0, err
	}
	userID, ownerID := invite.UserID, invite.OwnerID

	// Resolve and verify the business before the token is claimed: an
	// invitation whose tenant cannot be verified must not write anything at
	// all, so this runs ahead of every write below.
	businessID, err := s.resolveInviteBusinessID(ctx, invite)
	if err != nil {
		logger.Ctx(ctx).Error("failed to resolve business for accepted invitation",
			logger.Int("user_id", int(userID)), logger.Int("owner_id", int(ownerID)), logger.Error(err))
		return 0, fmt.Errorf("resolve business for accepted invitation: %w", err)
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
