package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Milestones ──────────────────────────────────────────────────────────

func ListMilestonesRoutes() []string { return []string{"/flow/milestones"} }

type ListMilestonesRequest struct {
	ProjectID   string `query:"project_id"`
	WorkspaceID string `query:"workspace_id"`
}

type ListMilestonesResponse struct {
	Milestones []milestonePayload `json:"milestones"`
}

type milestonePayload struct {
	ID               string                   `json:"id"`
	ProjectID        string                   `json:"projectId"`
	Name             string                   `json:"name"`
	TargetDate       string                   `json:"targetDate"`
	Progress         int                      `json:"progress"`
	ComputedProgress int                      `json:"computedProgress"`
	DoneCount        int                      `json:"doneCount"`
	TaskCount        int                      `json:"taskCount"`
	Status           string                   `json:"status"`
	Description      string                   `json:"description,omitempty"`
	Project          *milestoneProjectPayload `json:"project,omitempty"`
}

type milestoneProjectPayload struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (r *ListMilestonesRequest) Validate() *errorModels.ServerError { return nil }

type ListMilestonesHandler struct{ FlowService *application.Service }

func (h *ListMilestonesHandler) Serve(ctx context.Context, r *ListMilestonesRequest) (ListMilestonesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListMilestonesResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		if id, err := strconv.ParseInt(r.WorkspaceID, 10, 64); err == nil {
			wsID = id
		}
	}

	var projID *int64
	if r.ProjectID != "" {
		if id, err := strconv.ParseInt(r.ProjectID, 10, 64); err == nil {
			projID = &id
		}
	}

	milestones, err := h.FlowService.ListMilestones(ctx, wsID, projID)
	if err != nil {
		return ListMilestonesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	payloads := make([]milestonePayload, 0, len(milestones))
	for _, m := range milestones {
		p := milestonePayload{
			ID:               strconv.FormatInt(m.ID, 10),
			ProjectID:        strconv.FormatInt(m.ProjectID, 10),
			Name:             m.Name,
			TargetDate:       m.DueDate.Format("2006-01-02"),
			Progress:         m.Progress,
			ComputedProgress: m.Progress,
			DoneCount:        m.DoneCount,
			TaskCount:        m.TaskCount,
			Status:           m.Status,
			Description:      m.Description,
		}
		if m.Project != nil {
			p.Project = &milestoneProjectPayload{Name: m.Project.Name, Color: m.Project.Color}
		}
		payloads = append(payloads, p)
	}
	return ListMilestonesResponse{Milestones: payloads}, nil
}

// ─── Create Milestone ─────────────────────────────────────────────────────────

func CreateMilestoneRoutes() []string { return []string{"/flow/milestones"} }

type CreateMilestoneRequest struct {
	ProjectID   string `json:"projectId"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TargetDate  string `json:"targetDate"`
}

type CreateMilestoneResponse struct {
	Milestone milestonePayload `json:"milestone"`
}

func (r *CreateMilestoneRequest) Validate() *errorModels.ServerError {
	if r.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

type CreateMilestoneHandler struct{ FlowService *application.Service }

func (h *CreateMilestoneHandler) Serve(ctx context.Context, r *CreateMilestoneRequest) (CreateMilestoneResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateMilestoneResponse{}, errorModels.PermissionDeniedServerError()
	}

	projID, _ := strconv.ParseInt(r.ProjectID, 10, 64)
	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		if id, err := strconv.ParseInt(r.WorkspaceID, 10, 64); err == nil {
			wsID = id
		}
	}

	dueDate, _ := time.Parse("2006-01-02", r.TargetDate)
	if dueDate.IsZero() {
		dueDate = time.Now().AddDate(0, 1, 0)
	}

	milestone, err := h.FlowService.CreateMilestone(ctx, domain.MilestoneInput{
		ProjectID:   projID,
		WorkspaceID: wsID,
		Name:        r.Name,
		Description: r.Description,
		DueDate:     dueDate,
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return CreateMilestoneResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return CreateMilestoneResponse{Milestone: milestonePayload{
		ID:          strconv.FormatInt(milestone.ID, 10),
		ProjectID:   strconv.FormatInt(milestone.ProjectID, 10),
		Name:        milestone.Name,
		TargetDate:  milestone.DueDate.Format("2006-01-02"),
		Status:      milestone.Status,
		Description: milestone.Description,
	}}, nil
}

// ─── Update Milestone ─────────────────────────────────────────────────────────

func UpdateMilestoneRoutes() []string { return []string{"/flow/milestones/{id}"} }

type UpdateMilestoneRequest struct {
	ID          string `path:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	TargetDate  string `json:"targetDate"`
}

type UpdateMilestoneResponse struct{}

func (r *UpdateMilestoneRequest) Validate() *errorModels.ServerError { return nil }

type UpdateMilestoneHandler struct{ FlowService *application.Service }

func (h *UpdateMilestoneHandler) Serve(ctx context.Context, r *UpdateMilestoneRequest) (UpdateMilestoneResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateMilestoneResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := strconv.ParseInt(r.ID, 10, 64)
	if err != nil || id <= 0 {
		return UpdateMilestoneResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid milestone id"}
	}
	dueDate, _ := time.Parse("2006-01-02", r.TargetDate)

	if err := h.FlowService.UpdateMilestone(ctx, domain.UpdateMilestoneInput{
		ID:          id,
		ActorID:     authUser.UserID,
		Name:        r.Name,
		Description: r.Description,
		Status:      r.Status,
		DueDate:     dueDate,
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return UpdateMilestoneResponse{}, errorModels.PermissionDeniedServerError()
		}
		return UpdateMilestoneResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return UpdateMilestoneResponse{}, nil
}
