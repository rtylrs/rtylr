package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// ─── Session payload ──────────────────────────────────────────────────────────

type SessionPayload struct {
	ID               int64      `json:"id"`
	OwnerID          int64      `json:"owner_id"`
	LocationID       int64      `json:"location_id"`
	StaffID          int64      `json:"staff_id"`
	Status           string     `json:"status"`
	OpeningCash      float64    `json:"opening_cash"`
	ClosingCash      float64    `json:"closing_cash"`
	ExpectedCash     float64    `json:"expected_cash"`
	CashDifference   float64    `json:"cash_difference"`
	TotalSales       float64    `json:"total_sales"`
	TotalOrders      int        `json:"total_orders"`
	TotalCashSales   float64    `json:"total_cash_sales"`
	TotalCardSales   float64    `json:"total_card_sales"`
	TotalOnlineSales float64    `json:"total_online_sales"`
	TotalRefunds     float64    `json:"total_refunds"`
	OpenedAt         time.Time  `json:"opened_at"`
	ClosedAt         *time.Time `json:"closed_at"`
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

func (h *GetCurrentSessionHandler) Serve(ctx context.Context, _ *GetCurrentSessionRequest) (GetCurrentSessionResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:view")
	if authErr != nil {
		return GetCurrentSessionResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return GetCurrentSessionResponse{}, bizErr
	}
	staffID := user.UserID

	sess, err := h.Service.GetCurrentOpenSession(ctx, businessID, staffID)
	if err != nil {
		return GetCurrentSessionResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get current session")
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
	LocationID  int64   `json:"location_id"`
	OpeningCash float64 `json:"opening_cash"`
}

type OpenSessionResponse struct {
	SessionID int64  `json:"session_id"`
	Message   string `json:"message"`
}

func (h *OpenSessionHandler) Validate(_ context.Context, req *OpenSessionRequest) error {
	if req.OpeningCash < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Opening cash cannot be negative")
	}

	return nil
}

func (h *OpenSessionHandler) Serve(ctx context.Context, req *OpenSessionRequest) (OpenSessionResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return OpenSessionResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return OpenSessionResponse{}, bizErr
	}
	staffID := user.UserID

	id, err := h.Service.OpenSession(ctx, businessID, staffID, domain.OpenSessionInput{
		LocationID:  req.LocationID,
		OpeningCash: req.OpeningCash,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSessionAlreadyOpen) {
			return OpenSessionResponse{}, onekiterrors.New(http.StatusConflict, "You already have an open session")
		}

		return OpenSessionResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to open session")
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

func (h *CloseSessionHandler) Validate(_ context.Context, req *CloseSessionRequest) error {
	if req.SessionID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Session id is required")
	}

	return nil
}

func (h *CloseSessionHandler) Serve(ctx context.Context, req *CloseSessionRequest) (CloseSessionResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return CloseSessionResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return CloseSessionResponse{}, bizErr
	}

	if err := h.Service.CloseSession(ctx, businessID, user.UserID, domain.CloseSessionInput{
		SessionID:   req.SessionID,
		ClosingCash: req.ClosingCash,
		Notes:       req.Notes,
	}); err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return CloseSessionResponse{}, onekiterrors.New(http.StatusNotFound, "Session not found")
		}

		if errors.Is(err, domain.ErrSessionNotOpen) {
			return CloseSessionResponse{}, onekiterrors.New(http.StatusBadRequest, "Session is not open")
		}

		return CloseSessionResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to close session")
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
	Sessions []SessionPayload `json:"sessions"`
}

func (h *ListSessionsHandler) Validate(_ context.Context, req *ListSessionsRequest) error {
	const layout = "2006-01-02"
	if req.DateFrom != "" {
		if _, err := time.Parse(layout, req.DateFrom); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "date_from must be in YYYY-MM-DD format")
		}
	}
	if req.DateTo != "" {
		if _, err := time.Parse(layout, req.DateTo); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "date_to must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func (h *ListSessionsHandler) Serve(ctx context.Context, req *ListSessionsRequest) (ListSessionsResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:view")
	if authErr != nil {
		return ListSessionsResponse{}, authErr
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ListSessionsResponse{}, bizErr
	}

	sessions, err := h.Service.ListSessions(ctx, businessID, req.LocationID, req.DateFrom, req.DateTo)
	if err != nil {
		return ListSessionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list sessions")
	}

	payloads := make([]SessionPayload, 0, len(sessions))
	for _, s := range sessions {
		payloads = append(payloads, newSessionPayload(s))
	}

	return ListSessionsResponse{Sessions: payloads}, nil
}
