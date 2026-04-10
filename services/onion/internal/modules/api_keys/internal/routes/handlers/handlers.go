package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/api_keys/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// -- List API Keys --

func ListRoutes() []string { return []string{"/api-keys"} }

type ListHandler struct{ Service *application.Service }

type ListRequest struct{}

type APIKeyPayload struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"keyPrefix"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	IsActive   bool       `json:"isActive"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type ListResponse struct {
	Keys []APIKeyPayload `json:"keys"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	keys, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list API keys",
		}
	}

	payloads := make([]APIKeyPayload, 0, len(keys))
	for _, k := range keys {
		scopes := k.Scopes
		if scopes == nil {
			scopes = []string{}
		}
		payloads = append(payloads, APIKeyPayload{
			ID:         k.ID,
			Name:       k.Name,
			KeyPrefix:  k.KeyPrefix,
			Scopes:     scopes,
			LastUsedAt: k.LastUsedAt,
			ExpiresAt:  k.ExpiresAt,
			IsActive:   k.IsActive,
			CreatedAt:  k.CreatedAt,
		})
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

	id, rawKey, err := h.Service.Create(ctx, device.User.UserID, application.CreateAPIKeyInputFromRequest(req.Name, req.Scopes))
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

	if err := h.Service.Revoke(ctx, device.User.UserID, req.ID); err != nil {
		return RevokeResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to revoke API key",
		}
	}

	return RevokeResponse{Message: "API key revoked"}, nil
}
