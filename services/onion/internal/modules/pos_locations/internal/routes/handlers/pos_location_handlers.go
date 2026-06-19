package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type LocationPayload struct {
	ID        int64  `json:"id"`
	OwnerID   int64  `json:"owner_id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func newLocationPayload(l domain.Location) LocationPayload {
	return LocationPayload{
		ID:        l.ID,
		OwnerID:   l.OwnerID,
		Name:      l.Name,
		Address:   l.Address,
		Phone:     l.Phone,
		IsActive:  l.IsActive,
		CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: l.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func LocationListRoutes() []string { return []string{"/pos/locations"} }

type (
	ListLocationsHandler  struct{ Service *application.Service }
	ListLocationsRequest  struct{}
	ListLocationsResponse struct {
		Locations []LocationPayload `json:"locations"`
	}
)

func (h *ListLocationsHandler) Serve(ctx context.Context, _ *ListLocationsRequest) (ListLocationsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLocationsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLocationsResponse{}, errorModels.PermissionDeniedServerError()
	}
	locations, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListLocationsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list locations"}
	}
	payloads := make([]LocationPayload, len(locations))
	for i, l := range locations {
		payloads[i] = newLocationPayload(l)
	}
	return ListLocationsResponse{Locations: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func LocationCreateRoutes() []string { return []string{"/pos/locations"} }

type CreateLocationHandler struct{ Service *application.Service }

type CreateLocationRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	IsActive *bool  `json:"is_active"`
}

type CreateLocationResponse struct {
	LocationID int64  `json:"location_id"`
	Message    string `json:"message"`
}

func (h *CreateLocationHandler) Validate(_ context.Context, req *CreateLocationRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateLocationHandler) Serve(ctx context.Context, req *CreateLocationRequest) (CreateLocationResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	id, err := h.Service.Create(ctx, domain.LocationInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:       req.Name,
		Address:    req.Address,
		Phone:      req.Phone,
		IsActive:   isActive,
	})
	if err != nil {
		return CreateLocationResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create location"}
	}
	return CreateLocationResponse{LocationID: id, Message: "Location created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func LocationUpdateRoutes() []string { return []string{"/pos/locations/{id}"} }

type UpdateLocationHandler struct{ Service *application.Service }

type UpdateLocationRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	IsActive *bool  `json:"is_active"`
}

type UpdateLocationResponse struct {
	Message string `json:"message"`
}

func (h *UpdateLocationHandler) Validate(_ context.Context, req *UpdateLocationRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateLocationHandler) Serve(ctx context.Context, req *UpdateLocationRequest) (UpdateLocationResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateLocationResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid location ID"}
	}
	if updateErr := h.Service.Update(ctx, domain.LocationUpdateInput{
		ID:         id,
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:       req.Name,
		Address:    req.Address,
		Phone:      req.Phone,
		IsActive:   req.IsActive,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrLocationNotFound) {
			return UpdateLocationResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Location not found"}
		}
		return UpdateLocationResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update location"}
	}
	return UpdateLocationResponse{Message: "Location updated successfully"}, nil
}
