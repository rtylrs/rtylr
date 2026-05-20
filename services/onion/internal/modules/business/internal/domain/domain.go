package domain

type BusinessProfile struct {
	ID           int64  `json:"id"`
	SiteName     string `json:"site_name"`
	LogoURL      string `json:"logo_url"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Address      string `json:"address"`
	Currency     string `json:"currency"`
	Timezone     string `json:"timezone"`
	Plan         string `json:"plan"`
}

type PublicBusiness struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	LogoURL  string `json:"logo_url"`
	Currency string `json:"currency"`
	Timezone string `json:"timezone"`
}
