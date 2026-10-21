package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type AssignAssetHandler struct{ Service *application.Service }

type AssignAssetRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

type AssignAssetResponse struct {
	Success bool `json:"success"`
}

func (h *AssignAssetHandler) Validate(_ context.Context, req *AssignAssetRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	return nil
}
