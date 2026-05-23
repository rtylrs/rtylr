package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SubmitRoutes() []string { return []string{"/quote"} }

type SubmitHandler struct{ Service *application.Service }

type SubmitRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Company     string `json:"company"`
	Service     string `json:"service"`
	Project     string `json:"project"`
	Budget      string `json:"budget"`
	Timeline    string `json:"timeline"`
	Source      string `json:"source"`
}

type SubmitResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *SubmitHandler) Validate(_ context.Context, req *SubmitRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	if req.Service == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "service is required"}
	}
	return nil
}

func (h *SubmitHandler) Serve(ctx context.Context, req *SubmitRequest) (SubmitResponse, *errorModels.ServerError) {
	// Prepend timeline to the project brief so it surfaces in the notification email
	// and Google Sheets without requiring a dedicated column.
	project := req.Project
	if req.Timeline != "" {
		project = fmt.Sprintf("Timeline: %s\n\n%s", req.Timeline, strings.TrimSpace(req.Project))
	}

	id, err := h.Service.Submit(ctx, domain.LeadInput{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.PhoneNumber,
		Company: req.Company,
		Service: req.Service,
		Project: project,
		Budget:  req.Budget,
		Source:  req.Source,
	})
	if err != nil {
		return SubmitResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to submit quote request"}
	}

	return SubmitResponse{ID: id, Message: "Quote request received"}, nil
}
