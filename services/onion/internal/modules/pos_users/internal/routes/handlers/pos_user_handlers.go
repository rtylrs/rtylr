package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_users/internal/application"
	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	posuserv1 "github.com/voxire/voxie/gen/go/pos/user/v1"
)

// toProtoPOSUser intentionally does NOT carry the PIN or any hash of it
// into the proto payload. Every caller of this function sits behind an
// authenticated, owner-scoped handler and gets only the operational
// attributes needed to drive the POS UI.
func toProtoPOSUser(u domain.POSUser) *posuserv1.POSUser {
	return &posuserv1.POSUser{
		Id:               u.ID,
		BusinessId:       u.BusinessID,
		LocationId:       u.LocationID,
		Name:             u.Name,
		Role:             string(u.Role),
		IsActive:         u.IsActive,
		MaxDiscountPct:   u.MaxDiscountPct,
		CanVoid:          u.CanVoid,
		CanRefund:        u.CanRefund,
		CanOpenDrawer:    u.CanOpenDrawer,
		CanApplyDiscount: u.CanApplyDiscount,
	}
}

// ── List ────────────────────────────────────────────────────────

func POSUserListRoutes() []string { return []string{"/pos/users"} }

type ListPOSUsersHandler struct{ Service *application.Service }

type ListPOSUsersRequest struct {
	LocationID int64 `json:"location_id" schema:"location_id"`
}

type ListPOSUsersResponse struct {
	Users []*posuserv1.POSUser `json:"users"`
}

func (h *ListPOSUsersHandler) Serve(ctx context.Context, req *ListPOSUsersRequest) (ListPOSUsersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListPOSUsersResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListPOSUsersResponse{}, errorModels.PermissionDeniedServerError()
	}

	var users []domain.POSUser
	var err error
	if req.LocationID > 0 {
		users, err = h.Service.ListByLocation(ctx, businessID, req.LocationID)
	} else {
		users, err = h.Service.ListByOwner(ctx, businessID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("failed to list pos_users",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return ListPOSUsersResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list POS users",
		}
	}

	out := make([]*posuserv1.POSUser, 0, len(users))
	for _, u := range users {
		out = append(out, toProtoPOSUser(u))
	}
	return ListPOSUsersResponse{Users: out}, nil
}

// ── Create ──────────────────────────────────────────────────────

func POSUserCreateRoutes() []string { return []string{"/pos/users"} }

type CreatePOSUserHandler struct{ Service *application.Service }

type CreatePOSUserRequest struct {
	UserID           int64   `json:"user_id"`
	LocationID       int64   `json:"location_id"`
	Name             string  `json:"name"`
	PIN              string  `json:"pin"`
	Role             string  `json:"role"`
	HourlyRate       float64 `json:"hourly_rate"`
	MaxDiscountPct   float64 `json:"max_discount_pct"`
	CanVoid          bool    `json:"can_void"`
	CanRefund        bool    `json:"can_refund"`
	CanOpenDrawer    bool    `json:"can_open_drawer"`
	CanApplyDiscount bool    `json:"can_apply_discount"`
}

type CreatePOSUserResponse struct {
	ID int64 `json:"id"`
}

func (h *CreatePOSUserHandler) Validate(_ context.Context, req *CreatePOSUserRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Name is required"}
	}
	req.PIN = strings.TrimSpace(req.PIN)
	if req.PIN == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "PIN is required"}
	}
	if req.LocationID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Location is required"}
	}
	if req.UserID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Team member is required"}
	}
	return nil
}

func (h *CreatePOSUserHandler) Serve(ctx context.Context, req *CreatePOSUserRequest) (CreatePOSUserResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreatePOSUserResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreatePOSUserResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.Create(ctx, businessID, device.User.UserID, domain.CreatePOSUserInput{
		UserID:           req.UserID,
		LocationID:       req.LocationID,
		Name:             req.Name,
		PIN:              req.PIN,
		Role:             domain.POSUserRole(req.Role),
		HourlyRate:       req.HourlyRate,
		MaxDiscountPct:   req.MaxDiscountPct,
		CanVoid:          req.CanVoid,
		CanRefund:        req.CanRefund,
		CanOpenDrawer:    req.CanOpenDrawer,
		CanApplyDiscount: req.CanApplyDiscount,
	})
	// Zero the PIN on the request struct so downstream logging of the
	// request object (should it ever happen) cannot leak it.
	req.PIN = ""
	if err != nil {
		if errors.Is(err, application.ErrInvalidInput) {
			return CreatePOSUserResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid input"}
		}
		logger.Ctx(ctx).Error("failed to create pos_user",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return CreatePOSUserResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create POS user"}
	}
	return CreatePOSUserResponse{ID: id}, nil
}

// ── Update ──────────────────────────────────────────────────────

func POSUserUpdateRoutes() []string { return []string{"/pos/users/update"} }

type UpdatePOSUserHandler struct{ Service *application.Service }

type UpdatePOSUserRequest struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	PIN              string  `json:"pin"`
	Role             string  `json:"role"`
	IsActive         bool    `json:"is_active"`
	HourlyRate       float64 `json:"hourly_rate"`
	MaxDiscountPct   float64 `json:"max_discount_pct"`
	CanVoid          bool    `json:"can_void"`
	CanRefund        bool    `json:"can_refund"`
	CanOpenDrawer    bool    `json:"can_open_drawer"`
	CanApplyDiscount bool    `json:"can_apply_discount"`
}

type UpdatePOSUserResponse struct {
	Success bool `json:"success"`
}

func (h *UpdatePOSUserHandler) Validate(_ context.Context, req *UpdatePOSUserRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *UpdatePOSUserHandler) Serve(ctx context.Context, req *UpdatePOSUserRequest) (UpdatePOSUserResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdatePOSUserResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdatePOSUserResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.Update(ctx, businessID, req.ID, domain.UpdatePOSUserInput{
		Name:             req.Name,
		PIN:              req.PIN,
		Role:             domain.POSUserRole(req.Role),
		IsActive:         req.IsActive,
		HourlyRate:       req.HourlyRate,
		MaxDiscountPct:   req.MaxDiscountPct,
		CanVoid:          req.CanVoid,
		CanRefund:        req.CanRefund,
		CanOpenDrawer:    req.CanOpenDrawer,
		CanApplyDiscount: req.CanApplyDiscount,
	})
	req.PIN = ""
	if err != nil {
		logger.Ctx(ctx).Error("failed to update pos_user",
			logger.Any("business_id", businessID),
			logger.Any("pos_user_id", req.ID),
			logger.Error(err),
		)
		return UpdatePOSUserResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update POS user"}
	}
	return UpdatePOSUserResponse{Success: true}, nil
}

// ── Deactivate ──────────────────────────────────────────────────

func POSUserDeactivateRoutes() []string { return []string{"/pos/users/deactivate"} }

type DeactivatePOSUserHandler struct{ Service *application.Service }

type DeactivatePOSUserRequest struct {
	ID int64 `json:"id"`
}

type DeactivatePOSUserResponse struct {
	Success bool `json:"success"`
}

func (h *DeactivatePOSUserHandler) Validate(_ context.Context, req *DeactivatePOSUserRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *DeactivatePOSUserHandler) Serve(ctx context.Context, req *DeactivatePOSUserRequest) (DeactivatePOSUserResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeactivatePOSUserResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeactivatePOSUserResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.Deactivate(ctx, businessID, req.ID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to deactivate pos_user",
			logger.Any("business_id", businessID),
			logger.Any("pos_user_id", req.ID),
			logger.Error(err),
		)
		return DeactivatePOSUserResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to deactivate POS user"}
	}
	return DeactivatePOSUserResponse{Success: true}, nil
}

// ── Verify PIN ──────────────────────────────────────────────────

func POSUserVerifyPINRoutes() []string { return []string{"/pos/users/verify-pin"} }

type VerifyPINHandler struct{ Service *application.Service }

type VerifyPINRequest struct {
	LocationID int64  `json:"location_id"`
	UserID     int64  `json:"user_id"`
	PIN        string `json:"pin"`
}

type VerifyPINResponse struct {
	UserID           int64   `json:"user_id"`
	POSUserID        int64   `json:"pos_user_id"`
	LocationID       int64   `json:"location_id"`
	Role             string  `json:"role"`
	DisplayName      string  `json:"display_name"`
	CanVoid          bool    `json:"can_void"`
	CanRefund        bool    `json:"can_refund"`
	CanOpenDrawer    bool    `json:"can_open_drawer"`
	CanApplyDiscount bool    `json:"can_apply_discount"`
	MaxDiscountPct   float64 `json:"max_discount_pct"`
}

func (h *VerifyPINHandler) Validate(_ context.Context, req *VerifyPINRequest) *errorModels.ServerError {
	if req.LocationID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Location is required"}
	}
	if req.UserID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Operator is required"}
	}
	if req.PIN == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "PIN is required"}
	}
	return nil
}

func (h *VerifyPINHandler) Serve(ctx context.Context, req *VerifyPINRequest) (VerifyPINResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return VerifyPINResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return VerifyPINResponse{}, errorModels.PermissionDeniedServerError()
	}

	result, err := h.Service.VerifyPIN(ctx, businessID, req.LocationID, req.UserID, req.PIN)
	req.PIN = ""
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPIN) {
			return VerifyPINResponse{}, &errorModels.ServerError{HTTPCode: http.StatusUnauthorized, Code: http.StatusUnauthorized, Message: "Invalid credentials"}
		}
		logger.Ctx(ctx).Error("pos_users verify-pin failed",
			logger.Any("business_id", businessID),
			logger.Any("pos_user_target", req.UserID),
			logger.Error(err),
		)
		return VerifyPINResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to verify PIN"}
	}

	return VerifyPINResponse{
		UserID:           result.UserID,
		POSUserID:        result.POSUserID,
		LocationID:       result.LocationID,
		Role:             string(result.Role),
		DisplayName:      result.DisplayName,
		CanVoid:          result.CanVoid,
		CanRefund:        result.CanRefund,
		CanOpenDrawer:    result.CanOpenDrawer,
		CanApplyDiscount: result.CanApplyDiscount,
		MaxDiscountPct:   result.MaxDiscountPct,
	}, nil
}

func POSUserSetPINRoutes() []string { return []string{"/pos/users/set-pin"} }

type SetPINHandler struct{ Service *application.Service }

type SetPINRequest struct {
	UserID     int64  `json:"user_id"`
	LocationID int64  `json:"location_id"`
	Name       string `json:"name"`
	PIN        string `json:"pin"`
}

type SetPINResponse struct {
	Success bool `json:"success"`
}

func (h *SetPINHandler) Validate(_ context.Context, req *SetPINRequest) *errorModels.ServerError {
	if req.UserID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Operator is required"}
	}
	if req.LocationID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Location is required"}
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Name is required"}
	}
	req.PIN = strings.TrimSpace(req.PIN)
	if req.PIN == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "PIN is required"}
	}
	return nil
}

func (h *SetPINHandler) Serve(ctx context.Context, req *SetPINRequest) (SetPINResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SetPINResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SetPINResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.SetPIN(ctx, businessID, application.SetPINActor{
		UserID:   device.User.UserID,
		IsOwner:  device.User.UserID == businessID,
		TeamRole: device.User.Role,
	}, domain.SetPINInput{
		UserID:     req.UserID,
		LocationID: req.LocationID,
		Name:       req.Name,
		PIN:        req.PIN,
	})
	req.PIN = ""
	if err != nil {
		if errors.Is(err, application.ErrInvalidInput) {
			return SetPINResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Not assigned to this location"}
		}
		if errors.Is(err, application.ErrForbidden) {
			return SetPINResponse{}, errorModels.PermissionDeniedServerError()
		}
		if errors.Is(err, application.ErrPINAlreadySet) {
			return SetPINResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "A PIN is already set. Ask a manager to reset it."}
		}
		logger.Ctx(ctx).Error("pos_users set-pin failed",
			logger.Any("business_id", businessID),
			logger.Any("target_user", req.UserID),
			logger.Error(err),
		)
		return SetPINResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to set PIN"}
	}
	return SetPINResponse{Success: true}, nil
}
