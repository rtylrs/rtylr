package domain

type BusinessProfile struct {
	ID           int64  `json:"id"`
	SiteName     string `json:"siteName"`
	LogoURL      string `json:"logoUrl"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
	Address      string `json:"address"`
	Currency     string `json:"currency"`
	Timezone     string `json:"timezone"`
	Plan         string `json:"plan"`
}
