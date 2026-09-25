package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/application"
	"github.com/voxire/onion/internal/modules/pos_floors/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posfloorv1 "github.com/voxire/voxie/gen/go/pos/floor/v1"
)

type FloorAPI struct {
	Service *application.Service
}

func NewFloorAPI(service *application.Service) *FloorAPI {
	return &FloorAPI{Service: service}
}

func (a *FloorAPI) ListFloors(ctx context.Context, req *posfloorv1.ListFloorsRequest) (*posfloorv1.ListFloorsResponse, error) {
	user, err := authenticatedFloorUser(ctx)
	if err != nil {
		return nil, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	floors, listErr := a.Service.List(ctx, businessID, req.GetBranchId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list floors")
	}

	payloads := make([]*posfloorv1.Floor, 0, len(floors))
	for _, floor := range floors {
		payloads = append(payloads, toProtoFloor(floor))
	}

	return &posfloorv1.ListFloorsResponse{Floors: payloads}, nil
}

func (a *FloorAPI) CreateFloor(ctx context.Context, req *posfloorv1.CreateFloorRequest) (*posfloorv1.CreateFloorResponse, error) {
	user, err := authenticatedFloorUser(ctx)
	if err != nil {
		return nil, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	id, createErr := a.Service.Create(ctx, domain.FloorInput{
		BusinessID: businessID,
		BranchID:   req.GetBranchId(),
		Name:       req.GetName(),
		SortOrder:  int(req.GetSortOrder()),
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create floor")
	}

	return &posfloorv1.CreateFloorResponse{FloorId: id, Message: "Floor created successfully"}, nil
}

func (a *FloorAPI) UpdateFloor(ctx context.Context, req *posfloorv1.UpdateFloorRequest) (*posfloorv1.UpdateFloorResponse, error) {
	user, err := authenticatedFloorUser(ctx)
	if err != nil {
		return nil, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	if updateErr := a.Service.Update(ctx, domain.FloorUpdateInput{
		ID:         req.GetId(),
		BusinessID: businessID,
		Name:       body.GetName(),
		SortOrder:  int(body.GetSortOrder()),
	}); updateErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update floor")
	}

	return &posfloorv1.UpdateFloorResponse{Message: "Floor updated successfully"}, nil
}

func (a *FloorAPI) DeleteFloor(ctx context.Context, req *posfloorv1.DeleteFloorRequest) (*posfloorv1.DeleteFloorResponse, error) {
	user, err := authenticatedFloorUser(ctx)
	if err != nil {
		return nil, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if deleteErr := a.Service.Delete(ctx, businessID, req.GetId()); deleteErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete floor")
	}

	return &posfloorv1.DeleteFloorResponse{Message: "Floor deleted successfully"}, nil
}

func authenticatedFloorUser(ctx context.Context) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	return device.User, nil
}

func toProtoFloor(f domain.Floor) *posfloorv1.Floor {
	return &posfloorv1.Floor{
		Id:         f.ID,
		BusinessId: f.BusinessID,
		BranchId:   f.BranchID,
		Name:       f.Name,
		SortOrder:  int32(f.SortOrder), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IsActive:   f.IsActive,
		CreatedAt:  f.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
