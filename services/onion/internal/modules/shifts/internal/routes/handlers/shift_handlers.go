package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/shifts/internal/application"
	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type ShiftPayload struct {
	ID           int64  `json:"id"`
	OwnerID      int64  `json:"owner_id"`
	StaffID      int64  `json:"staff_id"`
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	BreakMinutes int64  `json:"break_minutes"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func ListShiftsRoutes() []string  { return []string{"/hr/shifts"} }
func CreateShiftRoutes() []string { return []string{"/hr/shifts"} }
func UpdateShiftRoutes() []string { return []string{"/hr/shifts"} }
func DeleteShiftRoutes() []string { return []string{"/hr/shifts"} }

func newShiftPayload(s domain.Shift) *ShiftPayload {
	return &ShiftPayload{
		ID: s.ID, OwnerID: s.OwnerID, StaffID: s.StaffID, Date: s.Date.Format("2006-01-02"),
		StartTime: s.StartTime, EndTime: s.EndTime, BreakMinutes: int64(s.BreakDurationMinutes),
		Status: s.Status, Notes: s.Notes,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ────────────────────────────────────────────────────────────────────

type ListShiftsHandler struct{ Service *application.Service }

type ListShiftsRequest struct {
	StaffID int64  `json:"staff_id"`
	Date    string `json:"date"`
}

type ListShiftsResponse struct {
	Shifts []*ShiftPayload `json:"shifts"`
}

func (h *ListShiftsHandler) Serve(ctx context.Context, req *ListShiftsRequest) (ListShiftsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListShiftsResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListShiftsResponse{}, onekiterrors.PermissionDenied()
	}
	shifts, err := h.Service.ListShifts(ctx, businessID, req.StaffID, req.Date)
	if err != nil {
		return ListShiftsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list shifts")
	}
	payloads := make([]*ShiftPayload, 0, len(shifts))
	for _, s := range shifts {
		payloads = append(payloads, newShiftPayload(s))
	}
	return ListShiftsResponse{Shifts: payloads}, nil
}

// ─── Create ──────────────────────────────────────────────────────────────────

type CreateShiftHandler struct{ Service *application.Service }

type CreateShiftRequest struct {
	StaffID      int64  `json:"staff_id"`
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	BreakMinutes int    `json:"break_minutes"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
}

type CreateShiftResponse struct {
	ID int64 `json:"id"`
}

func (h *CreateShiftHandler) Validate(_ context.Context, req *CreateShiftRequest) error {
	if req.StaffID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "staff_id is required")
	}
	if req.Date == "" {
		return onekiterrors.New(http.StatusBadRequest, "date is required")
	}
	if req.StartTime == "" {
		return onekiterrors.New(http.StatusBadRequest, "start_time is required")
	}
	if req.EndTime == "" {
		return onekiterrors.New(http.StatusBadRequest, "end_time is required")
	}
	return nil
}

func (h *CreateShiftHandler) Serve(ctx context.Context, req *CreateShiftRequest) (CreateShiftResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateShiftResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateShiftResponse{}, onekiterrors.PermissionDenied()
	}
	input := domain.ShiftInput{
		OwnerID: device.User.UserID, BusinessID: businessID, StaffID: req.StaffID, Date: req.Date,
		StartTime: req.StartTime, EndTime: req.EndTime,
		BreakDurationMinutes: req.BreakMinutes, Notes: req.Notes,
	}
	id, err := h.Service.CreateShift(ctx, input)
	if err != nil {
		return CreateShiftResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create shift")
	}
	return CreateShiftResponse{ID: id}, nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

type UpdateShiftHandler struct{ Service *application.Service }

type UpdateShiftRequest struct {
	ID           int64  `json:"id"`
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	BreakMinutes int    `json:"break_minutes"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
}

type UpdateShiftResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateShiftHandler) Validate(_ context.Context, req *UpdateShiftRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	return nil
}

func (h *UpdateShiftHandler) Serve(ctx context.Context, req *UpdateShiftRequest) (UpdateShiftResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateShiftResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateShiftResponse{}, onekiterrors.PermissionDenied()
	}
	input := domain.ShiftUpdateInput{
		ID: req.ID, Date: req.Date, StartTime: req.StartTime, EndTime: req.EndTime,
		BreakDurationMinutes: req.BreakMinutes, Status: req.Status, Notes: req.Notes,
	}
	if err := h.Service.UpdateShift(ctx, businessID, input); err != nil {
		return UpdateShiftResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update shift")
	}
	return UpdateShiftResponse{Success: true}, nil
}

// ─── Delete ──────────────────────────────────────────────────────────────────

type DeleteShiftHandler struct{ Service *application.Service }

type DeleteShiftRequest struct {
	ID int64 `json:"id"`
}

type DeleteShiftResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteShiftHandler) Validate(_ context.Context, req *DeleteShiftRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	return nil
}

func (h *DeleteShiftHandler) Serve(ctx context.Context, req *DeleteShiftRequest) (DeleteShiftResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteShiftResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteShiftResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.DeleteShift(ctx, businessID, req.ID); err != nil {
		return DeleteShiftResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete shift")
	}
	return DeleteShiftResponse{Success: true}, nil
}
