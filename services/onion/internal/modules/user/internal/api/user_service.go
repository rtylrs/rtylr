package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	authorizationtransport "github.com/voxire/onion/internal/platform/authorization/transport"
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
	response := &hubuserv1.GetMeResponse{
		Id:                 profile.ID,
		Email:              profile.Email,
		FullName:           profile.FullName,
		BusinessId:         profile.BusinessID,
		BusinessName:       profile.BusinessName,
		IsConsoleAdmin:     profile.IsConsoleAdmin,
		Onboarded:          profile.Onboarded,
		EffectiveAccess:    authorizationtransport.EffectiveAccessToProto(profile.EffectiveAccess),
		ProvisioningStatus: authorizationtransport.ProvisioningStatusToProto(profile.ProvisioningStatus),
	}
	return response
}

func stringPointer(value string) *string {
	return &value
}
