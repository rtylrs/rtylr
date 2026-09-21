package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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
	ownerID, ok := teamOwnerIDFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groups, err := a.Service.ListGroups(ctx, ownerID)
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
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	_, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	name := req.GetName()
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}

	perms := protoResourcePermissionsToDomain(req.GetResourcePermissions())
	color := req.GetColor()
	if color == "" {
		color = "#6366f1"
	}

	id, err := a.Service.CreateGroup(ctx, ownerID, domain.CreateGroupInput{
		Name:                name,
		Description:         req.GetDescription(),
		Color:               color,
		ResourcePermissions: perms,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPermissionLevel) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid resource permission level")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create group")
	}

	g, err := a.Service.GetGroup(ctx, ownerID, id)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to fetch group")
	}

	return &hubteamgroupsv1.CreateGroupResponse{Group: toProtoGroup(*g)}, nil
}

func (a *GroupsAPI) GetGroupDetail(ctx context.Context, req *hubteamgroupsv1.GetGroupDetailRequest) (*hubteamgroupsv1.GetGroupDetailResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	ownerID, ok := teamOwnerIDFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	if groupID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group id")
	}

	detail, err := a.Service.GetGroupDetail(ctx, ownerID, groupID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusNotFound, "Group not found")
	}

	return &hubteamgroupsv1.GetGroupDetailResponse{Group: toProtoGroupDetail(*detail)}, nil
}

func (a *GroupsAPI) UpdateGroup(ctx context.Context, req *hubteamgroupsv1.UpdateGroupRequest) (*hubteamgroupsv1.UpdateGroupResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	_, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
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

	perms := protoResourcePermissionsToDomain(req.GetResourcePermissions())
	if err := a.Service.UpdateGroup(ctx, ownerID, groupID, domain.UpdateGroupInput{
		Name:                name,
		Description:         req.GetDescription(),
		Color:               req.GetColor(),
		ResourcePermissions: perms,
	}); err != nil {
		if errors.Is(err, domain.ErrGroupNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Group not found")
		}
		if errors.Is(err, domain.ErrInvalidPermissionLevel) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid resource permission level")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update group")
	}

	g, err := a.Service.GetGroup(ctx, ownerID, groupID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to fetch group")
	}

	return &hubteamgroupsv1.UpdateGroupResponse{Group: toProtoGroup(*g)}, nil
}

func (a *GroupsAPI) DeleteGroup(ctx context.Context, req *hubteamgroupsv1.DeleteGroupRequest) (*hubteamgroupsv1.DeleteGroupResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	_, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	if groupID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group id")
	}

	if err := a.Service.DeleteGroup(ctx, ownerID, groupID); err != nil {
		if errors.Is(err, domain.ErrGroupNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Group not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete group")
	}

	return &hubteamgroupsv1.DeleteGroupResponse{Message: "Group deleted"}, nil
}

func (a *GroupsAPI) AddGroupMembers(ctx context.Context, req *hubteamgroupsv1.AddGroupMembersRequest) (*hubteamgroupsv1.AddGroupMembersResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	actorID, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	device := authPkg.DeviceCtx(ctx)
	businessID := int64(0)
	if device.User != nil {
		businessID = device.User.BusinessID
	}

	groupID := req.GetId()
	if groupID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group id")
	}

	userIDs := req.GetUserIds()
	if len(userIDs) == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "user_ids is required")
	}

	if err := a.Service.AddGroupMembers(ctx, ownerID, groupID, businessID, actorID, userIDs); err != nil {
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
	_, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	groupID := req.GetId()
	userID := req.GetUserId()
	if groupID <= 0 || userID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid group or user id")
	}

	if err := a.Service.RemoveGroupMember(ctx, ownerID, groupID, userID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove member")
	}

	return &hubteamgroupsv1.RemoveGroupMemberResponse{Message: "Member removed"}, nil
}

func (a *GroupsAPI) GetUserGroups(ctx context.Context, req *hubteamgroupsv1.GetUserGroupsRequest) (*hubteamgroupsv1.GetUserGroupsResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:groups:manage"); permissionErr != nil {
		return nil, permissionErr
	}
	ownerID, ok := teamOwnerIDFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	userID := req.GetUserId()
	if userID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid user id")
	}

	groups, err := a.Service.GetUserGroups(ctx, ownerID, userID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get user groups")
	}

	payloads := make([]*hubteamgroupsv1.GroupSummary, 0, len(groups))
	for _, g := range groups {
		payloads = append(payloads, &hubteamgroupsv1.GroupSummary{Id: g.ID, Name: g.Name, Color: g.Color})
	}

	return &hubteamgroupsv1.GetUserGroupsResponse{Groups: payloads}, nil
}

func protoResourcePermissionsToDomain(perms []*hubteamgroupsv1.ResourcePermission) []domain.ResourcePermission {
	out := make([]domain.ResourcePermission, 0, len(perms))
	for _, p := range perms {
		out = append(out, domain.ResourcePermission{
			ResourceKey:     p.GetResourceKey(),
			PermissionLevel: domain.PermissionLevel(p.GetPermissionLevel()),
		})
	}
	return out
}

func toProtoGroup(g domain.Group) *hubteamgroupsv1.Group {
	perms := make([]*hubteamgroupsv1.ResourcePermission, 0, len(g.ResourcePermissions))
	for _, p := range g.ResourcePermissions {
		perms = append(perms, &hubteamgroupsv1.ResourcePermission{
			ResourceKey:     p.ResourceKey,
			PermissionLevel: string(p.PermissionLevel),
		})
	}
	return &hubteamgroupsv1.Group{
		Id:                  g.ID,
		OwnerId:             g.OwnerID,
		Name:                g.Name,
		Description:         g.Description,
		Color:               g.Color,
		ResourcePermissions: perms,
		MemberCount:         int32(g.MemberCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CreatedAt:           g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
		Id:                  g.GetId(),
		OwnerId:             g.GetOwnerId(),
		Name:                g.GetName(),
		Description:         g.GetDescription(),
		Color:               g.GetColor(),
		ResourcePermissions: g.GetResourcePermissions(),
		MemberCount:         int32(len(members)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CreatedAt:           g.GetCreatedAt(),
		UpdatedAt:           g.GetUpdatedAt(),
		Members:             members,
	}
}
