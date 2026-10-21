package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/integrations/internal/application"
	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubintegrationv1 "github.com/voxire/voxie/gen/go/hub/integration/v1"
)

type IntegrationAPI struct {
	Service *application.Service
}

func NewIntegrationAPI(service *application.Service) *IntegrationAPI {
	return &IntegrationAPI{Service: service}
}

func (a *IntegrationAPI) ListIntegrations(ctx context.Context, _ *hubintegrationv1.ListIntegrationsRequest) (*hubintegrationv1.ListIntegrationsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	integrations, err := a.Service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list integrations")
	}

	payloads := make([]*hubintegrationv1.Integration, 0, len(integrations))
	for _, i := range integrations {
		ig := &hubintegrationv1.Integration{
			Id:        i.ID,
			Provider:  i.Provider,
			Status:    string(i.Status),
			Config:    configFor(device.User, i.Config),
			CreatedAt: i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if i.ConnectedAt != nil {
			ig.ConnectedAt = i.ConnectedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		payloads = append(payloads, ig)
	}

	return &hubintegrationv1.ListIntegrationsResponse{Integrations: payloads}, nil
}

func (a *IntegrationAPI) ConnectIntegration(ctx context.Context, req *hubintegrationv1.ConnectIntegrationRequest) (*hubintegrationv1.ConnectIntegrationResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !authPkg.IsBusinessAdmin(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	provider := strings.TrimSpace(req.GetProvider())
	if provider == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Provider is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.Service.Connect(ctx, businessID, provider, req.GetConfig())
	if errors.Is(err, domain.ErrInvalidProvider) || errors.Is(err, domain.ErrInvalidConfig) {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to connect integration")
	}

	return &hubintegrationv1.ConnectIntegrationResponse{Id: id, Message: "Integration connected"}, nil
}

func (a *IntegrationAPI) DisconnectIntegration(ctx context.Context, req *hubintegrationv1.DisconnectIntegrationRequest) (*hubintegrationv1.DisconnectIntegrationResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !authPkg.IsBusinessAdmin(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	id := req.GetId()
	if id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Integration id is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.Disconnect(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Integration not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to disconnect integration")
	}

	return &hubintegrationv1.DisconnectIntegrationResponse{Message: "Integration disconnected"}, nil
}

// configFor hides the stored config from ordinary members and masks secrets for
// everyone else.
func configFor(user *authPkg.AuthUser, config json.RawMessage) []byte {
	if !authPkg.IsBusinessAdmin(user) {
		return nil
	}
	return application.RedactConfig(config)
}
