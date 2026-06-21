package config

// Base URLs of the rtylr web apps. These are product constants — the same for
// every tenant — so use them instead of hardcoding "https://<app>.rtylr.com"
// anywhere in the codebase. (AuthAppURL is also the default for the optionally
// env-overridable Config.AuthAppURL field.)
const (
	AuthAppURL     = "https://auth.rtylr.com"
	HubAppURL      = "https://dash.rtylr.com"
	ERPAppURL      = "https://erp.rtylr.com"
	HRAppURL       = "https://hr.rtylr.com"
	CRMAppURL      = "https://crm.rtylr.com"
	FinanceAppURL  = "https://finance.rtylr.com"
	FlowAppURL     = "https://flow.rtylr.com"
	MenuAppURL     = "https://menu.rtylr.com"
	InsightsAppURL = "https://insights.rtylr.com"
	LandingURL     = "https://rtylr.com"
)

// DefaultLicenseAuthorityURL is the hosted license authority that self-host
// backends validate their RTYLR_LICENSE_KEY against. It is a product constant,
// not a customer-configurable value.
const DefaultLicenseAuthorityURL = "https://license.voxire.com"
