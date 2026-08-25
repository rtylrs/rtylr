package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/voxire/onion/internal/modules/oauth/internal/application"
	"github.com/voxire/onion/internal/modules/oauth/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	commonoauthv1 "github.com/voxire/voxie/gen/go/common/oauth/v1"
)

type OauthAPI struct {
	Service *application.Service
}

func NewOauthAPI(service *application.Service) *OauthAPI {
	return &OauthAPI{Service: service}
}

func (a *OauthAPI) Issue(ctx context.Context, req *commonoauthv1.IssueRequest) (*commonoauthv1.IssueResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if _, err := a.Service.ValidateAuthRequest(ctx, req.GetClientId(), req.GetRedirectUri(), req.GetScope()); err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	method := req.GetCodeChallengeMethod()
	if method == "" {
		method = "S256"
	}

	code, err := a.Service.IssueAuthCode(
		ctx,
		req.GetClientId(),
		device.User.UserID,
		req.GetRedirectUri(),
		req.GetScope(),
		req.GetCodeChallenge(),
		method,
	)
	if err != nil {
		logger.Ctx(ctx).Error("failed to issue auth code")
		if errors.Is(err, domain.ErrPKCERequired) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "failed to issue authorization code")
	}

	callbackURL, err := url.Parse(req.GetRedirectUri())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "invalid redirect_uri")
	}

	cq := callbackURL.Query()
	cq.Set("code", code)
	if req.GetState() != "" {
		cq.Set("state", req.GetState())
	}
	cq.Set("iss", "rtylr")
	callbackURL.RawQuery = cq.Encode()

	return &commonoauthv1.IssueResponse{RedirectUrl: callbackURL.String()}, nil
}

func (a *OauthAPI) Token(ctx context.Context, req *commonoauthv1.TokenRequest) (*commonoauthv1.TokenResponse, error) {
	grantType := req.GetGrantType()
	if grantType != "authorization_code" && grantType != "refresh_token" {
		return nil, oauthStatusError(http.StatusBadRequest, "unsupported grant_type")
	}
	if req.GetClientId() == "" {
		return nil, oauthStatusError(http.StatusBadRequest, "client_id is required")
	}

	switch grantType {
	case "authorization_code":
		at, rt, expiresIn, err := a.Service.ExchangeCode(ctx, req.GetClientId(), req.GetCode(), req.GetCodeVerifier(), req.GetRedirectUri())
		if err != nil {
			return nil, oauthTokenError(err)
		}
		return &commonoauthv1.TokenResponse{
			AccessToken:  at,
			RefreshToken: rt,
			TokenType:    "Bearer",
			ExpiresIn:    int32(expiresIn),
		}, nil
	case "refresh_token":
		at, rt, expiresIn, err := a.Service.RefreshTokens(ctx, req.GetClientId(), req.GetRefreshToken())
		if err != nil {
			return nil, oauthTokenError(err)
		}
		return &commonoauthv1.TokenResponse{
			AccessToken:  at,
			RefreshToken: rt,
			TokenType:    "Bearer",
			ExpiresIn:    int32(expiresIn),
		}, nil
	default:
		return nil, oauthStatusError(http.StatusBadRequest, "unsupported grant_type")
	}
}

func (a *OauthAPI) Revoke(ctx context.Context, req *commonoauthv1.RevokeRequest) (*commonoauthv1.RevokeResponse, error) {
	if req.GetToken() == "" {
		return nil, oauthStatusError(http.StatusBadRequest, "token is required")
	}
	if err := a.Service.RevokeToken(ctx, req.GetToken()); err != nil {
		logger.Ctx(ctx).Warn("oauth token revocation failed", logger.Error(err))
	}
	return &commonoauthv1.RevokeResponse{Ok: true}, nil
}

func (a *OauthAPI) Me(ctx context.Context, _ *commonoauthv1.MeRequest) (*commonoauthv1.MeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	return &commonoauthv1.MeResponse{
		UserId:     device.User.UserID,
		BusinessId: device.User.BusinessID,
		Role:       device.User.Role,
	}, nil
}

func oauthTokenError(err error) error {
	code := http.StatusBadRequest
	if errors.Is(err, domain.ErrCodeNotFound) ||
		errors.Is(err, domain.ErrTokenNotFound) ||
		errors.Is(err, domain.ErrCodeVerifierFailed) ||
		errors.Is(err, domain.ErrClientAuthenticationRequired) ||
		errors.Is(err, domain.ErrClientNotFound) ||
		errors.Is(err, domain.ErrUserNotFound) {
		code = http.StatusUnauthorized
	}
	return onekiterrors.New(code, err.Error())
}

func oauthStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
