package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// BusinessExistsRoutes powers inline name-availability validation on the
// onboarding "create your business" screen. Business names are globally unique,
// so the frontend probes this as the owner types. Returns only a boolean.
func BusinessExistsRoutes() []string {
	return []string{"/get-business-exists"}
}

type BusinessExistsHandler struct {
	UserService *application.Service
}

type BusinessExistsRequest struct {
	Name string `json:"name"`
}

type BusinessExistsResponse struct {
	Exists bool `json:"exists"`
}

func (h *BusinessExistsHandler) Validate(_ context.Context, req *BusinessExistsRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *BusinessExistsHandler) Serve(ctx context.Context, req *BusinessExistsRequest) (BusinessExistsResponse, *errorModels.ServerError) {
	exists, err := h.UserService.BusinessNameExists(ctx, req.Name)
	if err != nil {
		return BusinessExistsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to check business name"}
	}
	return BusinessExistsResponse{Exists: exists}, nil
}
