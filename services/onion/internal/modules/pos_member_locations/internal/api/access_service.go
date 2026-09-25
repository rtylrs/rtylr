package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/application"
	"github.com/voxire/onion/internal/modules/pos_member_locations/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posaccessv1 "github.com/voxire/voxie/gen/go/pos/access/v1"
)

type AccessAPI struct {
	Service *application.Service
}

func NewAccessAPI(service *application.Service) *AccessAPI {
	return &AccessAPI{Service: service}
}

func (a *AccessAPI) GetMyLocations(ctx context.Context, _ *posaccessv1.GetMyLocationsRequest) (*posaccessv1.GetMyLocationsResponse, error) {
	authUser, businessID, err := authenticatedAccessUser(ctx)
	if err != nil {
		return nil, err
	}

	var (
		locs    []infrastructure.MyLocation
		listErr error
	)
	if hasBusinessWidePOSAccess(authUser) {
		locs, listErr = a.Service.GetBusinessLocations(ctx, businessID)
	} else {
		locs, listErr = a.Service.GetMyLocations(ctx, businessID, authUser.UserID)
	}
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, listErr.Error())
	}

	out := make([]*posaccessv1.MyLocation, 0, len(locs))
	for _, l := range locs {
		out = append(out, toProtoMyLocation(l))
	}
	return &posaccessv1.GetMyLocationsResponse{Locations: out}, nil
}

func (a *AccessAPI) GetLocationRoster(ctx context.Context, req *posaccessv1.GetLocationRosterRequest) (*posaccessv1.GetLocationRosterResponse, error) {
	authUser, businessID, err := authenticatedAccessUser(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetLocationId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "location_id required")
	}

	if !hasBusinessWidePOSAccess(authUser) {
		assigned, accessErr := a.Service.IsAssignedToLocation(ctx, businessID, authUser.UserID, req.GetLocationId())
		if accessErr != nil {
			return nil, onekiterrors.New(http.StatusInternalServerError, accessErr.Error())
		}
		if !assigned {
			return nil, onekiterrors.PermissionDenied()
		}
	}

	roster, listErr := a.Service.GetLocationRoster(ctx, businessID, req.GetLocationId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, listErr.Error())
	}

	out := make([]*posaccessv1.RosterEntry, 0, len(roster))
	for _, entry := range roster {
		out = append(out, &posaccessv1.RosterEntry{
			UserId:    entry.UserID,
			Name:      entry.Name,
			HasPin:    entry.HasPIN,
			IsManager: entry.IsManager,
			PosRole:   entry.POSRole,
		})
	}
	return &posaccessv1.GetLocationRosterResponse{Roster: out}, nil
}

func (a *AccessAPI) GetMemberLocations(ctx context.Context, req *posaccessv1.GetMemberLocationsRequest) (*posaccessv1.GetMemberLocationsResponse, error) {
	authUser, businessID, err := authenticatedAccessUser(ctx)
	if err != nil {
		return nil, err
	}
	if !hasBusinessWidePOSAccess(authUser) {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetUserId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "user_id required")
	}

	locs, listErr := a.Service.ListForUser(ctx, businessID, req.GetUserId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, listErr.Error())
	}

	out := make([]*posaccessv1.MemberLocationEntry, 0, len(locs))
	for _, l := range locs {
		out = append(out, &posaccessv1.MemberLocationEntry{
			Id:                 l.ID,
			LocationId:         l.LocationID,
			CanManagePosUsers:  l.CanManagePOSUsers,
			CanManageTerminals: l.CanManageTerminals,
			CanViewReports:     l.CanViewReports,
		})
	}
	return &posaccessv1.GetMemberLocationsResponse{Locations: out}, nil
}

func (a *AccessAPI) AssignLocation(ctx context.Context, req *posaccessv1.AssignLocationRequest) (*posaccessv1.AssignLocationResponse, error) {
	authUser, businessID, err := authenticatedAccessUser(ctx)
	if err != nil {
		return nil, err
	}
	if !hasBusinessWidePOSAccess(authUser) {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetUserId() <= 0 || req.GetLocationId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "user_id and location_id required")
	}

	id, assignErr := a.Service.Assign(ctx, businessID, req.GetUserId(), req.GetLocationId(), req.GetCanManagePosUsers(), req.GetCanManageTerminals(), req.GetCanViewReports())
	if assignErr != nil {
		if errors.Is(assignErr, application.ErrInvalidAssignment) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid POS location assignment")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, assignErr.Error())
	}
	return &posaccessv1.AssignLocationResponse{Id: id}, nil
}

func (a *AccessAPI) RemoveLocation(ctx context.Context, req *posaccessv1.RemoveLocationRequest) (*posaccessv1.RemoveLocationResponse, error) {
	authUser, businessID, err := authenticatedAccessUser(ctx)
	if err != nil {
		return nil, err
	}
	if !hasBusinessWidePOSAccess(authUser) {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID required")
	}

	if removeErr := a.Service.Remove(ctx, businessID, req.GetId()); removeErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, removeErr.Error())
	}
	return &posaccessv1.RemoveLocationResponse{Success: true}, nil
}

func authenticatedAccessUser(ctx context.Context) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	return authUser, businessID, nil
}

func hasBusinessWidePOSAccess(u *authPkg.AuthUser) bool {
	return authPkg.HasAppPermission(u, authorization.AppPOS, "pos:employees:manage")
}

func toProtoMyLocation(l infrastructure.MyLocation) *posaccessv1.MyLocation {
	return &posaccessv1.MyLocation{
		LocationId: l.LocationID,
		Name:       l.Name,
		Address:    l.Address,
		Phone:      l.Phone,
	}
}
