package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/staff/internal/application"
	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdashboardv1 "github.com/voxire/voxie/gen/go/hr/dashboard/v1"
	hrstaffv1 "github.com/voxire/voxie/gen/go/hr/staff/v1"
)

// ─── List Staff ───────────────────────────────────────────────────────────────

func ListStaffRoutes() []string { return []string{"/hr/staff"} }

type ListStaffHandler struct {
	Service *application.Service
}

type ListStaffRequest struct {
	Search       string `json:"search"`
	DepartmentID string `json:"department_id"`
	Status       string `json:"status"`
}

// StaffCardPayload is the frontend-facing shape expected by the HR Staff page.
// It uses a flat structure with name, department string, and human-readable enums.
type StaffCardPayload struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	Department     string `json:"department"`
	Email          string `json:"email"`
	EmploymentType string `json:"employment_type"`
	Status         string `json:"status"`
}

// ListStaffResponse is a plain JSON array so the frontend can parse it directly.
type ListStaffResponse []StaffCardPayload

func toStaffCardPayload(s domain.Staff) StaffCardPayload {
	name := strings.TrimSpace(s.FirstName + " " + s.LastName)

	empType := "Full-time"
	switch s.EmploymentType {
	case "part_time":
		empType = "Part-time"
	case "contractor":
		empType = "Contractor"
	}

	statusLabel := "Active"
	switch s.Status {
	case "inactive", "terminated":
		statusLabel = "Inactive"
	}

	return StaffCardPayload{
		ID:             s.ID,
		Name:           name,
		Role:           s.Role,
		Department:     s.DepartmentName,
		Email:          s.Email,
		EmploymentType: empType,
		Status:         statusLabel,
	}
}

func newStaffPayload(s domain.Staff) *hrstaffv1.Staff {
	return &hrstaffv1.Staff{
		Id: s.ID, OwnerId: s.OwnerID, UserId: s.UserID, DepartmentId: s.DepartmentID,
		FirstName: s.FirstName, LastName: s.LastName, Email: s.Email, Phone: s.Phone,
		Role: s.Role, EmploymentType: s.EmploymentType, Status: s.Status,
		Salary: s.Salary, HourlyRate: s.HourlyRate, OvertimeRate: s.OvertimeRate,
		StartDate:            s.StartDate.Format("2006-01-02T15:04:05Z"),
		EndDate:              s.EndDate.Format("2006-01-02T15:04:05Z"),
		Address:              s.Address,
		EmergencyContactName: s.EmergencyContactName, EmergencyContactPhone: s.EmergencyContactPhone,
		Apps: s.Apps, Notes: s.Notes,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ListStaffHandler) Serve(ctx context.Context, req *ListStaffRequest) (ListStaffResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	staffList, err := h.Service.ListStaff(ctx, businessID, req.Search, req.DepartmentID, req.Status)
	if err != nil {
		return ListStaffResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list staff"}
	}

	payloads := make(ListStaffResponse, 0, len(staffList))
	for _, s := range staffList {
		payloads = append(payloads, toStaffCardPayload(s))
	}

	return payloads, nil
}

// ─── Create Staff ─────────────────────────────────────────────────────────────

func CreateStaffRoutes() []string { return []string{"/hr/staff"} }

type CreateStaffHandler struct {
	Service *application.Service
}

type CreateStaffRequest struct {
	UserID                int64   `json:"user_id"`
	DepartmentID          int64   `json:"department_id"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employment_type"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourly_rate"`
	OvertimeRate          float64 `json:"overtime_rate"`
	StartDate             string  `json:"start_date"`
	EndDate               string  `json:"end_date"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergency_contact_name"`
	EmergencyContactPhone string  `json:"emergency_contact_phone"`
	Apps                  string  `json:"apps"`
	Notes                 string  `json:"notes"`
}

type CreateStaffResponse struct {
	ID int64 `json:"id"`
}

func (h *CreateStaffHandler) Validate(_ context.Context, req *CreateStaffRequest) *errorModels.ServerError {
	if req.FirstName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "first_name is required"}
	}
	if req.LastName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "last_name is required"}
	}
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	if req.Role == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "role is required"}
	}
	if req.StartDate == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "start_date is required"}
	}
	return nil
}

func (h *CreateStaffHandler) Serve(ctx context.Context, req *CreateStaffRequest) (CreateStaffResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.StaffInput{
		OwnerID: device.User.UserID, BusinessID: businessID, UserID: req.UserID, DepartmentID: req.DepartmentID,
		FirstName: req.FirstName, LastName: req.LastName, Email: req.Email, Phone: req.Phone,
		Role: req.Role, EmploymentType: req.EmploymentType, Status: req.Status,
		Salary: req.Salary, HourlyRate: req.HourlyRate, OvertimeRate: req.OvertimeRate,
		StartDate: req.StartDate, EndDate: req.EndDate, Address: req.Address,
		EmergencyContactName: req.EmergencyContactName, EmergencyContactPhone: req.EmergencyContactPhone,
		Apps: req.Apps, Notes: req.Notes,
	}

	id, err := h.Service.CreateStaff(ctx, input)
	if err != nil {
		return CreateStaffResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create staff member"}
	}

	return CreateStaffResponse{ID: id}, nil
}

// ─── Update Staff ─────────────────────────────────────────────────────────────

func UpdateStaffRoutes() []string { return []string{"/hr/staff"} }

type UpdateStaffHandler struct {
	Service *application.Service
}

type UpdateStaffRequest struct {
	ID                    int64   `json:"id"`
	DepartmentID          int64   `json:"department_id"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employment_type"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourly_rate"`
	OvertimeRate          float64 `json:"overtime_rate"`
	EndDate               string  `json:"end_date"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergency_contact_name"`
	EmergencyContactPhone string  `json:"emergency_contact_phone"`
	Apps                  string  `json:"apps"`
	Notes                 string  `json:"notes"`
}

type UpdateStaffResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateStaffHandler) Validate(_ context.Context, req *UpdateStaffRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *UpdateStaffHandler) Serve(ctx context.Context, req *UpdateStaffRequest) (UpdateStaffResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.StaffUpdateInput{
		ID: req.ID, DepartmentID: req.DepartmentID, FirstName: req.FirstName, LastName: req.LastName,
		Email: req.Email, Phone: req.Phone, Role: req.Role, EmploymentType: req.EmploymentType,
		Status: req.Status, Salary: req.Salary, HourlyRate: req.HourlyRate, OvertimeRate: req.OvertimeRate,
		EndDate: req.EndDate, Address: req.Address,
		EmergencyContactName: req.EmergencyContactName, EmergencyContactPhone: req.EmergencyContactPhone,
		Apps: req.Apps, Notes: req.Notes,
	}

	if err := h.Service.UpdateStaff(ctx, businessID, input); err != nil {
		return UpdateStaffResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update staff member"}
	}

	return UpdateStaffResponse{Success: true}, nil
}

// ─── Delete Staff ─────────────────────────────────────────────────────────────

func DeleteStaffRoutes() []string { return []string{"/hr/staff"} }

type DeleteStaffHandler struct {
	Service *application.Service
}

type DeleteStaffRequest struct {
	ID int64 `json:"id"`
}

type DeleteStaffResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteStaffHandler) Validate(_ context.Context, req *DeleteStaffRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteStaffHandler) Serve(ctx context.Context, req *DeleteStaffRequest) (DeleteStaffResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.DeleteStaff(ctx, businessID, req.ID); err != nil {
		return DeleteStaffResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to terminate staff member"}
	}

	return DeleteStaffResponse{Success: true}, nil
}

// ─── Get Staff By ID ─────────────────────────────────────────────────────────

func GetStaffRoutes() []string { return []string{"/hr/staff/detail"} }

type GetStaffHandler struct {
	Service *application.Service
}

type GetStaffRequest struct {
	ID int64 `json:"id"`
}

type GetStaffResponse struct {
	Staff *hrstaffv1.Staff `json:"staff"`
}

func (h *GetStaffHandler) Validate(_ context.Context, req *GetStaffRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *GetStaffHandler) Serve(ctx context.Context, req *GetStaffRequest) (GetStaffResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	s, err := h.Service.GetStaffByID(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrStaffNotFound) {
			return GetStaffResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Staff member not found"}
		}
		return GetStaffResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get staff member"}
	}

	return GetStaffResponse{Staff: newStaffPayload(*s)}, nil
}

// ─── Clock In ────────────────────────────────────────────────────────────────

func ClockInRoutes() []string { return []string{"/hr/staff/clock-in"} }

type ClockInHandler struct {
	Service *application.Service
}

type ClockInRequest struct {
	StaffID int64 `json:"staff_id"`
}

type ClockInResponse struct {
	Success bool `json:"success"`
}

func (h *ClockInHandler) Validate(_ context.Context, req *ClockInRequest) *errorModels.ServerError {
	if req.StaffID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "staffId is required"}
	}
	return nil
}

func (h *ClockInHandler) Serve(ctx context.Context, req *ClockInRequest) (ClockInResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ClockInResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ClockInResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.ClockIn(ctx, businessID, req.StaffID); err != nil {
		return ClockInResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to clock in"}
	}

	return ClockInResponse{Success: true}, nil
}

// ─── Clock Out ───────────────────────────────────────────────────────────────

func ClockOutRoutes() []string { return []string{"/hr/staff/clock-out"} }

type ClockOutHandler struct {
	Service *application.Service
}

type ClockOutRequest struct {
	StaffID int64 `json:"staff_id"`
}

type ClockOutResponse struct {
	Success bool `json:"success"`
}

func (h *ClockOutHandler) Validate(_ context.Context, req *ClockOutRequest) *errorModels.ServerError {
	if req.StaffID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "staffId is required"}
	}
	return nil
}

func (h *ClockOutHandler) Serve(ctx context.Context, req *ClockOutRequest) (ClockOutResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ClockOutResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ClockOutResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.ClockOut(ctx, businessID, req.StaffID); err != nil {
		return ClockOutResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to clock out"}
	}

	return ClockOutResponse{Success: true}, nil
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func DashboardRoutes() []string { return []string{"/hr/dashboard"} }

type DashboardHandler struct {
	Service *application.Service
}

type DashboardRequest struct{}

type DashboardResponse struct {
	Stats *hrdashboardv1.DashboardStats `json:"stats"`
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	stats, err := h.Service.GetDashboardStats(ctx, businessID)
	if err != nil {
		return DashboardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get dashboard stats"}
	}

	return DashboardResponse{Stats: &hrdashboardv1.DashboardStats{
		TotalStaff: int64(stats.TotalStaff), OnShiftToday: int64(stats.OnShiftToday),
		PendingTimeOff: int64(stats.PendingTimeOff), PayrollDraft: int64(stats.PayrollDraft),
	}}, nil
}
