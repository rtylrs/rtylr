package handlers

import (
	"context"
	"net/http"

	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	automationsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/automations/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/application"
)

func CreateAutomationRoutes() []string {
	return []string{
		"/zontriq/automations",
	}
}

type CreateAutomationHandler struct {
	AutomationsService *application.Service
}

func (h *CreateAutomationHandler) Validate(_ context.Context, req *automationsv1.CreateAutomationRequest) *errorModels.ServerError {
	if req.GetName() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
	}

	if req.GetSeriesId() <= 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Series ID is required",
		}
	}

	if req.GetFrequency() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Frequency is required",
		}
	}

	return nil
}

func (h *CreateAutomationHandler) Serve(ctx context.Context, req *automationsv1.CreateAutomationRequest) (automationsv1.CreateAutomationResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return automationsv1.CreateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	automation, err := h.AutomationsService.CreateAutomation(ctx, int(zontriqUserID), int(req.GetSeriesId()), req.GetName(), req.GetFrequency(), req.GetChannelIds())
	if err != nil {
		return automationsv1.CreateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create automation",
		}
	}

	return automationsv1.CreateAutomationResponse{
		Automation: &automationsv1.Automation{
			Id:              automation.ID,
			SeriesId:        automation.SeriesID,
			Name:            automation.Name,
			Frequency:       automation.Frequency,
			State:           automation.State,
			NextRun:         automation.NextRun,
			LastRun:         automation.LastRun,
			VideosGenerated: automation.VideosGenerated,
			Status:          automation.Status,
			CreatedAt:       automation.CreatedAt,
			UpdatedAt:       automation.UpdatedAt,
		},
	}, nil
}
