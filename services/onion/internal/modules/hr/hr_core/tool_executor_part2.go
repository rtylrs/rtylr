package hr_core

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

func (m *Module) executeListHRRolesTool(ctx context.Context) (map[string]any, error) {
	businessID, _, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	roles, err := m.service.ListRoles(ctx, businessID)
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
	businessID, _, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	steps, err := m.service.GetSetupStatus(ctx, businessID)
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
	businessID, _, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
	}

	config, err := m.service.GetHRConfig(ctx, businessID)
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
	businessID, callerUserID, scopeErr := hrCoreToolScope(ctx)
	if scopeErr != nil {
		return nil, scopeErr
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
		SalaryCents:           money.DollarsToCents(parsed.Salary),
		HourlyRateCents:       money.DollarsToCents(parsed.HourlyRate),
		OvertimeRateCents:     money.DollarsToCents(parsed.OvertimeRate),
		StartDate:             parsed.StartDate,
		Address:               parsed.Address,
		EmergencyContactName:  parsed.EmergencyContactName,
		EmergencyContactPhone: parsed.EmergencyContactPhone,
		Notes:                 parsed.Notes,
	}

	created, err := m.service.CreateEmployeeForWriter(ctx, businessID, callerUserID, employee)
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
			"salary":                  money.CentsToDollars(created.SalaryCents),
			"hourly_rate":             money.CentsToDollars(created.HourlyRateCents),
			"overtime_rate":           money.CentsToDollars(created.OvertimeRateCents),
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
		SalaryCents:     money.DollarsToCents(parsed.Salary),
		HourlyRateCents: money.DollarsToCents(parsed.HourlyRate),
		EmploymentType:  parsed.EmploymentType,
	}); err != nil {
		return createEmployeeToolInput{}, err
	}
	return parsed, nil
}
