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

func GetAutomationRoutes() []string {
	return []string{
		"/zontriq/automations/{id}",
	}
}

type GetAutomationHandler struct {
	AutomationsService *application.Service
}

func (h *GetAutomationHandler) Serve(ctx context.Context, _ *automationsv1.GetAutomationRequest) (automationsv1.GetAutomationResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return automationsv1.GetAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := chi.URLParamFromCtx(ctx, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return automationsv1.GetAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid automation ID",
		}
	}

	automation, channels, err := h.AutomationsService.GetAutomation(ctx, id)
	if err != nil {
		return automationsv1.GetAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Automation not found",
		}
	}

	protoAutomation := &automationsv1.Automation{
		Id:              automation.ID,
		SeriesId:        automation.SeriesID,
		SeriesName:      automation.SeriesName,
		Name:            automation.Name,
		Frequency:       automation.Frequency,
		State:           automation.State,
		NextRun:         automation.NextRun,
		LastRun:         automation.LastRun,
		VideosGenerated: automation.VideosGenerated,
		Channels: func() []*automationsv1.AutomationChannel {
			result := make([]*automationsv1.AutomationChannel, 0, len(channels))
			for _, ch := range channels {
				result = append(result, &automationsv1.AutomationChannel{
					Id:          ch.ID,
					ChannelId:   ch.ChannelID,
					ChannelName: ch.ChannelName,
					Platform:    ch.Platform,
				})
			}

			return result
		}(),
		Status:    automation.Status,
		CreatedAt: automation.CreatedAt,
		UpdatedAt: automation.UpdatedAt,
	}

	return automationsv1.GetAutomationResponse{
		Automation: protoAutomation,
	}, nil
}
