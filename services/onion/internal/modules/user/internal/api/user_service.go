package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

type UserAPI struct {
	UserService       *application.Service
	ProfileResolver   hubprofile.Resolver
	ProvisioningStore authPkg.ProvisioningStateStore
}

func NewUserAPI(userService *application.Service, profileResolver hubprofile.Resolver, provisioningStores ...authPkg.ProvisioningStateStore) *UserAPI {
	var provisioningStore authPkg.ProvisioningStateStore
	if len(provisioningStores) > 0 {
		provisioningStore = provisioningStores[0]
	}
	return &UserAPI{
		UserService:       userService,
		ProfileResolver:   profileResolver,
		ProvisioningStore: provisioningStore,
	}
}

func (a *UserAPI) GetMe(ctx context.Context, _ *hubuserv1.GetMeRequest) (*hubuserv1.GetMeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if a.ProfileResolver == nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Profile resolver not configured")
	}

	profile, err := a.ProfileResolver.GetProfile(ctx, device.User.UserID, device.User.BusinessID)
	if err != nil {
		if errors.Is(err, hubprofile.ErrUserNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "User not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get user info")
	}

	return profileToGetMeResponse(profile), nil
}

func (a *UserAPI) UpdateProfile(ctx context.Context, req *hubuserv1.UpdateProfileRequest) (*hubuserv1.UpdateProfileResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	user, err := a.UserService.UpdateProfile(ctx, device.User.UserID, req.GetName())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update profile")
	}

	if a.ProfileResolver != nil {
		profile, profileErr := a.ProfileResolver.GetProfile(ctx, device.User.UserID, device.User.BusinessID)
		if profileErr == nil {
			profile.FullName = user.FullName
			return &hubuserv1.UpdateProfileResponse{User: profileToGetMeResponse(profile)}, nil
		}
	}

	return &hubuserv1.UpdateProfileResponse{
		User: &hubuserv1.GetMeResponse{
			Id:             user.ID,
			Email:          user.Email,
			FullName:       user.FullName,
			IsConsoleAdmin: user.IsConsoleAdmin,
		},
	}, nil
}

func (a *UserAPI) ChangePassword(ctx context.Context, req *hubuserv1.ChangePasswordRequest) (*hubuserv1.ChangePasswordResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	err := a.UserService.ChangePassword(ctx, device.User.UserID, req.GetCurrentPassword(), req.GetNewPassword())
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return nil, onekiterrors.New(http.StatusUnauthorized, "Current password is incorrect")
		}
		return nil, onekiterrors.New(http.StatusBadRequest, "Failed to change password")
	}
	if a.ProvisioningStore != nil {
		if err := a.ProvisioningStore.MarkPasswordChanged(ctx, device.User.UserID); err != nil {
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update account activation")
		}
	}

	return &hubuserv1.ChangePasswordResponse{}, nil
}

func (a *UserAPI) SwitchBusiness(ctx context.Context, req *hubuserv1.SwitchBusinessRequest) (*hubuserv1.SwitchBusinessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID := req.GetBusinessId()
	if businessID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "business_id is required")
	}

	token, err := a.UserService.SwitchBusiness(ctx, device.User.UserID, businessID)
	if err != nil {
		if errors.Is(err, application.ErrBusinessAccessDenied) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to switch business")
	}

	return &hubuserv1.SwitchBusinessResponse{Token: token}, nil
}

func profileToGetMeResponse(profile hubprofile.Profile) *hubuserv1.GetMeResponse {
	apps := profile.Apps
	if apps == nil {
		apps = []string{}
	}

	dashRole := profile.Role
	if profile.EffectiveAccess != nil && profile.EffectiveAccess.DashRole != "" {
		dashRole = profile.EffectiveAccess.DashRole
	}
	response := &hubuserv1.GetMeResponse{
		Id:                 profile.ID,
		Email:              profile.Email,
		FullName:           profile.FullName,
		Apps:               apps,
		Role:               profile.Role,
		BusinessId:         profile.BusinessID,
		BusinessName:       profile.BusinessName,
		IsOwner:            profile.IsOwner,
		IsConsoleAdmin:     profile.IsConsoleAdmin,
		Onboarded:          profile.Onboarded,
		EffectiveAccess:    effectiveAccessToProto(profile.EffectiveAccess),
		ProvisioningStatus: provisioningStatusToProto(profile.ProvisioningStatus),
	}
	if dashRole != "" {
		response.DashRole = stringPointer(dashRole)
	}
	return response
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
		result.DashRole = stringPointer(access.DashRole)
	}
	if access.DashRoleID != nil {
		value := strconv.FormatInt(*access.DashRoleID, 10)
		result.DashRoleId = &value
	}
	for _, appID := range access.SubscribedApps {
		result.SubscribedApps = append(result.SubscribedApps, string(appID))
	}
	for _, license := range access.Licenses {
		status := "revoked"
		if license.Active {
			status = "active"
		}
		result.Licenses = append(result.Licenses, &hubuserv1.LicenseAssignment{
			AppId:            string(license.AppID),
			Status:           status,
			AssignmentSource: string(license.Source),
			GrantedAt:        license.GrantedAt,
			RevokedAt:        license.RevokedAt,
		})
		if license.GroupID != nil {
			result.Licenses[len(result.Licenses)-1].GroupId = stringPointer(strconv.FormatInt(*license.GroupID, 10))
		}
	}
	for _, app := range access.Apps {
		appResult := &hubuserv1.EffectiveAppAccess{
			AppId:       string(app.AppID),
			Licensed:    app.Licensed,
			HasAccess:   app.HasAccess,
			Permissions: append([]string{}, app.Permissions...),
		}
		appResult.PermissionSources = make(map[string]string, len(app.PermissionSources))
		for permission, sources := range app.PermissionSources {
			appResult.PermissionSources[permission] = strings.Join(sources, ",")
		}
		for _, roleID := range app.RoleIDs {
			appResult.RoleIds = append(appResult.RoleIds, strconv.FormatInt(roleID, 10))
		}
		if app.PrimaryRoleID != nil {
			value := strconv.FormatInt(*app.PrimaryRoleID, 10)
			appResult.PrimaryRoleId = &value
		}
		if app.DisplayRole != "" {
			appResult.DisplayRole = stringPointer(app.DisplayRole)
		}
		for _, scope := range app.Scopes {
			appResult.Scopes = append(appResult.Scopes, &hubuserv1.RoleScope{
				ScopeType: scope.Type,
				ScopeId:   scope.ID,
			})
		}
		for _, source := range app.Sources {
			appResult.Sources = append(appResult.Sources, string(source))
		}
		result.Apps = append(result.Apps, appResult)
	}
	return result
}

func stringPointer(value string) *string {
	return &value
}

func provisioningStatusToProto(status *authorization.ProvisioningStatus) *hubuserv1.ProvisioningStatus {
	if status == nil {
		return nil
	}
	return &hubuserv1.ProvisioningStatus{
		Mode:                   stringPointer(string(status.Mode)),
		State:                  stringPointer(string(status.State)),
		EmailVerified:          boolPointer(status.EmailVerified),
		PasswordChangeRequired: boolPointer(status.PasswordChangeRequired),
		MfaRequired:            boolPointer(status.MFARequired),
		MfaEnrolled:            boolPointer(status.MFAEnrolled),
		MfaVerified:            boolPointer(status.MFAVerified),
		RecoveryAvailable:      boolPointer(status.RecoveryAvailable),
	}
}

func boolPointer(value bool) *bool {
	return &value
}
