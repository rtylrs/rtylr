package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/application"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func toProtoAuditLog(l domain.AuditLog) *posv1.AuditLog {
	p := &posv1.AuditLog{
		Id:         l.ID,
		BusinessId: l.BusinessID,
		BranchId:   l.BranchID,
		TerminalId: l.TerminalID,
		EmployeeId: l.EmployeeID,
		Action:     l.Action,
		EntityType: l.EntityType,
		EntityId:   l.EntityID,
		IpAddress:  l.IPAddress,
		CreatedAt:  l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if l.Details != nil {
		s := string(*l.Details)
		p.Details = &s
	}
	return p
}

// ─── List ─────────────────────────────────────────────────────────────────────

func AuditListRoutes() []string { return []string{"/pos/audit"} }

type ListAuditHandler struct{ Service *application.Service }
type ListAuditRequest struct {
	Action     *string `json:"action"`
	EntityType *string `json:"entity_type"`
	EmployeeID *int64  `json:"employee_id"`
	StartDate  *string `json:"start_date"`
	EndDate    *string `json:"end_date"`
}
type ListAuditResponse struct {
	Logs []*posv1.AuditLog `json:"logs"`
}

func (h *ListAuditHandler) Serve(ctx context.Context, req *ListAuditRequest) (ListAuditResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListAuditResponse{}, errorModels.PermissionDeniedServerError()
	}
	logs, err := h.Service.List(ctx, domain.AuditLogFilter{
		BusinessID: device.User.UserID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EmployeeID: req.EmployeeID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	})
	if err != nil {
		return ListAuditResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list audit logs"}
	}
	payloads := make([]*posv1.AuditLog, len(logs))
	for i, l := range logs {
		payloads[i] = toProtoAuditLog(l)
	}
	return ListAuditResponse{Logs: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func AuditCreateRoutes() []string { return []string{"/pos/audit"} }

type CreateAuditHandler struct{ Service *application.Service }

type CreateAuditRequest struct {
	BranchID   int64            `json:"branch_id"`
	TerminalID *int64           `json:"terminal_id"`
	EmployeeID int64            `json:"employee_id"`
	Action     string           `json:"action"`
	EntityType string           `json:"entity_type"`
	EntityID   *int64           `json:"entity_id"`
	Details    *json.RawMessage `json:"details"`
	IPAddress  *string          `json:"ip_address"`
}

type CreateAuditResponse struct {
	AuditLogID int64  `json:"audit_log_id"`
	Message    string `json:"message"`
}

func (h *CreateAuditHandler) Validate(_ context.Context, req *CreateAuditRequest) *errorModels.ServerError {
	if req.Action == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Action is required"}
	}
	if req.EntityType == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Entity type is required"}
	}
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Employee ID is required"}
	}
	return nil
}

func (h *CreateAuditHandler) Serve(ctx context.Context, req *CreateAuditRequest) (CreateAuditResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateAuditResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.AuditLogInput{
		BusinessID: device.User.UserID,
		BranchID:   req.BranchID,
		TerminalID: req.TerminalID,
		EmployeeID: req.EmployeeID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Details:    req.Details,
		IPAddress:  req.IPAddress,
	})
	if err != nil {
		return CreateAuditResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create audit log"}
	}
	return CreateAuditResponse{AuditLogID: id, Message: "Audit log created successfully"}, nil
}
