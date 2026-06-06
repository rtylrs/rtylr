package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── shared payload types ────────────────────────────────────────────────────

type ResourcePermissionPayload struct {
	ResourceKey     string `json:"resource_key"`
	PermissionLevel string `json:"permission_level"`
}

type GroupSummaryPayload struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type GroupPayload struct {
	ID                  int64                       `json:"id"`
	OwnerID             int64                       `json:"owner_id"`
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	Color               string                      `json:"color"`
	ResourcePermissions []ResourcePermissionPayload `json:"resource_permissions"`
	MemberCount         int                         `json:"member_count"`
	CreatedAt           string                      `json:"created_at"`
	UpdatedAt           string                      `json:"updated_at"`
}

type GroupMemberPayload struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	AddedAt  string `json:"added_at"`
}

type GroupDetailPayload struct {
	GroupPayload
	Members []GroupMemberPayload `json:"members"`
}

func toGroupPayload(g domain.Group) GroupPayload {
	perms := make([]ResourcePermissionPayload, 0, len(g.ResourcePermissions))
	for _, p := range g.ResourcePermissions {
		perms = append(perms, ResourcePermissionPayload{ResourceKey: p.ResourceKey, PermissionLevel: string(p.PermissionLevel)})
	}
	return GroupPayload{
		ID:                  g.ID,
		OwnerID:             g.OwnerID,
		Name:                g.Name,
		Description:         g.Description,
		Color:               g.Color,
		ResourcePermissions: perms,
		MemberCount:         g.MemberCount,
		CreatedAt:           g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func teamOwnerIDFromCtx(ctx context.Context, service *application.Service) (int64, bool) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, false
	}
	ownerID, err := service.BusinessOwnerID(ctx, device.User.UserID, device.User.BusinessID)
	return ownerID, err == nil && ownerID != 0
}

func teamManagerFromCtx(ctx context.Context, service *application.Service) (actorID, ownerID int64, ok bool) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, 0, false
	}
	ownerID, err := service.TeamManagerOwnerID(ctx, device.User.UserID, device.User.BusinessID)
	return device.User.UserID, ownerID, err == nil && ownerID != 0
}

// ─── List Groups ─────────────────────────────────────────────────────────────

func ListGroupsRoutes() []string { return []string{"/team/groups"} }

type (
	ListGroupsHandler  struct{ Service *application.Service }
	ListGroupsRequest  struct{}
	ListGroupsResponse struct {
		Groups []GroupPayload `json:"groups"`
	}
)

func (h *ListGroupsHandler) Serve(ctx context.Context, _ *ListGroupsRequest) (ListGroupsResponse, *errorModels.ServerError) {
	ownerID, ok := teamOwnerIDFromCtx(ctx, h.Service)
	if !ok {
		return ListGroupsResponse{}, errorModels.PermissionDeniedServerError()
	}
	groups, err := h.Service.ListGroups(ctx, ownerID)
	if err != nil {
		return ListGroupsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list groups"}
	}
	payloads := make([]GroupPayload, 0, len(groups))
	for _, g := range groups {
		payloads = append(payloads, toGroupPayload(g))
	}
	return ListGroupsResponse{Groups: payloads}, nil
}

// ─── Create Group ─────────────────────────────────────────────────────────────

type (
	CreateGroupHandler struct{ Service *application.Service }
	CreateGroupRequest struct {
		Name                string                      `json:"name"`
		Description         string                      `json:"description"`
		Color               string                      `json:"color"`
		ResourcePermissions []ResourcePermissionPayload `json:"resource_permissions"`
	}
)

type CreateGroupResponse struct {
	Group GroupPayload `json:"group"`
}

func (h *CreateGroupHandler) Validate(_ context.Context, req *CreateGroupRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateGroupHandler) Serve(ctx context.Context, req *CreateGroupRequest) (CreateGroupResponse, *errorModels.ServerError) {
	_, ownerID, ok := teamManagerFromCtx(ctx, h.Service)
	if !ok {
		return CreateGroupResponse{}, errorModels.PermissionDeniedServerError()
	}
	perms := make([]domain.ResourcePermission, 0, len(req.ResourcePermissions))
	for _, p := range req.ResourcePermissions {
		perms = append(perms, domain.ResourcePermission{ResourceKey: p.ResourceKey, PermissionLevel: domain.PermissionLevel(p.PermissionLevel)})
	}
	color := req.Color
	if color == "" {
		color = "#6366f1"
	}
	id, err := h.Service.CreateGroup(ctx, ownerID, domain.CreateGroupInput{
		Name: req.Name, Description: req.Description, Color: color, ResourcePermissions: perms,
	})
	if err != nil {
		return CreateGroupResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create group"}
	}
	g, err := h.Service.GetGroup(ctx, ownerID, id)
	if err != nil {
		return CreateGroupResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to fetch group"}
	}
	return CreateGroupResponse{Group: toGroupPayload(*g)}, nil
}

// ─── Get Group Detail ─────────────────────────────────────────────────────────

func GetGroupDetailRoutes() []string { return []string{"/team/groups/{id}"} }

type (
	GetGroupDetailHandler  struct{ Service *application.Service }
	GetGroupDetailRequest  struct{}
	GetGroupDetailResponse struct {
		Group GroupDetailPayload `json:"group"`
	}
)

func (h *GetGroupDetailHandler) Serve(ctx context.Context, _ *GetGroupDetailRequest) (GetGroupDetailResponse, *errorModels.ServerError) {
	ownerID, ok := teamOwnerIDFromCtx(ctx, h.Service)
	if !ok {
		return GetGroupDetailResponse{}, errorModels.PermissionDeniedServerError()
	}
	idRaw := chi.URLParamFromCtx(ctx, "id")
	groupID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || groupID <= 0 {
		return GetGroupDetailResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid group id"}
	}
	detail, err := h.Service.GetGroupDetail(ctx, ownerID, groupID)
	if err != nil {
		return GetGroupDetailResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Group not found"}
	}
	members := make([]GroupMemberPayload, 0, len(detail.Members))
	for _, m := range detail.Members {
		members = append(members, GroupMemberPayload{
			UserID: m.UserID, Email: m.Email, FullName: m.FullName,
			AddedAt: m.AddedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	payload := GroupDetailPayload{GroupPayload: toGroupPayload(detail.Group), Members: members}
	payload.MemberCount = len(members)
	return GetGroupDetailResponse{Group: payload}, nil
}

// ─── Update Group ─────────────────────────────────────────────────────────────

type (
	UpdateGroupHandler struct{ Service *application.Service }
	UpdateGroupRequest struct {
		Name                string                      `json:"name"`
		Description         string                      `json:"description"`
		Color               string                      `json:"color"`
		ResourcePermissions []ResourcePermissionPayload `json:"resource_permissions"`
	}
)

type UpdateGroupResponse struct {
	Group GroupPayload `json:"group"`
}

func (h *UpdateGroupHandler) Validate(_ context.Context, req *UpdateGroupRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateGroupHandler) Serve(ctx context.Context, req *UpdateGroupRequest) (UpdateGroupResponse, *errorModels.ServerError) {
	_, ownerID, ok := teamManagerFromCtx(ctx, h.Service)
	if !ok {
		return UpdateGroupResponse{}, errorModels.PermissionDeniedServerError()
	}
	idRaw := chi.URLParamFromCtx(ctx, "id")
	groupID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || groupID <= 0 {
		return UpdateGroupResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid group id"}
	}
	perms := make([]domain.ResourcePermission, 0, len(req.ResourcePermissions))
	for _, p := range req.ResourcePermissions {
		perms = append(perms, domain.ResourcePermission{ResourceKey: p.ResourceKey, PermissionLevel: domain.PermissionLevel(p.PermissionLevel)})
	}
	if err := h.Service.UpdateGroup(ctx, ownerID, groupID, domain.UpdateGroupInput{
		Name: req.Name, Description: req.Description, Color: req.Color, ResourcePermissions: perms,
	}); err != nil {
		return UpdateGroupResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update group"}
	}
	g, err := h.Service.GetGroup(ctx, ownerID, groupID)
	if err != nil {
		return UpdateGroupResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to fetch group"}
	}
	return UpdateGroupResponse{Group: toGroupPayload(*g)}, nil
}

// ─── Delete Group ─────────────────────────────────────────────────────────────

type (
	DeleteGroupHandler  struct{ Service *application.Service }
	DeleteGroupRequest  struct{}
	DeleteGroupResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteGroupHandler) Serve(ctx context.Context, _ *DeleteGroupRequest) (DeleteGroupResponse, *errorModels.ServerError) {
	_, ownerID, ok := teamManagerFromCtx(ctx, h.Service)
	if !ok {
		return DeleteGroupResponse{}, errorModels.PermissionDeniedServerError()
	}
	idRaw := chi.URLParamFromCtx(ctx, "id")
	groupID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || groupID <= 0 {
		return DeleteGroupResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid group id"}
	}
	if err := h.Service.DeleteGroup(ctx, ownerID, groupID); err != nil {
		return DeleteGroupResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete group"}
	}
	return DeleteGroupResponse{Message: "Group deleted"}, nil
}

// ─── Add Group Members ────────────────────────────────────────────────────────

func AddGroupMembersRoutes() []string { return []string{"/team/groups/{id}/members"} }

type (
	AddGroupMembersHandler struct{ Service *application.Service }
	AddGroupMembersRequest struct {
		UserIDs []int64 `json:"user_ids"`
	}
)

type AddGroupMembersResponse struct {
	Added   int    `json:"added"`
	Message string `json:"message"`
}

func (h *AddGroupMembersHandler) Validate(_ context.Context, req *AddGroupMembersRequest) *errorModels.ServerError {
	if len(req.UserIDs) == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "user_ids is required"}
	}
	return nil
}

func (h *AddGroupMembersHandler) Serve(ctx context.Context, req *AddGroupMembersRequest) (AddGroupMembersResponse, *errorModels.ServerError) {
	actorID, ownerID, ok := teamManagerFromCtx(ctx, h.Service)
	if !ok {
		return AddGroupMembersResponse{}, errorModels.PermissionDeniedServerError()
	}
	idRaw := chi.URLParamFromCtx(ctx, "id")
	groupID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || groupID <= 0 {
		return AddGroupMembersResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid group id"}
	}
	if err := h.Service.AddGroupMembers(ctx, ownerID, groupID, actorID, req.UserIDs); err != nil {
		return AddGroupMembersResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to add members"}
	}
	return AddGroupMembersResponse{Added: len(req.UserIDs), Message: "Members added"}, nil
}

// ─── Remove Group Member ──────────────────────────────────────────────────────

func RemoveGroupMemberRoutes() []string { return []string{"/team/groups/{id}/members/{userId}"} }

type (
	RemoveGroupMemberHandler  struct{ Service *application.Service }
	RemoveGroupMemberRequest  struct{}
	RemoveGroupMemberResponse struct {
		Message string `json:"message"`
	}
)

func (h *RemoveGroupMemberHandler) Serve(ctx context.Context, _ *RemoveGroupMemberRequest) (RemoveGroupMemberResponse, *errorModels.ServerError) {
	_, ownerID, ok := teamManagerFromCtx(ctx, h.Service)
	if !ok {
		return RemoveGroupMemberResponse{}, errorModels.PermissionDeniedServerError()
	}
	idRaw := chi.URLParamFromCtx(ctx, "id")
	groupID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || groupID <= 0 {
		return RemoveGroupMemberResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid group id"}
	}
	userIDRaw := chi.URLParamFromCtx(ctx, "userId")
	userID, err := strconv.ParseInt(userIDRaw, 10, 64)
	if err != nil || userID <= 0 {
		return RemoveGroupMemberResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid user id"}
	}
	if err := h.Service.RemoveGroupMember(ctx, ownerID, groupID, userID); err != nil {
		return RemoveGroupMemberResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to remove member"}
	}
	return RemoveGroupMemberResponse{Message: "Member removed"}, nil
}

// ─── Get User Groups ──────────────────────────────────────────────────────────

func GetUserGroupsRoutes() []string { return []string{"/team/users/{id}/groups"} }

type (
	GetUserGroupsHandler  struct{ Service *application.Service }
	GetUserGroupsRequest  struct{}
	GetUserGroupsResponse struct {
		Groups []GroupSummaryPayload `json:"groups"`
	}
)

func (h *GetUserGroupsHandler) Serve(ctx context.Context, _ *GetUserGroupsRequest) (GetUserGroupsResponse, *errorModels.ServerError) {
	ownerID, ok := teamOwnerIDFromCtx(ctx, h.Service)
	if !ok {
		return GetUserGroupsResponse{}, errorModels.PermissionDeniedServerError()
	}
	idRaw := chi.URLParamFromCtx(ctx, "id")
	userID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || userID <= 0 {
		return GetUserGroupsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid user id"}
	}
	groups, err := h.Service.GetUserGroups(ctx, ownerID, userID)
	if err != nil {
		return GetUserGroupsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get user groups"}
	}
	payloads := make([]GroupSummaryPayload, 0, len(groups))
	for _, g := range groups {
		payloads = append(payloads, GroupSummaryPayload{ID: g.ID, Name: g.Name, Color: g.Color})
	}
	return GetUserGroupsResponse{Groups: payloads}, nil
}

// ─── Bulk Invite ──────────────────────────────────────────────────────────────

func BulkInviteRoutes() []string { return []string{"/team/invite/bulk"} }

type BulkInviteHandler struct{ Service *application.Service }

type BulkInviteRequest struct {
	Emails  []string `json:"emails"`
	Role    string   `json:"role"`
	Apps    []string `json:"apps"`
	GroupID int64    `json:"group_id"`
}

type BulkInviteResultPayload struct {
	Email   string `json:"email"`
	UserID  int64  `json:"user_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type BulkInviteResponse struct {
	Results []BulkInviteResultPayload `json:"results"`
	Message string                    `json:"message"`
}

func (h *BulkInviteHandler) Validate(_ context.Context, req *BulkInviteRequest) *errorModels.ServerError {
	if len(req.Emails) == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "emails is required"}
	}
	return nil
}

func (h *BulkInviteHandler) Serve(ctx context.Context, req *BulkInviteRequest) (BulkInviteResponse, *errorModels.ServerError) {
	_, ownerID, ok := teamManagerFromCtx(ctx, h.Service)
	if !ok {
		return BulkInviteResponse{}, errorModels.PermissionDeniedServerError()
	}
	device := authPkg.DeviceCtx(ctx)
	businessID := int64(0)
	if device.User != nil {
		businessID = device.User.BusinessID
	}
	role := domain.Role(req.Role)
	if role == "" {
		role = domain.RoleStaff
	}
	apps := req.Apps
	if apps == nil {
		apps = []string{}
	}
	results, err := h.Service.BulkInvite(ctx, ownerID, businessID, domain.BulkInviteInput{
		Emails:  req.Emails,
		Role:    role,
		Apps:    apps,
		GroupID: req.GroupID,
	})
	if err != nil {
		return BulkInviteResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Bulk invite failed"}
	}
	payloads := make([]BulkInviteResultPayload, 0, len(results))
	for _, r := range results {
		payloads = append(payloads, BulkInviteResultPayload{Email: r.Email, UserID: r.UserID, Status: r.Status, Message: r.Message})
	}
	return BulkInviteResponse{Results: payloads, Message: "Bulk invite complete"}, nil
}
