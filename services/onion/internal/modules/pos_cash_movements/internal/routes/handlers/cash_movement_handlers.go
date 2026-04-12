package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/application"
	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/domain"
)

// ─── List Cash Movements ─────────────────────────────────────────────────────

func ListCashMovementsRoutes() []string { return []string{"/pos/cash/movements"} }

type ListCashMovementsHandler struct{ Service *application.Service }
type ListCashMovementsRequest struct {
	SessionID int64 `json:"session_id"`
}
type ListCashMovementsResponse struct {
	Movements []CashMovementPayload `json:"movements"`
}

type CashMovementPayload struct {
	ID         int64   `json:"id"`
	BusinessID int64   `json:"business_id"`
	BranchID   int64   `json:"branch_id"`
	TerminalID *int64  `json:"terminal_id"`
	SessionID  int64   `json:"session_id"`
	EmployeeID int64   `json:"employee_id"`
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	Reason     *string `json:"reason"`
	ApprovedBy *int64  `json:"approved_by"`
	CreatedAt  string  `json:"created_at"`
}

func (h *ListCashMovementsHandler) Serve(ctx context.Context, req *ListCashMovementsRequest) (ListCashMovementsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCashMovementsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.SessionID == 0 {
		return ListCashMovementsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "session_id is required"}
	}
	movements, err := h.Service.ListBySession(ctx, device.User.UserID, req.SessionID)
	if err != nil {
		return ListCashMovementsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list cash movements"}
	}
	payloads := make([]CashMovementPayload, len(movements))
	for i, m := range movements {
		payloads[i] = CashMovementPayload{
			ID:         m.ID,
			BusinessID: m.BusinessID,
			BranchID:   m.BranchID,
			TerminalID: m.TerminalID,
			SessionID:  m.SessionID,
			EmployeeID: m.EmployeeID,
			Type:       m.Type,
			Amount:     m.Amount,
			Reason:     m.Reason,
			ApprovedBy: m.ApprovedBy,
			CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return ListCashMovementsResponse{Movements: payloads}, nil
}

// ─── Cash Drop ───────────────────────────────────────────────────────────────

func CashDropRoutes() []string { return []string{"/pos/cash/drop"} }

type CashDropHandler struct{ Service *application.Service }

type CashDropRequest struct {
	BranchID   int64   `json:"branch_id"`
	TerminalID *int64  `json:"terminal_id"`
	SessionID  int64   `json:"session_id"`
	EmployeeID int64   `json:"employee_id"`
	Amount     float64 `json:"amount"`
	Reason     *string `json:"reason"`
}

type CashDropResponse struct {
	MovementID int64  `json:"movement_id"`
	Message    string `json:"message"`
}

func (h *CashDropHandler) Serve(ctx context.Context, req *CashDropRequest) (CashDropResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CashDropResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.SessionID == 0 || req.BranchID == 0 || req.Amount <= 0 {
		return CashDropResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "sessionId, branchId, and positive amount are required"}
	}
	id, err := h.Service.CashDrop(ctx, domain.CreateCashMovementInput{
		BusinessID: device.User.UserID,
		BranchID:   req.BranchID,
		TerminalID: req.TerminalID,
		SessionID:  req.SessionID,
		EmployeeID: req.EmployeeID,
		Amount:     req.Amount,
		Reason:     req.Reason,
	})
	if err != nil {
		return CashDropResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create cash drop"}
	}
	return CashDropResponse{MovementID: id, Message: "Cash drop recorded successfully"}, nil
}

// ─── Cash Pickup ─────────────────────────────────────────────────────────────

func CashPickupRoutes() []string { return []string{"/pos/cash/pickup"} }

type CashPickupHandler struct{ Service *application.Service }

type CashPickupRequest struct {
	BranchID   int64   `json:"branch_id"`
	TerminalID *int64  `json:"terminal_id"`
	SessionID  int64   `json:"session_id"`
	EmployeeID int64   `json:"employee_id"`
	Amount     float64 `json:"amount"`
	Reason     *string `json:"reason"`
}

type CashPickupResponse struct {
	MovementID int64  `json:"movement_id"`
	Message    string `json:"message"`
}

func (h *CashPickupHandler) Serve(ctx context.Context, req *CashPickupRequest) (CashPickupResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CashPickupResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.SessionID == 0 || req.BranchID == 0 || req.Amount <= 0 {
		return CashPickupResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "sessionId, branchId, and positive amount are required"}
	}
	id, err := h.Service.CashPickup(ctx, domain.CreateCashMovementInput{
		BusinessID: device.User.UserID,
		BranchID:   req.BranchID,
		TerminalID: req.TerminalID,
		SessionID:  req.SessionID,
		EmployeeID: req.EmployeeID,
		Amount:     req.Amount,
		Reason:     req.Reason,
	})
	if err != nil {
		return CashPickupResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create cash pickup"}
	}
	return CashPickupResponse{MovementID: id, Message: "Cash pickup recorded successfully"}, nil
}

// ─── Cash Adjust ─────────────────────────────────────────────────────────────

func CashAdjustRoutes() []string { return []string{"/pos/cash/adjust"} }

type CashAdjustHandler struct{ Service *application.Service }

type CashAdjustRequest struct {
	BranchID   int64   `json:"branch_id"`
	TerminalID *int64  `json:"terminal_id"`
	SessionID  int64   `json:"session_id"`
	EmployeeID int64   `json:"employee_id"`
	Amount     float64 `json:"amount"`
	Reason     *string `json:"reason"`
}

type CashAdjustResponse struct {
	MovementID int64  `json:"movement_id"`
	Message    string `json:"message"`
}

func (h *CashAdjustHandler) Serve(ctx context.Context, req *CashAdjustRequest) (CashAdjustResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CashAdjustResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.SessionID == 0 || req.BranchID == 0 {
		return CashAdjustResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "sessionId and branchId are required"}
	}
	if req.Reason == nil || *req.Reason == "" {
		return CashAdjustResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "reason is required for adjustments"}
	}
	id, err := h.Service.CashAdjust(ctx, domain.CreateCashMovementInput{
		BusinessID: device.User.UserID,
		BranchID:   req.BranchID,
		TerminalID: req.TerminalID,
		SessionID:  req.SessionID,
		EmployeeID: req.EmployeeID,
		Amount:     req.Amount,
		Reason:     req.Reason,
	})
	if err != nil {
		return CashAdjustResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create cash adjustment"}
	}
	return CashAdjustResponse{MovementID: id, Message: "Cash adjustment recorded successfully"}, nil
}
