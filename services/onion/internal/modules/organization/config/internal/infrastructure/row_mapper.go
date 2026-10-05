package infrastructure

import (
	"github.com/voxire/onion/internal/modules/organization/config/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/config/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

func configFromRow(row sqlc.GetConfigRow) domain.Config {
	return domain.Config{
		OwnerId:                   convert.Int32ToString(row.OwnerID),
		BusinessID:                row.BusinessID,
		SiteName:                  convert.ToString(row.SiteName),
		LogoUrl:                   convert.ToString(row.LogoUrl),
		FaviconUrl:                convert.ToString(row.FaviconUrl),
		MetaTags:                  stringSliceOrEmpty(row.MetaTags),
		ContactEmail:              row.ContactEmail,
		ContactNumber:             row.ContactNumber,
		SenderEmail:               convert.ToString(row.SenderEmail),
		EmailProvider:             row.EmailProvider,
		SenderName:                convert.ToString(row.SenderName),
		ReplyToEmail:              convert.ToString(row.ReplyToEmail),
		EmailNotificationsEnabled: row.EmailNotificationsEnabled,
		SMTPHost:                  convert.ToString(row.SmtpHost),
		SMTPPort:                  row.SmtpPort.Int32,
		SMTPUsername:              convert.ToString(row.SmtpUsername),
		SMTPPasswordConfigured:    convert.ToString(row.SmtpPasswordHash) != "",
		Website:                   convert.ToString(row.Website),
		MapsURL:                   convert.ToString(row.MapsUrl),
		Address:                   convert.ToString(row.Address),
		City:                      convert.ToString(row.City),
		Zipcode:                   convert.ToString(row.Zipcode),
		Country:                   convert.ToString(row.Country),
		Instagram:                 convert.ToString(row.Instagram),
		MaintenanceMode:           convert.ToBool(row.MaintenanceMode),
		MaintenanceNotice:         convert.ToString(row.MaintenanceNotice),
		AddonsEnabled:             convert.ToBool(row.AddonsEnabled),
		EnableReviews:             convert.ToBool(row.EnableReviews),
		EnableWishlist:            convert.ToBool(row.EnableWishlist),
		EnableNewsletterSignup:    convert.ToBool(row.EnableNewsletterSignup),
		AboutContent:              row.AboutContent,
		FooterSubText:             convert.ToString(row.FooterSubText),
	}
}

func stringSliceOrEmpty(raw []byte) []string {
	values := convert.ToStringSlice(raw)
	if values == nil {
		return []string{}
	}
	return values
}
