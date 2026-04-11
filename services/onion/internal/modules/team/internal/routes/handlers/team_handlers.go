package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Staff Users ─────────────────────────────────────────────────────────

func ListUsersRoutes() []string { return []string{"/team/users", "/team/members"} }

type ListUsersHandler struct{ Service *application.Service }

type ListUsersRequest struct{}

type StaffUserPayload struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
	Apps      []string  `json:"apps"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
}

type ListUsersResponse struct {
	Users []StaffUserPayload `json:"users"`
}

func newStaffUserPayload(u domain.StaffUser) StaffUserPayload {
	apps := u.Apps
	if apps == nil {
		apps = []string{}
	}

	return StaffUserPayload{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: u.CreatedAt,
		Apps:      apps,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
	}
}

func (h *ListUsersHandler) Serve(ctx context.Context, _ *ListUsersRequest) (ListUsersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListUsersResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	users, err := h.Service.ListStaffUsers(ctx, ownerID)
	if err != nil {
		return ListUsersResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list users",
		}
	}

	payloads := make([]StaffUserPayload, 0, len(users))
	for _, u := range users {
		payloads = append(payloads, newStaffUserPayload(u))
	}

	return ListUsersResponse{Users: payloads}, nil
}

func GetMemberByIDRoutes() []string { return []string{"/team/members/{id}"} }

type GetMemberByIDHandler struct{ Service *application.Service }

type GetMemberByIDRequest struct{}

type GetMemberByIDResponse struct {
	User StaffUserPayload `json:"user"`
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

	ownerID := device.User.UserID
	businessID := device.User.BusinessID

	member, err := h.Service.GetStaffUserByID(ctx, ownerID, businessID, memberID)
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

	return GetMemberByIDResponse{User: newStaffUserPayload(*member)}, nil
}

// ─── Invite User ──────────────────────────────────────────────────────────────

func InviteUserRoutes() []string { return []string{"/team/users/invite", "/team/invite"} }

type InviteUserHandler struct{ Service *application.Service }

type InviteUserRequest struct {
	Email    string   `json:"email"`
	FullName string   `json:"fullName"`
	Password string   `json:"password"`
	Apps     []string `json:"apps"`
	Role     string   `json:"role"`
}

type InviteUserResponse struct {
	UserID  int64  `json:"userId"`
	Message string `json:"message"`
}

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

	ownerID := device.User.UserID

	apps := req.Apps
	if apps == nil {
		apps = []string{}
	}

	role := domain.Role(req.Role)
	if role == "" {
		role = domain.RoleStaff
	}

	userID, token, err := h.Service.InviteUser(ctx, ownerID, domain.InviteUserInput{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		Apps:     apps,
		Role:     role,
	})
	if err != nil {
		return InviteUserResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to invite user",
		}
	}

	// Only send invite email if no password was set manually
	if req.Password == "" && token != "" {
		go func() {
			inviteLink := "https://auth.rtylr.com/invite/" + token
			emailPkg.SendInvitation(ctx, req.Email, "Your team", "rtylr", inviteLink)
		}()
		return InviteUserResponse{UserID: userID, Message: "Invitation sent"}, nil
	}

	return InviteUserResponse{UserID: userID, Message: "Member created"}, nil
}

// ─── Verify Invite Token ────────────────────────────────────────────────────

func VerifyInviteRoutes() []string { return []string{"/auth/invite/verify"} }

type VerifyInviteHandler struct{ Service *application.Service }

type VerifyInviteRequest struct {
	Token string `json:"token"`
}

type VerifyInviteResponse struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Valid    bool   `json:"valid"`
}

func (h *VerifyInviteHandler) Serve(_ context.Context, req *VerifyInviteRequest) (VerifyInviteResponse, *errorModels.ServerError) {
	_, email, fullName, err := h.Service.VerifyInviteToken(context.Background(), req.Token)
	if err != nil {
		return VerifyInviteResponse{Valid: false}, nil
	}

	return VerifyInviteResponse{Email: email, FullName: fullName, Valid: true}, nil
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

func (h *AcceptInviteHandler) Serve(_ context.Context, req *AcceptInviteRequest) (AcceptInviteResponse, *errorModels.ServerError) {
	_, err := h.Service.AcceptInvite(context.Background(), req.Token, req.Password)
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
	UserID int64    `json:"userId"`
	Apps   []string `json:"apps"`
	Role   string   `json:"role"`
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

	ownerID := device.User.UserID

	apps := req.Apps
	if apps == nil {
		apps = []string{}
	}

	if err := h.Service.UpdatePermissions(ctx, ownerID, domain.UpdatePermissionsInput{
		UserID: req.UserID,
		Apps:   apps,
		Role:   domain.Role(req.Role),
	}); err != nil {
		return UpdatePermissionsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update permissions",
		}
	}

	return UpdatePermissionsResponse{Message: "Permissions updated"}, nil
}

// ─── Deactivate User ──────────────────────────────────────────────────────────

type DeactivateUserHandler struct{ Service *application.Service }

type DeactivateUserRequest struct {
	UserID int64 `json:"userId"`
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

	ownerID := device.User.UserID

	if err := h.Service.DeactivateUser(ctx, ownerID, req.UserID); err != nil {
		return DeactivateUserResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to deactivate user",
		}
	}

	return DeactivateUserResponse{Message: "User deactivated"}, nil
}

// ─── Get Me ───────────────────────────────────────────────────────────────────

func MeRoutes() []string { return []string{"/user/me"} }

type MeHandler struct{ Service *application.Service }

type MeRequest struct {
	OwnerID int64 `json:"ownerId"`
}

type MeResponse struct {
	ID           int64    `json:"id"`
	Email        string   `json:"email"`
	FullName     string   `json:"fullName"`
	Apps         []string `json:"apps"`
	Role         string   `json:"role"`
	BusinessID   int64    `json:"businessId"`
	BusinessName string   `json:"businessName"`
	IsOwner      bool     `json:"isOwner"`
	Onboarded    bool     `json:"onboarded"`
}

func (h *MeHandler) Serve(ctx context.Context, req *MeRequest) (MeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MeResponse{}, errorModels.PermissionDeniedServerError()
	}

	userID := device.User.UserID
	businessID := device.User.BusinessID

	email, fullName, apps, role, err := h.Service.GetMeByBusiness(ctx, businessID, userID)
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

	if apps == nil {
		apps = []string{}
	}

	// Get business name
	businessName, isOwner := h.Service.GetBusinessInfo(ctx, businessID, userID)

	// If user is the owner, they have access to everything
	if isOwner {
		role = "owner"
		apps = []string{"pos", "erp", "hr", "crm", "finance", "flow"}
	}

	return MeResponse{
		ID:           userID,
		Email:        email,
		FullName:     fullName,
		Apps:         apps,
		Role:         string(role),
		BusinessID:   businessID,
		BusinessName: businessName,
		IsOwner:      isOwner,
		Onboarded:    h.Service.IsOnboarded(ctx, userID),
	}, nil
}

// ─── Mark Onboarded ─────────────────────────────────────────────────────────

func MarkOnboardedRoutes() []string { return []string{"/user/onboarded"} }

type MarkOnboardedHandler struct{ Service *application.Service }
type MarkOnboardedRequest struct{}
type MarkOnboardedResponse struct {
	Message string `json:"message"`
}

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

	// Get requesting user's name for the email
	requesterEmail, requesterName, _ := h.Service.GetUserInfo(ctx, device.User.UserID)
	if requesterName == "" {
		requesterName = requesterEmail
	}

	// Send email to business owner (non-blocking)
	go func() {
		ownerEmail := h.Service.GetBusinessOwnerEmail(context.Background(), device.User.BusinessID)
		if ownerEmail != "" {
			appUpper := strings.ToUpper(req.App)
			emailPkg.Send(context.Background(), emailPkg.SendParams{
				To:      []string{ownerEmail},
				Subject: requesterName + " requested access to " + appUpper,
				HTML: `<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:480px;margin:0 auto;padding:40px 20px;">
					<h2 style="font-size:20px;font-weight:600;color:#111;margin:0 0 8px 0;">Access request</h2>
					<p style="font-size:14px;color:#666;line-height:1.6;margin:0 0 24px 0;">
						<strong>` + requesterName + `</strong> (` + requesterEmail + `) is requesting access to <strong>` + appUpper + `</strong>.
					</p>
					<a href="https://hub.rtylr.com/team" style="display:inline-block;background:#4583C4;color:#fff;padding:12px 24px;border-radius:8px;text-decoration:none;font-size:14px;font-weight:600;">
						Go to Team
					</a>
				</div>`,
			})
		}
	}()

	return RequestAccessResponse{Message: "Access requested"}, nil
}
