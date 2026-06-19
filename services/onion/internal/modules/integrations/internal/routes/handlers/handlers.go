package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/integrations/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubintegrationv1 "github.com/voxire/voxie/gen/go/hub/integration/v1"
)

func IntegrationRoutes() []string { return []string{"/integrations"} }

// -- List --

type (
	ListHandler struct{ Service *application.Service }
	ListRequest struct{}
)

type ListResponse struct {
	Integrations []*hubintegrationv1.Integration `json:"integrations"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	integrations, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list integrations",
		}
	}

	payloads := make([]*hubintegrationv1.Integration, 0, len(integrations))
	for _, i := range integrations {
		ig := &hubintegrationv1.Integration{
			Id:        i.ID,
			Provider:  i.Provider,
			Status:    string(i.Status),
			Config:    i.Config,
			CreatedAt: i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if i.ConnectedAt != nil {
			ig.ConnectedAt = i.ConnectedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		payloads = append(payloads, ig)
	}

	return ListResponse{Integrations: payloads}, nil
}

// -- Connect --

type ConnectHandler struct{ Service *application.Service }

type ConnectRequest struct {
	Provider string          `json:"provider"`
	Config   json.RawMessage `json:"config"`
}

type ConnectResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *ConnectHandler) Validate(_ context.Context, req *ConnectRequest) *errorModels.ServerError {
	req.Provider = strings.TrimSpace(req.Provider)
	if req.Provider == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Provider is required",
		}
	}
	return nil
}

func (h *ConnectHandler) Serve(ctx context.Context, req *ConnectRequest) (ConnectResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ConnectResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ConnectResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.Connect(ctx, businessID, req.Provider, req.Config)
	if err != nil {
		return ConnectResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to connect integration",
		}
	}

	return ConnectResponse{ID: id, Message: "Integration connected"}, nil
}

// -- Disconnect --

type DisconnectHandler struct{ Service *application.Service }

type DisconnectRequest struct {
	ID int64 `json:"id"`
}

type DisconnectResponse struct {
	Message string `json:"message"`
}

func (h *DisconnectHandler) Validate(_ context.Context, req *DisconnectRequest) *errorModels.ServerError {
	if req.ID <= 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Integration id is required",
		}
	}
	return nil
}

func (h *DisconnectHandler) Serve(ctx context.Context, req *DisconnectRequest) (DisconnectResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DisconnectResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DisconnectResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Disconnect(ctx, businessID, req.ID); err != nil {
		return DisconnectResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to disconnect integration",
		}
	}

	return DisconnectResponse{Message: "Integration disconnected"}, nil
}
