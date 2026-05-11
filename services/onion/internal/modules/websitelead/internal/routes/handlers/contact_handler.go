package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ContactRoutes() []string { return []string{"/contact"} }

type ContactHandler struct{ Service *application.Service }

type ContactRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
}

type ContactResponse struct {
	Message string `json:"message"`
}

func (h *ContactHandler) Validate(_ context.Context, req *ContactRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	return nil
}

func (h *ContactHandler) Serve(ctx context.Context, req *ContactRequest) (ContactResponse, *errorModels.ServerError) {
	_, err := h.Service.Submit(ctx, domain.LeadInput{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.PhoneNumber,
		Service: "contact",
		Project: req.Message,
	})
	if err != nil {
		return ContactResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to submit contact request",
		}
	}

	return ContactResponse{Message: "Message received"}, nil
}
