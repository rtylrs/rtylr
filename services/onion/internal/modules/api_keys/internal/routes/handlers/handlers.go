package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/api_keys/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubdeveloperv1 "github.com/voxire/voxie/gen/go/hub/developer/v1"
)

// -- List API Keys --

func ListRoutes() []string { return []string{"/api-keys"} }

type ListHandler struct{ Service *application.Service }

type ListRequest struct{}

type ListResponse struct {
	Keys []*hubdeveloperv1.ApiKey `json:"keys"`
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

	keys, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list API keys",
		}
	}

	payloads := make([]*hubdeveloperv1.ApiKey, 0, len(keys))
	for _, k := range keys {
		scopes := k.Scopes
		if scopes == nil {
			scopes = []string{}
		}
		ak := &hubdeveloperv1.ApiKey{
			Id:        k.ID,
			Name:      k.Name,
			KeyPrefix: k.KeyPrefix,
			Scopes:    scopes,
			IsActive:  k.IsActive,
			CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if k.LastUsedAt != nil {
			ak.LastUsedAt = k.LastUsedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if k.ExpiresAt != nil {
			ak.ExpiresAt = k.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		}
		payloads = append(payloads, ak)
	}

	return ListResponse{Keys: payloads}, nil
}

// -- Create API Key --

type CreateHandler struct{ Service *application.Service }

type CreateRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type CreateResponse struct {
	ID      int64  `json:"id"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Name is required",
		}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, rawKey, err := h.Service.Create(ctx, device.User.UserID, businessID, application.CreateAPIKeyInputFromRequest(req.Name, req.Scopes))
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create API key",
		}
	}

	return CreateResponse{ID: id, Key: rawKey, Message: "API key created"}, nil
}

// -- Revoke API Key --

type RevokeHandler struct{ Service *application.Service }

type RevokeRequest struct {
	ID int64 `json:"id"`
}

type RevokeResponse struct {
	Message string `json:"message"`
}

func (h *RevokeHandler) Validate(_ context.Context, req *RevokeRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Key id is required",
		}
	}
	return nil
}

func (h *RevokeHandler) Serve(ctx context.Context, req *RevokeRequest) (RevokeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RevokeResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RevokeResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Revoke(ctx, businessID, req.ID); err != nil {
		return RevokeResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to revoke API key",
		}
	}

	return RevokeResponse{Message: "API key revoked"}, nil
}
