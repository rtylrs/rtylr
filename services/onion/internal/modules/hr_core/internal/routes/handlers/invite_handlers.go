package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

// --- Invite Employee ---

func InviteEmployeeRoutes() []string { return []string{"/hr/employees/invite"} }

type InviteEmployeeHandler struct{ Service *application.Service }

type InviteEmployeeRequest struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	EmploymentType string `json:"employment_type"`
	DepartmentID   int64  `json:"department_id"`
	LocationID     int64  `json:"location_id"`
}

type InviteEmployeeResponse struct {
	Token string `json:"token"`
}

func (h *InviteEmployeeHandler) Validate(_ context.Context, req *InviteEmployeeRequest) *errorModels.ServerError {
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	if req.FirstName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "first_name is required"}
	}
	if req.LastName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "last_name is required"}
	}
	return nil
}

func (h *InviteEmployeeHandler) Serve(ctx context.Context, req *InviteEmployeeRequest) (InviteEmployeeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return InviteEmployeeResponse{}, errorModels.PermissionDeniedServerError()
	}

	emp := domain.EmployeeRecord{
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		Phone:          req.Phone,
		EmploymentType: req.EmploymentType,
		DepartmentID:   req.DepartmentID,
		LocationID:     req.LocationID,
	}

	token, err := h.Service.InviteEmployee(ctx, device.User.UserID, emp)
	if err != nil {
		logger.Ctx(ctx).Error("failed to invite employee", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return InviteEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to invite employee"}
	}

	return InviteEmployeeResponse{Token: token}, nil
}

// --- Verify Invite ---

func VerifyInviteRoutes() []string { return []string{"/hr/employees/invite/verify"} }

type VerifyInviteHandler struct{ Service *application.Service }

type VerifyInviteRequest struct {
	Token string `json:"token"`
}

type VerifyInviteResponse struct {
	Valid      bool  `json:"valid"`
	EmployeeID int64 `json:"employee_id"`
	OwnerID    int64 `json:"owner_id"`
}

func (h *VerifyInviteHandler) Serve(ctx context.Context, req *VerifyInviteRequest) (VerifyInviteResponse, *errorModels.ServerError) {
	invite, err := h.Service.VerifyInviteToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, domain.ErrInviteTokenInvalid) || errors.Is(err, domain.ErrInviteTokenExpired) {
			return VerifyInviteResponse{Valid: false}, nil
		}
		return VerifyInviteResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to verify invite"}
	}

	return VerifyInviteResponse{Valid: true, EmployeeID: invite.EmployeeID, OwnerID: invite.OwnerID}, nil
}

// --- Accept Invite ---

func AcceptInviteRoutes() []string { return []string{"/hr/employees/invite/accept"} }

type AcceptInviteHandler struct{ Service *application.Service }

type AcceptInviteRequest struct {
	Token  string `json:"token"`
	UserID int64  `json:"user_id"`
}

type AcceptInviteResponse struct {
	Success bool `json:"success"`
}

func (h *AcceptInviteHandler) Validate(_ context.Context, req *AcceptInviteRequest) *errorModels.ServerError {
	if req.Token == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "token is required"}
	}
	return nil
}

func (h *AcceptInviteHandler) Serve(ctx context.Context, req *AcceptInviteRequest) (AcceptInviteResponse, *errorModels.ServerError) {
	err := h.Service.AcceptInvite(ctx, req.Token, req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrInviteTokenInvalid) {
			return AcceptInviteResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid invite token"}
		}
		if errors.Is(err, domain.ErrInviteTokenExpired) {
			return AcceptInviteResponse{}, &errorModels.ServerError{HTTPCode: http.StatusGone, Code: http.StatusGone, Message: "Invite token has expired"}
		}
		logger.Ctx(ctx).Error("failed to accept invite", logger.String("token", req.Token), logger.Error(err))
		return AcceptInviteResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to accept invite"}
	}

	return AcceptInviteResponse{Success: true}, nil
}
