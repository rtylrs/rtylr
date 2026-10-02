package config

// Base URLs of the rtylr web apps. These are product constants — the same for
// every tenant — so use them instead of hardcoding "https://<app>.rtylr.com"
// anywhere in the codebase. (AuthAppURL is also the default for the optionally
// env-overridable Config.AuthAppURL field.)
const (
	AuthAppURL     = "https://sso.rtylr.com"
	HubAppURL      = "https://portal.rtylr.com"
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
