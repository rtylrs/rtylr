package domain

type BusinessProfile struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
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
	Slug     string `json:"slug"`
	LogoURL  string `json:"logo_url"`
	Currency string `json:"currency"`
	Timezone string `json:"timezone"`
}

type PublicLocation struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Address  string `json:"address"`
	Timezone string `json:"timezone"`
	IsActive bool   `json:"is_active"`
}

type SlugLookupResult struct {
	Business  PublicBusiness   `json:"business"`
	Locations []PublicLocation `json:"locations"`
}

type OwnedBusiness struct {
	ID       int64
	Name     string
	LogoURL  string
	ParentID int64
	IsActive bool
	IsOwned  bool
}
