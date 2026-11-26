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
	ErrMenuProfileInvalidInput  = errors.New("invalid menu profile input")
	ErrMenuPresetNotFound       = errors.New("menu preset not found")
	ErrMenuPresetInvalidInput   = errors.New("invalid menu preset input")
	ErrBranchItemInvalidInput   = errors.New("invalid branch menu item input")
	ErrTranslationInvalidInput  = errors.New("invalid menu translation input")
	ErrCurrencyInvalidInput     = errors.New("invalid menu currency input")
	ErrStampCardNotFound        = errors.New("stamp card not found")
	ErrStampCardInvalidInput    = errors.New("invalid stamp card input")
	ErrTableActionInvalidInput  = errors.New("invalid table action input")
	ErrEventInvalidInput        = errors.New("invalid menu event input")
	ErrGuestCaptureInvalidInput = errors.New("invalid guest capture input")
	ErrFeatureDisabled          = errors.New("feature is disabled")
	ErrUnsafeCustomCSS          = errors.New("custom CSS contains disallowed directives")
)

const MaxCustomCSSBytes = 32 * 1024

var cssCommentPattern = regexp.MustCompile(`(?s)/\*.*?\*/`)

var disallowedCSSPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)@import\b`),
	regexp.MustCompile(`(?i)@charset\b`),
	regexp.MustCompile(`(?i)@font-face\b`),
	regexp.MustCompile(`(?i)@namespace\b`),
	regexp.MustCompile(`(?i)javascript\s*:`),
	regexp.MustCompile(`(?i)vbscript\s*:`),
	regexp.MustCompile(`(?i)expression\s*\(`),
	regexp.MustCompile(`(?i)behavior\s*:`),
	regexp.MustCompile(`(?i)-moz-binding\s*:`),
	regexp.MustCompile(`(?i)</\s*style`),
	regexp.MustCompile(`(?i)<\s*script`),
	regexp.MustCompile(`(?i)\burl\s*\(`),
}

var protectedSelectorPattern = regexp.MustCompile(`(?i)(poweredby|powered-by|data-rtylr-powered|\bfooter\b)`)

// SanitizeCustomCSS strips comments, blocks escape/URL/active-content bypasses,
// protects the RTYLR attribution footer, and scopes every selector strictly
// inside .rtylr-menu-root so tenant CSS cannot affect the surrounding page.
func SanitizeCustomCSS(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if len(trimmed) > MaxCustomCSSBytes {
		return "", fmt.Errorf("%w: custom_css exceeds %d bytes", ErrUnsafeCustomCSS, MaxCustomCSSBytes)
	}
	// Reject CSS backslash escapes (e.g. \75rl(...)) and unclosed comments before pattern checks.
	if strings.Contains(trimmed, `\`) {
		return "", fmt.Errorf("%w: CSS escape sequences are not allowed", ErrUnsafeCustomCSS)
	}
	noComments := strings.TrimSpace(cssCommentPattern.ReplaceAllString(trimmed, ""))
	if strings.Contains(noComments, "/*") || strings.Contains(noComments, "*/") {
		return "", fmt.Errorf("%w: unclosed CSS comment", ErrUnsafeCustomCSS)
	}
	if noComments == "" {
		return "", nil
	}
	for _, pattern := range disallowedCSSPatterns {
		if pattern.MatchString(noComments) {
			return "", fmt.Errorf("%w: matched forbidden CSS construct", ErrUnsafeCustomCSS)
		}
	}
	scoped, err := scopeCSSToMenuRoot(noComments)
	if err != nil {
		return "", err
	}
	return scoped, nil
}

func scopeCSSToMenuRoot(css string) (string, error) {
	var out strings.Builder
	i := 0
	n := len(css)
	for i < n {
		for i < n && (css[i] == ' ' || css[i] == '\t' || css[i] == '\n' || css[i] == '\r') {
			i++
		}
		if i >= n {
			break
		}
		if css[i] == '@' {
			// Only allow safe @media rules, recursively scoping their inner rules.
			openIdx := strings.IndexByte(css[i:], '{')
			if openIdx < 0 {
				return "", fmt.Errorf("%w: invalid at-rule", ErrUnsafeCustomCSS)
			}
			header := strings.TrimSpace(css[i : i+openIdx])
			if !strings.HasPrefix(strings.ToLower(header), "@media") {
				return "", fmt.Errorf("%w: unsupported at-rule %q", ErrUnsafeCustomCSS, header)
			}
			bodyStart := i + openIdx + 1
			depth := 1
			j := bodyStart
			for j < n && depth > 0 {
				if css[j] == '{' {
					depth++
				} else if css[j] == '}' {
					depth--
				}
				j++
			}
			if depth != 0 {
				return "", fmt.Errorf("%w: unbalanced braces in @media rule", ErrUnsafeCustomCSS)
			}
			innerScoped, err := scopeCSSToMenuRoot(css[bodyStart : j-1])
			if err != nil {
				return "", err
			}
			if out.Len() > 0 {
				out.WriteByte('\n')
			}
			out.WriteString(header)
			out.WriteString(" {\n")
			out.WriteString(innerScoped)
			out.WriteString("\n}")
			i = j
			continue
		}

		openIdx := strings.IndexByte(css[i:], '{')
		if openIdx < 0 {
			return "", fmt.Errorf("%w: malformed CSS rule", ErrUnsafeCustomCSS)
		}
		selectorList := strings.TrimSpace(css[i : i+openIdx])
		closeIdx := strings.IndexByte(css[i+openIdx+1:], '}')
		if closeIdx < 0 {
			return "", fmt.Errorf("%w: unclosed CSS rule", ErrUnsafeCustomCSS)
		}
		declarations := strings.TrimSpace(css[i+openIdx+1 : i+openIdx+1+closeIdx])
		if strings.Contains(declarations, "{") {
			return "", fmt.Errorf("%w: nested rule blocks are not supported", ErrUnsafeCustomCSS)
		}

		rawSelectors := strings.Split(selectorList, ",")
		scopedSelectors := make([]string, 0, len(rawSelectors))
		for _, sel := range rawSelectors {
			s := strings.TrimSpace(sel)
			if s == "" {
				continue
			}
			if protectedSelectorPattern.MatchString(s) {
				return "", fmt.Errorf("%w: selector targets protected attribution footer", ErrUnsafeCustomCSS)
			}
			lower := strings.ToLower(s)
			switch {
			case lower == ":root" || lower == "html" || lower == "body":
				s = ".rtylr-menu-root"
			case strings.HasPrefix(s, ".rtylr-menu-root"):
				// Already scoped.
			default:
				s = ".rtylr-menu-root " + s
			}
			scopedSelectors = append(scopedSelectors, s)
		}
		if len(scopedSelectors) == 0 {
			return "", fmt.Errorf("%w: empty CSS selector", ErrUnsafeCustomCSS)
		}
		if out.Len() > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(strings.Join(scopedSelectors, ", "))
		out.WriteString(" { ")
		out.WriteString(declarations)
		out.WriteString(" }")
		i = i + openIdx + 1 + closeIdx + 1
	}
	return out.String(), nil
}

// IsLoyaltyStampsEnabled reports whether a published profile config enables digital stamp cards.
func IsLoyaltyStampsEnabled(configJSON string) bool {
	if strings.TrimSpace(configJSON) == "" {
		return true
	}
	var parsed struct {
		Features struct {
			EnableLoyaltyStamps *bool `json:"enableLoyaltyStamps"`
		} `json:"features"`
	}
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		return true
	}
	if parsed.Features.EnableLoyaltyStamps == nil {
		return true
	}
	return *parsed.Features.EnableLoyaltyStamps
}

// IsCRMCaptureEnabled reports whether a published profile enables optional CRM guest capture.
func IsCRMCaptureEnabled(configJSON string) bool {
	if strings.TrimSpace(configJSON) == "" {
		return true
	}
	var parsed struct {
		Features struct {
			EnableCrmCapture *bool `json:"enableCrmCapture"`
		} `json:"features"`
	}
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		return true
	}
	if parsed.Features.EnableCrmCapture == nil {
		return true
	}
	return *parsed.Features.EnableCrmCapture
}

// IsOrderingEnabled reports whether a published profile enables customer ordering.
func IsOrderingEnabled(configJSON string) bool {
	if strings.TrimSpace(configJSON) == "" {
		return true
	}
	var parsed struct {
		Features struct {
			EnableOrdering *bool `json:"enableOrdering"`
		} `json:"features"`
	}
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		return true
	}
	if parsed.Features.EnableOrdering == nil {
		return true
	}
	return *parsed.Features.EnableOrdering
}

// DefaultDraftConfigJSON returns a baseline MenuLayoutConfig JSON payload for
// branches that have not saved a custom profile yet.
func DefaultDraftConfigJSON() string {
	return `{"pages":{"menu":true,"about":false,"gallery":false,"reviews":false,"offers":false},"header":{"coverHeight":"md","showLogo":true,"showDescription":true,"showMeta":true,"sticky":true},"nav":{"style":"underline","size":"md"},"items":{"layout":"grid","columns":2,"imagePosition":"top","imageSize":"md","showDescription":true,"showBadges":true,"cardRadius":"md"},"colors":{"primary":"#111827","secondary":"#374151","accent":"#4583C4","bg":"#f9f9f7","surface":"#ffffff","text":"#1a1a1a"},"typography":{"headingFont":"system","bodyFont":"system","fontSize":"md"},"mode":"light","features":{"enableOrdering":true,"fulfillmentModes":{"dineIn":true,"selfService":true,"pickup":true,"delivery":false},"tableActions":{"callWaiter":true,"requestBill":true,"splitBill":false},"enableLoyaltyStamps":true,"enableCrmCapture":true,"enableAiAssistant":true,"enableBanners":true,"enableUpsells":true,"enableDietaryFilters":true,"enableCurrencySwitcher":true,"enableLanguageSwitcher":true,"enableSearch":true},"payment":{"mode":"pos","redirectUrl":""}}`
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
