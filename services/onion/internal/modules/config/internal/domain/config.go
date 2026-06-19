package domain

import (
	"errors"

	sqlc "github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

var ErrConfigNotFound = errors.New("config not found")

type Config struct {
	OwnerId                string
	BusinessID             int64  `json:"business_id"`
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
	BusinessSlug           string
}

func (c *Config) FromRows(row sqlc.GetConfigRow) {
	c.OwnerId = convert.Int32ToString(row.OwnerID)
	c.BusinessID = row.BusinessID
	c.SiteName = convert.ToString(row.SiteName)
	c.LogoUrl = convert.ToString(row.LogoUrl)
	c.FaviconUrl = convert.ToString(row.FaviconUrl)
	if tags := convert.ToStringSlice(row.MetaTags); tags != nil {
		c.MetaTags = tags
	} else {
		c.MetaTags = []string{}
	}
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
	BusinessID             int64  `json:"business_id"`
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
