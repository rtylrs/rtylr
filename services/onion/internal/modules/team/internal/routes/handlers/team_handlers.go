package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/config"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
)

// ─── List Staff Users ─────────────────────────────────────────────────────────

func ListUsersRoutes() []string { return []string{"/team/users", "/team/members"} }

type ListUsersHandler struct{ Service *application.Service }

type ListUsersRequest struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

type StaffUserWithGroupsPayload struct {
	*hubteammembersv1.StaffUserPayload
	Groups []GroupSummaryPayload `json:"groups"`
}

type ListUsersResponse struct {
	Users    []StaffUserWithGroupsPayload `json:"users"`
	Page     int64                        `json:"page"`
	PageSize int64                        `json:"page_size"`
	HasMore  bool                         `json:"has_more"`
}

func newProtoStaffUser(u domain.StaffUser) *hubteammembersv1.StaffUserPayload {
	apps := u.Apps
	if apps == nil {
		apps = []string{}
	}

	return &hubteammembersv1.StaffUserPayload{
		Id:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Apps:      apps,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
	}
}

func (h *ListUsersHandler) Serve(ctx context.Context, req *ListUsersRequest) (ListUsersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListUsersResponse{}, errorModels.PermissionDeniedServerError()
	}

	users, hasMore, err := h.Service.ListStaffUsersByBusiness(ctx, device.User.BusinessID, req.Page, req.PageSize)
	if err != nil {
		return ListUsersResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list users",
		}
	}
	payloads := make([]StaffUserWithGroupsPayload, 0, len(users))
	for _, u := range users {
		groups := make([]GroupSummaryPayload, 0, len(u.Groups))
		for _, g := range u.Groups {
			groups = append(groups, GroupSummaryPayload{ID: g.ID, Name: g.Name, Color: g.Color})
		}
		payloads = append(payloads, StaffUserWithGroupsPayload{
			StaffUserPayload: newProtoStaffUser(u),
			Groups:           groups,
		})
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListUsersResponse{Users: payloads, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}

func GetMemberByIDRoutes() []string { return []string{"/team/members/{id}"} }

type GetMemberByIDHandler struct{ Service *application.Service }

type GetMemberByIDRequest struct{}

type GetMemberByIDResponse struct {
	User StaffUserWithGroupsPayload `json:"user"`
}

func (h *GetMemberByIDHandler) Serve(ctx context.Context, _ *GetMemberByIDRequest) (GetMemberByIDResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetMemberByIDResponse{}, errorModels.PermissionDeniedServerError()
	}

	memberIDRaw := chi.URLParamFromCtx(ctx, "id")
	memberID, err := strconv.ParseInt(memberIDRaw, 10, 64)
	if err != nil || memberID <= 0 {
		return GetMemberByIDResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid member id",
		}
	}

	actorID := device.User.UserID
	businessID := device.User.BusinessID

	member, err := h.Service.GetStaffUserByID(ctx, actorID, businessID, memberID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return GetMemberByIDResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Member not found",
			}
		}

		return GetMemberByIDResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get member",
		}
	}

	memberGroups := make([]GroupSummaryPayload, 0, len(member.Groups))
	for _, g := range member.Groups {
		memberGroups = append(memberGroups, GroupSummaryPayload{ID: g.ID, Name: g.Name, Color: g.Color})
	}
	return GetMemberByIDResponse{User: StaffUserWithGroupsPayload{
		StaffUserPayload: newProtoStaffUser(*member),
		Groups:           memberGroups,
	}}, nil
}

// ─── Invite User ──────────────────────────────────────────────────────────────

func InviteUserRoutes() []string { return []string{"/team/users/invite", "/team/invite"} }

type InviteUserHandler struct{ Service *application.Service }

type InviteUserRequest struct {
	Email    string   `json:"email"`
	FullName string   `json:"full_name"`
	Password string   `json:"password"`
	Apps     []string `json:"apps"`
	Role     string   `json:"role"`
	GroupID  int64    `json:"group_id"`
}

type InviteUserResponse = hubteaminvitationsv1.InviteUserResponse

func (h *InviteUserHandler) Validate(_ context.Context, req *InviteUserRequest) *errorModels.ServerError {
	if req.Email == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Email is required",
		}
	}

	return nil
}

func (h *InviteUserHandler) Serve(ctx context.Context, req *InviteUserRequest) (InviteUserResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return InviteUserResponse{}, errorModels.PermissionDeniedServerError()
	}

	actorID := device.User.UserID

	apps := req.Apps
	if apps == nil {
		apps = []string{}
	}

	role := domain.Role(req.Role)
	if role == "" {
		role = domain.RoleStaff
	}

	userID, token, err := h.Service.InviteUser(ctx, actorID, device.User.BusinessID, domain.InviteUserInput{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		Apps:     apps,
		Role:     role,
	})
	if err != nil {
		return InviteUserResponse{}, errorModels.FromError(err, "Failed to invite user")
	}

	// Record the seat change asynchronously so billing stays accurate.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("goroutine panicked", logger.Any("panic", r))
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		ownerID, err := h.Service.BusinessOwnerID(bgCtx, actorID, device.User.BusinessID)
		if err != nil {
			logger.Ctx(bgCtx).Error("failed to resolve owner for billing seat change",
				logger.Any("actor_id", actorID),
				logger.Error(err),
			)
			return
		}
		if err := h.Service.RecordSeatChange(bgCtx, ownerID); err != nil {
			logger.Ctx(bgCtx).Error("failed to record seat change for billing",
				logger.Any("owner_id", ownerID),
				logger.Error(err),
			)
		}
	}()

	if req.GroupID > 0 {
		ownerID, ownerErr := h.Service.BusinessOwnerID(ctx, actorID, device.User.BusinessID)
		if ownerErr != nil {
			return InviteUserResponse{}, errorModels.FromError(ownerErr, "Failed to invite user")
		}
		if grpErr := h.Service.AddGroupMembers(ctx, ownerID, req.GroupID, ownerID, []int64{userID}); grpErr != nil {
			logger.Ctx(ctx).Warn("failed to add invited user to group",
				logger.Int("group_id", int(req.GroupID)),
				logger.Int("user_id", int(userID)),
				logger.Error(grpErr),
			)
		}
	}

	// Only send invite email if no password was set manually. Enqueue is a fast
	// insert handled by the background worker, so no goroutine is needed.
	if req.Password == "" && token != "" {
		inviteLink := config.AuthAppURL + "/invite/" + token
		if emailErr := h.Service.EnqueueInviteEmail(ctx, req.Email, inviteLink); emailErr != nil {
			logger.Ctx(ctx).Warn("failed to enqueue invite email", logger.Error(emailErr))
		}
		return InviteUserResponse{UserId: userID, Message: "Invitation sent"}, nil
	}

	return InviteUserResponse{UserId: userID, Message: "Member created"}, nil
}

// ─── Verify Invite Token ────────────────────────────────────────────────────

func VerifyInviteRoutes() []string { return []string{"/auth/invite/verify"} }

type VerifyInviteHandler struct{ Service *application.Service }

type VerifyInviteRequest struct {
	Token string `json:"token"`
}

type VerifyInviteResponse struct {
	Email    string   `json:"email"`
	FullName string   `json:"full_name"`
	Apps     []string `json:"apps"`
	Valid    bool     `json:"valid"`
}

func (h *VerifyInviteHandler) Serve(ctx context.Context, req *VerifyInviteRequest) (VerifyInviteResponse, *errorModels.ServerError) {
	_, email, fullName, apps, err := h.Service.VerifyInviteToken(ctx, req.Token)
	if err != nil {
		return VerifyInviteResponse{Valid: false}, nil
	}

	return VerifyInviteResponse{Email: email, FullName: fullName, Apps: apps, Valid: true}, nil
}

// ─── Accept Invite ──────────────────────────────────────────────────────────

func AcceptInviteRoutes() []string { return []string{"/auth/invite/accept"} }

type AcceptInviteHandler struct{ Service *application.Service }

type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type AcceptInviteResponse struct {
	Message string `json:"message"`
}

func (h *AcceptInviteHandler) Validate(_ context.Context, req *AcceptInviteRequest) *errorModels.ServerError {
	if req.Token == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Token is required"}
	}
	if len(req.Password) < 6 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Password must be at least 6 characters"}
	}
	return nil
}

func (h *AcceptInviteHandler) Serve(ctx context.Context, req *AcceptInviteRequest) (AcceptInviteResponse, *errorModels.ServerError) {
	_, err := h.Service.AcceptInvite(ctx, req.Token, req.Password)
	if err != nil {
		msg := "Invalid or expired invitation"
		if errors.Is(err, domain.ErrTokenAlreadyUsed) {
			msg = "This invitation has already been accepted"
		} else if errors.Is(err, domain.ErrTokenExpired) {
			msg = "This invitation has expired"
		}
		return AcceptInviteResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  msg,
		}
	}

	return AcceptInviteResponse{Message: "Password set successfully. You can now sign in."}, nil
}

// ─── Update Permissions ───────────────────────────────────────────────────────

func UpdatePermissionsRoutes() []string { return []string{"/team/users/permissions"} }

type UpdatePermissionsHandler struct{ Service *application.Service }

type UpdatePermissionsRequest struct {
	UserID   int64     `json:"user_id"`
	FullName *string   `json:"full_name"`
	Apps     *[]string `json:"apps"`
	Role     string    `json:"role"`
}

type UpdatePermissionsResponse struct {
	Message string `json:"message"`
}

func (h *UpdatePermissionsHandler) Validate(_ context.Context, req *UpdatePermissionsRequest) *errorModels.ServerError {
	if req.UserID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "User id is required",
		}
	}

	return nil
}

func (h *UpdatePermissionsHandler) Serve(ctx context.Context, req *UpdatePermissionsRequest) (UpdatePermissionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdatePermissionsResponse{}, errorModels.PermissionDeniedServerError()
	}

	actorID := device.User.UserID

	if err := h.Service.UpdatePermissions(ctx, actorID, device.User.BusinessID, domain.UpdatePermissionsInput{
		UserID:   req.UserID,
		FullName: req.FullName,
		Apps:     req.Apps,
		Role:     domain.Role(req.Role),
	}); err != nil {
		return UpdatePermissionsResponse{}, errorModels.FromError(err, "Failed to update permissions")
	}

	return UpdatePermissionsResponse{Message: "Permissions updated"}, nil
}

// ─── Deactivate User ──────────────────────────────────────────────────────────

type DeactivateUserHandler struct{ Service *application.Service }

type DeactivateUserRequest struct {
	UserID int64 `json:"user_id"`
}

type DeactivateUserResponse struct {
	Message string `json:"message"`
}

func (h *DeactivateUserHandler) Validate(_ context.Context, req *DeactivateUserRequest) *errorModels.ServerError {
	if req.UserID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "User id is required",
		}
	}

	return nil
}

func (h *DeactivateUserHandler) Serve(ctx context.Context, req *DeactivateUserRequest) (DeactivateUserResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeactivateUserResponse{}, errorModels.PermissionDeniedServerError()
	}

	actorID := device.User.UserID

	if err := h.Service.DeactivateUser(ctx, actorID, device.User.BusinessID, req.UserID); err != nil {
		return DeactivateUserResponse{}, errorModels.FromError(err, "Failed to deactivate user")
	}

	// Record the seat change asynchronously so billing stays accurate.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("goroutine panicked", logger.Any("panic", r))
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		ownerID, err := h.Service.BusinessOwnerID(bgCtx, actorID, device.User.BusinessID)
		if err != nil {
			logger.Ctx(bgCtx).Error("failed to resolve owner for billing seat change",
				logger.Any("actor_id", actorID),
				logger.Error(err),
			)
			return
		}
		if err := h.Service.RecordSeatChange(bgCtx, ownerID); err != nil {
			logger.Ctx(bgCtx).Error("failed to record seat change for billing",
				logger.Any("owner_id", ownerID),
				logger.Error(err),
			)
		}
	}()

	return DeactivateUserResponse{Message: "User deactivated"}, nil
}

// ─── Deactivate Member by Path ID ─────────────────────────────────────────────
// POST /team/members/{id}/deactivate — convenience endpoint used by per-app UIs

func DeactivateMemberRoutes() []string { return []string{"/team/members/{id}/deactivate"} }

type DeactivateMemberHandler struct{ Service *application.Service }

type DeactivateMemberRequest struct{}

type DeactivateMemberResponse struct {
	Message string `json:"message"`
}

func (h *DeactivateMemberHandler) Serve(ctx context.Context, _ *DeactivateMemberRequest) (DeactivateMemberResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeactivateMemberResponse{}, errorModels.PermissionDeniedServerError()
	}

	memberIDRaw := chi.URLParamFromCtx(ctx, "id")
	memberID, err := strconv.ParseInt(memberIDRaw, 10, 64)
	if err != nil || memberID <= 0 {
		return DeactivateMemberResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid member id",
		}
	}

	actorID := device.User.UserID

	if err := h.Service.DeactivateUser(ctx, actorID, device.User.BusinessID, memberID); err != nil {
		return DeactivateMemberResponse{}, errorModels.FromError(err, "Failed to deactivate member")
	}

	// Record the seat change asynchronously so billing stays accurate.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("goroutine panicked", logger.Any("panic", r))
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		ownerID, err := h.Service.BusinessOwnerID(bgCtx, actorID, device.User.BusinessID)
		if err != nil {
			logger.Ctx(bgCtx).Error("failed to resolve owner for billing seat change",
				logger.Any("actor_id", actorID),
				logger.Error(err),
			)
			return
		}
		if err := h.Service.RecordSeatChange(bgCtx, ownerID); err != nil {
			logger.Ctx(bgCtx).Error("failed to record seat change for billing",
				logger.Any("owner_id", ownerID),
				logger.Error(err),
			)
		}
	}()

	return DeactivateMemberResponse{Message: "Member deactivated"}, nil
}

// ─── Get Me ───────────────────────────────────────────────────────────────────

func MeRoutes() []string { return []string{"/user/me"} }

type MeHandler struct{ Service *application.Service }

type MeRequest struct {
	OwnerID int64 `json:"owner_id"`
}

type MeResponse struct {
	ID                 int64             `json:"id"`
	Email              string            `json:"email"`
	FullName           string            `json:"full_name"`
	Apps               []string          `json:"apps"`
	Role               string            `json:"role"`
	BusinessID         int64             `json:"business_id"`
	BusinessName       string            `json:"business_name"`
	IsOwner            bool              `json:"is_owner"`
	IsConsoleAdmin     bool              `json:"is_console_admin"`
	Onboarded          bool              `json:"onboarded"`
	PlanSlug           string            `json:"plan_slug"`
	SubscriptionStatus string            `json:"subscription_status"`
	EffectiveResources map[string]string `json:"effective_resources"`
}

func (h *MeHandler) Serve(ctx context.Context, _ *MeRequest) (MeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MeResponse{}, errorModels.PermissionDeniedServerError()
	}

	profile, err := h.Service.GetMyProfile(ctx, device.User.UserID, device.User.BusinessID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return MeResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "User not found",
			}
		}
		return MeResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get user info",
		}
	}

	return MeResponse(profile), nil
}

// ─── Mark Onboarded ─────────────────────────────────────────────────────────

func MarkOnboardedRoutes() []string { return []string{"/user/onboarded"} }

type (
	MarkOnboardedHandler  struct{ Service *application.Service }
	MarkOnboardedRequest  struct{}
	MarkOnboardedResponse struct {
		Message string `json:"message"`
	}
)

func (h *MarkOnboardedHandler) Serve(ctx context.Context, _ *MarkOnboardedRequest) (MarkOnboardedResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MarkOnboardedResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.MarkOnboarded(ctx, device.User.UserID); err != nil {
		return MarkOnboardedResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update",
		}
	}

	return MarkOnboardedResponse{Message: "Onboarding complete"}, nil
}

// ─── Request Access ─────────────────────────────────────────────────────────

func RequestAccessRoutes() []string { return []string{"/team/request-access"} }

type RequestAccessHandler struct{ Service *application.Service }

type RequestAccessRequest struct {
	App string `json:"app"`
}

type RequestAccessResponse struct {
	Message string `json:"message"`
}

func (h *RequestAccessHandler) Validate(_ context.Context, req *RequestAccessRequest) *errorModels.ServerError {
	if req.App == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "App is required"}
	}
	return nil
}

func (h *RequestAccessHandler) Serve(ctx context.Context, req *RequestAccessRequest) (RequestAccessResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RequestAccessResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.RequestAccess(ctx, device.User.BusinessID, device.User.UserID, req.App)
	if err != nil {
		return RequestAccessResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to request access",
		}
	}

	// Notify the business owner. Enqueue is a fast insert; the worker delivers.
	requesterEmail, requesterName, _ := h.Service.GetUserInfo(ctx, device.User.UserID)
	if requesterName == "" {
		requesterName = requesterEmail
	}
	if emailErr := h.Service.EnqueueAccessRequestEmail(ctx, device.User.BusinessID, requesterName, requesterEmail, req.App); emailErr != nil {
		logger.Ctx(ctx).Warn("failed to enqueue access request email", logger.Error(emailErr))
	}

	return RequestAccessResponse{Message: "Access requested"}, nil
}
