package hr_core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
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
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListEmployeesToolInput(input)
	if err != nil {
		return nil, err
	}

	employees, total, err := m.service.ListEmployeesForViewer(ctx, authUser.UserID, authUser.UserID, parsed.Limit, parsed.Offset)
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
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	employeeID, err := parseGetEmployeeToolInput(input)
	if err != nil {
		return nil, err
	}

	employee, err := m.service.GetEmployeeForViewer(ctx, authUser.UserID, authUser.UserID, employeeID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"employee": employeeDetailToolPayload(*employee),
	}, nil
}

func (m *Module) executeListHRLocationsTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	locations, err := m.service.ListLocations(ctx, authUser.UserID)
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
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListHRDepartmentsToolInput(input)
	if err != nil {
		return nil, err
	}

	departments, err := m.service.ListDepartments(ctx, authUser.UserID, parsed.LocationID)
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
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	// The Caret tool summarises schedules for one business; 500 was the old
	// hard LIMIT on this query and stays the bound here.
	schedules, err := m.service.ListWorkSchedules(ctx, authUser.UserID, 500, 0)
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

func (m *Module) executeListHRRolesTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	roles, err := m.service.ListRoles(ctx, authUser.UserID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(roles))
	for _, role := range roles {
		payloads = append(payloads, roleToolPayload(role))
	}

	return map[string]any{
		"roles":    payloads,
		"returned": len(payloads),
	}, nil
}

func (m *Module) executeGetHRSetupStatusTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	steps, err := m.service.GetSetupStatus(ctx, authUser.UserID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(steps))
	completed := 0
	for _, step := range steps {
		if step.IsCompleted {
			completed++
		}
		payloads = append(payloads, setupStepToolPayload(step))
	}

	return map[string]any{
		"steps":      payloads,
		"total":      len(payloads),
		"completed":  completed,
		"incomplete": len(payloads) - completed,
	}, nil
}

func (m *Module) executeGetHRConfigTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	config, err := m.service.GetHRConfig(ctx, authUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrSetupNotComplete) {
			return map[string]any{
				"configured": false,
				"config":     nil,
			}, nil
		}
		return nil, err
	}

	return map[string]any{
		"configured": config != nil,
		"config":     hrConfigToolPayload(config),
	}, nil
}

func (m *Module) executeCreateEmployeeTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateEmployeeToolInput(input)
	if err != nil {
		return nil, err
	}

	employee := domain.EmployeeRecord{
		DepartmentID:          parsed.DepartmentID,
		LocationID:            parsed.LocationID,
		WorkScheduleID:        parsed.WorkScheduleID,
		FirstName:             parsed.FirstName,
		LastName:              parsed.LastName,
		Email:                 parsed.Email,
		Phone:                 parsed.Phone,
		EmploymentType:        parsed.EmploymentType,
		SalaryCents:           domain.DollarsToCents(parsed.Salary),
		HourlyRateCents:       domain.DollarsToCents(parsed.HourlyRate),
		OvertimeRateCents:     domain.DollarsToCents(parsed.OvertimeRate),
		StartDate:             parsed.StartDate,
		Address:               parsed.Address,
		EmergencyContactName:  parsed.EmergencyContactName,
		EmergencyContactPhone: parsed.EmergencyContactPhone,
		Notes:                 parsed.Notes,
	}

	created, err := m.service.CreateEmployeeForWriter(ctx, authUser.UserID, authUser.UserID, employee)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"employee": map[string]any{
			"id":                      created.ID,
			"first_name":              created.FirstName,
			"last_name":               created.LastName,
			"email":                   created.Email,
			"phone":                   created.Phone,
			"employment_type":         created.EmploymentType,
			"status":                  created.Status,
			"department_id":           created.DepartmentID,
			"location_id":             created.LocationID,
			"work_schedule_id":        created.WorkScheduleID,
			"salary":                  domain.CentsToDollars(created.SalaryCents),
			"hourly_rate":             domain.CentsToDollars(created.HourlyRateCents),
			"overtime_rate":           domain.CentsToDollars(created.OvertimeRateCents),
			"start_date":              formatEmployeeToolDate(created.StartDate),
			"address":                 created.Address,
			"emergency_contact_name":  created.EmergencyContactName,
			"emergency_contact_phone": created.EmergencyContactPhone,
			"notes":                   created.Notes,
		},
	}, nil
}

func parseCreateEmployeeToolInput(input map[string]any) (createEmployeeToolInput, error) {
	firstName := strings.TrimSpace(hrToolString(input["first_name"]))
	lastName := strings.TrimSpace(hrToolString(input["last_name"]))
	if firstName == "" && lastName == "" {
		firstName, lastName = splitEmployeeName(hrToolString(input["name"]))
	}

	employmentType := normalizeEmploymentType(input["employment_type"])
	if employmentType == "" {
		return createEmployeeToolInput{}, errors.New("employment_type must be one of full_time, part_time, contract, or freelance")
	}

	departmentID, err := hrToolNonNegativeInt64(input["department_id"])
	if err != nil {
		return createEmployeeToolInput{}, fmt.Errorf("invalid department_id: %w", err)
	}
	locationID, err := hrToolNonNegativeInt64(input["location_id"])
	if err != nil {
		return createEmployeeToolInput{}, fmt.Errorf("invalid location_id: %w", err)
	}
	workScheduleID, err := hrToolNonNegativeInt64(input["work_schedule_id"])
	if err != nil {
		return createEmployeeToolInput{}, fmt.Errorf("invalid work_schedule_id: %w", err)
	}

	salary, err := hrToolNonNegativeFloat64(input["salary"])
	if err != nil {
		return createEmployeeToolInput{}, fmt.Errorf("invalid salary: %w", err)
	}
	hourlyRate, err := hrToolNonNegativeFloat64(input["hourly_rate"])
	if err != nil {
		return createEmployeeToolInput{}, fmt.Errorf("invalid hourly_rate: %w", err)
	}
	overtimeRate, err := hrToolNonNegativeFloat64(input["overtime_rate"])
	if err != nil {
		return createEmployeeToolInput{}, fmt.Errorf("invalid overtime_rate: %w", err)
	}
	startDate, err := optionalEmployeeToolDate(input["start_date"])
	if err != nil {
		return createEmployeeToolInput{}, fmt.Errorf("invalid start_date: %w", err)
	}

	parsed := createEmployeeToolInput{
		DepartmentID:          departmentID,
		LocationID:            locationID,
		WorkScheduleID:        workScheduleID,
		FirstName:             firstName,
		LastName:              lastName,
		Email:                 strings.TrimSpace(hrToolString(input["email"])),
		Phone:                 strings.TrimSpace(hrToolString(input["phone"])),
		EmploymentType:        employmentType,
		Salary:                salary,
		HourlyRate:            hourlyRate,
		OvertimeRate:          overtimeRate,
		StartDate:             startDate,
		Address:               strings.TrimSpace(hrToolString(input["address"])),
		EmergencyContactName:  strings.TrimSpace(hrToolString(input["emergency_contact_name"])),
		EmergencyContactPhone: strings.TrimSpace(hrToolString(input["emergency_contact_phone"])),
		Notes:                 strings.TrimSpace(hrToolString(input["notes"])),
	}
	if err := domain.ValidateEmployee(domain.EmployeeRecord{
		FirstName:       parsed.FirstName,
		LastName:        parsed.LastName,
		Email:           parsed.Email,
		SalaryCents:     domain.DollarsToCents(parsed.Salary),
		HourlyRateCents: domain.DollarsToCents(parsed.HourlyRate),
		EmploymentType:  parsed.EmploymentType,
	}); err != nil {
		return createEmployeeToolInput{}, err
	}
	return parsed, nil
}

func parseGetEmployeeToolInput(input map[string]any) (int64, error) {
	employeeID, err := hrToolNonNegativeInt64(firstHRToolValue(input, "employee_id", "employee", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid employee_id: %w", err)
	}
	if employeeID <= 0 {
		return 0, errors.New("employee_id is required")
	}
	return employeeID, nil
}

func parseListHRDepartmentsToolInput(input map[string]any) (listHRDepartmentsToolInput, error) {
	locationID, err := hrToolNonNegativeInt64(firstHRToolValue(input, "location_id", "location"))
	if err != nil {
		return listHRDepartmentsToolInput{}, fmt.Errorf("invalid location_id: %w", err)
	}
	return listHRDepartmentsToolInput{LocationID: locationID}, nil
}

func parseListEmployeesToolInput(input map[string]any) (listEmployeesToolInput, error) {
	limit, err := hrToolNonNegativeInt64(firstHRToolValue(input, "limit", "page_size"))
	if err != nil {
		return listEmployeesToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListEmployeesToolLimit
	}
	if limit > maxListEmployeesToolLimit {
		limit = maxListEmployeesToolLimit
	}

	page, err := hrToolNonNegativeInt64(firstHRToolValue(input, "page"))
	if err != nil {
		return listEmployeesToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	var offset int64
	if offsetValue := firstHRToolValue(input, "offset"); offsetValue != nil {
		offset, err = hrToolNonNegativeInt64(offsetValue)
		if err != nil {
			return listEmployeesToolInput{}, fmt.Errorf("invalid offset: %w", err)
		}
	} else {
		offset = (page - 1) * limit
	}
	if offset > math.MaxInt32 {
		return listEmployeesToolInput{}, errors.New("offset is too large")
	}

	return listEmployeesToolInput{
		Limit:  int32(limit),
		Offset: int32(offset),           // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Page:   int32(offset/limit) + 1, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}, nil
}

func firstHRToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func employeeToolPayload(e domain.EmployeeRecord) map[string]any {
	fullName := strings.TrimSpace(strings.TrimSpace(e.FirstName) + " " + strings.TrimSpace(e.LastName))
	return map[string]any{
		"id":               e.ID,
		"full_name":        fullName,
		"first_name":       e.FirstName,
		"last_name":        e.LastName,
		"email":            e.Email,
		"phone":            e.Phone,
		"employment_type":  e.EmploymentType,
		"status":           e.Status,
		"department_id":    e.DepartmentID,
		"location_id":      e.LocationID,
		"work_schedule_id": e.WorkScheduleID,
		"salary":           domain.CentsToDollars(e.SalaryCents),
		"hourly_rate":      domain.CentsToDollars(e.HourlyRateCents),
		"start_date":       formatEmployeeToolDate(e.StartDate),
		"end_date":         formatEmployeeToolDate(e.EndDate),
	}
}

func employeeDetailToolPayload(e domain.EmployeeRecord) map[string]any {
	payload := employeeToolPayload(e)
	payload["overtime_rate"] = domain.CentsToDollars(e.OvertimeRateCents)
	payload["address"] = e.Address
	payload["emergency_contact_name"] = e.EmergencyContactName
	payload["emergency_contact_phone"] = e.EmergencyContactPhone
	payload["notes"] = e.Notes
	payload["created_at"] = formatHRToolTimestamp(e.CreatedAt)
	payload["updated_at"] = formatHRToolTimestamp(e.UpdatedAt)
	return payload
}

func locationToolPayload(location domain.LocationRecord) map[string]any {
	return map[string]any{
		"id":         location.ID,
		"name":       location.Name,
		"address":    location.Address,
		"city":       location.City,
		"country":    location.Country,
		"timezone":   location.Timezone,
		"latitude":   location.Latitude,
		"longitude":  location.Longitude,
		"is_default": location.IsDefault,
	}
}

func departmentToolPayload(department domain.DepartmentRecord) map[string]any {
	return map[string]any{
		"id":                       department.ID,
		"name":                     department.Name,
		"location_id":              hrToolOptionalID(department.LocationID),
		"parent_department_id":     hrToolOptionalID(department.ParentDepartmentID),
		"default_work_schedule_id": hrToolOptionalID(department.DefaultWorkScheduleID),
	}
}

func workScheduleToolPayload(schedule domain.WorkScheduleRecord) map[string]any {
	return map[string]any{
		"id":            schedule.ID,
		"name":          schedule.Name,
		"schedule_data": hrToolJSONPayload(schedule.ScheduleData),
		"is_default":    schedule.IsDefault,
	}
}

func roleToolPayload(role domain.RoleRecord) map[string]any {
	return map[string]any{
		"id":              role.ID,
		"name":            role.Name,
		"description":     role.Description,
		"is_system":       role.IsSystem,
		"permissions":     hrToolJSONPayload(role.Permissions),
		"permission_keys": roleToolPermissionKeys(role.Permissions),
	}
}

func setupStepToolPayload(step domain.SetupStepRecord) map[string]any {
	return map[string]any{
		"id":           step.ID,
		"step_key":     step.StepKey,
		"step_label":   step.StepLabel,
		"is_completed": step.IsCompleted,
		"completed_at": formatHRToolTimestamp(step.CompletedAt),
	}
}

func hrConfigToolPayload(config *domain.HRCompanyConfigRecord) map[string]any {
	if config == nil {
		return nil
	}
	return map[string]any{
		"id":                       config.ID,
		"company_name":             config.CompanyName,
		"fiscal_year_start_month":  config.FiscalYearStartMonth,
		"default_currency":         config.DefaultCurrency,
		"default_timezone":         config.DefaultTimezone,
		"default_work_schedule_id": hrToolOptionalID(config.DefaultWorkScheduleID),
		"overtime_enabled":         config.OvertimeEnabled,
		"overtime_multiplier":      config.OvertimeMultiplier,
		"created_at":               formatHRToolTimestamp(config.CreatedAt),
		"updated_at":               formatHRToolTimestamp(config.UpdatedAt),
	}
}

func hrToolOptionalID(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}

func hrToolJSONPayload(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return trimmed
	}
	return payload
}

func roleToolPermissionKeys(raw string) []string {
	var permissions map[string]bool
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &permissions); err != nil {
		return nil
	}

	keys := make([]string, 0, len(permissions))
	for key, enabled := range permissions {
		if enabled {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func splitEmployeeName(value string) (string, string) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", ""
	}
	parts := strings.Fields(name)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func normalizeEmploymentType(value any) string {
	employeeType := strings.ToLower(strings.TrimSpace(hrToolString(value)))
	if employeeType == "" {
		return domain.EmploymentTypeFullTime
	}
	employeeType = strings.ReplaceAll(employeeType, "-", "_")
	employeeType = strings.ReplaceAll(employeeType, " ", "_")
	switch employeeType {
	case domain.EmploymentTypeFullTime, domain.EmploymentTypePartTime, domain.EmploymentTypeContract, domain.EmploymentTypeFreelance:
		return employeeType
	default:
		return ""
	}
}

func hrToolNonNegativeInt64(value any) (int64, error) {
	out, err := hrToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func hrToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func hrToolNonNegativeFloat64(value any) (float64, error) {
	out, err := hrToolFloat64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func hrToolFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseFloat(raw, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func optionalEmployeeToolDate(value any) (time.Time, error) {
	switch v := value.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return time.Time{}, nil
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return time.Time{}, fmt.Errorf("unsupported value %T", value)
	}
}

func hrToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func formatEmployeeToolDate(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.Format("2006-01-02")
}

func formatHRToolTimestamp(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}
