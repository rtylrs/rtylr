package infrastructure

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func employeeToParams(businessID int64, emp domain.EmployeeRecord) sqlc.CreateEmployeeParams {
	return sqlc.CreateEmployeeParams{
		OwnerID:               businessID,
		BusinessID:            businessID,
		UserID:                sqlutil.NullInt64(emp.UserID),
		DepartmentID:          sqlutil.NullInt64(emp.DepartmentID),
		LocationID:            sqlutil.NullInt64(emp.LocationID),
		WorkScheduleID:        sqlutil.NullInt64(emp.WorkScheduleID),
		FirstName:             emp.FirstName,
		LastName:              emp.LastName,
		Email:                 emp.Email,
		Phone:                 sqlutil.NullString(emp.Phone),
		EmploymentType:        emp.EmploymentType,
		Status:                emp.Status,
		Salary:                centsToNullDecimalString(emp.SalaryCents),
		HourlyRate:            centsToNullDecimalString(emp.HourlyRateCents),
		OvertimeRate:          centsToNullDecimalString(emp.OvertimeRateCents),
		StartDate:             nullTime(emp.StartDate),
		EndDate:               nullTime(emp.EndDate),
		Address:               sqlutil.NullString(emp.Address),
		EmergencyContactName:  sqlutil.NullString(emp.EmergencyContactName),
		EmergencyContactPhone: sqlutil.NullString(emp.EmergencyContactPhone),
		Notes:                 sqlutil.NullString(emp.Notes),
	}
}

func locationFromRow(row sqlc.HrLocation) domain.LocationRecord {
	return domain.LocationRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Name:       row.Name,
		Address:    row.Address.String,
		City:       row.City.String,
		Country:    row.Country.String,
		Timezone:   row.Timezone.String,
		Latitude:   row.Latitude.Float64,
		Longitude:  row.Longitude.Float64,
		IsDefault:  row.IsDefault,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func departmentFromRow(row sqlc.HrDepartment) domain.DepartmentRecord {
	return domain.DepartmentRecord{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
		BusinessID:            row.BusinessID,
		LocationID:            row.LocationID,
		ParentDepartmentID:    row.ParentDepartmentID.Int64,
		Name:                  row.Name,
		DefaultWorkScheduleID: row.DefaultWorkScheduleID.Int64,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func workScheduleFromRow(row sqlc.HrWorkSchedule) domain.WorkScheduleRecord {
	return domain.WorkScheduleRecord{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		BusinessID:   row.BusinessID,
		Name:         row.Name,
		ScheduleData: string(row.ScheduleData),
		IsDefault:    row.IsDefault,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func roleFromRow(row sqlc.HrRole) domain.RoleRecord {
	return domain.RoleRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Name:        row.Name,
		Description: row.Description.String,
		Permissions: string(row.Permissions),
		IsSystem:    row.IsSystem,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func setupStepFromRow(row sqlc.HrSetupStep) domain.SetupStepRecord {
	return domain.SetupStepRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		StepKey:     row.StepKey,
		StepLabel:   row.StepLabel,
		IsCompleted: row.IsCompleted,
		CompletedAt: row.CompletedAt.Time,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// centsToNullDecimalString renders an integer cents amount as the DECIMAL(10,2)
// string sqlc expects, using only integer arithmetic so no float64 rounding
// error can be introduced on the way into the database. A zero amount is
// stored as NULL, matching the previous float-based helper's behaviour.
func centsToNullDecimalString(cents int64) sql.NullString {
	if cents == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: centsToDecimalString(cents), Valid: true}
}

func centsToDecimalString(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	whole := cents / 100
	fraction := cents % 100
	s := strconv.FormatInt(whole, 10) + "." + fmt.Sprintf("%02d", fraction)
	if negative {
		s = "-" + s
	}
	return s
}

func nullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: v != 0}
}

func nullTime(v time.Time) sql.NullTime {
	return sql.NullTime{Time: v, Valid: !v.IsZero()}
}

// nullDecimalStringToCents parses a MySQL DECIMAL(10,2) value read back as a
// string into integer cents. It works on the string digits directly rather
// than through strconv.ParseFloat, so no float64 rounding error can be
// introduced on the way out of the database.
func nullDecimalStringToCents(v sql.NullString) int64 {
	if !v.Valid {
		return 0
	}
	cents, err := decimalStringToCents(v.String)
	if err != nil {
		// DECIMAL values from MySQL should always parse; log corruption and return zero
		logger.Global().Warn("failed to parse nullable decimal from database", logger.String("value", v.String), logger.Error(err))
		return 0
	}
	return cents
}

func decimalStringToCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}

	whole, fraction, hasFraction := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if len(fraction) > 2 {
		fraction = fraction[:2]
	}
	for len(fraction) < 2 {
		fraction += "0"
	}
	if !hasFraction {
		fraction = "00"
	}

	wholeCents, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, err
	}
	fractionCents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, err
	}

	total := wholeCents*100 + fractionCents
	if negative {
		total = -total
	}
	return total, nil
}

func parseFloat(v string) float64 {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		// DECIMAL values from MySQL should always parse; log corruption and return zero
		logger.Global().Warn("failed to parse decimal from database", logger.String("value", v), logger.Error(err))
		return 0
	}
	return f
}

// defaultedScheduleData keeps an absent schedule out of the JSON column as an
// empty object rather than an empty string, which MySQL will not accept.
func defaultedScheduleData(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "{}"
	}

	return raw
}
