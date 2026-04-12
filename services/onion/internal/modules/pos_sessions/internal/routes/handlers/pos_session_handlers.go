package handlers

import (
	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/application"
	"github.com/voxire/onion/internal/modules/pos_sessions/internal/domain"
)

func sessionToProto(s domain.Session) *posv1.Session {
	p := &posv1.Session{
		Id:          s.ID,
		Status:      s.Status,
		OpeningCash: s.OpeningFloat,
		ClosingCash: s.ClosingFloat,
		TotalSales:  s.TotalSales,
		OpenedAt:    s.OpenedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.LocationID != nil {
		p.LocationId = *s.LocationID
	}
	if s.StaffID != nil {
		p.OpenedBy = *s.StaffID
	}
	if s.ClosedAt != nil {
		p.ClosedAt = s.ClosedAt.Format("2006-01-02T15:04:05Z")
	}
	return p
}

// ─── List ─────────────────────────────────────────────────────────────────────

func SessionListRoutes() []string { return []string{"/pos/sessions"} }

type ListSessionsHandler struct{ Service *application.Service }
type ListSessionsRequest struct{}
type ListSessionsResponse struct {
	Sessions []*posv1.Session `json:"sessions"`
}

func (h *ListSessionsHandler) Serve(ctx context.Context, _ *ListSessionsRequest) (ListSessionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListSessionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	sessions, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListSessionsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list sessions"}
	}
	payloads := make([]*posv1.Session, len(sessions))
	for i, s := range sessions {
		payloads[i] = sessionToProto(s)
	}
	return ListSessionsResponse{Sessions: payloads}, nil
}

// ─── Open ─────────────────────────────────────────────────────────────────────

func SessionOpenRoutes() []string { return []string{"/pos/sessions/open"} }

type OpenSessionHandler struct{ Service *application.Service }

type OpenSessionRequest struct {
	LocationID   *int64  `json:"location_id"`
	StaffID      *int64  `json:"staff_id"`
	OpeningFloat float64 `json:"opening_float"`
}

type OpenSessionResponse struct {
	SessionID int64  `json:"session_id"`
	Message   string `json:"message"`
}

func (h *OpenSessionHandler) Serve(ctx context.Context, req *OpenSessionRequest) (OpenSessionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return OpenSessionResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Open(ctx, domain.OpenSessionInput{
		OwnerID:      device.User.UserID,
		LocationID:   req.LocationID,
		StaffID:      req.StaffID,
		OpeningFloat: req.OpeningFloat,
	})
	if err != nil {
		return OpenSessionResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to open session"}
	}
	return OpenSessionResponse{SessionID: id, Message: "Session opened successfully"}, nil
}

// ─── Close ────────────────────────────────────────────────────────────────────

func SessionCloseRoutes() []string { return []string{"/pos/sessions/{id}/close"} }

type CloseSessionHandler struct{ Service *application.Service }

type CloseSessionRequest struct {
	ClosingFloat float64 `json:"closing_float"`
}

type CloseSessionResponse struct {
	Message string `json:"message"`
}

func (h *CloseSessionHandler) Serve(ctx context.Context, req *CloseSessionRequest) (CloseSessionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CloseSessionResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return CloseSessionResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid session ID"}
	}
	if closeErr := h.Service.Close(ctx, domain.CloseSessionInput{
		ID:           id,
		OwnerID:      device.User.UserID,
		ClosingFloat: req.ClosingFloat,
	}); closeErr != nil {
		if errors.Is(closeErr, domain.ErrSessionNotFound) {
			return CloseSessionResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Session not found or already closed"}
		}
		return CloseSessionResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to close session"}
	}
	return CloseSessionResponse{Message: "Session closed successfully"}, nil
}
