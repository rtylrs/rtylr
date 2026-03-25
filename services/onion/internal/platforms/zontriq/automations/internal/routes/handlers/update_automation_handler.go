package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	automationsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/automations/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/application"
)

func UpdateAutomationRoutes() []string {
	return []string{
		"/zontriq/automations/{id}",
	}
}

type UpdateAutomationHandler struct {
	AutomationsService *application.Service
}

func (h *UpdateAutomationHandler) Validate(_ context.Context, req *automationsv1.UpdateAutomationRequest) *errorModels.ServerError {
	if req.GetName() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
	}

	if req.GetFrequency() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Frequency is required",
		}
	}

	if req.GetState() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "State is required",
		}
	}

	return nil
}

func (h *UpdateAutomationHandler) Serve(ctx context.Context, req *automationsv1.UpdateAutomationRequest) (automationsv1.UpdateAutomationResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return automationsv1.UpdateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := chi.URLParamFromCtx(ctx, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return automationsv1.UpdateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid automation ID",
		}
	}

	automation, err := h.AutomationsService.UpdateAutomation(ctx, id, req.GetName(), req.GetFrequency(), req.GetState(), req.GetChannelIds())
	if err != nil {
		return automationsv1.UpdateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update automation",
		}
	}

	return automationsv1.UpdateAutomationResponse{
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
