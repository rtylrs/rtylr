package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	"github.com/voxire/onion/internal/modules/business/internal/domain"
	usermodule "github.com/voxire/onion/internal/modules/user"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubbusinessv1 "github.com/voxire/voxie/gen/go/hub/business/v1"
)

type BusinessAPI struct {
	Service     *application.Service
	UserService usermodule.OnboardingPort
}

func NewBusinessAPI(service *application.Service) *BusinessAPI {
	return &BusinessAPI{Service: service}
}

func (a *BusinessAPI) WithUserService(userService usermodule.OnboardingPort) *BusinessAPI {
	a.UserService = userService
	return a
}

func (a *BusinessAPI) GetBusiness(ctx context.Context, req *hubbusinessv1.GetBusinessRequest) (*hubbusinessv1.GetBusinessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	profileScopeID := req.GetBusinessId()
	if profileScopeID > 0 {
		// An explicit business id is attacker-controlled: verify membership
		// before reading another tenant's profile.
		if _, err := tenant.ResolveBusinessID(ctx, profileScopeID, device.User.UserID); err != nil {
			return nil, onekiterrors.PermissionDenied()
		}
	} else {
		profileScopeID = device.User.BusinessID
	}
	if profileScopeID == 0 {
		profileScopeID = device.User.UserID
	}

	profile, err := a.Service.GetProfile(ctx, profileScopeID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get business profile")
	}

	return &hubbusinessv1.GetBusinessResponse{
		Business: &hubbusinessv1.Business{
			Id:       profile.ID,
			Name:     profile.Name,
			LogoUrl:  profile.LogoURL,
			Address:  profile.Address,
			Phone:    profile.ContactPhone,
			Email:    profile.ContactEmail,
			Timezone: profile.Timezone,
			Currency: profile.Currency,
			Plan:     profile.Plan,
		},
	}, nil
}

func (a *BusinessAPI) LookupBusinessBySlug(ctx context.Context, req *hubbusinessv1.LookupBusinessBySlugRequest) (*hubbusinessv1.LookupBusinessBySlugResponse, error) {
	slug := strings.ToLower(strings.TrimSpace(req.GetSlug()))
	res, err := a.Service.LookupBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, onekiterrors.New(http.StatusNotFound, "Business not found")
		}
		logger.Ctx(ctx).Error("slug lookup failed", logger.String("slug", slug), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to lookup business")
	}

	locations := res.Locations
	if locations == nil {
		locations = []domain.PublicLocation{}
	}
	protoLocations := make([]*hubbusinessv1.PublicLocation, 0, len(locations))
	for _, loc := range locations {
		protoLocations = append(protoLocations, &hubbusinessv1.PublicLocation{
			Id:       loc.ID,
			Name:     loc.Name,
			Slug:     loc.Slug,
			Address:  loc.Address,
			Timezone: loc.Timezone,
			IsActive: loc.IsActive,
		})
	}

	return &hubbusinessv1.LookupBusinessBySlugResponse{
		Business: &hubbusinessv1.PublicBusinessSummary{
			Id:       res.Business.ID,
			Name:     res.Business.Name,
			Slug:     res.Business.Slug,
			LogoUrl:  res.Business.LogoURL,
			Currency: res.Business.Currency,
			Timezone: res.Business.Timezone,
		},
		Locations: protoLocations,
	}, nil
}

func (a *BusinessAPI) GetPublicBusiness(ctx context.Context, req *hubbusinessv1.GetPublicBusinessRequest) (*hubbusinessv1.GetPublicBusinessResponse, error) {
	res, err := a.Service.GetPublicBusiness(ctx, req.GetBusinessId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, onekiterrors.New(http.StatusNotFound, "Business not found")
		}
		logger.Ctx(ctx).Error("public business lookup failed",
			logger.Any("business_id", req.GetBusinessId()),
			logger.Error(err),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load business")
	}
	return &hubbusinessv1.GetPublicBusinessResponse{
		Business: &hubbusinessv1.PublicBusinessProfile{
			Id:       res.ID,
			Name:     res.Name,
			LogoUrl:  res.LogoURL,
			Currency: res.Currency,
			Timezone: res.Timezone,
		},
	}, nil
}

func (a *BusinessAPI) SetSlug(ctx context.Context, req *hubbusinessv1.SetSlugRequest) (*hubbusinessv1.SetSlugResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID := device.User.BusinessID
	if businessID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "No business attached to this user")
	}

	slug := strings.ToLower(strings.TrimSpace(req.GetSlug()))
	if err := a.Service.UpdateSlug(ctx, businessID, device.User.UserID, slug); err != nil {
		if errors.Is(err, application.ErrBusinessNotOwned) {
			return nil, onekiterrors.PermissionDenied()
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, onekiterrors.New(http.StatusConflict, "That slug is taken.")
		}
		logger.Ctx(ctx).Error("set slug failed",
			logger.Any("business_id", businessID),
			logger.String("slug", slug),
			logger.Error(err),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update slug")
	}

	return &hubbusinessv1.SetSlugResponse{Slug: slug}, nil
}

func (a *BusinessAPI) UpdateBusiness(_ context.Context, _ *hubbusinessv1.UpdateBusinessRequest) (*hubbusinessv1.UpdateBusinessResponse, error) {
	return nil, onekiterrors.New(http.StatusNotImplemented, "Update business is not implemented")
}

func (a *BusinessAPI) RenameBusiness(ctx context.Context, req *hubbusinessv1.RenameBusinessRequest) (*hubbusinessv1.RenameBusinessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if len(name) > 255 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name must be 255 characters or fewer")
	}

	if err := a.Service.RenameBusiness(ctx, req.GetBusinessId(), device.User.UserID, name); err != nil {
		switch {
		case errors.Is(err, application.ErrBusinessNameEmpty):
			return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
		case errors.Is(err, application.ErrBusinessNotOwned):
			return nil, onekiterrors.PermissionDenied()
		case strings.Contains(strings.ToLower(err.Error()), "duplicate"):
			return nil, onekiterrors.New(http.StatusConflict, "That business name is already taken")
		default:
			logger.Ctx(ctx).Error("rename business failed", logger.Error(err))
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to rename business")
		}
	}

	return &hubbusinessv1.RenameBusinessResponse{
		Business: &hubbusinessv1.Business{Id: req.GetBusinessId(), Name: name},
	}, nil
}

func (a *BusinessAPI) SetLogo(ctx context.Context, req *hubbusinessv1.SetLogoRequest) (*hubbusinessv1.SetLogoResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	logoURL := strings.TrimSpace(req.GetLogoUrl())
	if logoURL == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Logo URL is required")
	}
	if len(logoURL) > 512 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Logo URL must be 512 characters or fewer")
	}

	if err := a.Service.SetLogo(ctx, req.GetBusinessId(), device.User.UserID, logoURL); err != nil {
		switch {
		case errors.Is(err, application.ErrLogoURLEmpty):
			return nil, onekiterrors.New(http.StatusBadRequest, "Logo URL is required")
		case errors.Is(err, application.ErrBusinessNotOwned):
			return nil, onekiterrors.PermissionDenied()
		default:
			logger.Ctx(ctx).Error("set business logo failed", logger.Error(err))
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to set business logo")
		}
	}

	return &hubbusinessv1.SetLogoResponse{
		Business: &hubbusinessv1.Business{Id: req.GetBusinessId(), LogoUrl: logoURL},
	}, nil
}

func (a *BusinessAPI) DeleteBusiness(ctx context.Context, req *hubbusinessv1.DeleteBusinessRequest) (*hubbusinessv1.DeleteBusinessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteBusiness(ctx, req.GetBusinessId(), device.User.UserID); err != nil {
		switch {
		case errors.Is(err, application.ErrCannotDeleteOnlyBusiness):
			return nil, onekiterrors.New(http.StatusConflict, "You cannot delete your only remaining business")
		case errors.Is(err, application.ErrBusinessNotOwned):
			return nil, onekiterrors.PermissionDenied()
		default:
			logger.Ctx(ctx).Error("delete business failed", logger.Error(err))
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete business")
		}
	}

	return &hubbusinessv1.DeleteBusinessResponse{Message: "Business deleted"}, nil
}

func (a *BusinessAPI) ListMyBusinesses(ctx context.Context, _ *hubbusinessv1.ListMyBusinessesRequest) (*hubbusinessv1.ListMyBusinessesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businesses, err := a.Service.ListUserBusinesses(ctx, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list businesses")
	}

	entries := make([]*hubbusinessv1.BusinessEntry, 0, len(businesses))
	for _, b := range businesses {
		entries = append(entries, &hubbusinessv1.BusinessEntry{
			Id:       b.ID,
			Name:     b.Name,
			LogoUrl:  b.LogoURL,
			ParentId: b.ParentID,
			IsActive: b.IsActive,
			IsOwned:  b.IsOwned,
		})
	}

	return &hubbusinessv1.ListMyBusinessesResponse{Businesses: entries}, nil
}

func (a *BusinessAPI) CreateSubBusiness(ctx context.Context, req *hubbusinessv1.CreateSubBusinessRequest) (*hubbusinessv1.CreateSubBusinessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	parentID := req.GetParentBusinessId()
	if parentID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Parent business id is required")
	}

	id, err := a.Service.CreateSubBusiness(ctx, name, device.User.UserID, parentID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "Forbidden") {
			return nil, onekiterrors.PermissionDenied()
		}
		if strings.Contains(err.Error(), "not found") {
			return nil, onekiterrors.New(http.StatusNotFound, "Parent business not found")
		}
		logger.Ctx(ctx).Error("create sub-business failed", logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create sub-business")
	}

	return &hubbusinessv1.CreateSubBusinessResponse{Id: id}, nil
}

func (a *BusinessAPI) OnboardBusiness(ctx context.Context, req *hubbusinessv1.OnboardBusinessRequest) (*hubbusinessv1.OnboardBusinessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if a.UserService == nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Onboarding is not configured")
	}

	name := strings.TrimSpace(req.GetName())
	token, err := a.UserService.OnboardCreateBusiness(ctx, device.User.UserID, name)
	if err != nil {
		switch {
		case errors.Is(err, usermodule.ErrBusinessNameRequired):
			return nil, onekiterrors.New(http.StatusBadRequest, "Business name is required")
		case errors.Is(err, usermodule.ErrBusinessNameTaken):
			return nil, onekiterrors.New(http.StatusConflict, "That business name is already taken")
		default:
			logger.Ctx(ctx).Error("onboard business failed", logger.Error(err))
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create business")
		}
	}

	return &hubbusinessv1.OnboardBusinessResponse{Token: token}, nil
}

func (a *BusinessAPI) CheckBusinessNameExists(ctx context.Context, req *hubbusinessv1.CheckBusinessNameExistsRequest) (*hubbusinessv1.CheckBusinessNameExistsResponse, error) {
	if a.UserService == nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Business name check is not configured")
	}

	exists, err := a.UserService.BusinessNameExists(ctx, req.GetName())
	if err != nil {
		logger.Ctx(ctx).Error("business name exists check failed", logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to check business name")
	}

	return &hubbusinessv1.CheckBusinessNameExistsResponse{Exists: exists}, nil
}
