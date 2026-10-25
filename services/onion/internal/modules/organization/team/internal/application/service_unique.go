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

// GetBusinessApps returns the apps included in the business's plan.
func (s *Service) GetBusinessApps(ctx context.Context, businessID int64) ([]string, error) {
	return s.repo.GetBusinessApps(ctx, businessID)
}

// GetBusinessSubscriptionInfo returns the plan slug and status for the business's subscription.
func (s *Service) GetBusinessSubscriptionInfo(ctx context.Context, businessID int64) (string, string, error) {
	return s.repo.GetBusinessSubscriptionInfo(ctx, businessID)
}

// seatChangeStore is the repository surface needed to record a seat change.
// Keeping it as an interface lets the record-then-invalidate flow be tested
// without a database.
type seatChangeStore interface {
	CountActiveMembers(ctx context.Context, businessID int64) (int, error)
	RecordSeatChange(ctx context.Context, businessID int64, newCount int) error
}

// recordSeatChange persists the business's current active-member count as the
// billable seat count and then invalidates the cached subscription payload.
func recordSeatChange(ctx context.Context, store seatChangeStore, businessID int64) error {
	newCount, err := store.CountActiveMembers(ctx, businessID)
	if err != nil {
		return err
	}

	if err := store.RecordSeatChange(ctx, businessID, newCount); err != nil {
		return err
	}

	invalidateSeatSubscriptionScope(ctx, businessID)
	return nil
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// It is invoked asynchronously so it never blocks team operations, but it returns its error so the
// caller can log billing failures instead of silently discarding them (which would corrupt billing).
func (s *Service) RecordSeatChange(ctx context.Context, businessID int64) error {
	return recordSeatChange(ctx, s.repo, businessID)
}

// ─── Groups ──────────────────────────────────────────────────────────────────

func (s *Service) CreateGroup(ctx context.Context, actorID, businessID int64, input domain.CreateGroupInput) (int64, error) {
	if input.Name == "" || businessID <= 0 {
		return 0, errors.New("business and name are required")
	}
	if input.Color == "" {
		input.Color = "#6366f1"
	}
	input.BusinessID = businessID
	if err := s.enforceAuthorizationDelegation(ctx, actorID, businessID, domain.UpdatePermissionsInput{
		LicensedAppIDs: &input.LicensedAppIDs, PermissionProfileAssignments: &input.PermissionProfileAssignments,
	}, nil); err != nil {
		return 0, err
	}
	groupID, err := s.repo.CreateGroup(ctx, actorID, input)
	if err != nil {
		return 0, err
	}
	s.recordAudit(ctx, businessID, actorID, "group", groupID, "group.created", fmt.Sprintf("Created group %q", input.Name))
	return groupID, nil
}

func (s *Service) ListGroups(ctx context.Context, businessID int64) ([]domain.Group, error) {
	return s.repo.ListGroups(ctx, businessID)
}

func (s *Service) GetGroup(ctx context.Context, businessID, groupID int64) (*domain.Group, error) {
	return s.repo.GetGroup(ctx, businessID, groupID)
}

func (s *Service) GetGroupDetail(ctx context.Context, businessID, groupID int64) (*domain.GroupDetail, error) {
	return s.repo.GetGroupDetail(ctx, businessID, groupID)
}

func (s *Service) UpdateGroup(ctx context.Context, actorID, businessID, groupID int64, input domain.UpdateGroupInput) error {
	if input.Name == "" || businessID <= 0 {
		return errors.New("business and name are required")
	}
	input.BusinessID = businessID
	group, err := s.repo.GetGroup(ctx, businessID, groupID)
	if err != nil {
		return err
	}
	if err := s.enforceAuthorizationDelegation(ctx, actorID, businessID, domain.UpdatePermissionsInput{
		LicensedAppIDs: &input.LicensedAppIDs, PermissionProfileAssignments: &input.PermissionProfileAssignments,
	}, &delegationCurrent{Grants: grantsOfGroupAssignments(group.PermissionProfileAssignments)}); err != nil {
		return err
	}
	if err := s.repo.UpdateGroup(ctx, actorID, businessID, groupID, input); err != nil {
		return err
	}
	s.invalidateGroupMemberScopes(ctx, businessID, groupID)
	s.recordAudit(ctx, businessID, actorID, "group", groupID, "group.updated", fmt.Sprintf("Updated group %q", input.Name))
	return nil
}

func (s *Service) DeleteGroup(ctx context.Context, actorID, businessID, groupID int64) error {
	if group, lookupErr := s.repo.GetGroup(ctx, businessID, groupID); lookupErr == nil {
		if err := s.enforceGroupMembershipDelegation(ctx, actorID, businessID, group); err != nil {
			return err
		}
	}
	memberIDs := s.groupMemberIDs(ctx, businessID, groupID)
	description := "Deleted a group"
	if group, lookupErr := s.repo.GetGroup(ctx, businessID, groupID); lookupErr == nil && group != nil {
		description = fmt.Sprintf("Deleted group %q", group.Name)
	}
	if err := s.repo.DeleteGroup(ctx, businessID, groupID); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, businessID, memberIDs)
	s.recordAudit(ctx, businessID, 0, "group", groupID, "group.deleted", description)
	return nil
}

func (s *Service) AddGroupMembers(ctx context.Context, groupID, businessID, callerID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return errors.New("no user IDs provided")
	}
	// Verify the group belongs to this business before adding members —
	// otherwise a caller could target a group id owned by another business.
	group, err := s.repo.GetGroup(ctx, businessID, groupID)
	if err != nil {
		return err
	}
	if err := s.enforceGroupMembershipDelegation(ctx, callerID, businessID, group); err != nil {
		return err
	}
	// businessID is the tenant the group's grants are resolved in; a membership
	// row without it is invisible to the authorization snapshot readers.
	if err := s.repo.AddGroupMembers(ctx, groupID, businessID, callerID, userIDs); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, businessID, userIDs)
	noun := "members"
	if len(userIDs) == 1 {
		noun = "member"
	}
	s.recordAudit(ctx, businessID, callerID, "group", groupID, "group.members_added",
		fmt.Sprintf("Added %d %s to group %q", len(userIDs), noun, group.Name))
	return nil
}

func (s *Service) RemoveGroupMember(ctx context.Context, actorID, businessID, groupID, userID int64) error {
	group, err := s.repo.GetGroup(ctx, businessID, groupID)
	if err != nil {
		return domain.ErrPermissionNotFound
	}
	if err := s.enforceGroupMembershipDelegation(ctx, actorID, businessID, group); err != nil {
		return err
	}
	if err := s.repo.RemoveGroupMember(ctx, businessID, groupID, userID); err != nil {
		return err
	}
	s.invalidateMemberScopes(ctx, businessID, []int64{userID})
	s.recordAudit(ctx, businessID, 0, "group", groupID, "group.member_removed",
		fmt.Sprintf("Removed a member from group %q", group.Name))
	return nil
}

func (s *Service) GetUserGroups(ctx context.Context, businessID, userID int64) ([]domain.GroupSummary, error) {
	return s.repo.GetUserGroups(ctx, businessID, userID)
}

func (s *Service) invalidateGroupMemberScopes(ctx context.Context, businessID, groupID int64) {
	s.invalidateMemberScopes(ctx, businessID, s.groupMemberIDs(ctx, businessID, groupID))
}

func (s *Service) groupMemberIDs(ctx context.Context, businessID, groupID int64) []int64 {
	detail, err := s.repo.GetGroupDetail(ctx, businessID, groupID)
	if err != nil || detail == nil {
		return nil
	}
	userIDs := make([]int64, 0, len(detail.Members))
	for _, member := range detail.Members {
		userIDs = append(userIDs, member.UserID)
	}
	return userIDs
}

func (s *Service) invalidateMemberScopes(ctx context.Context, businessID int64, userIDs []int64) {
	for _, userID := range userIDs {
		if userID != 0 {
			invalidateUserScope(ctx, userID)
		}
	}
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
}

// BulkInvite invites multiple users with the same member type, licenses, profiles and groups.
func (s *Service) BulkInvite(ctx context.Context, actorID, businessID int64, input domain.BulkInviteInput) ([]domain.BulkInviteResult, error) {
	results := make([]domain.BulkInviteResult, 0, len(input.Emails))
	for _, email := range input.Emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}
		userID, token, err := s.InviteUser(ctx, actorID, businessID, domain.InviteUserInput{
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
