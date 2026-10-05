package api

import (
	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/application"
)

type UnassignAssetHandler struct{ Service *application.Service }

type UnassignAssetRequest struct{}

type UnassignAssetResponse struct {
	Success bool `json:"success"`
}
