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

func DeleteAutomationRoutes() []string {
	return []string{
		"/zontriq/automations/{id}",
	}
}

type DeleteAutomationHandler struct {
	AutomationsService *application.Service
}

func (h *DeleteAutomationHandler) Serve(ctx context.Context, _ *automationsv1.DeleteAutomationRequest) (automationsv1.DeleteAutomationResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return automationsv1.DeleteAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := chi.URLParamFromCtx(ctx, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return automationsv1.DeleteAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid automation ID",
		}
	}

	if err := h.AutomationsService.DeleteAutomation(ctx, id); err != nil {
		return automationsv1.DeleteAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete automation",
		}
	}

	return automationsv1.DeleteAutomationResponse{}, nil
}
