package auth

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/models"
	"github.com/voxire/onion/internal/platform/utils/stringutils"
)

var desktopAppVersionRegex = regexp.MustCompile(`anghami/([0-9.]+).*electron`)

type deviceCtxKey struct{}

type Device struct {
	User                   *AuthUser
	Language               string
	IsAnonymous            bool
	IP                     string
	UserAgent              string
	DeviceInfo             string
	Platform               string
	DeviceUniqueIdentifier string
	PlatformVersion        string
	HasProxyEnabled        bool
	TvOsVersion            string
	TvDeviceModel          string
	Browser                string
}

func (d *Device) Ctx(ctx context.Context) context.Context {
	return context.WithValue(ctx, deviceCtxKey{}, *d)
}

func DeviceCtx(ctx context.Context) Device {
	device, ok := ctx.Value(deviceCtxKey{}).(Device)
	if !ok {
		if authUser, authOK := GetAuthUser(ctx); authOK && authUser != nil {
			return Device{
				User:        authUser,
				Language:    "en",
				IsAnonymous: false,
				IP:          authUser.IP,
			}
		}
		return GetAnonymousDevice()
	}

	return device
}

func GetAnonymousDevice() Device {
	return Device{
		IsAnonymous: true,
		Language:    "en",
		User:        nil,
	}
}

func extractIP(r *http.Request) string {
	// Keep this trust boundary aligned with the rate limiter: Cloudflare's
	// header is accepted only when the socket peer is a configured proxy.
	// Direct connections must use the peer rather than an attacker-supplied
	// forwarding header.
	if config.IsTrustedProxyIP(r.RemoteAddr) {
		if ip := strings.TrimSpace(r.Header.Get("Cf-Connecting-Ip")); net.ParseIP(ip) != nil {
			return ip
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return strings.Trim(r.RemoteAddr, "[]")
}

func ExtractDeviceInfo(r *http.Request) Device {
	device := Device{
		IP:         extractIP(r),
		UserAgent:  r.UserAgent(),
		Language:   extractLanguage(r),
		DeviceInfo: extractDeviceInfo(r),
	}

	populateAdvancedDeviceInfo(r, &device, true)

	if authUser, ok := GetAuthUser(r.Context()); ok {
		device.User = authUser
		device.IsAnonymous = false
	} else {
		device.IsAnonymous = true
	}

	return device
}

func extractLanguage(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return "en"
	}

	languages := strings.Split(acceptLang, ",")
	if len(languages) > 0 {
		primaryLang := strings.Split(strings.TrimSpace(languages[0]), "-")[0]

		return strings.ToLower(primaryLang)
	}

	return "en"
}

func extractDeviceInfo(r *http.Request) string {
	userAgent := r.UserAgent()
	if userAgent == "" {
		return "unknown"
	}

	userAgent = strings.ToLower(userAgent)

	switch {
	case strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone"):
		return "mobile"
	case strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "ipad"):
		return "tablet"
	case strings.Contains(userAgent, "bot") || strings.Contains(userAgent, "crawler"):
		return "bot"
	default:
		return "desktop"
	}
}

func DeviceMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			device := ExtractDeviceInfo(r)

			ctx := device.Ctx(r.Context())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func IsAuthenticatedDevice(ctx context.Context) bool {
	device := DeviceCtx(ctx)

	return !device.IsAnonymous && device.User != nil
}

func GetUserFromDevice(ctx context.Context) (*AuthUser, bool) {
	device := DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, false
	}

	return device.User, true
}

func GetLanguageFromDevice(ctx context.Context) string {
	device := DeviceCtx(ctx)

	return device.Language
}
func (d *Device) IsDesktopApp() bool {
	return d.Platform == models.PlatformDesktopApp
}

const (
	clientPlatformKey         = "client_platform"
	languageKey               = "language"
	deviceUniqueIdentifierKey = "device_unique_identifier"
	clientVersionKey          = "client_version"
	tvOsVersionKey            = "tv_os_version"
	tvDeviceModelKey          = "tv_device_model"
	userAgentKey              = "User-Agent"
	fastlyProxyEnabledKey     = "Fastly-Proxy-Enabled"
	fastlyBrowserKey          = "Fastly-Browser"
)

var (
	iOSUserAgentPatterns                  = []string{"cfnetwork"}
	ipadUserAgentPatterns                 = []string{"ipad"}
	appleTVUserAgentPatterns              = []string{"appletv", "tvos"}
	androidUserAgentPatterns              = []string{"okhttp", "anghami android", "dalvik"}
	webUserAgentPatterns                  = []string{"applewebkit", "mozilla", "chrome", "opera"}
	googleBotUserAgentPatterns            = []string{"googlebot"}
	twitterBotUserAgentPatterns           = []string{"twitterbot"}
	phpAPIUserAgentPatterns               = []string{"anghami api", "php"}
	goClientUserAgentPatterns             = []string{models.PlatformGoKoussa}
	svixWebhooksUserAgentPatterns         = []string{"svix-webhooks"}
	octopusUserAgentPatterns              = []string{"python-urllib", "octopus"}
	androidTVUserAgentPatterns            = []string{"android tv"}
	samsungTVUserAgentPatterns            = []string{"tizen"}
	lgTVUserAgentPatterns                 = []string{"web0s"}
	desktopAppUserAgentPatterns           = []string{"electron"}
	smartTVUserAgentPatterns              = []string{"smart"}
	amazonFireTVUserAgentPatterns         = []string{"aftb", "afts"}
	amazonFireTVStickUserAgentPatterns    = []string{"aftm", "aftt"}
	appleWatchUserAgentPatterns           = []string{"anghami apple watch"}
	sonosUserAgentPatterns                = []string{"sonos"}
	androidWearOsUserAgentPatterns        = []string{"androidwearos"}
	androidHarmonyWearOsUserAgentPatterns = []string{"harmonywearos"}
	androidAutomotiveOsUserAgentPatterns  = []string{"automotiveos"}
	dashboardUserAgentPatterns            = []string{"dashboard"}
	watsonUserAgentPatterns               = []string{"watson"}
)

func detectPlatform(userAgent, deviceInfo string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, models.PlatformAlexa) {
		return models.PlatformAlexa
	}

	if platform := detectApplePlatforms(userAgent, deviceInfo); platform != models.PlatformNotAvailable {
		return platform
	}

	if platform := detectAndroidPlatforms(userAgent, deviceInfo); platform != models.PlatformNotAvailable {
		return platform
	}

	if platform := detectTVPlatforms(userAgent); platform != models.PlatformNotAvailable {
		return platform
	}

	if platform := detectWebPlatforms(userAgent, deviceInfo); platform != models.PlatformNotAvailable {
		return platform
	}

	return detectOtherPlatforms(userAgent)
}

func detectApplePlatforms(userAgent, deviceInfo string) string {
	if stringutils.ContainsOneOf(userAgent, appleTVUserAgentPatterns) {
		return models.PlatformAppleTV
	}

	if stringutils.ContainsOneOf(userAgent, appleWatchUserAgentPatterns) {
		return models.PlatformAppleWatch
	}

	if (strings.HasPrefix(deviceInfo, "i") && !strings.Contains(userAgent, models.PlatformAlexa)) || stringutils.ContainsOneOf(userAgent, iOSUserAgentPatterns) {
		if stringutils.ContainsOneOf(userAgent, ipadUserAgentPatterns) {
			return models.PlatformIpad
		}

		return models.PlatformIOS
	}

	return models.PlatformNotAvailable
}

func detectAndroidPlatforms(userAgent, deviceInfo string) string {
	if stringutils.ContainsOneOf(userAgent, androidTVUserAgentPatterns) {
		return models.PlatformAndroidTV
	}

	if stringutils.ContainsOneOf(userAgent, androidWearOsUserAgentPatterns) {
		return models.PlatformAndroidWearOS
	}

	if stringutils.ContainsOneOf(userAgent, androidHarmonyWearOsUserAgentPatterns) {
		return models.PlatformAndroidHarmonyOS
	}

	if stringutils.ContainsOneOf(userAgent, androidAutomotiveOsUserAgentPatterns) {
		return models.PlatformAndroidAutomotiveOS
	}

	if (strings.HasPrefix(deviceInfo, "a") && !strings.Contains(userAgent, models.PlatformAlexa)) || stringutils.ContainsOneOf(userAgent, androidUserAgentPatterns) {
		return models.PlatformAndroid
	}

	return models.PlatformNotAvailable
}

func detectTVPlatforms(userAgent string) string {
	if stringutils.ContainsOneOf(userAgent, samsungTVUserAgentPatterns) {
		return models.PlatformSamsungTV
	}

	if stringutils.ContainsOneOf(userAgent, lgTVUserAgentPatterns) {
		return models.PlatformLgTV
	}

	if stringutils.ContainsOneOf(userAgent, amazonFireTVUserAgentPatterns) {
		return models.PlatformAmazonFireTV
	}

	if stringutils.ContainsOneOf(userAgent, amazonFireTVStickUserAgentPatterns) {
		return models.PlatformAmazonFireTVStick
	}

	return models.PlatformNotAvailable
}

func detectWebPlatforms(userAgent, deviceInfo string) string {
	if (strings.HasPrefix(deviceInfo, "w") && !strings.Contains(userAgent, models.PlatformAlexa)) || stringutils.ContainsOneOf(userAgent, webUserAgentPatterns) {
		if stringutils.ContainsOneOf(userAgent, smartTVUserAgentPatterns) {
			return models.PlatformSmartTV
		}

		return models.PlatformWeb
	}

	return models.PlatformNotAvailable
}

func detectOtherPlatforms(userAgent string) string {
	platformChecks := []struct {
		patterns []string
		platform string
	}{
		{desktopAppUserAgentPatterns, models.PlatformDesktopApp},
		{phpAPIUserAgentPatterns, models.PlatformPhpAPI},
		{dashboardUserAgentPatterns, models.PlatformDashboard},
		{watsonUserAgentPatterns, models.PlatformWatson},
		{googleBotUserAgentPatterns, models.PlatformGoogleBot},
		{twitterBotUserAgentPatterns, models.PlatformTwitterBot},
		{sonosUserAgentPatterns, models.PlatformSonos},
		{octopusUserAgentPatterns, models.PlatformOctopus},
		{goClientUserAgentPatterns, models.PlatformGoKoussa},
		{svixWebhooksUserAgentPatterns, models.PlatformSvixWebhooks},
	}

	for _, check := range platformChecks {
		if stringutils.ContainsOneOf(userAgent, check.patterns) {
			return check.platform
		}
	}

	return models.PlatformNotAvailable
}

func detectPlatformVersion(userAgent, deviceInfo string, isDesktopApp bool) string {
	if isDesktopApp {
		desktopAppInfo := desktopAppVersionRegex.FindStringSubmatch(strings.ToLower(userAgent))
		if len(desktopAppInfo) > 1 {
			return desktopAppInfo[1]
		}
	}

	if len(deviceInfo) > 1 {
		return deviceInfo[1:]
	}

	return "N/A"
}

func populateAdvancedDeviceInfo(r *http.Request, device *Device, useLegacyFallback bool) {
	if platform := r.Header.Get(clientPlatformKey); platform != "" {
		device.Platform = platform
	}

	if device.Platform == "" && useLegacyFallback {
		device.Platform = detectPlatform(r.UserAgent(), device.DeviceInfo)
	}

	if language := r.Header.Get(languageKey); language != "" {
		device.Language = language
	}

	if deviceIdentifier := r.Header.Get(deviceUniqueIdentifierKey); deviceIdentifier != "" {
		device.DeviceUniqueIdentifier = deviceIdentifier
	}

	if clientVersion := r.Header.Get(clientVersionKey); clientVersion != "" {
		device.PlatformVersion = clientVersion
	}

	if device.PlatformVersion == "" && useLegacyFallback {
		device.PlatformVersion = detectPlatformVersion(r.UserAgent(), device.DeviceInfo, device.IsDesktopApp())
	}

	if userAgent := r.Header.Get(userAgentKey); userAgent != "" {
		device.UserAgent = userAgent
	}

	if proxyEnabled := r.Header.Get(fastlyProxyEnabledKey); proxyEnabled == "1" {
		device.HasProxyEnabled = true
	}

	if tvOsVersion := r.Header.Get(tvOsVersionKey); tvOsVersion != "" {
		device.TvOsVersion = tvOsVersion
	}

	if tvDeviceModel := r.Header.Get(tvDeviceModelKey); tvDeviceModel != "" {
		device.TvDeviceModel = tvDeviceModel
	}

	if browser := strings.ToLower(r.Header.Get(fastlyBrowserKey)); browser != "" {
		device.Browser = browser
	}
}
