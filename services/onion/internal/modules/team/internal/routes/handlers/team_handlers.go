package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Staff Users ─────────────────────────────────────────────────────────

func ListUsersRoutes() []string { return []string{"/team/users"} }

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

// ─── Invite User ──────────────────────────────────────────────────────────────

func InviteUserRoutes() []string { return []string{"/team/users/invite"} }

type InviteUserHandler struct{ Service *application.Service }

type InviteUserRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	FullName string   `json:"fullName"`
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

	if req.Password == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Password is required",
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

	userID, err := h.Service.InviteUser(ctx, ownerID, domain.InviteUserInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
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

	return InviteUserResponse{UserID: userID, Message: "User invited"}, nil
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
	ID       int64    `json:"id"`
	Email    string   `json:"email"`
	FullName string   `json:"fullName"`
	Apps     []string `json:"apps"`
	Role     string   `json:"role"`
}

func (h *MeHandler) Serve(ctx context.Context, req *MeRequest) (MeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MeResponse{}, errorModels.PermissionDeniedServerError()
	}

	userID := device.User.UserID

	ownerID := req.OwnerID
	if ownerID == 0 {
		ownerID = userID
	}

	email, fullName, apps, role, err := h.Service.GetMe(ctx, ownerID, userID)
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

	return MeResponse{
		ID:       userID,
		Email:    email,
		FullName: fullName,
		Apps:     apps,
		Role:     string(role),
	}, nil
}
