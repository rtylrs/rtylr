package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/tenant"
	hremployeev1 "github.com/voxire/voxie/gen/go/hr/employee/v1"
)

const (
	employeeListPath               = "/hr/employee/v1/employees"
	defaultEmployeeListLimit int32 = 20
	maxEmployeeListLimit     int32 = 100
)

// EmployeeListAPI is the compatibility surface for the contracted employee
// list. The hand-registered HR routes return hr.staff.v1.Staff, while Flow's
// capacity query consumes the newer hr.employee.v1.EmployeePayload shape.
// Keeping this adapter read-only avoids exposing a second write path while the
// remaining employee mutations continue to use their established handlers.
type EmployeeListAPI struct {
	Service *application.Service
}

func NewEmployeeListAPI(service *application.Service) *EmployeeListAPI {
	return &EmployeeListAPI{Service: service}
}

// RegisterEmployeeListAPI registers only the implemented compatibility method.
// The generated EmployeeService also contains write and invite methods, but
// those have different semantics from hr_core's existing authenticated
// handlers (notably invite acceptance needs a user id). Registering them here
// as no-op endpoints would turn a clear 404 into a misleading success path.
func RegisterEmployeeListAPI(mux *http.ServeMux, api *EmployeeListAPI) {
	mux.Handle("GET "+employeeListPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := parseEmployeeListRequest(r)
		if err != nil {
			serverHttp.WriteOneKitError(r.Context(), w, err)
			return
		}

		resp, err := api.ListEmployees(r.Context(), req)
		if err != nil {
			serverHttp.WriteOneKitError(r.Context(), w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Ctx(r.Context()).Error("failed to encode employee list response", logger.Error(err))
			return
		}
	}))
}

func parseEmployeeListRequest(r *http.Request) (*hremployeev1.ListEmployeesRequest, error) {
	query := r.URL.Query()
	req := &hremployeev1.ListEmployeesRequest{}

	if value := query.Get("department_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			return nil, onekitBadRequest("department_id must be a positive integer")
		}
		req.DepartmentId = &parsed
	}
	if value := query.Get("location_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			return nil, onekitBadRequest("location_id must be a positive integer")
		}
		req.LocationId = &parsed
	}
	if value := query.Get("status"); value != "" {
		req.Status = &value
	}
	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, onekitBadRequest("limit must be an integer")
		}
		req.Limit = int32(parsed) // #nosec G115 -- ParseInt is bounded to int32 above.
	}
	if value := query.Get("offset"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, onekitBadRequest("offset must be an integer")
		}
		req.Offset = int32(parsed) // #nosec G115 -- ParseInt is bounded to int32 above.
	}

	return req, nil
}

func onekitBadRequest(message string) error {
	return onekiterrors.New(http.StatusBadRequest, message)
}

func (a *EmployeeListAPI) ListEmployees(
	ctx context.Context,
	req *hremployeev1.ListEmployeesRequest,
) (*hremployeev1.ListEmployeesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:view") ||
		!authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = defaultEmployeeListLimit
	}
	if limit > maxEmployeeListLimit {
		limit = maxEmployeeListLimit
	}
	offset := req.GetOffset()
	if offset < 0 {
		offset = 0
	}

	params := domain.ListEmployeesParams{Limit: limit, Offset: offset}
	if req.DepartmentId != nil {
		params.DepartmentID = req.DepartmentId
	}
	if req.LocationId != nil {
		params.LocationID = req.LocationId
	}
	if req.Status != nil {
		params.Status = req.Status
	}
	employees, total, err := a.Service.ListEmployeesForViewerWithFilters(ctx, businessID, device.User.UserID, params)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list employees")
	}

	payloads := make([]*hremployeev1.EmployeePayload, 0, len(employees))
	for _, employee := range employees {
		payloads = append(payloads, employeeToPayload(employee))
	}

	return &hremployeev1.ListEmployeesResponse{
		Employees: payloads,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func employeeToPayload(employee domain.EmployeeRecord) *hremployeev1.EmployeePayload {
	payload := &hremployeev1.EmployeePayload{
		Id:                  employee.ID,
		OwnerId:             employee.OwnerID,
		UserId:              employee.UserID,
		DepartmentId:        employee.DepartmentID,
		LocationId:          employee.LocationID,
		ContractType:        employee.EmploymentType,
		Status:              employee.Status,
		HireDate:            formatEmployeeDate(employee.StartDate),
		Currency:            "",
		AccountCreationMode: employeeAccountCreationMode(employee.Status),
		CreatedAt:           employee.CreatedAt.Unix(),
		UpdatedAt:           employee.UpdatedAt.Unix(),
	}
	if employee.WorkScheduleID > 0 {
		workScheduleID := employee.WorkScheduleID
		payload.WorkScheduleId = &workScheduleID
	}
	if !employee.EndDate.IsZero() {
		endDate := formatEmployeeDate(employee.EndDate)
		payload.TerminationDate = &endDate
	}
	if employee.HourlyRateCents > 0 {
		rate := float64(employee.HourlyRateCents) / 100
		payload.HourlyRate = &rate
	}
	if employee.SalaryCents > 0 {
		salary := float64(employee.SalaryCents) / 100
		payload.MonthlySalary = &salary
	}
	return payload
}

func formatEmployeeDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func employeeAccountCreationMode(status string) string {
	if status == "invited" {
		return "invite_link"
	}
	return "hr_created"
}
