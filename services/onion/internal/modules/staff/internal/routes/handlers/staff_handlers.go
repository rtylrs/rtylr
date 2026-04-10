package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/staff/internal/application"
	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

type StaffPayload struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"ownerId"`
	UserID                int64     `json:"userId"`
	DepartmentID          int64     `json:"departmentId"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	Email                 string    `json:"email"`
	Phone                 string    `json:"phone"`
	Role                  string    `json:"role"`
	EmploymentType        string    `json:"employmentType"`
	Status                string    `json:"status"`
	Salary                float64   `json:"salary"`
	HourlyRate            float64   `json:"hourlyRate"`
	OvertimeRate          float64   `json:"overtimeRate"`
	StartDate             time.Time `json:"startDate"`
	EndDate               time.Time `json:"endDate"`
	Address               string    `json:"address"`
	EmergencyContactName  string    `json:"emergencyContactName"`
	EmergencyContactPhone string    `json:"emergencyContactPhone"`
	Apps                  string    `json:"apps"`
	Notes                 string    `json:"notes"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type ListStaffResponse struct {
	Staff []StaffPayload `json:"staff"`
}

func newStaffPayload(s domain.Staff) StaffPayload {
	return StaffPayload{
		ID: s.ID, OwnerID: s.OwnerID, UserID: s.UserID, DepartmentID: s.DepartmentID,
		FirstName: s.FirstName, LastName: s.LastName, Email: s.Email, Phone: s.Phone,
		Role: s.Role, EmploymentType: s.EmploymentType, Status: s.Status,
		Salary: s.Salary, HourlyRate: s.HourlyRate, OvertimeRate: s.OvertimeRate,
		StartDate: s.StartDate, EndDate: s.EndDate, Address: s.Address,
		EmergencyContactName: s.EmergencyContactName, EmergencyContactPhone: s.EmergencyContactPhone,
		Apps: s.Apps, Notes: s.Notes, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}

func (h *ListStaffHandler) Serve(ctx context.Context, req *ListStaffRequest) (ListStaffResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListStaffResponse{}, errorModels.PermissionDeniedServerError()
	}

	staffList, err := h.Service.ListStaff(ctx, device.User.UserID, req.Search, req.DepartmentID, req.Status)
	if err != nil {
		return ListStaffResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list staff"}
	}

	payloads := make([]StaffPayload, 0, len(staffList))
	for _, s := range staffList {
		payloads = append(payloads, newStaffPayload(s))
	}

	return ListStaffResponse{Staff: payloads}, nil
}

// ─── Create Staff ─────────────────────────────────────────────────────────────

func CreateStaffRoutes() []string { return []string{"/hr/staff"} }

type CreateStaffHandler struct {
	Service *application.Service
}

type CreateStaffRequest struct {
	UserID                int64   `json:"userId"`
	DepartmentID          int64   `json:"departmentId"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employmentType"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourlyRate"`
	OvertimeRate          float64 `json:"overtimeRate"`
	StartDate             string  `json:"startDate"`
	EndDate               string  `json:"endDate"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergencyContactName"`
	EmergencyContactPhone string  `json:"emergencyContactPhone"`
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

	input := domain.StaffInput{
		OwnerID: device.User.UserID, UserID: req.UserID, DepartmentID: req.DepartmentID,
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
	DepartmentID          int64   `json:"departmentId"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employmentType"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourlyRate"`
	OvertimeRate          float64 `json:"overtimeRate"`
	EndDate               string  `json:"endDate"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergencyContactName"`
	EmergencyContactPhone string  `json:"emergencyContactPhone"`
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

	input := domain.StaffUpdateInput{
		ID: req.ID, DepartmentID: req.DepartmentID, FirstName: req.FirstName, LastName: req.LastName,
		Email: req.Email, Phone: req.Phone, Role: req.Role, EmploymentType: req.EmploymentType,
		Status: req.Status, Salary: req.Salary, HourlyRate: req.HourlyRate, OvertimeRate: req.OvertimeRate,
		EndDate: req.EndDate, Address: req.Address,
		EmergencyContactName: req.EmergencyContactName, EmergencyContactPhone: req.EmergencyContactPhone,
		Apps: req.Apps, Notes: req.Notes,
	}

	if err := h.Service.UpdateStaff(ctx, device.User.UserID, input); err != nil {
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

	if err := h.Service.DeleteStaff(ctx, device.User.UserID, req.ID); err != nil {
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
	Staff StaffPayload `json:"staff"`
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

	s, err := h.Service.GetStaffByID(ctx, device.User.UserID, req.ID)
	if err != nil {
		if err == domain.ErrStaffNotFound {
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
	StaffID int64 `json:"staffId"`
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

	if err := h.Service.ClockIn(ctx, device.User.UserID, req.StaffID); err != nil {
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
	StaffID int64 `json:"staffId"`
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

	if err := h.Service.ClockOut(ctx, device.User.UserID, req.StaffID); err != nil {
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

type DashboardStatsPayload struct {
	TotalStaff     int `json:"totalStaff"`
	OnShiftToday   int `json:"onShiftToday"`
	PendingTimeOff int `json:"pendingTimeOff"`
	PayrollDraft   int `json:"payrollDraft"`
}

type DashboardResponse struct {
	Stats DashboardStatsPayload `json:"stats"`
}

func (h *DashboardHandler) Serve(ctx context.Context, _ *DashboardRequest) (DashboardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	stats, err := h.Service.GetDashboardStats(ctx, device.User.UserID)
	if err != nil {
		return DashboardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get dashboard stats"}
	}

	return DashboardResponse{Stats: DashboardStatsPayload{
		TotalStaff: stats.TotalStaff, OnShiftToday: stats.OnShiftToday,
		PendingTimeOff: stats.PendingTimeOff, PayrollDraft: stats.PayrollDraft,
	}}, nil
}
