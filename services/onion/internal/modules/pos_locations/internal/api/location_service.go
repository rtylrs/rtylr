package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	poslocationv1 "github.com/voxire/voxie/gen/go/pos/location/v1"
)

type LocationAPI struct {
	Service *application.Service
}

func NewLocationAPI(service *application.Service) *LocationAPI {
	return &LocationAPI{Service: service}
}

func (a *LocationAPI) ListLocations(ctx context.Context, _ *poslocationv1.ListLocationsRequest) (*poslocationv1.ListLocationsResponse, error) {
	businessID, err := requireLocationPermission(ctx, "pos:orders:view")
	if err != nil {
		return nil, err
	}

	locations, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list locations")
	}

	payloads := make([]*poslocationv1.Location, len(locations))
	for i, l := range locations {
		payloads[i] = toProtoLocation(l)
	}
	return &poslocationv1.ListLocationsResponse{Locations: payloads}, nil
}

func (a *LocationAPI) CreateLocation(ctx context.Context, req *poslocationv1.CreateLocationRequest) (*poslocationv1.CreateLocationResponse, error) {
	device, businessID, err := requireLocationDevice(ctx, "pos:employees:manage")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = req.GetIsActive()
	}

	id, createErr := a.Service.Create(ctx, domain.LocationInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:       req.GetName(),
		Address:    req.GetAddress(),
		Phone:      req.GetPhone(),
		IsActive:   isActive,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create location")
	}
	return &poslocationv1.CreateLocationResponse{LocationId: id, Message: "Location created successfully"}, nil
}

func (a *LocationAPI) UpdateLocation(ctx context.Context, req *poslocationv1.UpdateLocationRequest) (*poslocationv1.UpdateLocationResponse, error) {
	device, businessID, err := requireLocationDevice(ctx, "pos:employees:manage")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid location ID")
	}

	var isActive *bool
	if body.IsActive != nil {
		v := body.GetIsActive()
		isActive = &v
	}

	if updateErr := a.Service.Update(ctx, domain.LocationUpdateInput{
		ID:         req.GetId(),
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:       body.GetName(),
		Address:    body.GetAddress(),
		Phone:      body.GetPhone(),
		IsActive:   isActive,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrLocationNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Location not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update location")
	}
	return &poslocationv1.UpdateLocationResponse{Message: "Location updated successfully"}, nil
}

func requireLocationPermission(ctx context.Context, permission string) (int64, error) {
	_, businessID, err := requireLocationDevice(ctx, permission)
	return businessID, err
}

func requireLocationDevice(ctx context.Context, permission string) (authPkg.Device, int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppPOS, permission) {
		return device, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return device, 0, onekiterrors.PermissionDenied()
	}
	return device, businessID, nil
}

func toProtoLocation(l domain.Location) *poslocationv1.Location {
	return &poslocationv1.Location{
		Id:        l.ID,
		OwnerId:   l.OwnerID,
		Name:      l.Name,
		Address:   l.Address,
		Phone:     l.Phone,
		IsActive:  l.IsActive,
		CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: l.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
