package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrMenuProfileInvalidInput = errors.New("invalid menu profile input")
	ErrMenuPresetNotFound      = errors.New("menu preset not found")
	ErrMenuPresetInvalidInput  = errors.New("invalid menu preset input")
	ErrBranchItemInvalidInput  = errors.New("invalid branch menu item input")
	ErrTranslationInvalidInput = errors.New("invalid menu translation input")
	ErrCurrencyInvalidInput    = errors.New("invalid menu currency input")
	ErrStampCardNotFound       = errors.New("stamp card not found")
	ErrStampCardInvalidInput   = errors.New("invalid stamp card input")
	ErrTableActionInvalidInput = errors.New("invalid table action input")
	ErrEventInvalidInput       = errors.New("invalid menu event input")
	ErrUnsafeCustomCSS         = errors.New("custom CSS contains disallowed directives")
)

const MaxCustomCSSBytes = 32 * 1024

var disallowedCSSPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)@import\b`),
	regexp.MustCompile(`(?i)@charset\b`),
	regexp.MustCompile(`(?i)javascript\s*:`),
	regexp.MustCompile(`(?i)vbscript\s*:`),
	regexp.MustCompile(`(?i)expression\s*\(`),
	regexp.MustCompile(`(?i)behavior\s*:`),
	regexp.MustCompile(`(?i)-moz-binding\s*:`),
	regexp.MustCompile(`(?i)</\s*style`),
	regexp.MustCompile(`(?i)<\s*script`),
	regexp.MustCompile(`(?i)url\s*\(\s*['"]?\s*(?:https?:|//|data:text/html|javascript:)`),
}

// SanitizeCustomCSS enforces size and safety limits on tenant-authored CSS.
// External url(...) references and active-content constructs are rejected to
// prevent XSS and CSS-based data exfiltration on customer devices.
func SanitizeCustomCSS(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if len(trimmed) > MaxCustomCSSBytes {
		return "", fmt.Errorf("%w: custom_css exceeds %d bytes", ErrUnsafeCustomCSS, MaxCustomCSSBytes)
	}
	for _, pattern := range disallowedCSSPatterns {
		if pattern.MatchString(trimmed) {
			return "", fmt.Errorf("%w: matched forbidden CSS construct", ErrUnsafeCustomCSS)
		}
	}
	return trimmed, nil
}

// DefaultDraftConfigJSON returns a baseline MenuLayoutConfig JSON payload for
// branches that have not saved a custom profile yet.
func DefaultDraftConfigJSON() string {
	return `{"pages":{"menu":true,"about":false,"gallery":false,"reviews":false,"offers":false},"header":{"coverHeight":"md","showLogo":true,"showDescription":true,"showMeta":true,"sticky":true},"nav":{"style":"underline","size":"md"},"items":{"layout":"grid","columns":2,"imagePosition":"top","imageSize":"md","showDescription":true,"showBadges":true,"cardRadius":"md"},"colors":{"primary":"#111827","secondary":"#374151","accent":"#4583C4","bg":"#f9f9f7","surface":"#ffffff","text":"#1a1a1a"},"typography":{"headingFont":"system","bodyFont":"system","fontSize":"md"},"mode":"light","features":{"enableOrdering":true,"fulfillmentModes":{"dineIn":true,"selfService":true,"pickup":true,"delivery":false},"tableActions":{"callWaiter":true,"requestBill":true,"splitBill":false},"enableLoyaltyStamps":true,"enableAiAssistant":true,"enableBanners":true,"enableUpsells":true,"enableDietaryFilters":true,"enableCurrencySwitcher":true,"enableLanguageSwitcher":true,"enableSearch":true},"payment":{"mode":"pos","redirectUrl":""}}`
}

// ExtractPaymentRedirectURL parses the optional third-party payment redirect URL
// from a published profile's config JSON when payment.mode == "redirect".
func ExtractPaymentRedirectURL(configJSON string) string {
	if strings.TrimSpace(configJSON) == "" {
		return ""
	}
	var parsed struct {
		Payment struct {
			Mode        string `json:"mode"`
			RedirectURL string `json:"redirectUrl"`
		} `json:"payment"`
	}
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		return ""
	}
	if !strings.EqualFold(strings.TrimSpace(parsed.Payment.Mode), "redirect") {
		return ""
	}
	u := strings.TrimSpace(parsed.Payment.RedirectURL)
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		return ""
	}
	return u
}

type MenuProfile struct {
	ID                   int64
	BusinessID           int64
	LocationID           int64
	DraftConfigJSON      string
	PublishedConfigJSON  string
	CustomCSS            string
	PublishedCustomCSS   string
	WhiteLabelHideBadge  bool
	DefaultLocale        string
	EnabledLocales       []string
	DefaultCurrency      string
	IsPublished          bool
	PublishedAt          time.Time
	ClonedFromLocationID int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type SaveMenuProfileDraftInput struct {
	LocationID          int64
	DraftConfigJSON     string
	CustomCSS           *string
	WhiteLabelHideBadge *bool
	DefaultLocale       *string
	EnabledLocales      []string
	DefaultCurrency     *string
}

type MenuPreset struct {
	ID         int64
	BusinessID int64
	Name       string
	ConfigJSON string
	IsSystem   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type BranchMenuItem struct {
	ID               int64
	BusinessID       int64
	LocationID       int64
	MenuItemID       int64
	IsVisible        bool
	HasPriceOverride bool
	PriceOverride    float64
	DisplayOrder     int32
	FeaturedBadge    string
	DietaryTags      []string
	Allergens        []string
	Calories         int32
	MediaURLs        []string
	UpsellItemIDs    []int64
	UpdatedAt        time.Time
}

type MenuTranslation struct {
	ID         int64
	BusinessID int64
	LocationID int64
	EntityType string
	EntityID   string
	Locale     string
	Field      string
	Value      string
	Source     string
	UpdatedAt  time.Time
}

type TranslationSourceEntry struct {
	EntityType string
	EntityID   string
	Field      string
	SourceText string
}

type MenuCurrency struct {
	ID            int64
	BusinessID    int64
	LocationID    int64
	Code          string
	Symbol        string
	ExchangeRate  float64
	DecimalPlaces int32
	IsDefault     bool
	IsActive      bool
	Position      int32
}

type StampCard struct {
	ID                  int64
	BusinessID          int64
	LocationID          int64
	Title               string
	Description         string
	StampsRequired      int32
	RewardType          string
	RewardLabel         string
	RewardItemID        int64
	RewardValue         float64
	QualifyingMinAmount float64
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CustomerStampProgress struct {
	StampCardID     int64
	CardTitle       string
	StampsCount     int32
	StampsRequired  int32
	RewardsUnlocked int32
	RewardsRedeemed int32
	RewardLabel     string
}

type TableAction struct {
	ID         int64
	BusinessID int64
	LocationID int64
	TableID    int64
	TableLabel string
	ActionType string
	Status     string
	Note       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MenuEventInput struct {
	LocationID int64
	TableID    int64
	SessionID  string
	EventType  string
	ItemID     int64
	Locale     string
	Currency   string
}

type AnalyticsCount struct {
	EventType string
	Count     int64
}

type AnalyticsSummary struct {
	Counts         []AnalyticsCount
	QRScans        int64
	MenuViews      int64
	ItemViews      int64
	AddToCarts     int64
	CheckoutStarts int64
	OrdersPlaced   int64
}

type PublicHostResolution struct {
	Found        bool
	BusinessID   int64
	BusinessSlug string
	Domain       string
}

type PublicMenuBundle struct {
	Profile      MenuProfile
	Banners      []Banner
	BranchItems  []BranchMenuItem
	Currencies   []MenuCurrency
	Translations []MenuTranslation
	StampCards   []StampCard
}
