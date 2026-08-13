package domain

import (
	"errors"

	sqlc "github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

var ErrConfigNotFound = errors.New("config not found")

var ErrInvalidOwnerID = errors.New("invalid owner id")

type Config struct {
	OwnerId                   string
	BusinessID                int64 `json:"business_id"`
	SiteName                  string
	LogoUrl                   string
	FaviconUrl                string
	MetaTags                  []string
	ContactEmail              string
	ContactNumber             string
	SenderEmail               string
	EmailProvider             string
	SenderName                string
	ReplyToEmail              string
	EmailNotificationsEnabled bool
	SMTPHost                  string
	SMTPPort                  int32
	SMTPUsername              string
	SMTPPasswordConfigured    bool
	Website                   string
	MapsURL                   string
	Address                   string
	City                      string
	Zipcode                   string
	Country                   string
	Instagram                 string
	MaintenanceMode           bool
	MaintenanceNotice         string
	AddonsEnabled             bool
	EnableReviews             bool
	EnableWishlist            bool
	EnableNewsletterSignup    bool
	AboutContent              string
	FooterSubText             string
	BusinessSlug              string
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
	c.SenderEmail = convert.ToString(row.SenderEmail)
	c.EmailProvider = row.EmailProvider
	c.SenderName = convert.ToString(row.SenderName)
	c.ReplyToEmail = convert.ToString(row.ReplyToEmail)
	c.EmailNotificationsEnabled = row.EmailNotificationsEnabled
	c.SMTPHost = convert.ToString(row.SmtpHost)
	c.SMTPPort = row.SmtpPort.Int32
	c.SMTPUsername = convert.ToString(row.SmtpUsername)
	c.SMTPPasswordConfigured = convert.ToString(row.SmtpPasswordHash) != ""
	c.Website = convert.ToString(row.Website)
	c.MapsURL = convert.ToString(row.MapsUrl)
	c.Address = convert.ToString(row.Address)
	c.City = convert.ToString(row.City)
	c.Zipcode = convert.ToString(row.Zipcode)
	c.Country = convert.ToString(row.Country)
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
	OwnerId                   string
	BusinessID                int64 `json:"business_id"`
	SiteName                  string
	LogoUrl                   string
	FaviconUrl                string
	MetaTags                  []string
	ContactEmail              string
	ContactNumber             string
	SenderEmail               string
	EmailProvider             string
	SenderName                string
	ReplyToEmail              string
	EmailNotificationsEnabled bool
	SMTPHost                  string
	SMTPPort                  int32
	SMTPUsername              string
	SMTPPassword              string
	Website                   string
	MapsURL                   string
	Address                   string
	City                      string
	Zipcode                   string
	Country                   string
	Instagram                 string
	MaintenanceMode           bool
	MaintenanceNotice         string
	AddonsEnabled             bool
	EnableReviews             bool
	EnableWishlist            bool
	EnableNewsletterSignup    bool
	AboutContent              string
	FooterSubText             string
}
