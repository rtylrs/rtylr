package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	posaccessv1 "github.com/voxire/voxie/gen/go/pos/access/v1"
)

func toProtoMyLocation(l infrastructure.MyLocation) *posaccessv1.MyLocation {
	return &posaccessv1.MyLocation{
		LocationId: l.LocationID,
		Name:       l.Name,
		Address:    l.Address,
		Phone:      l.Phone,
	}
}

func MyLocationsRoutes() []string { return []string{"/pos/access/my-locations"} }

type (
	GetMyLocationsHandler  struct{ Service *application.Service }
	GetMyLocationsRequest  struct{}
	GetMyLocationsResponse struct {
		Locations []*posaccessv1.MyLocation `json:"locations"`
	}
)

func (h *GetMyLocationsHandler) Serve(ctx context.Context, _ *GetMyLocationsRequest) (GetMyLocationsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetMyLocationsResponse{}, errorModels.PermissionDeniedServerError()
	}
	var (
		locs []infrastructure.MyLocation
		err  error
	)
	if hasBusinessWidePOSAccess(ctx, authUser) {
		locs, err = h.Service.GetBusinessLocations(ctx, authUser.BusinessID)
	} else {
		locs, err = h.Service.GetMyLocations(ctx, authUser.BusinessID, authUser.UserID)
	}
	if err != nil {
		return GetMyLocationsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	out := make([]*posaccessv1.MyLocation, 0, len(locs))
	for _, l := range locs {
		out = append(out, toProtoMyLocation(l))
	}
	return GetMyLocationsResponse{Locations: out}, nil
}

func LocationRosterRoutes() []string { return []string{"/pos/access/location-roster"} }

type (
	GetLocationRosterHandler struct{ Service *application.Service }
	GetLocationRosterRequest struct {
		LocationID int64 `json:"location_id" schema:"location_id"`
	}
	GetLocationRosterResponse struct {
		Roster []infrastructure.RosterEntry `json:"roster"`
	}
)

func (h *GetLocationRosterHandler) Serve(ctx context.Context, req *GetLocationRosterRequest) (GetLocationRosterResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetLocationRosterResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.LocationID <= 0 {
		return GetLocationRosterResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "locationId required"}
	}

	if !hasBusinessWidePOSAccess(ctx, authUser) {
		assigned, accessErr := h.Service.IsAssignedToLocation(ctx, authUser.BusinessID, authUser.UserID, req.LocationID)
		if accessErr != nil {
			return GetLocationRosterResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: accessErr.Error()}
		}
		if !assigned {
			return GetLocationRosterResponse{}, errorModels.PermissionDeniedServerError()
		}
	}

	roster, err := h.Service.GetLocationRoster(ctx, authUser.BusinessID, req.LocationID)
	if err != nil {
		return GetLocationRosterResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return GetLocationRosterResponse{Roster: roster}, nil
}

func isOwnerOrAdmin(u *authPkg.AuthUser) bool {
	return u.Role == "owner" || u.Role == "admin"
}

func hasBusinessWidePOSAccess(ctx context.Context, u *authPkg.AuthUser) bool {
	if u == nil {
		return false
	}
	if isOwnerOrAdmin(u) {
		return true
	}
	if u.BusinessID <= 0 || u.UserID <= 0 {
		return false
	}
	return tenant.ResolveOwnerID(ctx, u.BusinessID, 0) == u.UserID
}

func MemberLocationsRoutes() []string { return []string{"/pos/access/member-locations"} }

type MemberLocationEntry struct {
	ID                 int64 `json:"id"`
	LocationID         int64 `json:"location_id"`
	CanManagePOSUsers  bool  `json:"can_manage_pos_users"`
	CanManageTerminals bool  `json:"can_manage_terminals"`
	CanViewReports     bool  `json:"can_view_reports"`
}

type (
	GetMemberLocationsHandler struct{ Service *application.Service }
	GetMemberLocationsRequest struct {
		UserID int64 `json:"user_id" schema:"user_id"`
	}
	GetMemberLocationsResponse struct {
		Locations []MemberLocationEntry `json:"locations"`
	}
)

func (h *GetMemberLocationsHandler) Serve(ctx context.Context, req *GetMemberLocationsRequest) (GetMemberLocationsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetMemberLocationsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !hasBusinessWidePOSAccess(ctx, authUser) {
		return GetMemberLocationsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.UserID <= 0 {
		return GetMemberLocationsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "userId required"}
	}
	locs, err := h.Service.ListForUser(ctx, authUser.BusinessID, req.UserID)
	if err != nil {
		return GetMemberLocationsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	out := make([]MemberLocationEntry, 0, len(locs))
	for _, l := range locs {
		out = append(out, MemberLocationEntry{
			ID:                 l.ID,
			LocationID:         l.LocationID,
			CanManagePOSUsers:  l.CanManagePOSUsers,
			CanManageTerminals: l.CanManageTerminals,
			CanViewReports:     l.CanViewReports,
		})
	}
	return GetMemberLocationsResponse{Locations: out}, nil
}

func AssignRoutes() []string { return []string{"/pos/access/assign"} }

type (
	AssignLocationHandler struct{ Service *application.Service }
	AssignLocationRequest struct {
		UserID             int64 `json:"user_id"`
		UserIDCamel        int64 `json:"userId"`
		LocationID         int64 `json:"location_id"`
		LocationIDCamel    int64 `json:"locationId"`
		CanManagePOSUsers  bool  `json:"can_manage_pos_users"`
		CanManageTerminals bool  `json:"can_manage_terminals"`
		CanViewReports     bool  `json:"can_view_reports"`

		CanManagePOSUsersCamel  *bool `json:"canManagePosUsers,omitempty"`
		CanManageTerminalsCamel *bool `json:"canManageTerminals,omitempty"`
		CanViewReportsCamel     *bool `json:"canViewReports,omitempty"`
	}
)

type AssignLocationResponse struct {
	ID int64 `json:"id"`
}

func (req *AssignLocationRequest) normalizeAliases() {
	if req.UserID == 0 {
		req.UserID = req.UserIDCamel
	}
	if req.LocationID == 0 {
		req.LocationID = req.LocationIDCamel
	}
	if req.CanManagePOSUsersCamel != nil {
		req.CanManagePOSUsers = *req.CanManagePOSUsersCamel
	}
	if req.CanManageTerminalsCamel != nil {
		req.CanManageTerminals = *req.CanManageTerminalsCamel
	}
	if req.CanViewReportsCamel != nil {
		req.CanViewReports = *req.CanViewReportsCamel
	}
}

func (h *AssignLocationHandler) Validate(_ context.Context, req *AssignLocationRequest) *errorModels.ServerError {
	req.normalizeAliases()
	if req.UserID <= 0 || req.LocationID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "userId and locationId required"}
	}
	return nil
}

func (h *AssignLocationHandler) Serve(ctx context.Context, req *AssignLocationRequest) (AssignLocationResponse, *errorModels.ServerError) {
	req.normalizeAliases()
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return AssignLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !hasBusinessWidePOSAccess(ctx, authUser) {
		return AssignLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Assign(ctx, authUser.BusinessID, req.UserID, req.LocationID, req.CanManagePOSUsers, req.CanManageTerminals, req.CanViewReports)
	if err != nil {
		if errors.Is(err, application.ErrInvalidAssignment) {
			return AssignLocationResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid POS location assignment"}
		}
		return AssignLocationResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return AssignLocationResponse{ID: id}, nil
}

func RemoveRoutes() []string { return []string{"/pos/access/remove"} }

type (
	RemoveLocationHandler struct{ Service *application.Service }
	RemoveLocationRequest struct {
		ID int64 `json:"id"`
	}
)

type RemoveLocationResponse struct {
	Success bool `json:"success"`
}

func (h *RemoveLocationHandler) Validate(_ context.Context, req *RemoveLocationRequest) *errorModels.ServerError {
	if req.ID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID required"}
	}
	return nil
}

func (h *RemoveLocationHandler) Serve(ctx context.Context, req *RemoveLocationRequest) (RemoveLocationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return RemoveLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !hasBusinessWidePOSAccess(ctx, authUser) {
		return RemoveLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Remove(ctx, authUser.BusinessID, req.ID)
	if err != nil {
		return RemoveLocationResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return RemoveLocationResponse{Success: true}, nil
}
