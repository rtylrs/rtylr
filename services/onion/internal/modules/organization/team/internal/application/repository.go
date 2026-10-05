package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
)

type repository interface {
	ActivateNormalizedAuthorization(ctx context.Context, businessID, userID int64) error
	AddGroupMembers(ctx context.Context, ownerID, groupID, businessID, addedBy int64, userIDs []int64) error
	AttachUserToBusiness(ctx context.Context, userID, businessID int64) (bool, error)
	CountActiveMembers(ctx context.Context, ownerID int64) (int, error)
	CountStaffUsersByBusinessID(ctx context.Context, businessID int64) (int64, error)
	CreateAccessRequest(ctx context.Context, businessID, userID int64, app string) error
	CreateGroup(ctx context.Context, ownerID int64, input domain.CreateGroupInput) (int64, error)
	CreateInviteToken(ctx context.Context, userID, ownerID, businessID int64, token string, expiresAt time.Time) error
	CreateUser(ctx context.Context, email, username, passwordHash, fullName string) (int64, error)
	DeactivateUser(ctx context.Context, businessID, ownerID, actorID, userID int64) error
	DeleteGroup(ctx context.Context, ownerID, businessID, groupID int64) error
	DeleteInvitedUser(ctx context.Context, businessID, ownerID, userID int64) error
	GetAdminsAndOwnerUserIDs(ctx context.Context, businessID int64) ([]int64, error)
	GetAuthorizationRoleGrant(ctx context.Context, businessID, roleID int64) (domain.AuthorizationRoleGrant, error)
	GetAuthorizationSnapshot(ctx context.Context, businessID, userID int64, businessApps []string) (authorization.Snapshot, bool, error)
	GetBusinessByID(ctx context.Context, businessID int64) (string, int64, error)
	GetBusinessIDByOwnerID(ctx context.Context, ownerID int64) (int64, error)
	GetGroup(ctx context.Context, ownerID, businessID, groupID int64) (*domain.Group, error)
	GetGroupDetail(ctx context.Context, ownerID, businessID, groupID int64) (*domain.GroupDetail, error)
	GetInviteToken(ctx context.Context, token string) (domain.InviteToken, error)
	GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error)
	GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error)
	GetProvisioningStatus(ctx context.Context, userID int64) (*authorization.ProvisioningStatus, bool, error)
	GetProvisioningStatusForBusiness(ctx context.Context, businessID, userID int64) (*authorization.ProvisioningStatus, bool, error)
	GetStaffUserByID(ctx context.Context, ownerID, businessID, userID int64) (*domain.StaffUser, error)
	GetUserBusinessID(ctx context.Context, id int64) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (int64, string, string, error)
	GetUserByID(ctx context.Context, id int64) (string, string, bool, error)
	GetUserGroups(ctx context.Context, ownerID, businessID, userID int64) ([]domain.GroupSummary, error)
	IsMFAVerified(ctx context.Context, userID int64) (bool, error)
	IsMFAVerifiedForBusiness(ctx context.Context, businessID, userID int64) (bool, error)
	IsOnboarded(ctx context.Context, userID int64) bool
	ListAccessRequests(ctx context.Context, businessID int64) ([]domain.AccessRequest, error)
	ListGroups(ctx context.Context, ownerID, businessID int64) ([]domain.Group, error)
	ListStaffUsersByBusinessID(ctx context.Context, businessID int64, limit, offset int32) ([]domain.StaffUser, error)
	MarkInviteTokenUsed(ctx context.Context, token string) error
	MarkOnboarded(ctx context.Context, userID int64) error
	MarkPasswordChanged(ctx context.Context, userID int64) error
	RecordSeatChange(ctx context.Context, ownerID int64, newCount int) error
	ReleaseInviteToken(ctx context.Context, token string) error
	RemoveGroupMember(ctx context.Context, ownerID, businessID, groupID, userID int64) error
	ResolveAccessRequest(ctx context.Context, id, businessID int64, status string) error
	RevokeBusinessUserPermissions(ctx context.Context, businessID, userID int64) (int64, error)
	SetNormalizedAuthorizationPending(ctx context.Context, businessID, userID int64) error
	SetPassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateGroup(ctx context.Context, ownerID, businessID, groupID int64, input domain.UpdateGroupInput) error
	UpdateUserProfile(ctx context.Context, userID int64, fullName string) error
	UpsertNormalizedAuthorization(ctx context.Context, businessID, ownerID, actorID, userID int64, apps []string, memberType, memberStatus string, assignments []domain.AppRoleAssignmentInput, groupIDs []int64) error
	UpsertProvisioningState(ctx context.Context, userID, businessID int64, mode, state string, emailVerified, passwordChangeRequired bool) error
}

type teamEmailNotifications interface {
	QueueInvite(ctx context.Context, toEmail, inviteLink string) error
	QueueAccessRequest(ctx context.Context, toEmail, requesterName, requesterEmail, app string) error
}
