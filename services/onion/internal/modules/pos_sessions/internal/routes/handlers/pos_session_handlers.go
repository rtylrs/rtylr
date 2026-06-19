package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/application"
	"github.com/voxire/onion/internal/modules/pos_sessions/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type SessionPayload struct {
	ID           int64   `json:"id"`
	OwnerID      int64   `json:"owner_id"`
	LocationID   *int64  `json:"location_id"`
	StaffID      *int64  `json:"staff_id"`
	OpeningFloat float64 `json:"opening_float"`
	ClosingFloat float64 `json:"closing_float"`
	TotalSales   float64 `json:"total_sales"`
	Status       string  `json:"status"`
	OpenedAt     string  `json:"opened_at"`
	ClosedAt     *string `json:"closed_at"`
	CreatedAt    string  `json:"created_at"`
}

func newSessionPayload(s domain.Session) SessionPayload {
	p := SessionPayload{
		ID:           s.ID,
		OwnerID:      s.OwnerID,
		LocationID:   s.LocationID,
		StaffID:      s.StaffID,
		OpeningFloat: s.OpeningFloat,
		ClosingFloat: s.ClosingFloat,
		TotalSales:   s.TotalSales,
		Status:       s.Status,
		OpenedAt:     s.OpenedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.ClosedAt != nil {
		t := s.ClosedAt.Format("2006-01-02T15:04:05Z")
		p.ClosedAt = &t
	}
	return p
}

// ─── List ─────────────────────────────────────────────────────────────────────

func SessionListRoutes() []string { return []string{"/pos/sessions"} }

type (
	ListSessionsHandler  struct{ Service *application.Service }
	ListSessionsRequest  struct{}
	ListSessionsResponse struct {
		Sessions []SessionPayload `json:"sessions"`
	}
)

func (h *ListSessionsHandler) Serve(ctx context.Context, _ *ListSessionsRequest) (ListSessionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListSessionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListSessionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	sessions, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListSessionsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list sessions"}
	}
	payloads := make([]SessionPayload, len(sessions))
	for i, s := range sessions {
		payloads[i] = newSessionPayload(s)
	}
	return ListSessionsResponse{Sessions: payloads}, nil
}

// ─── Get Session Detail ──────────────────────────────────────────────────────

func SessionGetRoutes() []string { return []string{"/pos/sessions/{id}"} }

type (
	GetSessionHandler  struct{ Service *application.Service }
	GetSessionRequest  struct{}
	GetSessionResponse struct {
		Session *SessionPayload `json:"session"`
	}
)

func (h *GetSessionHandler) Serve(ctx context.Context, _ *GetSessionRequest) (GetSessionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSessionResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetSessionResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return GetSessionResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid session ID"}
	}
	session, getErr := h.Service.Get(ctx, businessID, id)
	if getErr != nil {
		if errors.Is(getErr, domain.ErrSessionNotFound) {
			return GetSessionResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Session not found"}
		}
		return GetSessionResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get session"}
	}
	payload := newSessionPayload(*session)
	return GetSessionResponse{Session: &payload}, nil
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return OpenSessionResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Open(ctx, domain.OpenSessionInput{
		OwnerID:      device.User.UserID,
		BusinessID:   businessID,
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
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
		BusinessID:   businessID,
		ClosingFloat: req.ClosingFloat,
	}); closeErr != nil {
		if errors.Is(closeErr, domain.ErrSessionNotFound) {
			return CloseSessionResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Session not found or already closed"}
		}
		return CloseSessionResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to close session"}
	}
	return CloseSessionResponse{Message: "Session closed successfully"}, nil
}
