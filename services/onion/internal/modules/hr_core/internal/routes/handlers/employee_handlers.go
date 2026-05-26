package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func employeeToProto(e domain.EmployeeRecord) *hrv1.Staff {
	return &hrv1.Staff{
		Id:                    e.ID,
		OwnerId:               e.OwnerID,
		UserId:                e.UserID,
		DepartmentId:          e.DepartmentID,
		FirstName:             e.FirstName,
		LastName:              e.LastName,
		Email:                 e.Email,
		Phone:                 e.Phone,
		EmploymentType:        e.EmploymentType,
		Status:                e.Status,
		Salary:                e.Salary,
		HourlyRate:            e.HourlyRate,
		OvertimeRate:          e.OvertimeRate,
		StartDate:             e.StartDate.Format("2006-01-02"),
		EndDate:               e.EndDate.Format("2006-01-02"),
		Address:               e.Address,
		EmergencyContactName:  e.EmergencyContactName,
		EmergencyContactPhone: e.EmergencyContactPhone,
		Notes:                 e.Notes,
		CreatedAt:             e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:             e.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List Employees ---

func ListEmployeesRoutes() []string { return []string{"/hr/employees"} }

type ListEmployeesHandler struct{ Service *application.Service }

type ListEmployeesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListEmployeesResponse struct {
	Employees []*hrv1.Staff `json:"employees"`
	Total     int64         `json:"total"`
	Limit     int32         `json:"limit"`
	Offset    int32         `json:"offset"`
}

func (h *ListEmployeesHandler) Serve(ctx context.Context, req *ListEmployeesRequest) (ListEmployeesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListEmployeesResponse{}, errorModels.PermissionDeniedServerError()
	}

	employees, total, err := h.Service.ListEmployees(ctx, device.User.UserID, req.Limit, req.Offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list employees", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListEmployeesResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list employees"}
	}

	payloads := make([]*hrv1.Staff, 0, len(employees))
	for _, e := range employees {
		payloads = append(payloads, employeeToProto(e))
	}

	return ListEmployeesResponse{Employees: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}

// --- Get Employee ---

func GetEmployeeRoutes() []string { return []string{"/hr/employees/{id}"} }

type GetEmployeeHandler struct{ Service *application.Service }

type GetEmployeeRequest struct{}

type GetEmployeeResponse struct {
	Employee *hrv1.Staff `json:"employee"`
}

func (h *GetEmployeeHandler) Serve(ctx context.Context, _ *GetEmployeeRequest) (GetEmployeeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetEmployeeResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid employee id"}
	}

	emp, err := h.Service.GetEmployee(ctx, device.User.UserID, id)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			return GetEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Employee not found"}
		}
		logger.Ctx(ctx).Error("failed to get employee", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", id), logger.Error(err))
		return GetEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get employee"}
	}

	return GetEmployeeResponse{Employee: employeeToProto(*emp)}, nil
}

// --- Create Employee ---

func CreateEmployeeRoutes() []string { return []string{"/hr/employees"} }

type CreateEmployeeHandler struct{ Service *application.Service }

type CreateEmployeeRequest struct {
	DepartmentID   int64   `json:"department_id"`
	LocationID     int64   `json:"location_id"`
	WorkScheduleID int64   `json:"work_schedule_id"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	EmploymentType string  `json:"employment_type"`
	Salary         float64 `json:"salary"`
	HourlyRate     float64 `json:"hourly_rate"`
	OvertimeRate   float64 `json:"overtime_rate"`
	StartDate      string  `json:"start_date"`
	Address        string  `json:"address"`
}

type CreateEmployeeResponse struct {
	Employee *hrv1.Staff `json:"employee"`
}

func (h *CreateEmployeeHandler) Validate(_ context.Context, req *CreateEmployeeRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.FirstName) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "first_name is required"}
	}
	if strings.TrimSpace(req.LastName) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "last_name is required"}
	}
	if strings.TrimSpace(req.Email) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	if req.StartDate != "" {
		if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "start_date must be in YYYY-MM-DD format"}
		}
	}
	return nil
}

func (h *CreateEmployeeHandler) Serve(ctx context.Context, req *CreateEmployeeRequest) (CreateEmployeeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateEmployeeResponse{}, errorModels.PermissionDeniedServerError()
	}

	emp := domain.EmployeeRecord{
		DepartmentID:   req.DepartmentID,
		LocationID:     req.LocationID,
		WorkScheduleID: req.WorkScheduleID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		Phone:          req.Phone,
		EmploymentType: req.EmploymentType,
		Salary:         req.Salary,
		HourlyRate:     req.HourlyRate,
		OvertimeRate:   req.OvertimeRate,
		Address:        req.Address,
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			emp.StartDate = t
		}
	}

	created, err := h.Service.CreateEmployee(ctx, device.User.UserID, emp)
	if err != nil {
		logger.Ctx(ctx).Error("failed to create employee", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create employee"}
	}

	return CreateEmployeeResponse{Employee: employeeToProto(*created)}, nil
}

// --- Update Employee ---

func UpdateEmployeeRoutes() []string { return []string{"/hr/employees/{id}"} }

type UpdateEmployeeHandler struct{ Service *application.Service }

type UpdateEmployeeRequest struct {
	DepartmentID          int64   `json:"department_id"`
	LocationID            int64   `json:"location_id"`
	WorkScheduleID        int64   `json:"work_schedule_id"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
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
	Notes                 string  `json:"notes"`
}

type UpdateEmployeeResponse struct {
	Employee *hrv1.Staff `json:"employee"`
}

func (h *UpdateEmployeeHandler) Validate(_ context.Context, req *UpdateEmployeeRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.FirstName) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "first_name is required"}
	}
	if strings.TrimSpace(req.LastName) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "last_name is required"}
	}
	if req.StartDate != "" {
		if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "start_date must be in YYYY-MM-DD format"}
		}
	}
	if req.EndDate != "" {
		if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "end_date must be in YYYY-MM-DD format"}
		}
	}
	return nil
}

func (h *UpdateEmployeeHandler) Serve(ctx context.Context, req *UpdateEmployeeRequest) (UpdateEmployeeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateEmployeeResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid employee id"}
	}

	emp := domain.EmployeeRecord{
		ID:                    id,
		DepartmentID:          req.DepartmentID,
		LocationID:            req.LocationID,
		WorkScheduleID:        req.WorkScheduleID,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		EmploymentType:        req.EmploymentType,
		Status:                req.Status,
		Salary:                req.Salary,
		HourlyRate:            req.HourlyRate,
		OvertimeRate:          req.OvertimeRate,
		Address:               req.Address,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		Notes:                 req.Notes,
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			emp.StartDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			emp.EndDate = t
		}
	}

	updated, err := h.Service.UpdateEmployee(ctx, device.User.UserID, emp)
	if err != nil {
		logger.Ctx(ctx).Error("failed to update employee", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", id), logger.Error(err))
		return UpdateEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update employee"}
	}

	return UpdateEmployeeResponse{Employee: employeeToProto(*updated)}, nil
}

// --- Delete Employee ---

func DeleteEmployeeRoutes() []string { return []string{"/hr/employees/{id}"} }

type DeleteEmployeeHandler struct{ Service *application.Service }

type DeleteEmployeeRequest struct{}

type DeleteEmployeeResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteEmployeeHandler) Serve(ctx context.Context, _ *DeleteEmployeeRequest) (DeleteEmployeeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteEmployeeResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid employee id"}
	}

	emp := domain.EmployeeRecord{ID: id, Status: "terminated"}
	_, err := h.Service.UpdateEmployee(ctx, device.User.UserID, emp)
	if err != nil {
		logger.Ctx(ctx).Error("failed to delete employee", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", id), logger.Error(err))
		return DeleteEmployeeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete employee"}
	}

	return DeleteEmployeeResponse{Success: true}, nil
}
