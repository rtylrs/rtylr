package api

import (
	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
)

type (
	GetProfileHandler struct{ Service *application.Service }
	GetProfileRequest struct{}
)

type GetProfileResponse struct {
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
