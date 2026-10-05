package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
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

// ─── Get Current Session ──────────────────────────────────────────────────────

type GetCurrentSessionHandler struct{ Service *application.Service }

type GetCurrentSessionRequest struct{}

type GetCurrentSessionResponse struct {
	Session *SessionPayload `json:"session"`
}

// ─── Open Session ─────────────────────────────────────────────────────────────

type OpenSessionHandler struct{ Service *application.Service }

type OpenSessionRequest struct {
	LocationID  int64   `json:"location_id"`
	OpeningCash float64 `json:"opening_cash"`
}

type OpenSessionResponse struct {
	SessionID int64  `json:"session_id"`
	Message   string `json:"message"`
}

// ─── Close Session ────────────────────────────────────────────────────────────

type CloseSessionHandler struct{ Service *application.Service }

type CloseSessionRequest struct {
	SessionID   int64   `json:"session_id"`
	ClosingCash float64 `json:"closing_cash"`
	Notes       string  `json:"notes"`
}

type CloseSessionResponse struct {
	Message string `json:"message"`
}

// ─── List Sessions ────────────────────────────────────────────────────────────

type ListSessionsHandler struct{ Service *application.Service }

type ListSessionsRequest struct {
	LocationID int64  `json:"location_id"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
}

type ListSessionsResponse struct {
	Sessions []SessionPayload `json:"sessions"`
}
