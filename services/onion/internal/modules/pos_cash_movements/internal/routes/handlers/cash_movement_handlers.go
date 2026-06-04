package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/application"
	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	poscashv1 "github.com/voxire/voxie/gen/go/pos/cash/v1"
)

// ─── List Cash Movements ─────────────────────────────────────────────────────

func ListCashMovementsRoutes() []string { return []string{"/pos/cash/movements"} }

type (
	ListCashMovementsHandler struct{ Service *application.Service }
	ListCashMovementsRequest struct {
		SessionID int64 `json:"session_id"`
	}
)

type ListCashMovementsResponse struct {
	Movements []*poscashv1.CashMovement `json:"movements"`
}

func (h *ListCashMovementsHandler) Serve(ctx context.Context, req *ListCashMovementsRequest) (ListCashMovementsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCashMovementsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.SessionID == 0 {
		return ListCashMovementsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "session_id is required"}
	}
	movements, err := h.Service.ListBySession(ctx, device.User.BusinessID, req.SessionID)
	if err != nil {
		return ListCashMovementsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list cash movements"}
	}
	payloads := make([]*poscashv1.CashMovement, len(movements))
	for i, m := range movements {
		payloads[i] = &poscashv1.CashMovement{
			Id:         m.ID,
			BusinessId: m.BusinessID,
			BranchId:   m.BranchID,
			TerminalId: m.TerminalID,
			SessionId:  m.SessionID,
			EmployeeId: m.EmployeeID,
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
		BusinessID: device.User.BusinessID,
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
		BusinessID: device.User.BusinessID,
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
	if req.Amount == 0 {
		return CashAdjustResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "amount must be non-zero for adjustments"}
	}
	if req.Reason == nil || strings.TrimSpace(*req.Reason) == "" {
		return CashAdjustResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "reason is required for adjustments"}
	}
	id, err := h.Service.CashAdjust(ctx, domain.CreateCashMovementInput{
		BusinessID: device.User.BusinessID,
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
