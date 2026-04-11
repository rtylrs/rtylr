package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/infrastructure"
)

func MyLocationsRoutes() []string { return []string{"/pos/access/my-locations"} }

type GetMyLocationsHandler struct{ Service *application.Service }
type GetMyLocationsRequest struct{}
type GetMyLocationsResponse struct{ Locations []infrastructure.MyLocation `json:"locations"` }

func (h *GetMyLocationsHandler) Serve(ctx context.Context, _ *GetMyLocationsRequest) (GetMyLocationsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return GetMyLocationsResponse{}, errorModels.PermissionDeniedServerError() }
	locs, err := h.Service.GetMyLocations(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil { return GetMyLocationsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }
	return GetMyLocationsResponse{Locations: locs}, nil
}

func AssignRoutes() []string { return []string{"/pos/access/assign"} }

type AssignLocationHandler struct{ Service *application.Service }
type AssignLocationRequest struct {
	UserID             int64 `json:"userId"`
	LocationID         int64 `json:"locationId"`
	CanManagePOSUsers  bool  `json:"canManagePosUsers"`
	CanManageTerminals bool  `json:"canManageTerminals"`
	CanViewReports     bool  `json:"canViewReports"`
}
type AssignLocationResponse struct{ ID int64 `json:"id"` }

func (h *AssignLocationHandler) Validate(_ context.Context, req *AssignLocationRequest) *errorModels.ServerError {
	if req.UserID == 0 || req.LocationID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "userId and locationId required"} }
	return nil
}

func (h *AssignLocationHandler) Serve(ctx context.Context, req *AssignLocationRequest) (AssignLocationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return AssignLocationResponse{}, errorModels.PermissionDeniedServerError() }
	id, err := h.Service.Assign(ctx, authUser.BusinessID, req.UserID, req.LocationID, req.CanManagePOSUsers, req.CanManageTerminals, req.CanViewReports)
	if err != nil { return AssignLocationResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }
	return AssignLocationResponse{ID: id}, nil
}

func RemoveRoutes() []string { return []string{"/pos/access/remove"} }

type RemoveLocationHandler struct{ Service *application.Service }
type RemoveLocationRequest struct{ ID int64 `json:"id"` }
type RemoveLocationResponse struct{ Success bool `json:"success"` }

func (h *RemoveLocationHandler) Validate(_ context.Context, req *RemoveLocationRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID required"} }
	return nil
}

func (h *RemoveLocationHandler) Serve(ctx context.Context, req *RemoveLocationRequest) (RemoveLocationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return RemoveLocationResponse{}, errorModels.PermissionDeniedServerError() }
	err := h.Service.Remove(ctx, authUser.BusinessID, req.ID)
	if err != nil { return RemoveLocationResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()} }
	return RemoveLocationResponse{Success: true}, nil
}
