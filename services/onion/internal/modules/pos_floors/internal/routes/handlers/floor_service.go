package handlers

import (
	"context"
	"net/http"
	"strings"

	posfloorv1 "github.com/voxire/voxie/gen/go/pos/floor/v1"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/application"
	"github.com/voxire/onion/internal/modules/pos_floors/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

	floors, listErr := a.Service.List(ctx, user.BusinessID, req.GetBranchId())
	if listErr != nil {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list floors"}
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
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Name is required"}
	}

	id, createErr := a.Service.Create(ctx, domain.FloorInput{
		BusinessID: user.BusinessID,
		BranchID:   req.GetBranchId(),
		Name:       req.GetName(),
		SortOrder:  int(req.GetSortOrder()),
	})
	if createErr != nil {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create floor"}
	}

	return &posfloorv1.CreateFloorResponse{FloorId: id, Message: "Floor created successfully"}, nil
}

func (a *FloorAPI) UpdateFloor(ctx context.Context, req *posfloorv1.UpdateFloorRequest) (*posfloorv1.UpdateFloorResponse, error) {
	user, err := authenticatedFloorUser(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Name is required"}
	}

	if updateErr := a.Service.Update(ctx, domain.FloorUpdateInput{
		ID:         req.GetId(),
		BusinessID: user.BusinessID,
		Name:       req.GetName(),
		SortOrder:  int(req.GetSortOrder()),
	}); updateErr != nil {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update floor"}
	}

	return &posfloorv1.UpdateFloorResponse{Message: "Floor updated successfully"}, nil
}

func (a *FloorAPI) DeleteFloor(ctx context.Context, req *posfloorv1.DeleteFloorRequest) (*posfloorv1.DeleteFloorResponse, error) {
	user, err := authenticatedFloorUser(ctx)
	if err != nil {
		return nil, err
	}

	if deleteErr := a.Service.Delete(ctx, user.BusinessID, req.GetId()); deleteErr != nil {
		return nil, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete floor"}
	}

	return &posfloorv1.DeleteFloorResponse{Message: "Floor deleted successfully"}, nil
}

func authenticatedFloorUser(ctx context.Context) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, errorModels.PermissionDeniedServerError()
	}
	return device.User, nil
}
