package infrastructure

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
)

const dateLayout = "2006-01-02"

// parseDate reads a YYYY-MM-DD contract value into a time.Time.
func parseDate(value string) (time.Time, error) {
	return time.Parse(dateLayout, value)
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v > 0}
}

func int64Value(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}

func stringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

// sumToInt64 narrows the interface{} that sqlc returns for COALESCE(SUM(...)).
// MySQL hands a SUM back as a decimal, which the driver may deliver as []byte,
// int64 or float64 depending on the column and the connection; all three are
// handled rather than assuming one and panicking on the others.
func sumToInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		parsed, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func timeValue(v sql.NullTime) time.Time {
	if v.Valid {
		return v.Time
	}
	return time.Time{}
}

// The narrowed worktime queries each generate a distinct Row struct even
// though they select identical column lists; these explicit adapters let the
// shared mappers stay single-sourced.

func workDayFromGet(r sqlc.GetWorkDayRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func workDayFromGetByID(r sqlc.GetWorkDayByIDRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func lineFromList(r sqlc.ListWorkDayLinesRow) sqlc.HrTimeEntry {
	return sqlc.HrTimeEntry{
		ID: r.ID, WorkDayID: r.WorkDayID, ProjectID: r.ProjectID, TaskID: r.TaskID,
		WorkTypeID: r.WorkTypeID, WorkMode: r.WorkMode, Description: r.Description,
		DurationMinutes: r.DurationMinutes, IsBillable: r.IsBillable, Status: r.Status,
		ClientKey: r.ClientKey, LoggedBy: r.LoggedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func workDayFromEmployeeIDs(r sqlc.ListWorkDaysByEmployeeIDsRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func workDayFromList(r sqlc.ListWorkDaysRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func lineFromGet(r sqlc.GetTimeEntryRow) sqlc.HrTimeEntry {
	return sqlc.HrTimeEntry{
		ID: r.ID, WorkDayID: r.WorkDayID, ProjectID: r.ProjectID, TaskID: r.TaskID,
		WorkTypeID: r.WorkTypeID, WorkMode: r.WorkMode, Description: r.Description,
		DurationMinutes: r.DurationMinutes, IsBillable: r.IsBillable, Status: r.Status,
		ClientKey: r.ClientKey, LoggedBy: r.LoggedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func lineFromClientKey(r sqlc.GetTimeEntryByClientKeyRow) sqlc.HrTimeEntry {
	return sqlc.HrTimeEntry{
		ID: r.ID, WorkDayID: r.WorkDayID, ProjectID: r.ProjectID, TaskID: r.TaskID,
		WorkTypeID: r.WorkTypeID, WorkMode: r.WorkMode, Description: r.Description,
		DurationMinutes: r.DurationMinutes, IsBillable: r.IsBillable, Status: r.Status,
		ClientKey: r.ClientKey, LoggedBy: r.LoggedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func workTypeFromGet(r sqlc.GetWorkTypeRow) sqlc.FlowWorkType {
	return sqlc.FlowWorkType{
		ID: r.ID, ParentID: r.ParentID, Name: r.Name, Description: r.Description,
		HrCategoryID: r.HrCategoryID, SortOrder: r.SortOrder, Archived: r.Archived,
	}
}

func workTypeFromList(r sqlc.ListWorkTypesRow) sqlc.FlowWorkType {
	return sqlc.FlowWorkType{
		ID: r.ID, ParentID: r.ParentID, Name: r.Name, Description: r.Description,
		HrCategoryID: r.HrCategoryID, SortOrder: r.SortOrder, Archived: r.Archived,
	}
}

func toDomainWorkDay(row sqlc.FlowWorkDay) domain.WorkDay {
	day := domain.WorkDay{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		Date:         row.Date,
		Timezone:     row.Timezone,
		WorkMode:     row.WorkMode,
		StartTime:    stringValue(row.StartTime),
		EndTime:      stringValue(row.EndTime),
		BreakMinutes: row.BreakMinutes,
		Status:       row.Status,
		LoggedBy:     int64Value(row.LoggedBy),
		ApprovedBy:   int64Value(row.ApprovedBy),
		ApprovedAt:   timeValue(row.ApprovedAt),
		ReopenedBy:   int64Value(row.ReopenedBy),
		ReopenedAt:   timeValue(row.ReopenedAt),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	day.ClockedMinutes = domain.ClockedMinutes(day.StartTime, day.EndTime, day.BreakMinutes)
	return day
}

func toDomainLine(row sqlc.HrTimeEntry) domain.WorkDayLine {
	return domain.WorkDayLine{
		ID:              row.ID,
		WorkDayID:       int64Value(row.WorkDayID),
		ProjectID:       int64Value(row.ProjectID),
		TaskID:          int64Value(row.TaskID),
		WorkTypeID:      int64Value(row.WorkTypeID),
		WorkMode:        stringValue(row.WorkMode),
		Description:     stringValue(row.Description),
		DurationMinutes: row.DurationMinutes,
		IsBillable:      row.IsBillable,
		Status:          row.Status,
		ClientKey:       stringValue(row.ClientKey),
		LoggedBy:        int64Value(row.LoggedBy),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func toDomainWorkType(row sqlc.FlowWorkType) domain.WorkType {
	return domain.WorkType{
		ID:           row.ID,
		ParentID:     int64Value(row.ParentID),
		Name:         row.Name,
		Description:  stringValue(row.Description),
		HRCategoryID: int64Value(row.HrCategoryID),
		SortOrder:    row.SortOrder,
		Archived:     row.Archived,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
