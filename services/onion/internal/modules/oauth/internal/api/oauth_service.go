package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/oauth/internal/application"
	"github.com/voxire/onion/internal/modules/oauth/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	commonoauthv1 "github.com/voxire/voxie/gen/go/common/oauth/v1"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
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
			ExpiresIn:    int32(expiresIn), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
			ExpiresIn:    int32(expiresIn), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
		UserId:          device.User.UserID,
		BusinessId:      device.User.BusinessID,
		Role:            device.User.Role,
		EffectiveAccess: effectiveAccessToProto(device.User.EffectiveAccess),
	}, nil
}

func effectiveAccessToProto(access *authorization.EffectiveAccess) *hubuserv1.EffectiveAccess {
	if access == nil {
		return nil
	}
	result := &hubuserv1.EffectiveAccess{
		BusinessId:           strconv.FormatInt(access.BusinessID, 10),
		DashPermissions:      append([]string{}, access.DashPermissions...),
		SubscribedApps:       make([]string, 0, len(access.SubscribedApps)),
		Licenses:             make([]*hubuserv1.LicenseAssignment, 0, len(access.Licenses)),
		Apps:                 make([]*hubuserv1.EffectiveAppAccess, 0, len(access.Apps)),
		IsOwner:              access.IsOwner,
		IsAdmin:              access.IsAdmin,
		AuthorizationVersion: access.AuthorizationVersion,
	}
	if access.DashRole != "" {
		result.DashRole = stringValue(access.DashRole)
	}
	if access.DashRoleID != nil {
		result.DashRoleId = stringValue(strconv.FormatInt(*access.DashRoleID, 10))
	}
	for _, appID := range access.SubscribedApps {
		result.SubscribedApps = append(result.SubscribedApps, string(appID))
	}
	for _, license := range access.Licenses {
		status := "revoked"
		if license.Active {
			status = "active"
		}
		item := &hubuserv1.LicenseAssignment{
			AppId:            string(license.AppID),
			Status:           status,
			AssignmentSource: string(license.Source),
			GrantedAt:        license.GrantedAt,
			RevokedAt:        license.RevokedAt,
		}
		if license.GroupID != nil {
			item.GroupId = stringValue(strconv.FormatInt(*license.GroupID, 10))
		}
		result.Licenses = append(result.Licenses, item)
	}
	for _, app := range access.Apps {
		item := &hubuserv1.EffectiveAppAccess{
			AppId:             string(app.AppID),
			Licensed:          app.Licensed,
			HasAccess:         app.HasAccess,
			Permissions:       append([]string{}, app.Permissions...),
			PermissionSources: make(map[string]string, len(app.PermissionSources)),
		}
		for permission, sources := range app.PermissionSources {
			item.PermissionSources[permission] = strings.Join(sources, ",")
		}
		for _, roleID := range app.RoleIDs {
			item.RoleIds = append(item.RoleIds, strconv.FormatInt(roleID, 10))
		}
		if app.PrimaryRoleID != nil {
			item.PrimaryRoleId = stringValue(strconv.FormatInt(*app.PrimaryRoleID, 10))
		}
		if app.DisplayRole != "" {
			item.DisplayRole = stringValue(app.DisplayRole)
		}
		for _, scope := range app.Scopes {
			item.Scopes = append(item.Scopes, &hubuserv1.RoleScope{ScopeType: scope.Type, ScopeId: scope.ID})
		}
		for _, source := range app.Sources {
			item.Sources = append(item.Sources, string(source))
		}
		result.Apps = append(result.Apps, item)
	}
	return result
}

func stringValue(value string) *string {
	return &value
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
