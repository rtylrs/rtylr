package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_users/internal/application"
	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
)

type POSUserPayload struct {
	ID               int64   `json:"id"`
	BusinessID       int64   `json:"business_id"`
	LocationID       int64   `json:"location_id"`
	Name             string  `json:"name"`
	Role             string  `json:"role"`
	IsActive         bool    `json:"is_active"`
	MaxDiscountPct   float64 `json:"max_discount_pct"`
	CanVoid          bool    `json:"can_void"`
	CanRefund        bool    `json:"can_refund"`
	CanOpenDrawer    bool    `json:"can_open_drawer"`
	CanApplyDiscount bool    `json:"can_apply_discount"`
}

func toPayload(u domain.POSUser) POSUserPayload {
	return POSUserPayload{
		ID: u.ID, BusinessID: u.BusinessID, LocationID: u.LocationID,
		Name: u.Name, Role: string(u.Role), IsActive: u.IsActive,
		MaxDiscountPct: u.MaxDiscountPct, CanVoid: u.CanVoid, CanRefund: u.CanRefund,
		CanOpenDrawer: u.CanOpenDrawer, CanApplyDiscount: u.CanApplyDiscount,
	}
}

// ── List ────────────────────────────────────────────────────────

func POSUserListRoutes() []string { return []string{"/pos/users"} }

type ListPOSUsersHandler struct{ Service *application.Service }
type ListPOSUsersRequest struct{ LocationID int64 `json:"location_id" schema:"location_id"` }
type ListPOSUsersResponse struct{ Users []POSUserPayload `json:"users"` }

func (h *ListPOSUsersHandler) Serve(ctx context.Context, req *ListPOSUsersRequest) (ListPOSUsersResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return ListPOSUsersResponse{}, errorModels.PermissionDeniedServerError() }

	var users []domain.POSUser
	var err error
	if req.LocationID > 0 {
		users, err = h.Service.ListByLocation(ctx, authUser.BusinessID, req.LocationID)
	} else {
		users, err = h.Service.ListByBusiness(ctx, authUser.BusinessID)
	}
	if err != nil { return ListPOSUsersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }

	out := make([]POSUserPayload, 0, len(users))
	for _, u := range users { out = append(out, toPayload(u)) }
	return ListPOSUsersResponse{Users: out}, nil
}

// ── Create ──────────────────────────────────────────────────────

func POSUserCreateRoutes() []string { return []string{"/pos/users"} }

type CreatePOSUserHandler struct{ Service *application.Service }
type CreatePOSUserRequest struct {
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
type CreatePOSUserResponse struct{ ID int64 `json:"id"` }

func (h *CreatePOSUserHandler) Validate(_ context.Context, req *CreatePOSUserRequest) *errorModels.ServerError {
	if req.Name == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"} }
	if req.PIN == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "PIN is required"} }
	if req.LocationID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Location is required"} }
	return nil
}

func (h *CreatePOSUserHandler) Serve(ctx context.Context, req *CreatePOSUserRequest) (CreatePOSUserResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreatePOSUserResponse{}, errorModels.PermissionDeniedServerError() }

	id, err := h.Service.Create(ctx, authUser.BusinessID, authUser.UserID, domain.CreatePOSUserInput{
		LocationID: req.LocationID, Name: req.Name, PIN: req.PIN,
		Role: domain.POSUserRole(req.Role), HourlyRate: req.HourlyRate,
		MaxDiscountPct: req.MaxDiscountPct, CanVoid: req.CanVoid, CanRefund: req.CanRefund,
		CanOpenDrawer: req.CanOpenDrawer, CanApplyDiscount: req.CanApplyDiscount,
	})
	if err != nil { return CreatePOSUserResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }
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
type UpdatePOSUserResponse struct{ Success bool `json:"success"` }

func (h *UpdatePOSUserHandler) Validate(_ context.Context, req *UpdatePOSUserRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"} }
	return nil
}

func (h *UpdatePOSUserHandler) Serve(ctx context.Context, req *UpdatePOSUserRequest) (UpdatePOSUserResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdatePOSUserResponse{}, errorModels.PermissionDeniedServerError() }
	err := h.Service.Update(ctx, authUser.BusinessID, req.ID, domain.UpdatePOSUserInput{
		Name: req.Name, PIN: req.PIN, Role: domain.POSUserRole(req.Role),
		IsActive: req.IsActive, HourlyRate: req.HourlyRate, MaxDiscountPct: req.MaxDiscountPct,
		CanVoid: req.CanVoid, CanRefund: req.CanRefund,
		CanOpenDrawer: req.CanOpenDrawer, CanApplyDiscount: req.CanApplyDiscount,
	})
	if err != nil { return UpdatePOSUserResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }
	return UpdatePOSUserResponse{Success: true}, nil
}

// ── Deactivate ──────────────────────────────────────────────────

func POSUserDeactivateRoutes() []string { return []string{"/pos/users/deactivate"} }

type DeactivatePOSUserHandler struct{ Service *application.Service }
type DeactivatePOSUserRequest struct{ ID int64 `json:"id"` }
type DeactivatePOSUserResponse struct{ Success bool `json:"success"` }

func (h *DeactivatePOSUserHandler) Validate(_ context.Context, req *DeactivatePOSUserRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"} }
	return nil
}

func (h *DeactivatePOSUserHandler) Serve(ctx context.Context, req *DeactivatePOSUserRequest) (DeactivatePOSUserResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return DeactivatePOSUserResponse{}, errorModels.PermissionDeniedServerError() }
	err := h.Service.Deactivate(ctx, authUser.BusinessID, req.ID)
	if err != nil { return DeactivatePOSUserResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }
	return DeactivatePOSUserResponse{Success: true}, nil
}

// ── Verify PIN ──────────────────────────────────────────────────

func POSUserVerifyPINRoutes() []string { return []string{"/pos/users/verify-pin"} }

type VerifyPINHandler struct{ Service *application.Service }
type VerifyPINRequest struct{ PIN string `json:"pin"` }
type VerifyPINResponse struct{ User *POSUserPayload `json:"user"` }

func (h *VerifyPINHandler) Validate(_ context.Context, req *VerifyPINRequest) *errorModels.ServerError {
	if req.PIN == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "PIN is required"} }
	return nil
}

func (h *VerifyPINHandler) Serve(ctx context.Context, req *VerifyPINRequest) (VerifyPINResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return VerifyPINResponse{}, errorModels.PermissionDeniedServerError() }
	user, err := h.Service.VerifyPIN(ctx, authUser.BusinessID, req.PIN)
	if err != nil { return VerifyPINResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }
	if user == nil { return VerifyPINResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Invalid PIN"} }
	p := toPayload(*user)
	return VerifyPINResponse{User: &p}, nil
}
