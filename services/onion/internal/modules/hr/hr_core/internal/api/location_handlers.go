package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type LocationPayload struct {
	ID        int64  `json:"id"`
	OwnerID   int64  `json:"owner_id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Timezone  string `json:"timezone"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func locationToPayload(loc domain.LocationRecord) *LocationPayload {
	return &LocationPayload{
		ID: loc.ID, OwnerID: loc.OwnerID, Name: loc.Name,
		Address: loc.Address, City: loc.City, Country: loc.Country,
		Timezone: loc.Timezone, IsDefault: loc.IsDefault,
		CreatedAt: loc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: loc.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List Locations ---

func ListLocationsRoutes() []string { return []string{"/hr/locations"} }

type (
	ListLocationsHandler  struct{ Service *application.Service }
	ListLocationsRequest  struct{}
	ListLocationsResponse struct {
		Locations []*LocationPayload `json:"locations"`
	}
)

func (h *ListLocationsHandler) Serve(ctx context.Context, _ *ListLocationsRequest) (ListLocationsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListLocationsResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLocationsResponse{}, onekiterrors.PermissionDenied()
	}

	locations, err := h.Service.ListLocations(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list locations", logger.Any("business_id", businessID), logger.Error(err))
		return ListLocationsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list locations")
	}

	payloads := make([]*LocationPayload, 0, len(locations))
	for _, loc := range locations {
		payloads = append(payloads, locationToPayload(loc))
	}
	return ListLocationsResponse{Locations: payloads}, nil
}

// --- Create Location ---

func CreateLocationRoutes() []string { return []string{"/hr/locations"} }

type (
	CreateLocationHandler struct{ Service *application.Service }
	CreateLocationRequest struct {
		Name      string `json:"name"`
		Address   string `json:"address"`
		City      string `json:"city"`
		Country   string `json:"country"`
		Timezone  string `json:"timezone"`
		IsDefault bool   `json:"is_default"`
	}
)

type CreateLocationResponse struct {
	Location *LocationPayload `json:"location"`
}

func (h *CreateLocationHandler) Validate(_ context.Context, req *CreateLocationRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *CreateLocationHandler) Serve(ctx context.Context, req *CreateLocationRequest) (CreateLocationResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateLocationResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:manage"); err != nil {
		return CreateLocationResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateLocationResponse{}, onekiterrors.PermissionDenied()
	}

	loc := domain.LocationRecord{
		Name: req.Name, Address: req.Address, City: req.City,
		Country: req.Country, Timezone: req.Timezone, IsDefault: req.IsDefault,
	}
	created, err := h.Service.CreateLocation(ctx, businessID, loc)
	if err != nil {
		logger.Ctx(ctx).Error("failed to create location", logger.Any("business_id", businessID), logger.Error(err))
		return CreateLocationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create location")
	}

	return CreateLocationResponse{Location: locationToPayload(*created)}, nil
}

// --- Update Location ---

func UpdateLocationRoutes() []string { return []string{"/hr/locations/{id}"} }

type (
	UpdateLocationHandler struct{ Service *application.Service }
	UpdateLocationRequest struct {
		Name      string `json:"name"`
		Address   string `json:"address"`
		City      string `json:"city"`
		Country   string `json:"country"`
		Timezone  string `json:"timezone"`
		IsDefault bool   `json:"is_default"`
	}
)

type UpdateLocationResponse struct {
	Location *LocationPayload `json:"location"`
}

func (h *UpdateLocationHandler) Validate(_ context.Context, req *UpdateLocationRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *UpdateLocationHandler) Serve(ctx context.Context, req *UpdateLocationRequest) (UpdateLocationResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateLocationResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:manage"); err != nil {
		return UpdateLocationResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLocationResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateLocationResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid location id")
	}

	loc := domain.LocationRecord{
		ID: id, Name: req.Name, Address: req.Address, City: req.City,
		Country: req.Country, Timezone: req.Timezone, IsDefault: req.IsDefault,
	}
	updated, err := h.Service.UpdateLocation(ctx, businessID, loc)
	if err != nil {
		if errors.Is(err, domain.ErrLocationNotFound) {
			return UpdateLocationResponse{}, onekiterrors.New(http.StatusNotFound, "Location not found")
		}
		logger.Ctx(ctx).Error("failed to update location", logger.Any("business_id", businessID), logger.Any("location_id", id), logger.Error(err))
		return UpdateLocationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update location")
	}

	return UpdateLocationResponse{Location: locationToPayload(*updated)}, nil
}
