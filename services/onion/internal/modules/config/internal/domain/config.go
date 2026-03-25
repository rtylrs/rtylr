package domain

import (
	"errors"

	sqlc "github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	configv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/config/v1"
	"github.com/voxire/onion/internal/platform/utils/convert"
	"google.golang.org/protobuf/proto"
)

var ErrConfigNotFound = errors.New("config not found")

type Config struct {
	OwnerId                string
	SiteName               string
	LogoUrl                string
	FaviconUrl             string
	MetaTags               []string
	ContactEmail           string
	ContactNumber          string
	Instagram              string
	MaintenanceMode        bool
	MaintenanceNotice      string
	AddonsEnabled          bool
	EnableReviews          bool
	EnableWishlist         bool
	EnableNewsletterSignup bool
	AboutContent           string
	FooterSubText          string
}

func (c *Config) FromRows(row sqlc.GetConfigRow) {
	c.OwnerId = convert.Int32ToString(row.OwnerID)
	c.SiteName = convert.ToString(row.SiteName)
	c.LogoUrl = convert.ToString(row.LogoUrl)
	c.FaviconUrl = convert.ToString(row.FaviconUrl)
	c.MetaTags = convert.ToStringSlice(row.MetaTags)
	c.ContactEmail = row.ContactEmail
	c.ContactNumber = row.ContactNumber
	c.Instagram = convert.ToString(row.Instagram)
	c.MaintenanceMode = convert.ToBool(row.MaintenanceMode)
	c.MaintenanceNotice = convert.ToString(row.MaintenanceNotice)
	c.AddonsEnabled = convert.ToBool(row.AddonsEnabled)
	c.EnableReviews = convert.ToBool(row.EnableReviews)
	c.EnableWishlist = convert.ToBool(row.EnableWishlist)
	c.EnableNewsletterSignup = convert.ToBool(row.EnableNewsletterSignup)
	c.AboutContent = row.AboutContent
	c.FooterSubText = convert.ToString(row.FooterSubText)
}

type ConfigInput struct {
	OwnerId                string
	SiteName               string
	LogoUrl                string
	FaviconUrl             string
	MetaTags               []string
	ContactEmail           string
	ContactNumber          string
	Instagram              string
	MaintenanceMode        bool
	MaintenanceNotice      string
	AddonsEnabled          bool
	EnableReviews          bool
	EnableWishlist         bool
	EnableNewsletterSignup bool
	AboutContent           string
	FooterSubText          string
}

func (c *ConfigInput) FromRows(row sqlc.CreateConfigRow) {
	c.OwnerId = convert.Int32ToString(row.OwnerID)
	c.SiteName = convert.ToString(row.SiteName)
	c.LogoUrl = convert.ToString(row.LogoUrl)
	c.FaviconUrl = convert.ToString(row.FaviconUrl)
	c.MetaTags = convert.ToStringSlice(row.MetaTags)
	c.ContactEmail = row.ContactEmail
	c.ContactNumber = row.ContactNumber
	c.Instagram = convert.ToString(row.Instagram)
	c.MaintenanceMode = convert.ToBool(row.MaintenanceMode)
	c.MaintenanceNotice = convert.ToString(row.MaintenanceNotice)
	c.AddonsEnabled = convert.ToBool(row.AddonsEnabled)
	c.EnableReviews = convert.ToBool(row.EnableReviews)
	c.EnableWishlist = convert.ToBool(row.EnableWishlist)
	c.EnableNewsletterSignup = convert.ToBool(row.EnableNewsletterSignup)
	c.AboutContent = row.AboutContent
	c.FooterSubText = convert.ToString(row.FooterSubText)
}

// ToCache converts Config to protobuf message for caching.
func (c Config) ToCache() proto.Message {
	return &configv1.Config{
		OwnerId:                c.OwnerId,
		SiteName:               c.SiteName,
		LogoUrl:                c.LogoUrl,
		FaviconUrl:             c.FaviconUrl,
		MetaTags:               c.MetaTags,
		ContactEmail:           c.ContactEmail,
		ContactNumber:          c.ContactNumber,
		Instagram:              c.Instagram,
		MaintenanceMode:        c.MaintenanceMode,
		MaintenanceNotice:      c.MaintenanceNotice,
		AddonsEnabled:          c.AddonsEnabled,
		EnableReviews:          c.EnableReviews,
		EnableWishlist:         c.EnableWishlist,
		EnableNewsletterSignup: c.EnableNewsletterSignup,
		AboutContent:           c.AboutContent,
		FooterSubText:          c.FooterSubText,
	}
}

// FromCache converts protobuf message back to Config.
func (c Config) FromCache(message proto.Message) Config {
	if protoConfig, ok := message.(*configv1.Config); ok {
		return Config{
			OwnerId:                protoConfig.OwnerId,
			SiteName:               protoConfig.SiteName,
			LogoUrl:                protoConfig.LogoUrl,
			FaviconUrl:             protoConfig.FaviconUrl,
			MetaTags:               protoConfig.MetaTags,
			ContactEmail:           protoConfig.ContactEmail,
			ContactNumber:          protoConfig.ContactNumber,
			Instagram:              protoConfig.Instagram,
			MaintenanceMode:        protoConfig.MaintenanceMode,
			MaintenanceNotice:      protoConfig.MaintenanceNotice,
			AddonsEnabled:          protoConfig.AddonsEnabled,
			EnableReviews:          protoConfig.EnableReviews,
			EnableWishlist:         protoConfig.EnableWishlist,
			EnableNewsletterSignup: protoConfig.EnableNewsletterSignup,
			AboutContent:           protoConfig.AboutContent,
			FooterSubText:          protoConfig.FooterSubText,
		}
	}

	return c
}
