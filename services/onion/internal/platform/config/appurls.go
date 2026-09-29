package config

// Base URLs of the rtylr web apps. These are product constants — the same for
// every tenant — so use them instead of hardcoding "https://<app>.rtylr.com"
// anywhere in the codebase. (AuthAppURL is also the default for the optionally
// env-overridable Config.AuthAppURL field.)
const (
	AuthAppURL     = "https://sso.rtylr.com"
	HubAppURL      = "https://dash.rtylr.com"
	POSAppURL      = "https://pos.rtylr.com"
	ERPAppURL      = "https://erp.rtylr.com"
	HRAppURL       = "https://hr.rtylr.com"
	CRMAppURL      = "https://crm.rtylr.com"
	FinanceAppURL  = "https://finance.rtylr.com"
	FlowAppURL     = "https://flow.rtylr.com"
	MenuAppURL     = "https://menu.rtylr.com"
	InsightsAppURL = "https://insights.rtylr.com"
	RecruitAppURL  = "https://recruit.rtylr.com"
	DefaultAPIURL  = "https://api.voxire.com"
	LandingURL     = "https://rtylr.com"
)

// DefaultLicenseAuthorityURL is the hosted license authority that self-host
// backends validate their RTYLR_LICENSE_KEY against. It is a product constant,
// not a customer-configurable value.
const DefaultLicenseAuthorityURL = "https://api.voxire.com"

// RuntimeURLs returns the per-deployment app URLs that the backend advertises to
// frontends via /runtime-config. Each is an env override (e.g. DASH_URL) when set,
// otherwise the hosted default. In self-host these come from the customer's env;
// hosted keeps the *.rtylr.com / api.voxire.com defaults.
func RuntimeURLs(cfg *AppConfig) map[string]string {
	pick := func(env, def string) string {
		if v := envFirst(env); v != "" {
			return v
		}
		return def
	}
	auth := AuthAppURL
	upload := "https://api.voxire.com"
	if cfg != nil {
		if cfg.AuthAppURL != "" {
			auth = cfg.AuthAppURL
		}
		if cfg.CDNBaseURL != "" {
			upload = cfg.CDNBaseURL
		}
	}
	return map[string]string{
		"auth":     auth,
		"dash":     pick("DASH_URL", HubAppURL),
		"pos":      pick("POS_URL", POSAppURL),
		"erp":      pick("ERP_URL", ERPAppURL),
		"hr":       pick("HR_URL", HRAppURL),
		"crm":      pick("CRM_URL", CRMAppURL),
		"finance":  pick("FINANCE_URL", FinanceAppURL),
		"flow":     pick("FLOW_URL", FlowAppURL),
		"insights": pick("INSIGHTS_URL", InsightsAppURL),
		"recruit":  pick("RECRUIT_URL", RecruitAppURL),
		"menu":     pick("MENU_URL", MenuAppURL),
		"api":      pick("API_URL", DefaultAPIURL),
		"upload":   upload,
	}
}
