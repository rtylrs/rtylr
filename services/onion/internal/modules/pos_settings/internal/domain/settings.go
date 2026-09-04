package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	ProfileRetail       = "retail"
	ProfileRestaurant   = "restaurant"
	ProfileQuickService = "quick_service"

	ThemeDark     = "dark"
	ThemeLight    = "light"
	ThemeMidnight = "midnight"
	ThemeSlate    = "slate"
	ThemeAurora   = "aurora"
)

var ErrSettingsNotFound = errors.New("pos settings not found")

type FeatureSettings struct {
	Tables         bool `json:"tables"`
	KitchenDisplay bool `json:"kitchen_display"`
	Modifiers      bool `json:"modifiers"`
	Combos         bool `json:"combos"`
	Tips           bool `json:"tips"`
	GiftCards      bool `json:"gift_cards"`
	Customers      bool `json:"customers"`
	Inventory      bool `json:"inventory"`
	CashDrawer     bool `json:"cash_drawer"`
}

type Settings struct {
	ID            int64
	BusinessID    int64
	Profile       string
	Theme         string
	Features      FeatureSettings
	ReceiptFooter string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UpdateSettingsInput struct {
	BusinessID    int64
	Profile       string
	Theme         string
	Features      FeatureSettings
	HasFeatures   bool
	ReceiptFooter string
}

func DefaultFeatures(profile string) FeatureSettings {
	switch profile {
	case ProfileRestaurant:
		return FeatureSettings{
			Tables:         true,
			KitchenDisplay: true,
			Modifiers:      true,
			Combos:         true,
			Tips:           true,
			GiftCards:      true,
			Customers:      true,
			Inventory:      true,
			CashDrawer:     true,
		}
	case ProfileQuickService:
		return FeatureSettings{
			Tables:         false,
			KitchenDisplay: true,
			Modifiers:      true,
			Combos:         true,
			Tips:           true,
			GiftCards:      true,
			Customers:      true,
			Inventory:      true,
			CashDrawer:     true,
		}
	default:
		return FeatureSettings{
			Tables:         false,
			KitchenDisplay: false,
			Modifiers:      false,
			Combos:         false,
			Tips:           false,
			GiftCards:      true,
			Customers:      true,
			Inventory:      true,
			CashDrawer:     true,
		}
	}
}

func DefaultSettings(businessID int64) Settings {
	now := time.Now().UTC()
	return Settings{
		BusinessID:    businessID,
		Profile:       ProfileRetail,
		Theme:         ThemeDark,
		Features:      DefaultFeatures(ProfileRetail),
		ReceiptFooter: "Thank you for your purchase!",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func NormalizeProfile(profile string) string {
	switch strings.TrimSpace(profile) {
	case ProfileRetail, ProfileRestaurant, ProfileQuickService:
		return strings.TrimSpace(profile)
	default:
		return ProfileRetail
	}
}

func NormalizeTheme(theme string) string {
	switch strings.TrimSpace(theme) {
	case ThemeDark, ThemeLight, ThemeMidnight, ThemeSlate, ThemeAurora:
		return strings.TrimSpace(theme)
	default:
		return ThemeDark
	}
}

func NormalizeSettings(settings Settings) Settings {
	settings.Profile = NormalizeProfile(settings.Profile)
	settings.Theme = NormalizeTheme(settings.Theme)
	if strings.TrimSpace(settings.ReceiptFooter) == "" {
		settings.ReceiptFooter = "Thank you for your purchase!"
	}
	return settings
}
