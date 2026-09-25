package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *ContactHandler) Validate(_ context.Context, req *ContactRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	if req.Email == "" {
		return onekiterrors.New(http.StatusBadRequest, "email is required")
	}
	return nil
}

func (h *ContactHandler) Serve(ctx context.Context, req *ContactRequest) (ContactResponse, error) {
	_, err := h.Service.Submit(ctx, domain.LeadInput{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.PhoneNumber,
		Service: "contact",
		Project: req.Message,
	})
	if err != nil {
		return ContactResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to submit contact request")
	}

	return ContactResponse{Message: "Message received"}, nil
}
