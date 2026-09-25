package hr_core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

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
		"salary":           money.CentsToDollars(e.SalaryCents),
		"hourly_rate":      money.CentsToDollars(e.HourlyRateCents),
		"start_date":       formatEmployeeToolDate(e.StartDate),
		"end_date":         formatEmployeeToolDate(e.EndDate),
	}
}

func employeeDetailToolPayload(e domain.EmployeeRecord) map[string]any {
	payload := employeeToolPayload(e)
	payload["overtime_rate"] = money.CentsToDollars(e.OvertimeRateCents)
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
