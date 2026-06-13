package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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
	locs, err := h.Service.GetMyLocations(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return GetMyLocationsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	out := make([]*posaccessv1.MyLocation, 0, len(locs))
	for _, l := range locs {
		out = append(out, toProtoMyLocation(l))
	}
	return GetMyLocationsResponse{Locations: out}, nil
}

// ── Location roster ─────────────────────────────────────────────

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
	if req.LocationID == 0 {
		return GetLocationRosterResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "locationId required"}
	}

	// The roster exposes co-worker names + PIN status, so restrict it to the
	// location's own people: owner/admin (manage all locations) or a member
	// assigned to this location. Prevents enumerating other locations' staff.
	if !isOwnerOrAdmin(authUser) {
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

// isOwnerOrAdmin reports whether the caller manages the whole business and so
// may view any location. The owner's user id equals the business id; admins
// carry the team admin role.
func isOwnerOrAdmin(u *authPkg.AuthUser) bool {
	return u.UserID == u.BusinessID || u.Role == "owner" || u.Role == "admin"
}

// ── Member locations (admin: view one member's assignments) ─────

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
	// Viewing another member's assignments is a team-admin action.
	if !isOwnerOrAdmin(authUser) {
		return GetMemberLocationsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.UserID == 0 {
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
		LocationID         int64 `json:"location_id"`
		CanManagePOSUsers  bool  `json:"can_manage_pos_users"`
		CanManageTerminals bool  `json:"can_manage_terminals"`
		CanViewReports     bool  `json:"can_view_reports"`
	}
)

type AssignLocationResponse struct {
	ID int64 `json:"id"`
}

func (h *AssignLocationHandler) Validate(_ context.Context, req *AssignLocationRequest) *errorModels.ServerError {
	if req.UserID == 0 || req.LocationID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "userId and locationId required"}
	}
	return nil
}

func (h *AssignLocationHandler) Serve(ctx context.Context, req *AssignLocationRequest) (AssignLocationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return AssignLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Assign(ctx, authUser.BusinessID, req.UserID, req.LocationID, req.CanManagePOSUsers, req.CanManageTerminals, req.CanViewReports)
	if err != nil {
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
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID required"}
	}
	return nil
}

func (h *RemoveLocationHandler) Serve(ctx context.Context, req *RemoveLocationRequest) (RemoveLocationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return RemoveLocationResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Remove(ctx, authUser.BusinessID, req.ID)
	if err != nil {
		return RemoveLocationResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return RemoveLocationResponse{Success: true}, nil
}
