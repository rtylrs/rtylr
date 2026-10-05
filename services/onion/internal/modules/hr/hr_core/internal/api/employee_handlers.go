package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrstaffv1 "github.com/voxire/voxie/gen/go/hr/staff/v1"
)

func employeeToProto(e domain.EmployeeRecord, includeSensitive bool) *hrstaffv1.Staff {
	staff := &hrstaffv1.Staff{
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
		StartDate:             e.StartDate.Format("2006-01-02"),
		EndDate:               e.EndDate.Format("2006-01-02"),
		Address:               e.Address,
		EmergencyContactName:  e.EmergencyContactName,
		EmergencyContactPhone: e.EmergencyContactPhone,
		Notes:                 e.Notes,
		CreatedAt:             e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:             e.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if includeSensitive {
		staff.Salary = money.CentsToDollars(e.SalaryCents)
		staff.HourlyRate = money.CentsToDollars(e.HourlyRateCents)
		staff.OvertimeRate = money.CentsToDollars(e.OvertimeRateCents)
	}
	return staff
}

// --- List Employees ---

func ListEmployeesRoutes() []string { return []string{"/hr/employees"} }

type ListEmployeesHandler struct{ Service *application.Service }

type ListEmployeesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListEmployeesResponse struct {
	Employees []*hrstaffv1.Staff `json:"employees"`
	Total     int64              `json:"total"`
	Limit     int32              `json:"limit"`
	Offset    int32              `json:"offset"`
}

func (h *ListEmployeesHandler) Serve(ctx context.Context, req *ListEmployeesRequest) (ListEmployeesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListEmployeesResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:view"); err != nil {
		return ListEmployeesResponse{}, err
	}
	includeSensitive := authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:sensitive_view")
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListEmployeesResponse{}, onekiterrors.PermissionDenied()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// businessID is the tenant scope; device.User.UserID is the viewer whose
	// compensation entitlement is resolved.
	employees, total, err := h.Service.ListEmployeesForViewer(ctx, businessID, device.User.UserID, limit, offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list employees", logger.Any("business_id", businessID), logger.Error(err))
		return ListEmployeesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list employees")
	}

	payloads := make([]*hrstaffv1.Staff, 0, len(employees))
	for _, e := range employees {
		payloads = append(payloads, employeeToProto(e, includeSensitive))
	}

	return ListEmployeesResponse{Employees: payloads, Total: total, Limit: limit, Offset: offset}, nil
}

// --- Get Employee ---

func GetEmployeeRoutes() []string { return []string{"/hr/employees/{id}"} }

type GetEmployeeHandler struct{ Service *application.Service }

type GetEmployeeRequest struct{}

type GetEmployeeResponse struct {
	Employee *hrstaffv1.Staff `json:"employee"`
}

func (h *GetEmployeeHandler) Serve(ctx context.Context, _ *GetEmployeeRequest) (GetEmployeeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetEmployeeResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:view"); err != nil {
		return GetEmployeeResponse{}, err
	}
	includeSensitive := authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:sensitive_view")

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetEmployeeResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid employee id")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetEmployeeResponse{}, onekiterrors.PermissionDenied()
	}

	emp, err := h.Service.GetEmployeeForViewer(ctx, businessID, device.User.UserID, id)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			return GetEmployeeResponse{}, onekiterrors.New(http.StatusNotFound, "Employee not found")
		}
		logger.Ctx(ctx).Error("failed to get employee", logger.Any("business_id", businessID), logger.Any("employee_id", id), logger.Error(err))
		return GetEmployeeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get employee")
	}

	return GetEmployeeResponse{Employee: employeeToProto(*emp, includeSensitive)}, nil
}

// --- Get My Employee ---

// GetMyEmployeeRoutes is a fixed path, not a param: without it, chi's router
// would try to parse "me" as the {id} segment GetEmployeeRoutes expects.
func GetMyEmployeeRoutes() []string { return []string{"/hr/employees/me"} }

type GetMyEmployeeHandler struct{ Service *application.Service }

type GetMyEmployeeRequest struct{}

type GetMyEmployeeResponse struct {
	Employee *hrstaffv1.Staff `json:"employee"`
}

// Serve exists because nothing else lets a caller resolve their own
// hr_employees.id: ListEmployeesRequest/GetEmployeeRequest have no user_id
// filter, so every self-service caller (HR's own leave/overtime/equipment
// forms, Flow's employee-hub front door) would otherwise have to page
// through the entire roster hoping to find themselves.
func (h *GetMyEmployeeHandler) Serve(ctx context.Context, _ *GetMyEmployeeRequest) (GetMyEmployeeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetMyEmployeeResponse{}, onekiterrors.PermissionDenied()
	}
	includeSensitive := authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:sensitive_view")
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetMyEmployeeResponse{}, onekiterrors.New(http.StatusNotFound, "No employee record is linked to this account")
	}

	employeeID, err := h.Service.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if err != nil || employeeID <= 0 {
		return GetMyEmployeeResponse{}, onekiterrors.New(http.StatusNotFound, "No employee record is linked to this account")
	}

	emp, err := h.Service.GetEmployeeForViewer(ctx, businessID, device.User.UserID, employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			return GetMyEmployeeResponse{}, onekiterrors.New(http.StatusNotFound, "No employee record is linked to this account")
		}
		logger.Ctx(ctx).Error("failed to get own employee record", logger.Any("business_id", businessID), logger.Error(err))
		return GetMyEmployeeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get employee record")
	}

	return GetMyEmployeeResponse{Employee: employeeToProto(*emp, includeSensitive)}, nil
}

// --- Create Employee ---

func CreateEmployeeRoutes() []string { return []string{"/hr/employees"} }

type CreateEmployeeHandler struct{ Service *application.Service }

type CreateEmployeeRequest struct {
	UserID         int64   `json:"user_id"`
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
	Employee *hrstaffv1.Staff `json:"employee"`
}

func (h *CreateEmployeeHandler) Validate(_ context.Context, req *CreateEmployeeRequest) error {
	if strings.TrimSpace(req.FirstName) == "" {
		return onekiterrors.New(http.StatusBadRequest, "first_name is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		return onekiterrors.New(http.StatusBadRequest, "last_name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return onekiterrors.New(http.StatusBadRequest, "email is required")
	}
	if req.StartDate != "" {
		if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "start_date must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func (h *CreateEmployeeHandler) Serve(ctx context.Context, req *CreateEmployeeRequest) (CreateEmployeeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateEmployeeResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:manage"); err != nil {
		return CreateEmployeeResponse{}, err
	}
	includeSensitive := authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:sensitive_view")
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateEmployeeResponse{}, onekiterrors.PermissionDenied()
	}

	emp := domain.EmployeeRecord{
		UserID:            req.UserID,
		DepartmentID:      req.DepartmentID,
		LocationID:        req.LocationID,
		WorkScheduleID:    req.WorkScheduleID,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Email:             req.Email,
		Phone:             req.Phone,
		EmploymentType:    req.EmploymentType,
		SalaryCents:       money.DollarsToCents(req.Salary),
		HourlyRateCents:   money.DollarsToCents(req.HourlyRate),
		OvertimeRateCents: money.DollarsToCents(req.OvertimeRate),
		Address:           req.Address,
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			emp.StartDate = t
		}
	}

	created, err := h.Service.CreateEmployeeForWriter(ctx, businessID, device.User.UserID, emp)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return CreateEmployeeResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrUserAlreadyLinked) {
			return CreateEmployeeResponse{}, onekiterrors.New(http.StatusConflict, "This user is already linked to another employee record")
		}
		logger.Ctx(ctx).Error("failed to create employee", logger.Any("business_id", businessID), logger.Error(err))
		return CreateEmployeeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create employee")
	}

	return CreateEmployeeResponse{Employee: employeeToProto(*created, includeSensitive)}, nil
}

// --- Update Employee ---

func UpdateEmployeeRoutes() []string { return []string{"/hr/employees/{id}"} }

type UpdateEmployeeHandler struct{ Service *application.Service }

type UpdateEmployeeRequest struct {
	UserID                int64   `json:"user_id"`
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
	Employee *hrstaffv1.Staff `json:"employee"`
}

func (h *UpdateEmployeeHandler) Validate(_ context.Context, req *UpdateEmployeeRequest) error {
	if strings.TrimSpace(req.FirstName) == "" {
		return onekiterrors.New(http.StatusBadRequest, "first_name is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		return onekiterrors.New(http.StatusBadRequest, "last_name is required")
	}
	if req.StartDate != "" {
		if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "start_date must be in YYYY-MM-DD format")
		}
	}
	if req.EndDate != "" {
		if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "end_date must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func (h *UpdateEmployeeHandler) Serve(ctx context.Context, req *UpdateEmployeeRequest) (UpdateEmployeeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateEmployeeResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:manage"); err != nil {
		return UpdateEmployeeResponse{}, err
	}
	includeSensitive := authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:sensitive_view")
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateEmployeeResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateEmployeeResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid employee id")
	}

	emp := domain.EmployeeRecord{
		ID:                    id,
		UserID:                req.UserID,
		DepartmentID:          req.DepartmentID,
		LocationID:            req.LocationID,
		WorkScheduleID:        req.WorkScheduleID,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		EmploymentType:        req.EmploymentType,
		Status:                req.Status,
		SalaryCents:           money.DollarsToCents(req.Salary),
		HourlyRateCents:       money.DollarsToCents(req.HourlyRate),
		OvertimeRateCents:     money.DollarsToCents(req.OvertimeRate),
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

	updated, err := h.Service.UpdateEmployeeForWriter(ctx, businessID, device.User.UserID, emp)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return UpdateEmployeeResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			return UpdateEmployeeResponse{}, onekiterrors.New(http.StatusNotFound, "Employee not found")
		}
		if errors.Is(err, domain.ErrUserAlreadyLinked) {
			return UpdateEmployeeResponse{}, onekiterrors.New(http.StatusConflict, "This user is already linked to another employee record")
		}
		logger.Ctx(ctx).Error("failed to update employee", logger.Any("business_id", businessID), logger.Any("employee_id", id), logger.Error(err))
		return UpdateEmployeeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update employee")
	}

	return UpdateEmployeeResponse{Employee: employeeToProto(*updated, includeSensitive)}, nil
}

// --- Delete Employee ---

func DeleteEmployeeRoutes() []string { return []string{"/hr/employees/{id}"} }

type DeleteEmployeeHandler struct{ Service *application.Service }

type DeleteEmployeeRequest struct{}

type DeleteEmployeeResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteEmployeeHandler) Serve(ctx context.Context, _ *DeleteEmployeeRequest) (DeleteEmployeeResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteEmployeeResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:manage"); err != nil {
		return DeleteEmployeeResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteEmployeeResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteEmployeeResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid employee id")
	}

	err := h.Service.DeactivateEmployeeForWriter(ctx, businessID, device.User.UserID, id)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return DeleteEmployeeResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			return DeleteEmployeeResponse{}, onekiterrors.New(http.StatusNotFound, "Employee not found")
		}
		logger.Ctx(ctx).Error("failed to delete employee", logger.Any("business_id", businessID), logger.Any("employee_id", id), logger.Error(err))
		return DeleteEmployeeResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete employee")
	}

	return DeleteEmployeeResponse{Success: true}, nil
}
