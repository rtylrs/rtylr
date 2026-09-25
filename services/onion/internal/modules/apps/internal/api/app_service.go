package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/hubapp"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubappv1 "github.com/voxire/voxie/gen/go/hub/app/v1"
)

const addonsCacheScope = "addons"

func invalidateAddonsCache(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(addonsCacheScope, businessID))
}

type AppAPI struct {
	AppsService   *application.Service
	AddonsService hubapp.AddonProvider
}

func NewAppAPI(appsService *application.Service, addonsService hubapp.AddonProvider) *AppAPI {
	return &AppAPI{
		AppsService:   appsService,
		AddonsService: addonsService,
	}
}

func (a *AppAPI) ListApps(ctx context.Context, _ *hubappv1.ListAppsRequest) (*hubappv1.ListAppsResponse, error) {
	var businessID, userID int64
	var role string
	if u, ok := authPkg.GetAuthUser(ctx); ok {
		businessID, userID, role = u.BusinessID, u.UserID, u.Role
	}

	apps, err := a.AppsService.GetVisibleAppsForUser(ctx, businessID, userID, role)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load apps")
	}

	payloads := make([]*hubappv1.App, 0, len(apps))
	for _, app := range apps {
		payloads = append(payloads, toProtoApp(app))
	}
	return &hubappv1.ListAppsResponse{Apps: payloads}, nil
}

func (a *AppAPI) ListAddons(ctx context.Context, _ *hubappv1.ListAddonsRequest) (*hubappv1.ListAddonsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	res, cacheErr := cacher.Of[hubappv1.ListAddonsResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(addonsCacheScope, authUser.BusinessID, "list"),
		func(ctx context.Context) (hubappv1.ListAddonsResponse, error) {
			return a.computeListAddons(ctx, authUser.BusinessID)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return &res, nil
	}

	res, computeErr := a.computeListAddons(ctx, authUser.BusinessID)
	if computeErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list addons")
	}
	return &res, nil
}

func (a *AppAPI) computeListAddons(ctx context.Context, businessID int64) (hubappv1.ListAddonsResponse, error) {
	addons, err := a.AddonsService.ListForBusiness(ctx, businessID)
	if err != nil {
		return hubappv1.ListAddonsResponse{}, err
	}

	payloads := make([]*hubappv1.Addon, 0, len(addons))
	for _, addon := range addons {
		payloads = append(payloads, toProtoAddon(addon))
	}
	return hubappv1.ListAddonsResponse{Addons: payloads}, nil
}

func (a *AppAPI) AddonStatus(ctx context.Context, req *hubappv1.AddonStatusRequest) (*hubappv1.AddonStatusResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	enabled := a.AddonsService.IsEnabled(ctx, authUser.BusinessID, req.GetAddonId())
	return &hubappv1.AddonStatusResponse{IsEnabled: enabled}, nil
}

func (a *AppAPI) EnableAddon(ctx context.Context, req *hubappv1.EnableAddonRequest) (*hubappv1.EnableAddonResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	addonID := req.GetAddonId()
	if addonID == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Addon ID is required")
	}

	if err := a.AddonsService.Enable(ctx, authUser.BusinessID, addonID); err != nil {
		if errors.Is(err, hubapp.ErrAddonNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Addon not found")
		}
		if errors.Is(err, hubapp.ErrAddonNotEntitled) {
			return nil, onekiterrors.New(http.StatusForbidden, "Addon requires an active entitlement")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to enable addon")
	}

	invalidateAddonsCache(ctx, authUser.BusinessID)
	return &hubappv1.EnableAddonResponse{Message: "Addon enabled"}, nil
}

func (a *AppAPI) DisableAddon(ctx context.Context, req *hubappv1.DisableAddonRequest) (*hubappv1.DisableAddonResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	addonID := req.GetAddonId()
	if addonID == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Addon ID is required")
	}

	if err := a.AddonsService.Disable(ctx, authUser.BusinessID, addonID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to disable addon")
	}

	invalidateAddonsCache(ctx, authUser.BusinessID)
	return &hubappv1.DisableAddonResponse{Message: "Addon disabled"}, nil
}

func toProtoAddon(a hubapp.BusinessAddon) *hubappv1.Addon {
	p := &hubappv1.Addon{
		Id:          a.AddonID,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		Price:       a.Price,
		IsFree:      a.IsFree,
		IsEnabled:   a.IsEnabled,
	}
	if a.EnabledAt != nil {
		s := a.EnabledAt.Format("2006-01-02T15:04:05Z07:00")
		p.EnabledAt = &s
	}
	if a.DisabledAt != nil {
		s := a.DisabledAt.Format("2006-01-02T15:04:05Z07:00")
		p.DisabledAt = &s
	}
	return p
}
