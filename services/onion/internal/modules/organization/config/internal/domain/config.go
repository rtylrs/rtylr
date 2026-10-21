package domain

import "errors"

var ErrConfigNotFound = errors.New("config not found")

var ErrInvalidOwnerID = errors.New("invalid owner id")

type Config struct {
	OwnerId                   string
	BusinessID                int64
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

type ConfigInput struct {
	OwnerId                   string
	BusinessID                int64
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
