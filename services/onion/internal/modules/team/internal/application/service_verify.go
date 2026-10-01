package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
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

	return invite.UserID, email, fullName, []string{}, nil
}

// resolveInviteBusinessID returns the business recorded on the invitation and
// verifies that the inviter owns it.
//
// The invite row is the authoritative binding: InviteUser records the business
// it granted access in, so that is what the accept activates. users.business_id
// is deliberately not consulted. It is the member's *own* default business, not
// the invitation's, so it names a different tenant whenever somebody who
// already belongs to one is invited to another — the previous code substituted
// it whenever the owner lookup failed and then activated the member and granted
// its authorization in that unrelated business.
//
// The invitation itself grants pending membership. Requiring an active
// membership before accepting would make new invitations impossible to accept.
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
	_, ownerID, err := s.repo.GetBusinessByID(ctx, candidate)
	if err != nil {
		return 0, err
	}
	if ownerID != invite.OwnerID {
		return 0, tenant.ErrNoBusinessScope
	}
	return candidate, nil
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
	apps := append([]string{}, existing.LicensedAppIDs...)
	for _, licensedApp := range apps {
		if licensedApp == app {
			return s.repo.ResolveAccessRequest(ctx, requestID, businessID, "approved")
		}
	}
	apps = append(apps, app)
	assignments := make([]domain.AppRoleAssignmentInput, 0, len(existing.PermissionProfileAssignments))
	for _, assignment := range existing.PermissionProfileAssignments {
		if assignment.GroupID != nil || assignment.Source != authorization.SourceExplicit {
			continue
		}
		assignments = append(assignments, domain.AppRoleAssignmentInput{
			AppID: string(assignment.AppID), RoleIDs: append([]int64{}, assignment.RoleIDs...),
			PrimaryRoleID: assignment.PrimaryRoleID, Source: string(authorization.SourceExplicit),
			Status: "active", Scopes: append([]authorization.Scope{}, assignment.Scopes...),
		})
	}
	groupIDs := make([]int64, 0, len(existing.Groups))
	for _, group := range existing.Groups {
		groupIDs = append(groupIDs, group.ID)
	}
	memberType := existing.MemberType
	if memberType == "" {
		memberType = "member"
	}
	if err := s.repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, actorID, targetUserID, apps, memberType, "active", assignments, groupIDs); err != nil {
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
