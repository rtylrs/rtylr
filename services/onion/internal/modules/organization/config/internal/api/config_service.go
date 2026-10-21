package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/config/internal/application"
	"github.com/voxire/onion/internal/modules/organization/config/internal/domain"
	custom_domains "github.com/voxire/onion/internal/modules/organization/custom_domains"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubbusinessv1 "github.com/voxire/voxie/gen/go/hub/business/v1"
	hubconfigv1 "github.com/voxire/voxie/gen/go/hub/config/v1"
)

type ConfigAPI struct {
	ConfigService *application.Service
	CustomDomains *custom_domains.API
}

func NewConfigAPI(configService *application.Service, customDomains *custom_domains.API) *ConfigAPI {
	return &ConfigAPI{ConfigService: configService, CustomDomains: customDomains}
}

func (a *ConfigAPI) GetCustomDomain(ctx context.Context, req *hubconfigv1.GetCustomDomainRequest) (*hubconfigv1.GetCustomDomainResponse, error) {
	return a.CustomDomains.GetCustomDomain(ctx, req)
}

func (a *ConfigAPI) RegisterCustomDomain(ctx context.Context, req *hubconfigv1.RegisterCustomDomainRequest) (*hubconfigv1.RegisterCustomDomainResponse, error) {
	return a.CustomDomains.RegisterCustomDomain(ctx, req)
}

func (a *ConfigAPI) VerifyCustomDomain(ctx context.Context, req *hubconfigv1.VerifyCustomDomainRequest) (*hubconfigv1.VerifyCustomDomainResponse, error) {
	return a.CustomDomains.VerifyCustomDomain(ctx, req)
}

func (a *ConfigAPI) RemoveCustomDomain(ctx context.Context, req *hubconfigv1.RemoveCustomDomainRequest) (*hubconfigv1.RemoveCustomDomainResponse, error) {
	return a.CustomDomains.RemoveCustomDomain(ctx, req)
}

func (a *ConfigAPI) GetConfig(ctx context.Context, _ *hubconfigv1.GetConfigRequest) (*hubconfigv1.GetConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	req := &GetConfigByOwnerRequest{}
	businessID, resolveErr := resolveConfigBusinessID(ctx, device.User, req)
	if resolveErr != nil {
		return nil, resolveErr
	}

	businessIDStr := strconv.FormatInt(businessID, 10)
	config, err := a.ConfigService.GetConfig(ctx, businessIDStr)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get config")
	}

	payload := newConfigPayload(config)
	return &hubconfigv1.GetConfigResponse{
		Business: &hubbusinessv1.Business{
			Id:      businessID,
			Name:    payload.SiteName,
			LogoUrl: payload.LogoURL,
			Address: payload.Address,
			Phone:   payload.ContactNumber,
			Email:   payload.ContactEmail,
			Website: payload.Website,
			Country: payload.Country,
			Slug:    payload.Slug,
		},
		Maintenance: payload.MaintenanceMode,
	}, nil
}

func (a *ConfigAPI) CreateConfig(ctx context.Context, req *hubconfigv1.CreateConfigRequest) (*hubconfigv1.CreateConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !authPkg.IsBusinessAdmin(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	if req.GetKey() == "" || req.GetValue() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Key and value are required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessIDStr := strconv.FormatInt(businessID, 10)
	configInput := domain.ConfigInput{
		OwnerId:      businessIDStr,
		BusinessID:   businessID,
		SiteName:     req.GetKey(),
		ContactEmail: req.GetValue(),
	}

	if err := a.ConfigService.CreateConfig(ctx, configInput); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create config")
	}

	return &hubconfigv1.CreateConfigResponse{Message: "Config created successfully"}, nil
}

func (a *ConfigAPI) UpdateConfig(ctx context.Context, req *hubconfigv1.UpdateConfigRequest) (*hubconfigv1.UpdateConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !authPkg.IsBusinessAdmin(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	ownerID := strconv.FormatInt(businessID, 10)
	current, err := a.ConfigService.GetConfig(ctx, ownerID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load current config")
	}

	input := domain.ConfigInput{
		OwnerId:                   ownerID,
		SiteName:                  current.SiteName,
		LogoUrl:                   current.LogoUrl,
		FaviconUrl:                current.FaviconUrl,
		MetaTags:                  current.MetaTags,
		ContactEmail:              current.ContactEmail,
		ContactNumber:             current.ContactNumber,
		SenderEmail:               current.SenderEmail,
		EmailProvider:             current.EmailProvider,
		SenderName:                current.SenderName,
		ReplyToEmail:              current.ReplyToEmail,
		EmailNotificationsEnabled: current.EmailNotificationsEnabled,
		SMTPHost:                  current.SMTPHost,
		SMTPPort:                  current.SMTPPort,
		SMTPUsername:              current.SMTPUsername,
		Website:                   current.Website,
		MapsURL:                   current.MapsURL,
		Address:                   current.Address,
		City:                      current.City,
		Zipcode:                   current.Zipcode,
		Country:                   current.Country,
		Instagram:                 current.Instagram,
		MaintenanceMode:           current.MaintenanceMode,
		MaintenanceNotice:         current.MaintenanceNotice,
		AddonsEnabled:             current.AddonsEnabled,
		EnableReviews:             current.EnableReviews,
		EnableWishlist:            current.EnableWishlist,
		EnableNewsletterSignup:    current.EnableNewsletterSignup,
		AboutContent:              current.AboutContent,
		FooterSubText:             current.FooterSubText,
	}

	if req.SiteName != nil {
		input.SiteName = *req.SiteName
	}
	if req.LogoUrl != nil {
		input.LogoUrl = *req.LogoUrl
	}
	if req.FaviconUrl != nil {
		input.FaviconUrl = *req.FaviconUrl
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
	if req.EmailProvider != nil {
		input.EmailProvider = strings.TrimSpace(*req.EmailProvider)
		if input.EmailProvider != "resend" && input.EmailProvider != "smtp" {
			return nil, onekiterrors.New(http.StatusBadRequest, "email_provider must be resend or smtp")
		}
	}
	if input.EmailProvider == "" {
		input.EmailProvider = "resend"
	}
	if req.SenderName != nil {
		input.SenderName = strings.TrimSpace(*req.SenderName)
	}
	if req.ReplyToEmail != nil {
		input.ReplyToEmail = strings.TrimSpace(*req.ReplyToEmail)
	}
	if req.EmailNotificationsEnabled != nil {
		input.EmailNotificationsEnabled = *req.EmailNotificationsEnabled
	}
	if req.SmtpHost != nil {
		input.SMTPHost = strings.TrimSpace(*req.SmtpHost)
	}
	if req.SmtpPort != nil {
		if *req.SmtpPort <= 0 || *req.SmtpPort > 65535 {
			return nil, onekiterrors.New(http.StatusBadRequest, "smtp_port must be between 1 and 65535")
		}
		input.SMTPPort = *req.SmtpPort
	}
	if req.SmtpUsername != nil {
		input.SMTPUsername = strings.TrimSpace(*req.SmtpUsername)
	}
	if req.SmtpPassword != nil {
		input.SMTPPassword = *req.SmtpPassword
	}
	if req.Website != nil {
		input.Website = *req.Website
	}
	if req.MapsUrl != nil {
		input.MapsURL = *req.MapsUrl
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

	if updateErr := a.ConfigService.UpdateConfig(ctx, ownerID, input); updateErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update config")
	}

	updated := domain.Config{
		OwnerId:                   input.OwnerId,
		BusinessID:                businessID,
		SiteName:                  input.SiteName,
		LogoUrl:                   input.LogoUrl,
		FaviconUrl:                input.FaviconUrl,
		MetaTags:                  input.MetaTags,
		ContactEmail:              input.ContactEmail,
		ContactNumber:             input.ContactNumber,
		SenderEmail:               input.SenderEmail,
		EmailProvider:             input.EmailProvider,
		SenderName:                input.SenderName,
		ReplyToEmail:              input.ReplyToEmail,
		EmailNotificationsEnabled: input.EmailNotificationsEnabled,
		SMTPHost:                  input.SMTPHost,
		SMTPPort:                  input.SMTPPort,
		SMTPUsername:              input.SMTPUsername,
		SMTPPasswordConfigured:    current.SMTPPasswordConfigured || input.SMTPPassword != "",
		Website:                   input.Website,
		MapsURL:                   input.MapsURL,
		Address:                   input.Address,
		City:                      input.City,
		Zipcode:                   input.Zipcode,
		Country:                   input.Country,
		Instagram:                 input.Instagram,
		MaintenanceMode:           input.MaintenanceMode,
		MaintenanceNotice:         input.MaintenanceNotice,
		AddonsEnabled:             input.AddonsEnabled,
		EnableReviews:             input.EnableReviews,
		EnableWishlist:            input.EnableWishlist,
		EnableNewsletterSignup:    input.EnableNewsletterSignup,
		AboutContent:              input.AboutContent,
		FooterSubText:             input.FooterSubText,
		BusinessSlug:              current.BusinessSlug,
	}

	return &hubconfigv1.UpdateConfigResponse{
		Message: "Config updated successfully",
		Config:  toProtoConfigPayload(newConfigPayload(updated)),
	}, nil
}

func (a *ConfigAPI) SendTestEmail(ctx context.Context, req *hubconfigv1.SendTestEmailRequest) (*hubconfigv1.SendTestEmailResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !authPkg.IsBusinessAdmin(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	if strings.TrimSpace(req.GetRecipient()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Recipient is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	cfg, err := a.ConfigService.GetConfig(ctx, strconv.FormatInt(businessID, 10))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load email settings")
	}
	if !cfg.EmailNotificationsEnabled {
		return nil, onekiterrors.New(http.StatusConflict, "Email notifications are disabled")
	}
	from := cfg.SenderEmail
	if cfg.SenderName != "" && from != "" {
		from = cfg.SenderName + " <" + from + ">"
	}
	if err := a.ConfigService.QueueTestEmail(ctx, application.TestEmail{To: req.GetRecipient(), Subject: "Email settings test", HTML: "<p>Your tenant email settings are working.</p>", From: from, ReplyTo: cfg.ReplyToEmail}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to queue test email")
	}
	return &hubconfigv1.SendTestEmailResponse{Message: "Test email queued", Queued: true}, nil
}

func (a *ConfigAPI) GetIndustries(_ context.Context, _ *hubconfigv1.GetIndustriesRequest) (*hubconfigv1.GetIndustriesResponse, error) {
	groups := make([]*hubconfigv1.IndustryGroup, 0, len(industryGroups))
	for _, group := range industryGroups {
		types := make([]*hubconfigv1.IndustryType, 0, len(group.Types))
		for _, t := range group.Types {
			types = append(types, &hubconfigv1.IndustryType{
				Slug:  t.Slug,
				Label: t.Label,
				Icon:  t.Icon,
			})
		}
		groups = append(groups, &hubconfigv1.IndustryGroup{
			Category: group.Category,
			Types:    types,
		})
	}
	return &hubconfigv1.GetIndustriesResponse{Groups: groups}, nil
}

func toProtoConfigPayload(payload *ConfigPayload) *hubconfigv1.ConfigPayload {
	if payload == nil {
		return nil
	}
	metaTags := payload.MetaTags
	if metaTags == nil {
		metaTags = []string{}
	}
	return &hubconfigv1.ConfigPayload{
		OwnerId:                   payload.OwnerID,
		SiteName:                  payload.SiteName,
		LogoUrl:                   payload.LogoURL,
		FaviconUrl:                payload.FaviconURL,
		MetaTags:                  metaTags,
		ContactEmail:              payload.ContactEmail,
		ContactNumber:             payload.ContactNumber,
		SenderEmail:               payload.SenderEmail,
		EmailProvider:             payload.EmailProvider,
		SenderName:                payload.SenderName,
		ReplyToEmail:              payload.ReplyToEmail,
		EmailNotificationsEnabled: payload.EmailNotificationsEnabled,
		SmtpHost:                  payload.SMTPHost,
		SmtpPort:                  payload.SMTPPort,
		SmtpUsername:              payload.SMTPUsername,
		SmtpPasswordConfigured:    payload.SMTPPasswordConfigured,
		Website:                   payload.Website,
		MapsUrl:                   payload.MapsURL,
		Address:                   payload.Address,
		City:                      payload.City,
		Zipcode:                   payload.Zipcode,
		Country:                   payload.Country,
		Instagram:                 payload.Instagram,
		MaintenanceMode:           payload.MaintenanceMode,
		MaintenanceNotice:         payload.MaintenanceNotice,
		AddonsEnabled:             payload.AddonsEnabled,
		EnableReviews:             payload.EnableReviews,
		EnableWishlist:            payload.EnableWishlist,
		EnableNewsletterSignup:    payload.EnableNewsletterSignup,
		AboutContent:              payload.AboutContent,
		FooterSubText:             payload.FooterSubText,
		Slug:                      payload.Slug,
	}
}
