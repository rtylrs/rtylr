package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
)

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// GetOwnerApps returns the list of apps the owner is subscribed to.
func (s *Service) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	return s.repo.GetOwnerApps(ctx, ownerID)
}

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (s *Service) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	return s.repo.GetOwnerSubscriptionInfo(ctx, ownerID)
}

// seatChangeStore is the repository surface needed to record a seat change.
// Keeping it as an interface lets the record-then-invalidate flow be tested
// without a database.
type seatChangeStore interface {
	CountActiveMembers(ctx context.Context, ownerID int64) (int, error)
	RecordSeatChange(ctx context.Context, ownerID int64, newCount int) error
}

// recordSeatChange persists the owner's current active-member count as the
// billable seat count and then invalidates the cached subscription payload.
func recordSeatChange(ctx context.Context, store seatChangeStore, ownerID int64) error {
	newCount, err := store.CountActiveMembers(ctx, ownerID)
	if err != nil {
		return err
	}

	if err := store.RecordSeatChange(ctx, ownerID, newCount); err != nil {
		return err
	}

	invalidateSeatSubscriptionScope(ctx, ownerID)
	return nil
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// It is invoked asynchronously so it never blocks team operations, but it returns its error so the
// caller can log billing failures instead of silently discarding them (which would corrupt billing).
func (s *Service) RecordSeatChange(ctx context.Context, ownerID int64) error {
	return recordSeatChange(ctx, s.repo, ownerID)
}

// ─── Groups ──────────────────────────────────────────────────────────────────

func (s *Service) CreateGroup(ctx context.Context, actorID, ownerID, businessID int64, input domain.CreateGroupInput) (int64, error) {
	if input.Name == "" || businessID <= 0 {
		return 0, errors.New("business and name are required")
	}
	if input.Color == "" {
		input.Color = "#6366f1"
	}
	input.BusinessID = businessID
	if err := s.enforceAuthorizationDelegation(ctx, actorID, ownerID, businessID, domain.UpdatePermissionsInput{
		LicensedAppIDs: &input.LicensedAppIDs, PermissionProfileAssignments: &input.PermissionProfileAssignments,
	}); err != nil {
		return 0, err
	}
	return s.repo.CreateGroup(ctx, ownerID, input)
}

func (s *Service) ListGroups(ctx context.Context, ownerID, businessID int64) ([]domain.Group, error) {
	return s.repo.ListGroups(ctx, ownerID, businessID)
}

func (s *Service) GetGroup(ctx context.Context, ownerID, businessID, groupID int64) (*domain.Group, error) {
	return s.repo.GetGroup(ctx, ownerID, businessID, groupID)
}

func (s *Service) GetGroupDetail(ctx context.Context, ownerID, businessID, groupID int64) (*domain.GroupDetail, error) {
	return s.repo.GetGroupDetail(ctx, ownerID, businessID, groupID)
}

func (s *Service) UpdateGroup(ctx context.Context, actorID, ownerID, businessID, groupID int64, input domain.UpdateGroupInput) error {
	if input.Name == "" || businessID <= 0 {
		return errors.New("business and name are required")
	}
	input.BusinessID = businessID
	if err := s.enforceAuthorizationDelegation(ctx, actorID, ownerID, businessID, domain.UpdatePermissionsInput{
		LicensedAppIDs: &input.LicensedAppIDs, PermissionProfileAssignments: &input.PermissionProfileAssignments,
	}); err != nil {
		return err
	}
	if err := s.repo.UpdateGroup(ctx, ownerID, businessID, groupID, input); err != nil {
		return err
	}
	s.invalidateGroupMemberScopes(ctx, ownerID, businessID, groupID)
	return nil
}

func (s *Service) DeleteGroup(ctx context.Context, ownerID, businessID, groupID int64) error {
	memberIDs := s.groupMemberIDs(ctx, ownerID, businessID, groupID)
	if err := s.repo.DeleteGroup(ctx, ownerID, businessID, groupID); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, ownerID, memberIDs)
	return nil
}

func (s *Service) AddGroupMembers(ctx context.Context, ownerID, groupID, businessID, callerID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return errors.New("no user IDs provided")
	}
	// Verify the group belongs to this owner before adding members — otherwise
	// a caller could target a group id owned by another business.
	if _, err := s.repo.GetGroup(ctx, ownerID, businessID, groupID); err != nil {
		return err
	}
	// businessID is the tenant the group's grants are resolved in; a membership
	// row without it is invisible to the authorization snapshot readers.
	if err := s.repo.AddGroupMembers(ctx, ownerID, groupID, businessID, callerID, userIDs); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, ownerID, userIDs)
	return nil
}

func (s *Service) RemoveGroupMember(ctx context.Context, ownerID, businessID, groupID, userID int64) error {
	if _, err := s.repo.GetGroup(ctx, ownerID, businessID, groupID); err != nil {
		return domain.ErrPermissionNotFound
	}
	if err := s.repo.RemoveGroupMember(ctx, ownerID, businessID, groupID, userID); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, ownerID, []int64{userID})
	return nil
}

func (s *Service) GetUserGroups(ctx context.Context, ownerID, businessID, userID int64) ([]domain.GroupSummary, error) {
	return s.repo.GetUserGroups(ctx, ownerID, businessID, userID)
}

func (s *Service) invalidateGroupMemberScopes(ctx context.Context, ownerID, businessID, groupID int64) {
	s.invalidateMemberScopes(ctx, ownerID, s.groupMemberIDs(ctx, ownerID, businessID, groupID))
}

func (s *Service) groupMemberIDs(ctx context.Context, ownerID, businessID, groupID int64) []int64 {
	detail, err := s.repo.GetGroupDetail(ctx, ownerID, businessID, groupID)
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

// BulkInvite invites multiple users with the same member type, licenses, profiles and groups.
func (s *Service) BulkInvite(ctx context.Context, ownerID, businessID int64, input domain.BulkInviteInput) ([]domain.BulkInviteResult, error) {
	results := make([]domain.BulkInviteResult, 0, len(input.Emails))
	for _, email := range input.Emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}
		userID, token, err := s.InviteUser(ctx, ownerID, businessID, domain.InviteUserInput{
			Email: email, MemberType: input.MemberType, LicensedAppIDs: input.LicensedAppIDs,
			PermissionProfileAssignments: input.PermissionProfileAssignments,
			GroupIDs:                     input.GroupIDs,
		})
		if err != nil {
			message := err.Error()
			if len(positiveInt64s(input.GroupIDs)) > 0 && strings.Contains(strings.ToLower(message), "group") {
				message = fmt.Sprintf("group assignment failed: %v", err)
			}
			results = append(results, domain.BulkInviteResult{
				Email:   email,
				Status:  "error",
				Message: message,
			})
			continue
		}
		results = append(results, domain.BulkInviteResult{Email: email, UserID: userID, Token: token, Status: "invited", Message: "Invitation created"})
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
