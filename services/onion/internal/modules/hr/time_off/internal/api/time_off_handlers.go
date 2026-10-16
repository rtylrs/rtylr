package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/time_off/internal/application"
	"github.com/voxire/onion/internal/modules/hr/time_off/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListTimeOffRoutes() []string    { return []string{"/hr/time-off"} }
func SubmitTimeOffRoutes() []string  { return []string{"/hr/time-off"} }
func ApproveTimeOffRoutes() []string { return []string{"/hr/time-off/approve"} }

type TimeOffEntryDTO struct {
	ID        int64  `json:"id"`
	StaffID   int64  `json:"staffId"`
	StaffName string `json:"staffName"`
	Type      string `json:"type"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Reason    string `json:"reason"`
	Status    string `json:"status"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"createdAt"`
}

// ─── List ────────────────────────────────────────────────────────────────────

type ListTimeOffRequest struct{}

type ListTimeOffHandler struct{ Service *application.Service }

func (h *ListTimeOffHandler) Serve(ctx context.Context, _ *ListTimeOffRequest) ([]TimeOffEntryDTO, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	items, err := h.Service.ListTimeOffRequests(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list time off requests")
	}
	dtos := make([]TimeOffEntryDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, toTimeOffEntryDTO(item))
	}
	return dtos, nil
}

// ─── Submit ──────────────────────────────────────────────────────────────────

type SubmitTimeOffRequest struct {
	StaffID        int64   `json:"staffId"`
	StaffIDSnake   int64   `json:"staff_id"`
	StaffName      string  `json:"staffName"`
	Type           string  `json:"type"`
	StartDate      string  `json:"startDate"`
	StartDateSnake string  `json:"start_date"`
	EndDate        string  `json:"endDate"`
	EndDateSnake   string  `json:"end_date"`
	DaysCount      float64 `json:"daysCount"`
	Reason         string  `json:"reason"`
	Notes          string  `json:"notes"`
	Note           string  `json:"note"`
}

func (h *SubmitTimeOffHandler) Validate(_ context.Context, req *SubmitTimeOffRequest) error {
	staffID := req.StaffID
	if staffID == 0 {
		staffID = req.StaffIDSnake
	}
	if staffID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "staffId is required")
	}
	startDate := req.StartDate
	if startDate == "" {
		startDate = req.StartDateSnake
	}
	endDate := req.EndDate
	if endDate == "" {
		endDate = req.EndDateSnake
	}
	if startDate == "" || endDate == "" {
		return onekiterrors.New(http.StatusBadRequest, "startDate and endDate are required")
	}
	start, startErr := time.Parse("2006-01-02", startDate)
	end, endErr := time.Parse("2006-01-02", endDate)
	if startErr != nil || endErr != nil {
		return onekiterrors.New(http.StatusBadRequest, "startDate and endDate must be YYYY-MM-DD")
	}
	if end.Before(start) {
		return onekiterrors.New(http.StatusBadRequest, "endDate cannot be before startDate")
	}
	return nil
}

type SubmitTimeOffHandler struct{ Service *application.Service }

func (h *SubmitTimeOffHandler) Serve(ctx context.Context, req *SubmitTimeOffRequest) (TimeOffEntryDTO, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return TimeOffEntryDTO{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return TimeOffEntryDTO{}, onekiterrors.PermissionDenied()
	}

	staffID := req.StaffID
	if staffID == 0 {
		staffID = req.StaffIDSnake
	}
	startDate := req.StartDate
	if startDate == "" {
		startDate = req.StartDateSnake
	}
	endDate := req.EndDate
	if endDate == "" {
		endDate = req.EndDateSnake
	}
	notes := req.Notes
	if notes == "" {
		notes = req.Note
	}

	// Normalize type for MySQL enum ('vacation','sick','personal','unpaid','other')
	reqType := strings.ToLower(strings.TrimSpace(req.Type))
	switch reqType {
	case "vacation", "sick", "personal", "unpaid":
	default:
		reqType = "other"
	}

	daysCount := req.DaysCount
	if daysCount <= 0 {
		start, err1 := time.Parse("2006-01-02", startDate)
		end, err2 := time.Parse("2006-01-02", endDate)
		if err1 == nil && err2 == nil {
			daysCount = float64(int(end.Sub(start).Hours()/24) + 1)
		}
		if daysCount <= 0 {
			daysCount = 1
		}
	}

	input := domain.TimeOffInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		StaffID:    staffID,
		Type:       reqType,
		StartDate:  startDate,
		EndDate:    endDate,
		DaysCount:  daysCount,
		Reason:     req.Reason,
		Notes:      notes,
	}

	id, err := h.Service.SubmitTimeOffRequest(ctx, input)
	if errors.Is(err, domain.ErrStaffNotFound) {
		return TimeOffEntryDTO{}, onekiterrors.New(http.StatusNotFound, "Staff member not found")
	}
	if err != nil {
		return TimeOffEntryDTO{}, onekiterrors.New(http.StatusInternalServerError, "Failed to submit time off request")
	}

	return TimeOffEntryDTO{
		ID:        id,
		StaffID:   staffID,
		StaffName: req.StaffName,
		Type:      formatType(req.Type),
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    req.Reason,
		Status:    "pending",
		Note:      notes,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ─── Approve ─────────────────────────────────────────────────────────────────

type ApproveTimeOffHandler struct{ Service *application.Service }

type ApproveTimeOffRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type ApproveTimeOffResponse struct {
	Success bool `json:"success"`
}

func (h *ApproveTimeOffHandler) Validate(_ context.Context, req *ApproveTimeOffRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	if req.Status != "approved" && req.Status != "rejected" {
		return onekiterrors.New(http.StatusBadRequest, "status must be 'approved' or 'rejected'")
	}
	return nil
}

func (h *ApproveTimeOffHandler) Serve(ctx context.Context, req *ApproveTimeOffRequest) (ApproveTimeOffResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	input := domain.TimeOffApprovalInput{
		ID: req.ID, Status: req.Status, ApprovedBy: device.User.UserID, Notes: req.Notes,
	}
	if err := h.Service.ApproveTimeOffRequest(ctx, businessID, input); err != nil {
		if errors.Is(err, domain.ErrTimeOffNotPending) {
			return ApproveTimeOffResponse{}, onekiterrors.New(http.StatusConflict, "Time off request not found or already decided")
		}
		return ApproveTimeOffResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update time off request")
	}
	return ApproveTimeOffResponse{Success: true}, nil
}

func toTimeOffEntryDTO(t domain.TimeOffRequest) TimeOffEntryDTO {
	status := t.Status
	if status == "" {
		status = "pending"
	}
	startDate := ""
	if !t.StartDate.IsZero() {
		startDate = t.StartDate.Format("2006-01-02")
	}
	endDate := ""
	if !t.EndDate.IsZero() {
		endDate = t.EndDate.Format("2006-01-02")
	}
	createdAt := ""
	if !t.CreatedAt.IsZero() {
		createdAt = t.CreatedAt.Format(time.RFC3339)
	}
	return TimeOffEntryDTO{
		ID:        t.ID,
		StaffID:   t.StaffID,
		StaffName: t.StaffName,
		Type:      formatType(t.Type),
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    t.Reason,
		Status:    status,
		Note:      t.Notes,
		CreatedAt: createdAt,
	}
}

func formatType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "vacation":
		return "Vacation"
	case "sick":
		return "Sick"
	case "personal":
		return "Personal"
	case "unpaid":
		return "Unpaid"
	case "bereavement":
		return "Bereavement"
	case "other":
		return "Other"
	default:
		if len(t) > 0 {
			return strings.ToUpper(t[:1]) + strings.ToLower(t[1:])
		}
		return t
	}
}
