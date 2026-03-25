package application

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/authutils"
)

const UnknownValue = "unknown"

func (s *Service) BuildFeatureContext(ctx context.Context) *domain.FeatureContext {
	authUser := authutils.GetAuthUserFromContext(ctx)

	if authUser == nil {
		return s.BuildAnonymousContext(ctx, nil)
	}

	return s.BuildUserContext(ctx, int32(authUser.UserID), nil)
}

func (s *Service) BuildUserContext(_ context.Context, userID int32, req *http.Request) *domain.FeatureContext {
	featureContext := &domain.FeatureContext{
		UserID:      userID,
		IsAnonymous: false,
		CustomAttrs: make(map[string]interface{}),
	}

	if req != nil {
		featureContext.IPAddress = s.extractIPAddress(req)
		featureContext.UserAgent = req.UserAgent()
		featureContext.Platform = s.extractPlatform(req.UserAgent())
		featureContext.AppVersion = s.extractAppVersion(req.Header.Get("X-App-Version"))
		featureContext.Country = s.extractCountry(req.Header.Get("X-Country"))
		featureContext.SessionID = s.extractSessionID(req)
	}

	return featureContext
}

func (s *Service) BuildAnonymousContext(_ context.Context, req *http.Request) *domain.FeatureContext {
	featureContext := &domain.FeatureContext{
		UserID:      0,
		IsAnonymous: true,
		CustomAttrs: make(map[string]interface{}),
	}

	if req != nil {
		featureContext.IPAddress = s.extractIPAddress(req)
		featureContext.UserAgent = req.UserAgent()
		featureContext.Platform = s.extractPlatform(req.UserAgent())
		featureContext.AppVersion = s.extractAppVersion(req.Header.Get("X-App-Version"))
		featureContext.Country = s.extractCountry(req.Header.Get("X-Country"))
		featureContext.SessionID = s.extractSessionID(req)
	}

	return featureContext
}

func (s *Service) EnrichUserContext(_ context.Context, userContext *domain.FeatureContext, userID int32) *domain.FeatureContext {
	userContext.UserID = userID
	userContext.IsAnonymous = false

	return userContext
}

func (s *Service) extractIPAddress(req *http.Request) string {
	ip := req.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")

		return strings.TrimSpace(ips[0])
	}

	ip = req.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	return req.RemoteAddr
}

func (s *Service) extractPlatform(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		return "ios"
	}

	if strings.Contains(userAgent, "android") {
		return "android"
	}

	if strings.Contains(userAgent, "windows") {
		return "windows"
	}

	if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os") {
		return "macos"
	}

	if strings.Contains(userAgent, "linux") {
		return "linux"
	}

	return UnknownValue
}

func (s *Service) extractAppVersion(versionHeader string) string {
	if versionHeader == "" {
		return UnknownValue
	}

	return versionHeader
}

func (s *Service) extractCountry(countryHeader string) string {
	if countryHeader == "" {
		return UnknownValue
	}

	return strings.ToUpper(countryHeader)
}

func (s *Service) extractSessionID(req *http.Request) string {
	sessionCookie, err := req.Cookie("session_id")
	if err == nil && sessionCookie.Value != "" {
		return sessionCookie.Value
	}

	sessionHeader := req.Header.Get("X-Session-ID")
	if sessionHeader != "" {
		return sessionHeader
	}

	return ""
}
