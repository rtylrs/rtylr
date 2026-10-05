package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type ApproveEquipmentRequestHandler struct{ Service *application.Service }

type ApproveEquipmentRequestRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason"`
}

type ApproveEquipmentRequestResponse struct {
	Success bool `json:"success"`
}

func (h *ApproveEquipmentRequestHandler) Validate(_ context.Context, req *ApproveEquipmentRequestRequest) error {
	if !req.Approve && req.Reason == "" {
		return onekiterrors.New(http.StatusBadRequest, "reason is required when rejecting")
	}
	return nil
}
