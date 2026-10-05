package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos/terminals/internal/application"
	"github.com/voxire/onion/internal/modules/pos/terminals/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posterminalv1 "github.com/voxire/voxie/gen/go/pos/terminal/v1"
)

type TerminalAPI struct {
	Service *application.Service
}

func NewTerminalAPI(service *application.Service) *TerminalAPI {
	return &TerminalAPI{Service: service}
}

func (a *TerminalAPI) ListTerminals(ctx context.Context, _ *posterminalv1.ListTerminalsRequest) (*posterminalv1.ListTerminalsResponse, error) {
	user, err := requireTerminalPermission(ctx, "pos:terminals:manage")
	if err != nil {
		return nil, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	terminals, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list terminals")
	}

	payloads := make([]*posterminalv1.Terminal, len(terminals))
	for i, t := range terminals {
		payloads[i] = toProtoTerminal(t)
	}
	return &posterminalv1.ListTerminalsResponse{Terminals: payloads}, nil
}

func (a *TerminalAPI) CreateTerminal(ctx context.Context, req *posterminalv1.CreateTerminalRequest) (*posterminalv1.CreateTerminalResponse, error) {
	user, err := requireTerminalPermission(ctx, "pos:terminals:manage")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if req.GetBranchId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Branch ID is required")
	}

	deviceID := req.GetDeviceId()
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, createErr := a.Service.Create(ctx, domain.TerminalInput{
		BusinessID: businessID,
		BranchID:   req.GetBranchId(),
		Name:       req.GetName(),
		DeviceID:   &deviceID,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create terminal")
	}
	return &posterminalv1.CreateTerminalResponse{TerminalId: id, Message: "Terminal created successfully"}, nil
}

func (a *TerminalAPI) UpdateTerminal(ctx context.Context, req *posterminalv1.UpdateTerminalRequest) (*posterminalv1.UpdateTerminalResponse, error) {
	user, err := requireTerminalPermission(ctx, "pos:terminals:manage")
	if err != nil {
		return nil, err
	}

	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	deviceID := body.GetDeviceId()
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if updateErr := a.Service.Update(ctx, domain.TerminalUpdateInput{
		ID:         req.GetId(),
		BusinessID: businessID,
		BranchID:   body.GetBranchId(),
		Name:       body.GetName(),
		DeviceID:   &deviceID,
		IsActive:   body.GetIsActive(),
	}); updateErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update terminal")
	}
	return &posterminalv1.UpdateTerminalResponse{Message: "Terminal updated successfully"}, nil
}

func (a *TerminalAPI) HeartbeatTerminal(ctx context.Context, req *posterminalv1.HeartbeatTerminalRequest) (*posterminalv1.HeartbeatTerminalResponse, error) {
	user, err := requireTerminalPermission(ctx, "pos:terminals:manage")
	if err != nil {
		return nil, err
	}
	if req.GetTerminalId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Terminal ID is required")
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if heartbeatErr := a.Service.Heartbeat(ctx, req.GetTerminalId(), businessID); heartbeatErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update heartbeat")
	}
	return &posterminalv1.HeartbeatTerminalResponse{Message: "Heartbeat updated successfully"}, nil
}

func requireTerminalPermission(ctx context.Context, permission string) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppPOS, permission) || !authPkg.HasAnyAppScope(device.User, authorization.AppPOS) {
		return nil, onekiterrors.PermissionDenied()
	}
	return device.User, nil
}

func toProtoTerminal(t domain.Terminal) *posterminalv1.Terminal {
	p := &posterminalv1.Terminal{
		Id:         t.ID,
		BusinessId: t.BusinessID,
		BranchId:   t.BranchID,
		Name:       t.Name,
		IsActive:   t.IsActive,
		CreatedAt:  t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if t.DeviceID != nil {
		p.DeviceId = t.DeviceID
	}
	if t.LastSeenAt != nil {
		s := t.LastSeenAt.Format("2006-01-02T15:04:05Z")
		p.LastSeenAt = &s
	}
	return p
}
