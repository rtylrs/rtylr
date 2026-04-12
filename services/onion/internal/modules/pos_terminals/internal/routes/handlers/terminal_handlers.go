package handlers

import (
	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/application"
	"github.com/voxire/onion/internal/modules/pos_terminals/internal/domain"
)

func terminalToProto(t domain.Terminal) *posv1.Terminal {
	p := &posv1.Terminal{
		Id:       t.ID,
		Name:     t.Name,
		IsActive: t.IsActive,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if t.DeviceID != nil {
		p.DeviceId = *t.DeviceID
	}
	if t.LastSeenAt != nil {
		p.LastHeartbeatAt = t.LastSeenAt.Format("2006-01-02T15:04:05Z")
	}
	return p
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TerminalListRoutes() []string { return []string{"/pos/terminals"} }

type ListTerminalsHandler struct{ Service *application.Service }
type ListTerminalsRequest struct{}
type ListTerminalsResponse struct {
	Terminals []*posv1.Terminal `json:"terminals"`
}

func (h *ListTerminalsHandler) Serve(ctx context.Context, _ *ListTerminalsRequest) (ListTerminalsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTerminalsResponse{}, errorModels.PermissionDeniedServerError()
	}
	terminals, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListTerminalsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list terminals"}
	}
	payloads := make([]*posv1.Terminal, len(terminals))
	for i, t := range terminals {
		payloads[i] = terminalToProto(t)
	}
	return ListTerminalsResponse{Terminals: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TerminalCreateRoutes() []string { return []string{"/pos/terminals"} }

type CreateTerminalHandler struct{ Service *application.Service }

type CreateTerminalRequest struct {
	BranchID int64   `json:"branch_id"`
	Name     string  `json:"name"`
	DeviceID *string `json:"device_id"`
}

type CreateTerminalResponse struct {
	TerminalID int64  `json:"terminal_id"`
	Message    string `json:"message"`
}

func (h *CreateTerminalHandler) Validate(_ context.Context, req *CreateTerminalRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.BranchID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Branch ID is required"}
	}
	return nil
}

func (h *CreateTerminalHandler) Serve(ctx context.Context, req *CreateTerminalRequest) (CreateTerminalResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTerminalResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.TerminalInput{
		BusinessID: device.User.UserID,
		BranchID:   req.BranchID,
		Name:       req.Name,
		DeviceID:   req.DeviceID,
	})
	if err != nil {
		return CreateTerminalResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create terminal"}
	}
	return CreateTerminalResponse{TerminalID: id, Message: "Terminal created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TerminalUpdateRoutes() []string { return []string{"/pos/terminals/{id}"} }

type UpdateTerminalHandler struct{ Service *application.Service }

type UpdateTerminalRequest struct {
	BranchID int64   `json:"branch_id"`
	Name     string  `json:"name"`
	DeviceID *string `json:"device_id"`
	IsActive bool    `json:"is_active"`
}

type UpdateTerminalResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTerminalHandler) Validate(_ context.Context, req *UpdateTerminalRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateTerminalHandler) Serve(ctx context.Context, req *UpdateTerminalRequest) (UpdateTerminalResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTerminalResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateTerminalResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid terminal ID"}
	}
	if updateErr := h.Service.Update(ctx, domain.TerminalUpdateInput{
		ID:         id,
		BusinessID: device.User.UserID,
		BranchID:   req.BranchID,
		Name:       req.Name,
		DeviceID:   req.DeviceID,
		IsActive:   req.IsActive,
	}); updateErr != nil {
		return UpdateTerminalResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update terminal"}
	}
	return UpdateTerminalResponse{Message: "Terminal updated successfully"}, nil
}

// ─── Heartbeat ────────────────────────────────────────────────────────────────

func TerminalHeartbeatRoutes() []string { return []string{"/pos/terminals/heartbeat"} }

type HeartbeatTerminalHandler struct{ Service *application.Service }

type HeartbeatTerminalRequest struct {
	TerminalID int64 `json:"terminal_id"`
}

type HeartbeatTerminalResponse struct {
	Message string `json:"message"`
}

func (h *HeartbeatTerminalHandler) Validate(_ context.Context, req *HeartbeatTerminalRequest) *errorModels.ServerError {
	if req.TerminalID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Terminal ID is required"}
	}
	return nil
}

func (h *HeartbeatTerminalHandler) Serve(ctx context.Context, req *HeartbeatTerminalRequest) (HeartbeatTerminalResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return HeartbeatTerminalResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Heartbeat(ctx, req.TerminalID, device.User.UserID); err != nil {
		return HeartbeatTerminalResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update heartbeat"}
	}
	return HeartbeatTerminalResponse{Message: "Heartbeat updated successfully"}, nil
}
