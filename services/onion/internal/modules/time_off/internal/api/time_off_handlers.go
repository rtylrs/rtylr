package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/time_off/internal/application"
	"github.com/voxire/onion/internal/modules/time_off/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeoffv1 "github.com/voxire/voxie/gen/go/hr/time_off/v1"
)

func isValidTimeOffDate(s string) bool {
	_, err1 := time.Parse("2006-01-02", s)
	_, err2 := time.Parse(time.RFC3339, s)
	return err1 == nil || err2 == nil
}

// timeOffDateBefore reports whether a is earlier than b. Both must be valid
// time-off dates (callers check first); unparseable values compare equal.
func timeOffDateBefore(a, b string) bool {
	pa, errA := parseTimeOffDate(a)
	pb, errB := parseTimeOffDate(b)
	if errA != nil || errB != nil {
		return false
	}
	return pa.Before(pb)
}

func parseTimeOffDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

func ListTimeOffRoutes() []string    { return []string{"/hr/time-off"} }
func SubmitTimeOffRoutes() []string  { return []string{"/hr/time-off"} }
func ApproveTimeOffRoutes() []string { return []string{"/hr/time-off/approve"} }

func newTimeOffPayload(t domain.TimeOffRequest) *hrtimeoffv1.TimeOffRequest {
	return &hrtimeoffv1.TimeOffRequest{
		Id: &t.ID, OwnerId: &t.OwnerID, StaffId: &t.StaffID, Type: t.Type,
		StartDate: t.StartDate.Format("2006-01-02"), EndDate: t.EndDate.Format("2006-01-02"),
		DaysCount: &t.DaysCount, Status: t.Status, Reason: t.Reason,
		ApprovedBy: t.ApprovedBy,
		ApprovedAt: t.ApprovedAt.Format("2006-01-02T15:04:05Z"),
		Notes:      t.Notes,
		CreatedAt:  t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ────────────────────────────────────────────────────────────────────

type (
	ListTimeOffHandler  struct{ Service *application.Service }
	ListTimeOffRequest  struct{}
	ListTimeOffResponse struct {
		Requests []*hrtimeoffv1.TimeOffRequest `json:"requests"`
	}
)

func (h *ListTimeOffHandler) Serve(ctx context.Context, _ *ListTimeOffRequest) (ListTimeOffResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	requests, err := h.Service.ListTimeOffRequests(ctx, businessID)
	if err != nil {
		return ListTimeOffResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list time off requests")
	}
	payloads := make([]*hrtimeoffv1.TimeOffRequest, 0, len(requests))
	for _, t := range requests {
		payloads = append(payloads, newTimeOffPayload(t))
	}
	return ListTimeOffResponse{Requests: payloads}, nil
}

// ─── Submit ──────────────────────────────────────────────────────────────────

type SubmitTimeOffHandler struct{ Service *application.Service }

type SubmitTimeOffRequest struct {
	StaffID   int64   `json:"staff_id"`
	Type      string  `json:"type"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	DaysCount float64 `json:"days_count"`
	Reason    string  `json:"reason"`
	Notes     string  `json:"notes"`
}

type SubmitTimeOffResponse struct {
	ID int64 `json:"id"`
}

func (h *SubmitTimeOffHandler) Validate(_ context.Context, req *SubmitTimeOffRequest) error {
	if req.StaffID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "staff_id is required")
	}
	if req.Type == "" {
		return onekiterrors.New(http.StatusBadRequest, "type is required")
	}
	if req.StartDate == "" {
		return onekiterrors.New(http.StatusBadRequest, "start_date is required")
	}
	if !isValidTimeOffDate(req.StartDate) {
		return onekiterrors.New(http.StatusBadRequest, "start_date is not a valid date")
	}
	if req.EndDate == "" {
		return onekiterrors.New(http.StatusBadRequest, "end_date is required")
	}
	if !isValidTimeOffDate(req.EndDate) {
		return onekiterrors.New(http.StatusBadRequest, "end_date is not a valid date")
	}
	if timeOffDateBefore(req.EndDate, req.StartDate) {
		return onekiterrors.New(http.StatusBadRequest, "end_date cannot be before start_date")
	}
	if req.DaysCount <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "days_count must be greater than 0")
	}
	return nil
}

func (h *SubmitTimeOffHandler) Serve(ctx context.Context, req *SubmitTimeOffRequest) (SubmitTimeOffResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SubmitTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SubmitTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	input := domain.TimeOffInput{
		OwnerID: device.User.UserID, BusinessID: businessID, StaffID: req.StaffID, Type: req.Type,
		StartDate: req.StartDate, EndDate: req.EndDate, DaysCount: req.DaysCount,
		Reason: req.Reason, Notes: req.Notes,
	}
	id, err := h.Service.SubmitTimeOffRequest(ctx, input)
	if err != nil {
		return SubmitTimeOffResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to submit time off request")
	}
	return SubmitTimeOffResponse{ID: id}, nil
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
		return ApproveTimeOffResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update time off request")
	}
	return ApproveTimeOffResponse{Success: true}, nil
}
