package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos/sessions/internal/application"
	"github.com/voxire/onion/internal/modules/pos/sessions/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	possessionv1 "github.com/voxire/voxie/gen/go/pos/session/v1"
)

type SessionAPI struct {
	Service *application.Service
}

func NewSessionAPI(service *application.Service) *SessionAPI {
	return &SessionAPI{Service: service}
}

func (a *SessionAPI) ListSessions(ctx context.Context, _ *possessionv1.ListSessionsRequest) (*possessionv1.ListSessionsResponse, error) {
	businessID, err := requireSessionPermission(ctx, "pos:orders:view")
	if err != nil {
		return nil, err
	}

	sessions, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list sessions")
	}

	payloads := make([]*possessionv1.Session, len(sessions))
	for i, s := range sessions {
		payloads[i] = toProtoSession(s)
	}
	return &possessionv1.ListSessionsResponse{Sessions: payloads}, nil
}

func (a *SessionAPI) GetCurrentSession(ctx context.Context, _ *possessionv1.GetCurrentSessionRequest) (*possessionv1.GetCurrentSessionResponse, error) {
	businessID, err := requireSessionPermission(ctx, "pos:orders:view")
	if err != nil {
		return nil, err
	}

	device := authPkg.DeviceCtx(ctx)
	sessions, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get current session")
	}

	for _, s := range sessions {
		if s.Status != "open" {
			continue
		}
		if s.StaffID != nil && *s.StaffID == device.User.UserID {
			return &possessionv1.GetCurrentSessionResponse{Session: toProtoCurrentSession(s)}, nil
		}
	}
	return &possessionv1.GetCurrentSessionResponse{}, nil
}

func (a *SessionAPI) GetSession(ctx context.Context, req *possessionv1.GetSessionRequest) (*possessionv1.GetSessionResponse, error) {
	businessID, err := requireSessionPermission(ctx, "pos:orders:view")
	if err != nil {
		return nil, err
	}
	if req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid session ID")
	}

	session, getErr := a.Service.Get(ctx, businessID, req.GetId())
	if getErr != nil {
		if errors.Is(getErr, domain.ErrSessionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Session not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get session")
	}
	return &possessionv1.GetSessionResponse{Session: toProtoSession(*session)}, nil
}

func (a *SessionAPI) OpenSession(ctx context.Context, req *possessionv1.OpenSessionRequest) (*possessionv1.OpenSessionResponse, error) {
	businessID, err := requireSessionPermission(ctx, "pos:orders:create")
	if err != nil {
		return nil, err
	}

	locationID := req.GetLocationId()
	device := authPkg.DeviceCtx(ctx)
	// The drawer belongs to the authenticated operator; never trust a client
	// supplied staff_id to open a session as another cashier.
	staffID := device.User.UserID
	id, openErr := a.Service.Open(ctx, domain.OpenSessionInput{
		OwnerID:      device.User.UserID,
		BusinessID:   businessID,
		LocationID:   &locationID,
		StaffID:      &staffID,
		OpeningFloat: req.GetOpeningFloat(),
	})
	if openErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to open session")
	}
	return &possessionv1.OpenSessionResponse{SessionId: id, Message: "Session opened successfully"}, nil
}

func (a *SessionAPI) CloseSession(ctx context.Context, req *possessionv1.CloseSessionRequest) (*possessionv1.CloseSessionResponse, error) {
	businessID, err := requireSessionPermission(ctx, "pos:orders:create")
	if err != nil {
		return nil, err
	}

	device := authPkg.DeviceCtx(ctx)
	if closeErr := a.Service.Close(ctx, domain.CloseSessionInput{
		ID:           req.GetId(),
		OwnerID:      device.User.UserID,
		BusinessID:   businessID,
		StaffID:      device.User.UserID,
		ClosingFloat: req.GetClosingFloat(),
	}); closeErr != nil {
		if errors.Is(closeErr, domain.ErrSessionNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Session not found or already closed")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to close session")
	}
	return &possessionv1.CloseSessionResponse{Message: "Session closed successfully"}, nil
}

func requireSessionPermission(ctx context.Context, permission string) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppPOS, permission) {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func toProtoCurrentSession(s domain.Session) *possessionv1.CurrentSession {
	locationID := int64(0)
	if s.LocationID != nil {
		locationID = *s.LocationID
	}
	staffID := int64(0)
	if s.StaffID != nil {
		staffID = *s.StaffID
	}

	p := &possessionv1.CurrentSession{
		Id:          s.ID,
		OwnerId:     s.OwnerID,
		LocationId:  locationID,
		StaffId:     staffID,
		Status:      s.Status,
		OpeningCash: s.OpeningFloat,
		ClosingCash: s.ClosingFloat,
		TotalSales:  s.TotalSales,
		OpenedAt:    s.OpenedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.ClosedAt != nil {
		closedAt := s.ClosedAt.Format("2006-01-02T15:04:05Z")
		p.ClosedAt = &closedAt
	}
	return p
}

func toProtoSession(s domain.Session) *possessionv1.Session {
	p := &possessionv1.Session{
		Id:           s.ID,
		OwnerId:      s.OwnerID,
		LocationId:   s.LocationID,
		StaffId:      s.StaffID,
		OpeningFloat: s.OpeningFloat,
		ClosingFloat: s.ClosingFloat,
		TotalSales:   s.TotalSales,
		Status:       s.Status,
		OpenedAt:     s.OpenedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.ClosedAt != nil {
		closedAt := s.ClosedAt.Format("2006-01-02T15:04:05Z")
		p.ClosedAt = &closedAt
	}
	return p
}
