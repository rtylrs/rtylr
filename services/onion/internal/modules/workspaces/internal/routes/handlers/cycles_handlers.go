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

// ─── List Cycles ──────────────────────────────────────────────────────────────

func ListCyclesRoutes() []string { return []string{"/flow/cycles"} }

type ListCyclesRequest struct {
	WorkspaceID string `query:"workspace_id"`
}

type ListCyclesResponse struct {
	Cycles []cyclePayload `json:"cycles"`
}

type cyclePayload struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Status      string `json:"status"`
	Scope       int    `json:"scope"`
	Completed   int    `json:"completed"`
}

func (r *ListCyclesRequest) Validate() *errorModels.ServerError { return nil }

type ListCyclesHandler struct{ FlowService *application.Service }

func (h *ListCyclesHandler) Serve(ctx context.Context, r *ListCyclesRequest) (ListCyclesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListCyclesResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		resolved, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, r.WorkspaceID)
		if err != nil {
			return ListCyclesResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid workspace_id"}
		}
		if resolved != 0 {
			wsID = resolved
		}
	}

	cycles, err := h.FlowService.ListCycles(ctx, wsID)
	if err != nil {
		return ListCyclesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	payloads := make([]cyclePayload, 0, len(cycles))
	for _, c := range cycles {
		payloads = append(payloads, cyclePayload{
			ID:          strconv.FormatInt(c.ID, 10),
			WorkspaceID: strconv.FormatInt(c.WorkspaceID, 10),
			Name:        c.Name,
			StartDate:   c.StartDate.Format("2006-01-02"),
			EndDate:     c.EndDate.Format("2006-01-02"),
			Status:      c.Status,
			Scope:       c.Scope,
			Completed:   c.Completed,
		})
	}
	return ListCyclesResponse{Cycles: payloads}, nil
}

// ─── Create Cycle ─────────────────────────────────────────────────────────────

func CreateCycleRoutes() []string { return []string{"/flow/cycles"} }

type CreateCycleRequest struct {
	Name        string `json:"name"`
	WorkspaceID string `json:"workspace_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type CreateCycleResponse struct {
	Cycle cyclePayload `json:"cycle"`
}

func (r *CreateCycleRequest) Validate() *errorModels.ServerError {
	if r.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

type CreateCycleHandler struct{ FlowService *application.Service }

func (h *CreateCycleHandler) Serve(ctx context.Context, r *CreateCycleRequest) (CreateCycleResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateCycleResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		resolved, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, r.WorkspaceID)
		if err != nil {
			return CreateCycleResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid workspace_id"}
		}
		if resolved != 0 {
			wsID = resolved
		}
	}

	start, _ := time.Parse("2006-01-02", r.StartDate)
	end, _ := time.Parse("2006-01-02", r.EndDate)
	if end.IsZero() {
		end = start.AddDate(0, 0, 14)
	}

	cycle, err := h.FlowService.CreateCycle(ctx, domain.CycleInput{
		WorkspaceID: wsID,
		Name:        r.Name,
		StartDate:   start,
		EndDate:     end,
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return CreateCycleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return CreateCycleResponse{Cycle: cyclePayload{
		ID:          strconv.FormatInt(cycle.ID, 10),
		WorkspaceID: strconv.FormatInt(cycle.WorkspaceID, 10),
		Name:        cycle.Name,
		StartDate:   cycle.StartDate.Format("2006-01-02"),
		EndDate:     cycle.EndDate.Format("2006-01-02"),
		Status:      cycle.Status,
	}}, nil
}

// ─── Update Cycle ─────────────────────────────────────────────────────────────

func UpdateCycleRoutes() []string { return []string{"/flow/cycles/{id}"} }

type UpdateCycleRequest struct {
	ID      string `path:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	EndDate string `json:"end_date"`
}

type UpdateCycleResponse struct{}

func (r *UpdateCycleRequest) Validate() *errorModels.ServerError { return nil }

type UpdateCycleHandler struct{ FlowService *application.Service }

func (h *UpdateCycleHandler) Serve(ctx context.Context, r *UpdateCycleRequest) (UpdateCycleResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateCycleResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := strconv.ParseInt(r.ID, 10, 64)
	if err != nil || id <= 0 {
		return UpdateCycleResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid cycle id"}
	}
	end, _ := time.Parse("2006-01-02", r.EndDate)

	if err := h.FlowService.UpdateCycle(ctx, domain.UpdateCycleInput{
		ID:      id,
		ActorID: authUser.UserID,
		Name:    r.Name,
		Status:  r.Status,
		EndDate: end,
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return UpdateCycleResponse{}, errorModels.PermissionDeniedServerError()
		}
		return UpdateCycleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return UpdateCycleResponse{}, nil
}
