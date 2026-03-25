package domain

import (
	sqlc "github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

type Theme struct {
	ID                    int64
	PrimaryColor          string
	SecondaryColor        string
	AccentColor           string
	DarkGreenColor        string
	BackgroundColor       string
	BgColor               string
	SurfaceColor          string
	LightColor            string
	DarkColor             string
	ItemBackgroundColor   string
	ErrorColor            string
	SuccessColor          string
	WarningColor          string
	InfoColor             string
	TextPrimaryColor      string
	TextSecondaryColor    string
	SubtextColor          string
	LightGrayColor        string
	ActiveColor           string
	AdminBackgroundColor  string
	AdminPrimaryColor     string
	VendorBackgroundColor string
	BorderColor           string
	FontFamilyPrimary     string
	FontFamilySecondary   string
	FontSizeBase          string
	FontSizeHeading       string
	MobileBreakpoint      string
}

func (t *Theme) FromRows(row sqlc.Theme) {
	t.ID = int64(row.ID)
	t.PrimaryColor = convert.ToString(row.PrimaryColor)
	t.SecondaryColor = convert.ToString(row.SecondaryColor)
	t.AccentColor = convert.ToString(row.AccentColor)
	t.DarkGreenColor = convert.ToString(row.DarkGreenColor)
	t.BackgroundColor = convert.ToString(row.BackgroundColor)
	t.BgColor = convert.ToString(row.BgColor)
	t.SurfaceColor = convert.ToString(row.SurfaceColor)
	t.LightColor = convert.ToString(row.LightColor)
	t.DarkColor = convert.ToString(row.DarkColor)
	t.ItemBackgroundColor = convert.ToString(row.ItemBackgroundColor)
	t.ErrorColor = convert.ToString(row.ErrorColor)
	t.SuccessColor = convert.ToString(row.SuccessColor)
	t.WarningColor = convert.ToString(row.WarningColor)
	t.InfoColor = convert.ToString(row.InfoColor)
	t.TextPrimaryColor = convert.ToString(row.TextPrimaryColor)
	t.TextSecondaryColor = convert.ToString(row.TextSecondaryColor)
	t.SubtextColor = convert.ToString(row.SubtextColor)
	t.LightGrayColor = convert.ToString(row.LightGrayColor)
	t.ActiveColor = convert.ToString(row.ActiveColor)
	t.AdminBackgroundColor = convert.ToString(row.AdminBackgroundColor)
	t.AdminPrimaryColor = convert.ToString(row.AdminPrimaryColor)
	t.VendorBackgroundColor = convert.ToString(row.VendorBackgroundColor)
	t.BorderColor = convert.ToString(row.BorderColor)
	t.FontFamilyPrimary = convert.ToString(row.FontFamilyPrimary)
	t.FontFamilySecondary = convert.ToString(row.FontFamilySecondary)
	t.FontSizeBase = convert.ToString(row.FontSizeBase)
	t.FontSizeHeading = convert.ToString(row.FontSizeHeading)
	t.MobileBreakpoint = convert.ToString(row.MobileBreakpoint)
}
