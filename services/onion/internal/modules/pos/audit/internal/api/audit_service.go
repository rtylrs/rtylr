package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos/audit/internal/application"
	"github.com/voxire/onion/internal/modules/pos/audit/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cursor"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	posauditv1 "github.com/voxire/voxie/gen/go/pos/audit/v1"
)

type AuditAPI struct {
	Service *application.Service
}

func NewAuditAPI(service *application.Service) *AuditAPI {
	return &AuditAPI{Service: service}
}

func (a *AuditAPI) ListAudit(ctx context.Context, req *posauditv1.ListAuditRequest) (*posauditv1.ListAuditResponse, error) {
	user, err := requireAuditPermission(ctx, "pos:reports:view")
	if err != nil {
		return nil, err
	}

	const layout = "2006-01-02"
	var startDate, endDate *string
	if s := strings.TrimSpace(req.GetStartDate()); s != "" {
		if _, parseErr := time.Parse(layout, s); parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD")
		}
		startDate = &s
	}
	if e := strings.TrimSpace(req.GetEndDate()); e != "" {
		if _, parseErr := time.Parse(layout, e); parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid end_date format, expected YYYY-MM-DD")
		}
		endDate = &e
	}
	if startDate != nil && endDate != nil {
		start, startErr := time.Parse(layout, *startDate)
		if startErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD")
		}
		end, endErr := time.Parse(layout, *endDate)
		if endErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid end_date format, expected YYYY-MM-DD")
		}
		if end.Before(start) {
			return nil, onekiterrors.New(http.StatusBadRequest, "end_date must not be before start_date")
		}
	}

	var action, entityType *string
	if v := req.GetAction(); v != "" {
		action = &v
	}
	if v := req.GetEntityType(); v != "" {
		entityType = &v
	}
	var employeeID *int64
	if v := req.GetEmployeeId(); v != 0 {
		employeeID = &v
	}

	businessID := user.BusinessID
	if businessID <= 0 {
		businessID = user.UserID
	}

	filter := domain.AuditLogFilter{
		BusinessID: businessID,
		Action:     action,
		EntityType: entityType,
		EmployeeID: employeeID,
		StartDate:  startDate,
		EndDate:    endDate,
		Limit:      req.GetLimit(),
	}
	if token := strings.TrimSpace(req.GetCursor()); token != "" {
		boundary, decodeErr := cursor.DecodeCreatedAtID(token)
		if decodeErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid cursor")
		}
		filter.CursorCreatedAt = &boundary.CreatedAt
		filter.CursorID = &boundary.ID
	}

	page, listErr := a.Service.ListPage(ctx, filter)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list audit logs")
	}

	payloads := make([]*posauditv1.AuditLog, len(page.Logs))
	for i, l := range page.Logs {
		payloads[i] = toProtoAuditLog(l)
	}
	return &posauditv1.ListAuditResponse{Logs: payloads, NextCursor: page.NextCursor}, nil
}

func (a *AuditAPI) CreateAudit(ctx context.Context, req *posauditv1.CreateAuditRequest) (*posauditv1.CreateAuditResponse, error) {
	user, err := requireAuditPermission(ctx, "pos:reports:view")
	if err != nil {
		return nil, err
	}

	action := strings.TrimSpace(req.GetAction())
	if action == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Action is required")
	}
	entityType := strings.TrimSpace(req.GetEntityType())
	if entityType == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Entity type is required")
	}
	if req.GetBranchId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Branch ID is required")
	}
	if req.GetEmployeeId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Employee ID is required")
	}

	var terminalID, entityID *int64
	if v := req.GetTerminalId(); v != 0 {
		terminalID = &v
	}
	if v := req.GetEntityId(); v != 0 {
		entityID = &v
	}
	var details *json.RawMessage
	if v := req.GetDetails(); v != "" {
		raw := json.RawMessage(v)
		details = &raw
	}
	var ipAddress *string
	if v := req.GetIpAddress(); v != "" {
		ipAddress = &v
	}

	businessID := user.BusinessID
	if businessID <= 0 {
		businessID = user.UserID
	}

	id, createErr := a.Service.Create(ctx, domain.AuditLogInput{
		BusinessID: businessID,
		BranchID:   req.GetBranchId(),
		TerminalID: terminalID,
		EmployeeID: req.GetEmployeeId(),
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		IPAddress:  ipAddress,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create audit log")
	}
	return &posauditv1.CreateAuditResponse{AuditLogId: id, Message: "Audit log created successfully"}, nil
}

func requireAuditPermission(ctx context.Context, permission string) (*authPkg.AuthUser, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppPOS, permission) {
		return nil, onekiterrors.PermissionDenied()
	}
	return device.User, nil
}

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
