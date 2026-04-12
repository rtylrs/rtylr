package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── Session payload ──────────────────────────────────────────────────────────

func sessionToProto(s domain.Session) *posv1.Session {
	p := &posv1.Session{
		Id:             s.ID,
		LocationId:     s.LocationID,
		OpenedBy:       s.StaffID,
		Status:         string(s.Status),
		OpeningCash:    s.OpeningCash,
		ClosingCash:    s.ClosingCash,
		ExpectedCash:   s.ExpectedCash,
		CashDifference: s.CashDifference,
		TotalOrders:    int32(s.TotalOrders),
		TotalSales:     s.TotalSales,
		TotalRefunds:   s.TotalRefunds,
		OpenedAt:       s.OpenedAt.Format(time.RFC3339),
	}
	if s.ClosedAt != nil {
		p.ClosedAt = s.ClosedAt.Format(time.RFC3339)
	}
	_ = fmt.Sprintf // ensure fmt is used
	_ = errors.New  // ensure errors is used
	return p
}

// ─── Get Current Session ──────────────────────────────────────────────────────

func GetCurrentSessionRoutes() []string { return []string{"/pos/sessions/current"} }

type GetCurrentSessionHandler struct{ Service *application.Service }

type GetCurrentSessionRequest struct{}

type GetCurrentSessionResponse struct {
	Session *posv1.Session `json:"session"`
}

func (h *GetCurrentSessionHandler) Serve(ctx context.Context, _ *GetCurrentSessionRequest) (GetCurrentSessionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCurrentSessionResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID
	staffID := device.User.UserID

	sess, err := h.Service.GetCurrentOpenSession(ctx, ownerID, staffID)
	if err != nil {
		return GetCurrentSessionResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get current session",
		}
	}

	if sess == nil {
		return GetCurrentSessionResponse{Session: nil}, nil
	}

	payload := sessionToProto(*sess)

	return GetCurrentSessionResponse{Session: payload}, nil
}

// ─── Open Session ─────────────────────────────────────────────────────────────

func OpenSessionRoutes() []string { return []string{"/pos/sessions/open"} }

type OpenSessionHandler struct{ Service *application.Service }

type OpenSessionRequest struct {
	LocationID  int64   `json:"location_id"`
	OpeningCash float64 `json:"opening_cash"`
}

type OpenSessionResponse struct {
	SessionID int64  `json:"session_id"`
	Message   string `json:"message"`
}

func (h *OpenSessionHandler) Validate(_ context.Context, req *OpenSessionRequest) *errorModels.ServerError {
	if req.OpeningCash < 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Opening cash cannot be negative",
		}
	}

	return nil
}

func (h *OpenSessionHandler) Serve(ctx context.Context, req *OpenSessionRequest) (OpenSessionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return OpenSessionResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID
	staffID := device.User.UserID

	id, err := h.Service.OpenSession(ctx, ownerID, staffID, domain.OpenSessionInput{
		LocationID:  req.LocationID,
		OpeningCash: req.OpeningCash,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSessionAlreadyOpen) {
			return OpenSessionResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusConflict,
				Code:     http.StatusConflict,
				Message:  "You already have an open session",
			}
		}

		return OpenSessionResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to open session",
		}
	}

	return OpenSessionResponse{SessionID: id, Message: "Session opened"}, nil
}

// ─── Close Session ────────────────────────────────────────────────────────────

func CloseSessionRoutes() []string { return []string{"/pos/sessions/close"} }

type CloseSessionHandler struct{ Service *application.Service }

type CloseSessionRequest struct {
	SessionID   int64   `json:"session_id"`
	ClosingCash float64 `json:"closing_cash"`
	Notes       string  `json:"notes"`
}

type CloseSessionResponse struct {
	Message string `json:"message"`
}

func (h *CloseSessionHandler) Validate(_ context.Context, req *CloseSessionRequest) *errorModels.ServerError {
	if req.SessionID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Session id is required",
		}
	}

	return nil
}

func (h *CloseSessionHandler) Serve(ctx context.Context, req *CloseSessionRequest) (CloseSessionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CloseSessionResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	if err := h.Service.CloseSession(ctx, ownerID, domain.CloseSessionInput{
		SessionID:   req.SessionID,
		ClosingCash: req.ClosingCash,
		Notes:       req.Notes,
	}); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return CloseSessionResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Session not found",
			}
		}

		if errors.Is(err, domain.ErrSessionNotOpen) {
			return CloseSessionResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Session is not open",
			}
		}

		return CloseSessionResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to close session",
		}
	}

	return CloseSessionResponse{Message: "Session closed"}, nil
}

// ─── List Sessions ────────────────────────────────────────────────────────────

func ListSessionsRoutes() []string { return []string{"/pos/sessions"} }

type ListSessionsHandler struct{ Service *application.Service }

type ListSessionsRequest struct {
	LocationID int64  `json:"location_id"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
}

type ListSessionsResponse struct {
	Sessions []*posv1.Session `json:"sessions"`
}

func (h *ListSessionsHandler) Serve(ctx context.Context, req *ListSessionsRequest) (ListSessionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListSessionsResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	sessions, err := h.Service.ListSessions(ctx, ownerID, req.LocationID, req.DateFrom, req.DateTo)
	if err != nil {
		return ListSessionsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list sessions",
		}
	}

	payloads := make([]*posv1.Session, 0, len(sessions))
	for _, s := range sessions {
		payloads = append(payloads, sessionToProto(s))
	}

	return ListSessionsResponse{Sessions: payloads}, nil
}
