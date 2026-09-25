package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/application"
	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"
)

type MenuAPI struct {
	Service *application.Service
}

func NewMenuAPI(service *application.Service) *MenuAPI {
	return &MenuAPI{Service: service}
}

func (a *MenuAPI) ListMenu(ctx context.Context, _ *hubmenuv1.ListMenuRequest) (*hubmenuv1.ListMenuResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:view")
	if err != nil {
		return nil, err
	}
	items, err := a.Service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list hub menu")
	}
	return &hubmenuv1.ListMenuResponse{Items: items}, nil
}

func (a *MenuAPI) UpdateMenu(ctx context.Context, req *hubmenuv1.UpdateMenuRequest) (*hubmenuv1.UpdateMenuResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	if err := a.Service.Update(ctx, businessID, req.GetItems()); err != nil {
		if errors.Is(err, application.ErrInvalidMenuItems) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update hub menu")
	}
	return &hubmenuv1.UpdateMenuResponse{Success: true}, nil
}

func (a *MenuAPI) ListMenuBanners(ctx context.Context, _ *hubmenuv1.ListMenuBannersRequest) (*hubmenuv1.ListMenuBannersResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	banners, err := a.Service.ListBanners(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list menu banners")
	}
	payloads := make([]*hubmenuv1.MenuBanner, 0, len(banners))
	for _, banner := range banners {
		payloads = append(payloads, toProtoBanner(banner))
	}
	return &hubmenuv1.ListMenuBannersResponse{Banners: payloads}, nil
}

func (a *MenuAPI) CreateMenuBanner(ctx context.Context, req *hubmenuv1.CreateMenuBannerRequest) (*hubmenuv1.CreateMenuBannerResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	input, err := createBannerInput(req)
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	banner, err := a.Service.CreateBanner(ctx, businessID, input)
	if err != nil {
		return nil, bannerError(err, "Failed to create menu banner")
	}
	return &hubmenuv1.CreateMenuBannerResponse{Banner: toProtoBanner(banner)}, nil
}

func (a *MenuAPI) UpdateMenuBanner(ctx context.Context, req *hubmenuv1.UpdateMenuBannerRequest) (*hubmenuv1.UpdateMenuBannerResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	input, err := updateBannerInput(req)
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	banner, err := a.Service.UpdateBanner(ctx, businessID, req.GetId(), input)
	if err != nil {
		return nil, bannerError(err, "Failed to update menu banner")
	}
	return &hubmenuv1.UpdateMenuBannerResponse{Banner: toProtoBanner(banner)}, nil
}

func (a *MenuAPI) DeleteMenuBanner(ctx context.Context, req *hubmenuv1.DeleteMenuBannerRequest) (*hubmenuv1.DeleteMenuBannerResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteBanner(ctx, businessID, req.GetId()); err != nil {
		return nil, bannerError(err, "Failed to delete menu banner")
	}
	return &hubmenuv1.DeleteMenuBannerResponse{Success: true}, nil
}

func authenticatedBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func requireMenuPermission(ctx context.Context, permission string) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || !authPkg.HasAppPermission(authUser, authorization.AppMenu, permission) {
		return 0, onekiterrors.PermissionDenied()
	}
	return authenticatedBusinessID(ctx)
}

func createBannerInput(req *hubmenuv1.CreateMenuBannerRequest) (domain.BannerInput, error) {
	startsAt, err := parseOptionalTime(req.GetStartsAt(), "starts_at")
	if err != nil {
		return domain.BannerInput{}, err
	}
	endsAt, err := parseOptionalTime(req.GetEndsAt(), "ends_at")
	if err != nil {
		return domain.BannerInput{}, err
	}
	return domain.BannerInput{
		Title:    req.GetTitle(),
		Body:     req.GetBody(),
		ImageURL: req.GetImageUrl(),
		LinkURL:  req.GetLinkUrl(),
		Position: req.GetPosition(),
		IsActive: req.IsActive == nil || req.GetIsActive(),
		StartsAt: startsAt,
		EndsAt:   endsAt,
	}, nil
}

func updateBannerInput(req *hubmenuv1.UpdateMenuBannerRequest) (domain.BannerInput, error) {
	startsAt, err := parseOptionalTime(req.GetStartsAt(), "starts_at")
	if err != nil {
		return domain.BannerInput{}, err
	}
	endsAt, err := parseOptionalTime(req.GetEndsAt(), "ends_at")
	if err != nil {
		return domain.BannerInput{}, err
	}
	return domain.BannerInput{
		Title:    req.GetTitle(),
		Body:     req.GetBody(),
		ImageURL: req.GetImageUrl(),
		LinkURL:  req.GetLinkUrl(),
		Position: req.GetPosition(),
		IsActive: req.GetIsActive(),
		StartsAt: startsAt,
		EndsAt:   endsAt,
	}, nil
}

func parseOptionalTime(value, field string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, errors.New(field + " must be RFC3339")
	}
	return parsed, nil
}

func toProtoBanner(banner domain.Banner) *hubmenuv1.MenuBanner {
	return &hubmenuv1.MenuBanner{
		Id:        banner.ID,
		Title:     banner.Title,
		Body:      banner.Body,
		ImageUrl:  banner.ImageURL,
		LinkUrl:   banner.LinkURL,
		Position:  banner.Position,
		IsActive:  &banner.IsActive,
		StartsAt:  formatOptionalTime(banner.StartsAt),
		EndsAt:    formatOptionalTime(banner.EndsAt),
		CreatedAt: formatOptionalTime(banner.CreatedAt),
		UpdatedAt: formatOptionalTime(banner.UpdatedAt),
	}
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func bannerError(err error, fallback string) error {
	if errors.Is(err, domain.ErrBannerNotFound) {
		return onekiterrors.New(http.StatusNotFound, "Menu banner not found")
	}
	if errors.Is(err, domain.ErrBannerInvalidInput) {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return onekiterrors.New(http.StatusInternalServerError, fallback)
}
