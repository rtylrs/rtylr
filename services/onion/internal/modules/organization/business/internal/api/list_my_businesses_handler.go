package api

import (
	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
)

type BusinessEntry struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	LogoURL  string `json:"logo_url"`
	ParentID int64  `json:"parent_id,omitempty"`
	IsActive bool   `json:"is_active"`
}

type (
	ListMyBusinessesHandler  struct{ Service *application.Service }
	ListMyBusinessesRequest  struct{}
	ListMyBusinessesResponse struct {
		Businesses []BusinessEntry `json:"businesses"`
	}
)
