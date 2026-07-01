package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/oauth/internal/application"
	"github.com/voxire/onion/internal/modules/oauth/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

// TokenRoutes returns the paths for the token endpoint.
func TokenRoutes() []string { return []string{"/oauth/token"} }

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenHandler struct{ Service *application.Service }

func (h *TokenHandler) Validate(_ context.Context, req *TokenRequest) *errorModels.ServerError {
	if req.GrantType != "authorization_code" && req.GrantType != "refresh_token" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "unsupported grant_type",
		}
	}
	if req.ClientID == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "client_id is required",
		}
	}
	return nil
}

func (h *TokenHandler) Serve(ctx context.Context, req *TokenRequest) (TokenResponse, *errorModels.ServerError) {
	switch req.GrantType {
	case "authorization_code":
		at, rt, expiresIn, err := h.Service.ExchangeCode(ctx, req.ClientID, req.Code, req.CodeVerifier, req.RedirectURI)
		if err != nil {
			return TokenResponse{}, tokenError(err)
		}
		return TokenResponse{AccessToken: at, RefreshToken: rt, TokenType: "Bearer", ExpiresIn: expiresIn}, nil

	case "refresh_token":
		at, rt, expiresIn, err := h.Service.RefreshTokens(ctx, req.ClientID, req.RefreshToken)
		if err != nil {
			return TokenResponse{}, tokenError(err)
		}
		return TokenResponse{AccessToken: at, RefreshToken: rt, TokenType: "Bearer", ExpiresIn: expiresIn}, nil

	default:
		return TokenResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "unsupported grant_type",
		}
	}
}

func tokenError(err error) *errorModels.ServerError {
	code := http.StatusBadRequest
	if errors.Is(err, domain.ErrCodeNotFound) ||
		errors.Is(err, domain.ErrTokenNotFound) ||
		errors.Is(err, domain.ErrCodeVerifierFailed) {
		code = http.StatusUnauthorized
	}
	return &errorModels.ServerError{HTTPCode: code, Code: code, Message: err.Error()}
}

// RevokeRoutes returns the paths for the revoke endpoint.
func RevokeRoutes() []string { return []string{"/oauth/revoke"} }

type RevokeRequest struct {
	Token    string `json:"token"`
	ClientID string `json:"client_id"`
}

type RevokeResponse struct {
	OK bool `json:"ok"`
}

type RevokeHandler struct{ Service *application.Service }

func (h *RevokeHandler) Validate(_ context.Context, req *RevokeRequest) *errorModels.ServerError {
	if req.Token == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "token is required",
		}
	}
	return nil
}

func (h *RevokeHandler) Serve(ctx context.Context, req *RevokeRequest) (RevokeResponse, *errorModels.ServerError) {
	if err := h.Service.RevokeToken(ctx, req.Token); err != nil {
		// RFC 7009: always return 200 so token existence is not leaked.
		logger.Ctx(ctx).Warn("oauth token revocation failed", logger.Error(err))
	}
	return RevokeResponse{OK: true}, nil
}

// MeRoutes returns paths for the OAuth user-info endpoint.
func MeRoutes() []string { return []string{"/oauth/me"} }

type MeRequest struct{}

type MeResponse struct {
	UserID     int64  `json:"user_id"`
	BusinessID int64  `json:"business_id,omitempty"`
	Role       string `json:"role,omitempty"`
}

type MeHandler struct{}

func (h *MeHandler) Serve(ctx context.Context, _ *MeRequest) (MeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MeResponse{}, errorModels.PermissionDeniedServerError()
	}
	return MeResponse{
		UserID:     device.User.UserID,
		BusinessID: device.User.BusinessID,
		Role:       device.User.Role,
	}, nil
}
