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

// VerifyInviteResponse reports only whether a token can still be accepted. The
// employee and business a token points at are tenant join keys, and a token is
// readable by everyone it was ever shown to, so answering with those ids would
// hand that caller another tenant's internal records.
type VerifyInviteResponse struct {
	Valid bool `json:"valid"`
}

func (h *VerifyInviteHandler) Serve(ctx context.Context, req *VerifyInviteRequest) (VerifyInviteResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return VerifyInviteResponse{}, onekiterrors.PermissionDenied()
	}

	if _, err := h.Service.VerifyInviteToken(ctx, req.Token); err != nil {
		if errors.Is(err, domain.ErrInviteTokenInvalid) || errors.Is(err, domain.ErrInviteTokenExpired) {
			return VerifyInviteResponse{Valid: false}, nil
		}
		return VerifyInviteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to verify invite")
	}

	return VerifyInviteResponse{Valid: true}, nil
}

// --- Accept Invite ---

func AcceptInviteRoutes() []string { return []string{"/hr/employees/invite/accept"} }

type AcceptInviteHandler struct{ Service *application.Service }

// AcceptInviteRequest carries the token only. The account that becomes the
// employee is the authenticated caller's, so no request field can select it.
type AcceptInviteRequest struct {
	Token string `json:"token"`
}

type AcceptInviteResponse struct {
	Success bool `json:"success"`
}

func (h *AcceptInviteHandler) Validate(_ context.Context, req *AcceptInviteRequest) error {
	if req.Token == "" {
		return onekiterrors.New(http.StatusBadRequest, "token is required")
	}
	return nil
}

func (h *AcceptInviteHandler) Serve(ctx context.Context, req *AcceptInviteRequest) (AcceptInviteResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return AcceptInviteResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.AcceptInviteForCaller(ctx, req.Token, device.User.UserID); err != nil {
		if errors.Is(err, domain.ErrInviteTokenInvalid) {
			return AcceptInviteResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid invite token")
		}
		if errors.Is(err, domain.ErrInviteTokenExpired) {
			return AcceptInviteResponse{}, onekiterrors.New(http.StatusGone, "Invite token has expired")
		}
		if errors.Is(err, domain.ErrInviteRecipientMismatch) {
			return AcceptInviteResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrUserAlreadyLinked) || errors.Is(err, domain.ErrEmployeeNotFound) {
			return AcceptInviteResponse{}, onekiterrors.New(http.StatusConflict, "Invite cannot be accepted")
		}
		// The token is a credential, so it is not part of the log line.
		logger.Ctx(ctx).Error("failed to accept invite", logger.Error(err))
		return AcceptInviteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to accept invite")
	}

	return AcceptInviteResponse{Success: true}, nil
}
