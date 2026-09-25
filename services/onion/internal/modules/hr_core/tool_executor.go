package hr_core

import (
	"context"
	"errors"
	"fmt"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listEmployeesToolName         = "hr.list_employees"
	getEmployeeToolName           = "hr.get_employee"
	listHRLocationsToolName       = "hr.list_locations"
	listHRDepartmentsToolName     = "hr.list_departments"
	listHRWorkSchedulesToolName   = "hr.list_work_schedules"
	listHRRolesToolName           = "hr.list_roles"
	getHRSetupStatusToolName      = "hr.get_setup_status"
	getHRConfigToolName           = "hr.get_config"
	createEmployeeToolName        = "hr.create_employee"
	defaultListEmployeesToolLimit = int64(20)
	maxListEmployeesToolLimit     = int64(50)
)

type listEmployeesToolInput struct {
	Limit  int32
	Offset int32
	Page   int32
}

type listHRDepartmentsToolInput struct {
	LocationID int64
}

type createEmployeeToolInput struct {
	DepartmentID          int64
	LocationID            int64
	WorkScheduleID        int64
	FirstName             string
	LastName              string
	Email                 string
	Phone                 string
	EmploymentType        string
	Salary                float64
	HourlyRate            float64
	OvertimeRate          float64
	StartDate             time.Time
	Address               string
	EmergencyContactName  string
	EmergencyContactPhone string
	Notes                 string
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listEmployeesToolName:
		output, err := m.executeListEmployeesTool(ctx, input)
		return output, true, err
	case getEmployeeToolName:
		output, err := m.executeGetEmployeeTool(ctx, input)
		return output, true, err
	case listHRLocationsToolName:
		output, err := m.executeListHRLocationsTool(ctx)
		return output, true, err
	case listHRDepartmentsToolName:
		output, err := m.executeListHRDepartmentsTool(ctx, input)
		return output, true, err
	case listHRWorkSchedulesToolName:
		output, err := m.executeListHRWorkSchedulesTool(ctx)
		return output, true, err
	case listHRRolesToolName:
		output, err := m.executeListHRRolesTool(ctx)
		return output, true, err
	case getHRSetupStatusToolName:
		output, err := m.executeGetHRSetupStatusTool(ctx)
		return output, true, err
	case getHRConfigToolName:
		output, err := m.executeGetHRConfigTool(ctx)
		return output, true, err
	case createEmployeeToolName:
		output, err := m.executeCreateEmployeeTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListEmployeesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, callerUserID, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	parsed, err := parseListEmployeesToolInput(input)
	if err != nil {
		return nil, err
	}

	employees, total, err := m.service.ListEmployeesForViewer(ctx, businessID, callerUserID, parsed.Limit, parsed.Offset)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(employees))
	for _, employee := range employees {
		payloads = append(payloads, employeeToolPayload(employee))
	}

	return map[string]any{
		"employees": payloads,
		"total":     total,
		"returned":  len(payloads),
		"limit":     parsed.Limit,
		"offset":    parsed.Offset,
		"page":      parsed.Page,
	}, nil
}

func (m *Module) executeGetEmployeeTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, callerUserID, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	employeeID, err := parseGetEmployeeToolInput(input)
	if err != nil {
		return nil, err
	}

	employee, err := m.service.GetEmployeeForViewer(ctx, businessID, callerUserID, employeeID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"employee": employeeDetailToolPayload(*employee),
	}, nil
}

func (m *Module) executeListHRLocationsTool(ctx context.Context) (map[string]any, error) {
	businessID, _, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	locations, err := m.service.ListLocations(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(locations))
	for _, location := range locations {
		payloads = append(payloads, locationToolPayload(location))
	}

	return map[string]any{
		"locations": payloads,
		"returned":  len(payloads),
	}, nil
}

func (m *Module) executeListHRDepartmentsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, _, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	parsed, err := parseListHRDepartmentsToolInput(input)
	if err != nil {
		return nil, err
	}

	departments, err := m.service.ListDepartments(ctx, businessID, parsed.LocationID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(departments))
	for _, department := range departments {
		payloads = append(payloads, departmentToolPayload(department))
	}

	return map[string]any{
		"departments": payloads,
		"returned":    len(payloads),
		"filters": map[string]any{
			"location_id": parsed.LocationID,
		},
	}, nil
}

func (m *Module) executeListHRWorkSchedulesTool(ctx context.Context) (map[string]any, error) {
	businessID, _, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	// The Caret tool summarises schedules for one business; 500 was the old
	// hard LIMIT on this query and stays the bound here.
	schedules, err := m.service.ListWorkSchedules(ctx, businessID, 500, 0)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(schedules))
	for _, schedule := range schedules {
		payloads = append(payloads, workScheduleToolPayload(schedule))
	}

	return map[string]any{
		"work_schedules": payloads,
		"returned":       len(payloads),
	}, nil
}

func hrCoreToolScope(ctx context.Context) (int64, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, 0, fmt.Errorf("business scope could not be resolved: %w", err)
	}
	return businessID, authUser.UserID, nil
}
