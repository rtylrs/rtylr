package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	authorizationtransport "github.com/voxire/onion/internal/platform/authorization/transport"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubteamgroupsv1 "github.com/voxire/voxie/gen/go/hub/team/groups/v1"
)

type GroupsAPI struct {
	Service *application.Service
}

func NewGroupsAPI(service *application.Service) *GroupsAPI {
	return &GroupsAPI{Service: service}
}

func (a *GroupsAPI) ListGroups(ctx context.Context, _ *hubteamgroupsv1.ListGroupsRequest) (*hubteamgroupsv1.ListGroupsResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	businessID, ok := teamBusinessIDFromCtx(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	ownerID, ok := teamOwnerIDFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groups, err := a.Service.ListGroups(ctx, ownerID, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list groups")
	}

	payloads := make([]*hubteamgroupsv1.Group, 0, len(groups))
	for _, g := range groups {
		payloads = append(payloads, toProtoGroup(g))
	}

	return &hubteamgroupsv1.ListGroupsResponse{Groups: payloads}, nil
}

func (a *GroupsAPI) CreateGroup(ctx context.Context, req *hubteamgroupsv1.CreateGroupRequest) (*hubteamgroupsv1.CreateGroupResponse, error) {
	_, permissionErr := requireHubPermission(ctx, "hub:groups:manage")
	if permissionErr != nil {
		return nil, permissionErr
	}
	businessID, businessOK := teamBusinessIDFromCtx(ctx)
	if !businessOK {
		return nil, onekiterrors.PermissionDenied()
	}
	actorID, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	name := req.GetName()
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}

	assignments, assignmentErr := permissionProfileAssignmentsFromProto(req.GetPermissionProfileAssignments())
	if assignmentErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, assignmentErr.Error())
	}
	color := req.GetColor()
	if color == "" {
		color = "#6366f1"
	}

	id, err := a.Service.CreateGroup(ctx, actorID, ownerID, businessID, domain.CreateGroupInput{
		BusinessID:                   businessID,
		Name:                         name,
		Description:                  req.GetDescription(),
		Color:                        color,
		LicensedAppIDs:               req.GetLicensedAppIds(),
		PermissionProfileAssignments: assignments,
	})
	if err != nil {
		if errors.Is(err, domain.ErrDelegationCeiling) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create group")
	}

	g, err := a.Service.GetGroup(ctx, ownerID, businessID, id)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to fetch group")
	}

	return &hubteamgroupsv1.CreateGroupResponse{Group: toProtoGroup(*g)}, nil
}

func (a *GroupsAPI) GetGroupDetail(ctx context.Context, req *hubteamgroupsv1.GetGroupDetailRequest) (*hubteamgroupsv1.GetGroupDetailResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	businessID, ok := teamBusinessIDFromCtx(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	ownerID, ok := teamOwnerIDFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	if groupID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group id")
	}

	detail, err := a.Service.GetGroupDetail(ctx, ownerID, businessID, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) ||
			errors.Is(err, domain.ErrGroupNotFound) ||
			errors.Is(err, domain.ErrPermissionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Group not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get group detail")
	}

	return &hubteamgroupsv1.GetGroupDetailResponse{Group: toProtoGroupDetail(*detail)}, nil
}

func (a *GroupsAPI) UpdateGroup(ctx context.Context, req *hubteamgroupsv1.UpdateGroupRequest) (*hubteamgroupsv1.UpdateGroupResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	businessID, businessOK := teamBusinessIDFromCtx(ctx)
	if !businessOK {
		return nil, onekiterrors.PermissionDenied()
	}
	actorID, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	if groupID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group id")
	}

	name := req.GetName()
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}

	assignments, assignmentErr := permissionProfileAssignmentsFromProto(req.GetPermissionProfileAssignments())
	if assignmentErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, assignmentErr.Error())
	}
	if err := a.Service.UpdateGroup(ctx, actorID, ownerID, businessID, groupID, domain.UpdateGroupInput{
		BusinessID:                   businessID,
		Name:                         name,
		Description:                  req.GetDescription(),
		Color:                        req.GetColor(),
		LicensedAppIDs:               req.GetLicensedAppIds(),
		PermissionProfileAssignments: assignments,
	}); err != nil {
		if errors.Is(err, domain.ErrDelegationCeiling) {
			return nil, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrGroupNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Group not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update group")
	}

	g, err := a.Service.GetGroup(ctx, ownerID, businessID, groupID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to fetch group")
	}

	return &hubteamgroupsv1.UpdateGroupResponse{Group: toProtoGroup(*g)}, nil
}

func (a *GroupsAPI) DeleteGroup(ctx context.Context, req *hubteamgroupsv1.DeleteGroupRequest) (*hubteamgroupsv1.DeleteGroupResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	businessID, businessOK := teamBusinessIDFromCtx(ctx)
	if !businessOK {
		return nil, onekiterrors.PermissionDenied()
	}
	actorID, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	if groupID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group id")
	}

	if err := a.Service.DeleteGroup(ctx, actorID, ownerID, businessID, groupID); err != nil {
		if errors.Is(err, domain.ErrDelegationCeiling) {
			return nil, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrGroupNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Group not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete group")
	}

	return &hubteamgroupsv1.DeleteGroupResponse{Message: "Group deleted"}, nil
}

func (a *GroupsAPI) AddGroupMembers(ctx context.Context, req *hubteamgroupsv1.AddGroupMembersRequest) (*hubteamgroupsv1.AddGroupMembersResponse, error) {
	authUser, permissionErr := requireHubPermission(ctx, "hub:groups:manage")
	if permissionErr != nil {
		return nil, permissionErr
	}
	actorID, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	if groupID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group id")
	}

	userIDs := req.GetUserIds()
	if len(userIDs) == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "user_ids is required")
	}
	for _, userID := range userIDs {
		if userID <= 0 {
			return nil, onekiterrors.New(http.StatusBadRequest, "user_ids must contain positive ids")
		}
		if _, err := a.Service.GetStaffUserByID(ctx, ownerID, businessID, userID); err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				return nil, onekiterrors.New(http.StatusBadRequest, "user_ids must belong to the current business")
			}
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to validate members")
		}
	}

	if err := a.Service.AddGroupMembers(ctx, ownerID, groupID, businessID, actorID, userIDs); err != nil {
		if errors.Is(err, domain.ErrDelegationCeiling) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add members")
	}

	return &hubteamgroupsv1.AddGroupMembersResponse{
		Added:   int32(len(userIDs)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Message: "Members added",
	}, nil
}

func (a *GroupsAPI) RemoveGroupMember(ctx context.Context, req *hubteamgroupsv1.RemoveGroupMemberRequest) (*hubteamgroupsv1.RemoveGroupMemberResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	businessID, businessOK := teamBusinessIDFromCtx(ctx)
	if !businessOK {
		return nil, onekiterrors.PermissionDenied()
	}
	actorID, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	userID := req.GetUserId()
	if groupID <= 0 || userID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group or user id")
	}

	if err := a.Service.RemoveGroupMember(ctx, actorID, ownerID, businessID, groupID, userID); err != nil {
		if errors.Is(err, domain.ErrDelegationCeiling) {
			return nil, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrPermissionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Group not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove member")
	}

	return &hubteamgroupsv1.RemoveGroupMemberResponse{Message: "Member removed"}, nil
}

func (a *GroupsAPI) GetUserGroups(ctx context.Context, req *hubteamgroupsv1.GetUserGroupsRequest) (*hubteamgroupsv1.GetUserGroupsResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	businessID, ok := teamBusinessIDFromCtx(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	ownerID, ok := teamOwnerIDFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	userID := req.GetUserId()
	if userID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid user id")
	}

	groups, err := a.Service.GetUserGroups(ctx, ownerID, businessID, userID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get user groups")
	}

	payloads := make([]*hubteamgroupsv1.GroupSummary, 0, len(groups))
	for _, g := range groups {
		payloads = append(payloads, &hubteamgroupsv1.GroupSummary{Id: g.ID, Name: g.Name, Color: g.Color})
	}

	return &hubteamgroupsv1.GetUserGroupsResponse{Groups: payloads}, nil
}

func toProtoGroup(g domain.Group) *hubteamgroupsv1.Group {
	return &hubteamgroupsv1.Group{
		Id:                           g.ID,
		OwnerId:                      g.OwnerID,
		Name:                         g.Name,
		Description:                  g.Description,
		Color:                        g.Color,
		LicensedAppIds:               append([]string{}, g.LicensedAppIDs...),
		PermissionProfileAssignments: authorizationtransport.PermissionProfileAssignmentsToProto(g.PermissionProfileAssignments),
		MemberCount:                  int32(g.MemberCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CreatedAt:                    g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:                    g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toProtoGroupDetail(detail domain.GroupDetail) *hubteamgroupsv1.GroupDetail {
	members := make([]*hubteamgroupsv1.GroupMember, 0, len(detail.Members))
	for _, m := range detail.Members {
		members = append(members, &hubteamgroupsv1.GroupMember{
			UserId:   m.UserID,
			Email:    m.Email,
			FullName: m.FullName,
			AddedAt:  m.AddedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	g := toProtoGroup(detail.Group)
	return &hubteamgroupsv1.GroupDetail{
		Id:                           g.GetId(),
		OwnerId:                      g.GetOwnerId(),
		Name:                         g.GetName(),
		Description:                  g.GetDescription(),
		Color:                        g.GetColor(),
		LicensedAppIds:               g.GetLicensedAppIds(),
		PermissionProfileAssignments: g.GetPermissionProfileAssignments(),
		MemberCount:                  int32(len(members)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CreatedAt:                    g.GetCreatedAt(),
		UpdatedAt:                    g.GetUpdatedAt(),
		Members:                      members,
	}
}
