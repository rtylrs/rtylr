package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── Session payload ──────────────────────────────────────────────────────────

type SessionPayload struct {
	ID               int64      `json:"id"`
	OwnerID          int64      `json:"ownerId"`
	LocationID       int64      `json:"locationId"`
	StaffID          int64      `json:"staffId"`
	Status           string     `json:"status"`
	OpeningCash      float64    `json:"openingCash"`
	ClosingCash      float64    `json:"closingCash"`
	ExpectedCash     float64    `json:"expectedCash"`
	CashDifference   float64    `json:"cashDifference"`
	TotalSales       float64    `json:"totalSales"`
	TotalOrders      int        `json:"totalOrders"`
	TotalCashSales   float64    `json:"totalCashSales"`
	TotalCardSales   float64    `json:"totalCardSales"`
	TotalOnlineSales float64    `json:"totalOnlineSales"`
	TotalRefunds     float64    `json:"totalRefunds"`
	OpenedAt         time.Time  `json:"openedAt"`
	ClosedAt         *time.Time `json:"closedAt"`
	Notes            string     `json:"notes"`
}

func newSessionPayload(s domain.Session) SessionPayload {
	return SessionPayload{
		ID:               s.ID,
		OwnerID:          s.OwnerID,
		LocationID:       s.LocationID,
		StaffID:          s.StaffID,
		Status:           string(s.Status),
		OpeningCash:      s.OpeningCash,
		ClosingCash:      s.ClosingCash,
		ExpectedCash:     s.ExpectedCash,
		CashDifference:   s.CashDifference,
		TotalSales:       s.TotalSales,
		TotalOrders:      s.TotalOrders,
		TotalCashSales:   s.TotalCashSales,
		TotalCardSales:   s.TotalCardSales,
		TotalOnlineSales: s.TotalOnlineSales,
		TotalRefunds:     s.TotalRefunds,
		OpenedAt:         s.OpenedAt,
		ClosedAt:         s.ClosedAt,
		Notes:            s.Notes,
	}
}

// ─── Get Current Session ──────────────────────────────────────────────────────

func GetCurrentSessionRoutes() []string { return []string{"/pos/sessions/current"} }

type GetCurrentSessionHandler struct{ Service *application.Service }

type GetCurrentSessionRequest struct{}

type GetCurrentSessionResponse struct {
	Session *SessionPayload `json:"session"`
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

	payload := newSessionPayload(*sess)

	return GetCurrentSessionResponse{Session: &payload}, nil
}

// ─── Open Session ─────────────────────────────────────────────────────────────

func OpenSessionRoutes() []string { return []string{"/pos/sessions/open"} }

type OpenSessionHandler struct{ Service *application.Service }

type OpenSessionRequest struct {
	LocationID  int64   `json:"locationId"`
	OpeningCash float64 `json:"openingCash"`
}

type OpenSessionResponse struct {
	SessionID int64  `json:"sessionId"`
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
	SessionID   int64   `json:"sessionId"`
	ClosingCash float64 `json:"closingCash"`
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
	LocationID int64  `json:"locationId"`
	DateFrom   string `json:"dateFrom"`
	DateTo     string `json:"dateTo"`
}

type ListSessionsResponse struct {
	Sessions []SessionPayload `json:"sessions"`
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

	payloads := make([]SessionPayload, 0, len(sessions))
	for _, s := range sessions {
		payloads = append(payloads, newSessionPayload(s))
	}

	return ListSessionsResponse{Sessions: payloads}, nil
}
