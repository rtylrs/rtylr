package domain

type BusinessProfile struct {
	ID           int64
	Name         string
	SiteName     string
	LogoURL      string
	ContactEmail string
	ContactPhone string
	Address      string
	Currency     string
	Timezone     string
	Plan         string
}

type PublicBusiness struct {
	ID       int64
	Name     string
	Slug     string
	LogoURL  string
	Currency string
	Timezone string
}

type PublicLocation struct {
	ID       int64
	Name     string
	Slug     string
	Address  string
	Timezone string
	IsActive bool
}

type SlugLookupResult struct {
	Business  PublicBusiness
	Locations []PublicLocation
}

type OwnedBusiness struct {
	ID       int64
	Name     string
	LogoURL  string
	ParentID int64
	IsActive bool
	IsOwned  bool
}
