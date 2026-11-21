package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/tenant"
)

type BreakPayload struct {
	ID          int64  `json:"id"`
	EmployeeID  int64  `json:"employee_id"`
	StaffID     int64  `json:"staff_id"`
	StaffName   string `json:"staff_name"`
	Type        string `json:"type"`
	Start       string `json:"start"`
	End         string `json:"end,omitempty"`
	DurationMin *int32 `json:"duration_min,omitempty"`
}

func breakToPayload(b domain.BreakRecord) *BreakPayload {
	p := &BreakPayload{
		ID:         b.ID,
		EmployeeID: b.EmployeeID,
		StaffID:    b.StaffID,
		StaffName:  b.StaffName,
		Type:       b.BreakType,
		Start:      b.StartedAt.UTC().Format(time.RFC3339),
	}
	if b.EndedAt != nil {
		p.End = b.EndedAt.UTC().Format(time.RFC3339)
	}
	if b.DurationMinutes != nil {
		d := *b.DurationMinutes
		p.DurationMin = &d
	}
	return p
}

func breakCaller(ctx context.Context) (businessID, userID int64, canManage bool, err error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, 0, false, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppAccess(device.User, authorization.AppHR) {
		return 0, 0, false, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, 0, false, onekiterrors.PermissionDenied()
	}
	canManage = authPkg.IsBusinessAdmin(device.User) ||
		authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:schedules:manage") ||
		authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:manage")
	return businessID, device.User.UserID, canManage, nil
}

func breakError(ctx context.Context, businessID int64, op string, err error) error {
	switch {
	case errors.Is(err, domain.ErrPermissionDenied):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrBreakNotFound):
		return onekiterrors.New(http.StatusNotFound, "Break not found")
	case errors.Is(err, domain.ErrBreakAlreadyOpen):
		return onekiterrors.New(http.StatusConflict, "Break is already in progress")
	case errors.Is(err, domain.ErrBreakAlreadyEnded):
		return onekiterrors.New(http.StatusConflict, "Break has already ended")
	case errors.Is(err, domain.ErrInvalidBreakType):
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	logger.Ctx(ctx).Error("break operation failed",
		logger.Any("business_id", businessID), logger.String("operation", op), logger.Error(err))
	return onekiterrors.New(http.StatusInternalServerError, "Failed to "+op)
}

func ListBreaksRoutes() []string {
	return []string{"/hr/shifts/breaks", "/hr/shift/v1/breaks"}
}

type ListBreaksHandler struct{ Service *application.Service }

type ListBreaksRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListBreaksResponse struct {
	Breaks []*BreakPayload `json:"breaks"`
}

func (h *ListBreaksHandler) Serve(ctx context.Context, req *ListBreaksRequest) (ListBreaksResponse, error) {
	businessID, userID, canManage, err := breakCaller(ctx)
	if err != nil {
		return ListBreaksResponse{}, err
	}
	var limit, offset int32
	if req != nil {
		limit = req.Limit
		offset = req.Offset
	}
	items, err := h.Service.ListBreaks(ctx, businessID, userID, canManage, limit, offset)
	if err != nil {
		return ListBreaksResponse{}, breakError(ctx, businessID, "list breaks", err)
	}
	payloads := make([]*BreakPayload, 0, len(items))
	for _, b := range items {
		payloads = append(payloads, breakToPayload(b))
	}
	return ListBreaksResponse{Breaks: payloads}, nil
}

func StartBreakRoutes() []string {
	return []string{"/hr/shifts/breaks", "/hr/shift/v1/breaks"}
}

type StartBreakHandler struct{ Service *application.Service }

type StartBreakRequest struct {
	EmployeeID int64  `json:"employee_id"`
	StaffID    int64  `json:"staff_id"`
	StaffName  string `json:"staff_name"`
	Type       string `json:"type"`
}

type StartBreakResponse struct {
	Break *BreakPayload `json:"break"`
}

func (h *StartBreakHandler) Validate(_ context.Context, req *StartBreakRequest) error {
	if req == nil {
		return nil
	}
	if _, err := domain.NormalizeBreakType(req.Type); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (h *StartBreakHandler) Serve(ctx context.Context, req *StartBreakRequest) (StartBreakResponse, error) {
	businessID, userID, canManage, err := breakCaller(ctx)
	if err != nil {
		return StartBreakResponse{}, err
	}
	var input domain.StartBreakInput
	if req != nil {
		input = domain.StartBreakInput{
			EmployeeID: req.EmployeeID,
			StaffID:    req.StaffID,
			StaffName:  req.StaffName,
			BreakType:  req.Type,
		}
	}
	created, err := h.Service.StartBreak(ctx, businessID, userID, canManage, input)
	if err != nil {
		return StartBreakResponse{}, breakError(ctx, businessID, "start break", err)
	}
	return StartBreakResponse{Break: breakToPayload(*created)}, nil
}

func EndBreakRoutes() []string {
	return []string{"/hr/shifts/breaks/{id}/end", "/hr/shift/v1/breaks/{id}/end"}
}

type EndBreakHandler struct{ Service *application.Service }

type EndBreakRequest struct {
	ID int64 `json:"id"`
}

type EndBreakResponse struct {
	Break   *BreakPayload `json:"break"`
	Success bool          `json:"success"`
}

func (h *EndBreakHandler) Validate(ctx context.Context, req *EndBreakRequest) error {
	id, err := resolveBreakID(ctx, req)
	if err != nil || id <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "invalid break id")
	}
	return nil
}

func (h *EndBreakHandler) Serve(ctx context.Context, req *EndBreakRequest) (EndBreakResponse, error) {
	businessID, userID, canManage, err := breakCaller(ctx)
	if err != nil {
		return EndBreakResponse{}, err
	}
	id, err := resolveBreakID(ctx, req)
	if err != nil || id <= 0 {
		return EndBreakResponse{}, onekiterrors.New(http.StatusBadRequest, "invalid break id")
	}
	ended, err := h.Service.EndBreak(ctx, businessID, userID, id, canManage)
	if err != nil {
		return EndBreakResponse{}, breakError(ctx, businessID, "end break", err)
	}
	return EndBreakResponse{Break: breakToPayload(*ended), Success: true}, nil
}

func resolveBreakID(ctx context.Context, req *EndBreakRequest) (int64, error) {
	if raw := serverHttp.PathParam(ctx, "id"); raw != "" {
		return strconv.ParseInt(raw, 10, 64)
	}
	if req != nil && req.ID > 0 {
		return req.ID, nil
	}
	return 0, errors.New("missing break id")
}
