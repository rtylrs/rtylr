package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func (h *InviteEmployeeHandler) Validate(_ context.Context, req *InviteEmployeeRequest) error {
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	if req.Email == "" {
		return onekiterrors.New(http.StatusBadRequest, "email is required")
	}
	if req.FirstName == "" {
		return onekiterrors.New(http.StatusBadRequest, "first_name is required")
	}
	if req.LastName == "" {
		return onekiterrors.New(http.StatusBadRequest, "last_name is required")
	}
	return nil
}

func (h *InviteEmployeeHandler) Serve(ctx context.Context, req *InviteEmployeeRequest) (InviteEmployeeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return InviteEmployeeResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return InviteEmployeeResponse{}, onekiterrors.PermissionDenied()
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

	token, err := h.Service.InviteEmployeeForWriter(ctx, businessID, device.User.UserID, emp)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return InviteEmployeeResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to invite employee", logger.Any("business_id", businessID), logger.Error(err))
		return InviteEmployeeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to invite employee")
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

func (h *VerifyInviteHandler) Serve(ctx context.Context, req *VerifyInviteRequest) (VerifyInviteResponse, error) {
	invite, err := h.Service.VerifyInviteToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, domain.ErrInviteTokenInvalid) || errors.Is(err, domain.ErrInviteTokenExpired) {
			return VerifyInviteResponse{Valid: false}, nil
		}
		return VerifyInviteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to verify invite")
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

func (h *AcceptInviteHandler) Validate(_ context.Context, req *AcceptInviteRequest) error {
	if req.Token == "" {
		return onekiterrors.New(http.StatusBadRequest, "token is required")
	}
	if req.UserID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "user_id is required")
	}
	return nil
}

func (h *AcceptInviteHandler) Serve(ctx context.Context, req *AcceptInviteRequest) (AcceptInviteResponse, error) {
	err := h.Service.AcceptInvite(ctx, req.Token, req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrInviteTokenInvalid) {
			return AcceptInviteResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid invite token")
		}
		if errors.Is(err, domain.ErrInviteTokenExpired) {
			return AcceptInviteResponse{}, onekiterrors.New(http.StatusGone, "Invite token has expired")
		}
		logger.Ctx(ctx).Error("failed to accept invite", logger.String("token", req.Token), logger.Error(err))
		return AcceptInviteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to accept invite")
	}

	return AcceptInviteResponse{Success: true}, nil
}
