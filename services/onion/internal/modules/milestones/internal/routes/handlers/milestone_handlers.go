package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/milestones/internal/application"
	"github.com/voxire/onion/internal/modules/milestones/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type MilestonePayload struct {
	ID          int64   `json:"id"`
	OwnerID     int64   `json:"owner_id"`
	ProjectID   int64   `json:"project_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	TargetDate  *string `json:"target_date"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func newMilestonePayload(m domain.Milestone) MilestonePayload {
	p := MilestonePayload{
		ID:          m.ID,
		OwnerID:     m.OwnerID,
		ProjectID:   m.ProjectID,
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if m.TargetDate != nil {
		s := m.TargetDate.Format("2006-01-02")
		p.TargetDate = &s
	}
	return p
}

// --- List Milestones ---

func MilestoneListRoutes() []string { return []string{"/flow/milestones"} }

type (
	ListMilestonesHandler  struct{ Service *application.Service }
	ListMilestonesRequest  struct{}
	ListMilestonesResponse struct {
		Milestones []MilestonePayload `json:"milestones"`
	}
)

func (h *ListMilestonesHandler) Validate(_ context.Context, _ *ListMilestonesRequest) *errorModels.ServerError {
	return nil
}

func (h *ListMilestonesHandler) Serve(ctx context.Context, _ *ListMilestonesRequest) (ListMilestonesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListMilestonesResponse{}, errorModels.PermissionDeniedServerError()
	}

	var projectID *int64
	if pidStr := chi.URLParamFromCtx(ctx, "projectId"); pidStr != "" {
		if pid, err := strconv.ParseInt(pidStr, 10, 64); err == nil {
			projectID = &pid
		}
	}

	milestones, err := h.Service.List(ctx, device.User.UserID, projectID)
	if err != nil {
		return ListMilestonesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list milestones"}
	}
	payloads := make([]MilestonePayload, len(milestones))
	for i, m := range milestones {
		payloads[i] = newMilestonePayload(m)
	}
	return ListMilestonesResponse{Milestones: payloads}, nil
}

// --- Create Milestone ---

func MilestoneCreateRoutes() []string { return []string{"/flow/milestones"} }

type CreateMilestoneHandler struct{ Service *application.Service }

type CreateMilestoneRequest struct {
	ProjectID   int64  `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TargetDate  string `json:"target_date"`
}

type CreateMilestoneResponse struct {
	MilestoneID int64  `json:"milestone_id"`
	Message     string `json:"message"`
}

func (h *CreateMilestoneHandler) Validate(_ context.Context, req *CreateMilestoneRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.ProjectID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Project ID is required"}
	}
	return nil
}

func (h *CreateMilestoneHandler) Serve(ctx context.Context, req *CreateMilestoneRequest) (CreateMilestoneResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateMilestoneResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.MilestoneInput{
		OwnerID:     device.User.UserID,
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
	}
	if req.TargetDate != "" {
		if t, err := time.Parse("2006-01-02", req.TargetDate); err == nil {
			input.TargetDate = &t
		}
	}
	id, err := h.Service.Create(ctx, input)
	if err != nil {
		return CreateMilestoneResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create milestone"}
	}
	return CreateMilestoneResponse{MilestoneID: id, Message: "Milestone created successfully"}, nil
}

// --- Update Milestone ---

func MilestoneUpdateRoutes() []string { return []string{"/flow/milestones/{id}"} }

type UpdateMilestoneHandler struct{ Service *application.Service }

type UpdateMilestoneRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TargetDate  string `json:"target_date"`
	Status      string `json:"status"`
}

type UpdateMilestoneResponse struct {
	Message string `json:"message"`
}

func (h *UpdateMilestoneHandler) Validate(_ context.Context, req *UpdateMilestoneRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateMilestoneHandler) Serve(ctx context.Context, req *UpdateMilestoneRequest) (UpdateMilestoneResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateMilestoneResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateMilestoneResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid milestone ID"}
	}
	input := domain.MilestoneUpdateInput{
		ID:          id,
		OwnerID:     device.User.UserID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}
	if req.TargetDate != "" {
		if t, parseErr := time.Parse("2006-01-02", req.TargetDate); parseErr == nil {
			input.TargetDate = &t
		}
	}
	if updateErr := h.Service.Update(ctx, input); updateErr != nil {
		if errors.Is(updateErr, domain.ErrMilestoneNotFound) {
			return UpdateMilestoneResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Milestone not found"}
		}
		return UpdateMilestoneResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update milestone"}
	}
	return UpdateMilestoneResponse{Message: "Milestone updated successfully"}, nil
}
