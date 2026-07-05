package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateConfigRoutes() []string {
	return []string{"/config/update"}
}

type UpdateConfigHandler struct {
	ConfigService *application.Service
}

type UpdateConfigRequest struct {
	SiteName               *string `json:"site_name"`
	LogoURL                *string `json:"logo_url"`
	FaviconURL             *string `json:"favicon_url"`
	ContactEmail           *string `json:"contact_email"`
	ContactNumber          *string `json:"contact_number"`
	SenderEmail            *string `json:"sender_email"`
	Website                *string `json:"website"`
	MapsURL                *string `json:"maps_url"`
	Address                *string `json:"address"`
	City                   *string `json:"city"`
	Zipcode                *string `json:"zipcode"`
	Country                *string `json:"country"`
	Instagram              *string `json:"instagram"`
	MaintenanceMode        *bool   `json:"maintenance_mode"`
	MaintenanceNotice      *string `json:"maintenance_notice"`
	AddonsEnabled          *bool   `json:"addons_enabled"`
	EnableReviews          *bool   `json:"enable_reviews"`
	EnableWishlist         *bool   `json:"enable_wishlist"`
	EnableNewsletterSignup *bool   `json:"enable_newsletter_signup"`
	AboutContent           *string `json:"about_content"`
	FooterSubText          *string `json:"footer_sub_text"`
}

type UpdateConfigResponse struct {
	Message string         `json:"message"`
	Config  *ConfigPayload `json:"config,omitempty"`
}

func (h *UpdateConfigHandler) Serve(ctx context.Context, req *UpdateConfigRequest) (UpdateConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateConfigResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateConfigResponse{}, onekiterrors.PermissionDenied()
	}

	ownerID := strconv.FormatInt(businessID, 10)

	current, err := h.ConfigService.GetConfig(ctx, ownerID)
	if err != nil {
		return UpdateConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to load current config")
	}

	input := domain.ConfigInput{
		OwnerId:                ownerID,
		SiteName:               current.SiteName,
		LogoUrl:                current.LogoUrl,
		FaviconUrl:             current.FaviconUrl,
		MetaTags:               current.MetaTags,
		ContactEmail:           current.ContactEmail,
		ContactNumber:          current.ContactNumber,
		SenderEmail:            current.SenderEmail,
		Website:                current.Website,
		MapsURL:                current.MapsURL,
		Address:                current.Address,
		City:                   current.City,
		Zipcode:                current.Zipcode,
		Country:                current.Country,
		Instagram:              current.Instagram,
		MaintenanceMode:        current.MaintenanceMode,
		MaintenanceNotice:      current.MaintenanceNotice,
		AddonsEnabled:          current.AddonsEnabled,
		EnableReviews:          current.EnableReviews,
		EnableWishlist:         current.EnableWishlist,
		EnableNewsletterSignup: current.EnableNewsletterSignup,
		AboutContent:           current.AboutContent,
		FooterSubText:          current.FooterSubText,
	}

	if req.SiteName != nil {
		input.SiteName = *req.SiteName
	}
	if req.LogoURL != nil {
		input.LogoUrl = *req.LogoURL
	}
	if req.FaviconURL != nil {
		input.FaviconUrl = *req.FaviconURL
	}
	if req.ContactEmail != nil {
		input.ContactEmail = *req.ContactEmail
	}
	if req.ContactNumber != nil {
		input.ContactNumber = *req.ContactNumber
	}
	if req.SenderEmail != nil {
		input.SenderEmail = *req.SenderEmail
	}
	if req.Website != nil {
		input.Website = *req.Website
	}
	if req.MapsURL != nil {
		input.MapsURL = *req.MapsURL
	}
	if req.Address != nil {
		input.Address = *req.Address
	}
	if req.City != nil {
		input.City = *req.City
	}
	if req.Zipcode != nil {
		input.Zipcode = *req.Zipcode
	}
	if req.Country != nil {
		input.Country = *req.Country
	}
	if req.Instagram != nil {
		input.Instagram = *req.Instagram
	}
	if req.MaintenanceMode != nil {
		input.MaintenanceMode = *req.MaintenanceMode
	}
	if req.MaintenanceNotice != nil {
		input.MaintenanceNotice = *req.MaintenanceNotice
	}
	if req.AddonsEnabled != nil {
		input.AddonsEnabled = *req.AddonsEnabled
	}
	if req.EnableReviews != nil {
		input.EnableReviews = *req.EnableReviews
	}
	if req.EnableWishlist != nil {
		input.EnableWishlist = *req.EnableWishlist
	}
	if req.EnableNewsletterSignup != nil {
		input.EnableNewsletterSignup = *req.EnableNewsletterSignup
	}
	if req.AboutContent != nil {
		input.AboutContent = *req.AboutContent
	}
	if req.FooterSubText != nil {
		input.FooterSubText = *req.FooterSubText
	}

	if updateErr := h.ConfigService.UpdateConfig(ctx, ownerID, input); updateErr != nil {
		return UpdateConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update config")
	}

	updated := domain.Config{
		OwnerId:                input.OwnerId,
		BusinessID:             businessID,
		SiteName:               input.SiteName,
		LogoUrl:                input.LogoUrl,
		FaviconUrl:             input.FaviconUrl,
		MetaTags:               input.MetaTags,
		ContactEmail:           input.ContactEmail,
		ContactNumber:          input.ContactNumber,
		SenderEmail:            input.SenderEmail,
		Website:                input.Website,
		MapsURL:                input.MapsURL,
		Address:                input.Address,
		City:                   input.City,
		Zipcode:                input.Zipcode,
		Country:                input.Country,
		Instagram:              input.Instagram,
		MaintenanceMode:        input.MaintenanceMode,
		MaintenanceNotice:      input.MaintenanceNotice,
		AddonsEnabled:          input.AddonsEnabled,
		EnableReviews:          input.EnableReviews,
		EnableWishlist:         input.EnableWishlist,
		EnableNewsletterSignup: input.EnableNewsletterSignup,
		AboutContent:           input.AboutContent,
		FooterSubText:          input.FooterSubText,
		BusinessSlug:           current.BusinessSlug,
	}

	return UpdateConfigResponse{
		Message: "Config updated successfully",
		Config:  newConfigPayload(updated),
	}, nil
}
