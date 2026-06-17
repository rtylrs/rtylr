package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/application"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	posauditv1 "github.com/voxire/voxie/gen/go/pos/audit/v1"
)

func toProtoAuditLog(l domain.AuditLog) *posauditv1.AuditLog {
	p := &posauditv1.AuditLog{
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

type (
	ListAuditHandler struct{ Service *application.Service }
	ListAuditRequest struct {
		Action     *string `json:"action"`
		EntityType *string `json:"entity_type"`
		EmployeeID *int64  `json:"employee_id"`
		StartDate  *string `json:"start_date"`
		EndDate    *string `json:"end_date"`
	}
)

type ListAuditResponse struct {
	Logs []*posauditv1.AuditLog `json:"logs"`
}

func (h *ListAuditHandler) Validate(_ context.Context, req *ListAuditRequest) *errorModels.ServerError {
	const layout = "2006-01-02"
	if req.StartDate != nil {
		*req.StartDate = strings.TrimSpace(*req.StartDate)
		if _, err := time.Parse(layout, *req.StartDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid start_date format, expected YYYY-MM-DD"}
		}
	}
	if req.EndDate != nil {
		*req.EndDate = strings.TrimSpace(*req.EndDate)
		if _, err := time.Parse(layout, *req.EndDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid end_date format, expected YYYY-MM-DD"}
		}
	}
	if req.StartDate != nil && req.EndDate != nil {
		start, _ := time.Parse(layout, *req.StartDate)
		end, _ := time.Parse(layout, *req.EndDate)
		if end.Before(start) {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "end_date must not be before start_date"}
		}
	}
	return nil
}

func (h *ListAuditHandler) Serve(ctx context.Context, req *ListAuditRequest) (ListAuditResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListAuditResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID := device.User.BusinessID
	if businessID <= 0 {
		businessID = device.User.UserID
	}
	logs, err := h.Service.List(ctx, domain.AuditLogFilter{
		BusinessID: businessID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EmployeeID: req.EmployeeID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	})
	if err != nil {
		return ListAuditResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list audit logs"}
	}
	payloads := make([]*posauditv1.AuditLog, len(logs))
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
	req.Action = strings.TrimSpace(req.Action)
	if req.Action == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Action is required"}
	}
	req.EntityType = strings.TrimSpace(req.EntityType)
	if req.EntityType == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Entity type is required"}
	}
	if req.BranchID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Branch ID is required"}
	}
	if req.EmployeeID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Employee ID is required"}
	}
	return nil
}

func (h *CreateAuditHandler) Serve(ctx context.Context, req *CreateAuditRequest) (CreateAuditResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateAuditResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID := device.User.BusinessID
	if businessID <= 0 {
		businessID = device.User.UserID
	}
	id, err := h.Service.Create(ctx, domain.AuditLogInput{
		BusinessID: businessID,
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
