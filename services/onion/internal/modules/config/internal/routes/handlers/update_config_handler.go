package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/config/internal/application"
	"github.com/voxire/onion/internal/modules/config/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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
	Message string `json:"message"`
}

func (h *UpdateConfigHandler) Serve(ctx context.Context, req *UpdateConfigRequest) (UpdateConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateConfigResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateConfigResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := strconv.FormatInt(businessID, 10)

	current, err := h.ConfigService.GetConfig(ctx, ownerID)
	if err != nil {
		return UpdateConfigResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load current config",
		}
	}

	input := domain.ConfigInput{
		OwnerId:                ownerID,
		SiteName:               current.SiteName,
		LogoUrl:                current.LogoUrl,
		FaviconUrl:             current.FaviconUrl,
		MetaTags:               current.MetaTags,
		ContactEmail:           current.ContactEmail,
		ContactNumber:          current.ContactNumber,
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
		return UpdateConfigResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update config",
		}
	}

	return UpdateConfigResponse{Message: "Config updated successfully"}, nil
}
