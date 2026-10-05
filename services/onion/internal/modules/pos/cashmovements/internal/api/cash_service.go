package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos/cashmovements/internal/application"
	"github.com/voxire/onion/internal/modules/pos/cashmovements/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	poscashv1 "github.com/voxire/voxie/gen/go/pos/cash/v1"
)

type CashAPI struct {
	Service *application.Service
}

func NewCashAPI(service *application.Service) *CashAPI {
	return &CashAPI{Service: service}
}

func (a *CashAPI) ListCashMovements(ctx context.Context, req *poscashv1.ListCashMovementsRequest) (*poscashv1.ListCashMovementsResponse, error) {
	user, err := requireCashPermission(ctx, "pos:cash:view")
	if err != nil {
		return nil, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	movements, listErr := a.Service.ListBySession(ctx, businessID, req.GetSessionId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list cash movements")
	}

	payloads := make([]*poscashv1.CashMovement, 0, len(movements))
	for _, movement := range movements {
		payloads = append(payloads, cashMovementToProto(movement))
	}

	return &poscashv1.ListCashMovementsResponse{Movements: payloads}, nil
}

func (a *CashAPI) CashDrop(ctx context.Context, req *poscashv1.CashDropRequest) (*poscashv1.CashDropResponse, error) {
	user, err := requireCashPermission(ctx, "pos:cash:drop")
	if err != nil {
		return nil, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, createErr := a.Service.CashDrop(ctx, cashMovementInput(businessID, req.GetBranchId(), req.GetTerminalId(), req.GetSessionId(), user.UserID, req.GetAmount(), req.GetReason()))
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create cash drop")
	}

	return &poscashv1.CashDropResponse{MovementId: id, Message: "Cash drop recorded successfully"}, nil
}

func (a *CashAPI) CashPickup(ctx context.Context, req *poscashv1.CashPickupRequest) (*poscashv1.CashPickupResponse, error) {
	user, err := requireCashPermission(ctx, "pos:cash:pickup")
	if err != nil {
		return nil, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, createErr := a.Service.CashPickup(ctx, cashMovementInput(businessID, req.GetBranchId(), req.GetTerminalId(), req.GetSessionId(), user.UserID, req.GetAmount(), req.GetReason()))
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create cash pickup")
	}

	return &poscashv1.CashPickupResponse{MovementId: id, Message: "Cash pickup recorded successfully"}, nil
}

func (a *CashAPI) CashAdjust(ctx context.Context, req *poscashv1.CashAdjustRequest) (*poscashv1.CashAdjustResponse, error) {
	user, err := requireCashPermission(ctx, "pos:cash:approve")
	if err != nil {
		return nil, err
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if strings.TrimSpace(req.GetReason()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "reason is required for adjustments")
	}

	input := cashMovementInput(businessID, req.GetBranchId(), req.GetTerminalId(), req.GetSessionId(), user.UserID, req.GetAmount(), req.GetReason())
	input.ApprovedBy = &user.UserID
	id, createErr := a.Service.CashAdjust(ctx, input)
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create cash adjustment")
	}

	return &poscashv1.CashAdjustResponse{MovementId: id, Message: "Cash adjustment recorded successfully"}, nil
}

func requireCashPermission(ctx context.Context, permission string) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppPOS, permission) || !authPkg.HasAnyAppScope(device.User, authorization.AppPOS) {
		return nil, onekiterrors.PermissionDenied()
	}
	return device.User, nil
}

func cashMovementInput(businessID, branchID, terminalID, sessionID, employeeID int64, amount float64, reason string) domain.CreateCashMovementInput {
	return domain.CreateCashMovementInput{
		BusinessID: businessID,
		BranchID:   branchID,
		TerminalID: ptrInt64(terminalID),
		SessionID:  sessionID,
		EmployeeID: employeeID,
		Amount:     amount,
		Reason:     ptrString(reason),
	}
}

func cashMovementToProto(movement domain.CashMovement) *poscashv1.CashMovement {
	return &poscashv1.CashMovement{
		Id:         movement.ID,
		BusinessId: movement.BusinessID,
		BranchId:   movement.BranchID,
		TerminalId: movement.TerminalID,
		SessionId:  movement.SessionID,
		EmployeeId: movement.EmployeeID,
		Type:       movement.Type,
		Amount:     movement.Amount,
		Reason:     movement.Reason,
		ApprovedBy: movement.ApprovedBy,
		CreatedAt:  movement.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ptrInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func ptrString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
