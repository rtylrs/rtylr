package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowtimev1 "github.com/voxire/voxie/gen/go/flow/time/v1"
)

type WorkTimeAPI struct {
	Service *application.Service
}

func NewWorkTimeAPI(service *application.Service) *WorkTimeAPI {
	return &WorkTimeAPI{Service: service}
}

// scope resolves the tenant for a request, and the employee the call acts on.
//
// A caller may pass employee_id to record or read a day for someone else,
// which is how a manager fills in for a field worker who could not, or reviews
// a report's day. That path requires a real permission check, which does not
// exist yet: no caller in this codebase can currently resolve "does this user
// have standing over that employee" without new cross-module plumbing (there
// is no employee-to-workspace resolution in this module today, and no
// business-wide permission check is actually enforced anywhere in the backend
// despite the frontend's permission matrix suggesting otherwise - see the
// amendment in docs/flow time section design for the full finding).
//
// Until that lands, any employee_id that differs from the caller is denied
// rather than trusted. This changes nothing a real user could reach: the one
// frontend path that would have exercised it (TaskDetail's canLogForOthers)
// already hardcodes false. Denying it here closes an API-level hole where any
// authenticated user could read or write any other employee's work day simply
// by passing their id.
//
// ErrNoBusinessScope is denied rather than defaulted: treating an unresolved
// scope as "no filter" is what turns a tenancy bug into a cross-tenant read, and
// the platform's own tenant package says so explicitly.
// callerMayActFor reports whether a caller may read or write the work day of the
// given employee id. A zero id means "whoever I am", which is always allowed;
// any other value must be the caller's own.
//
// Split out of scope() because the rest of scope() needs a database: it goes on
// to call tenant.ResolveBusinessID, which fails closed when the pool is nil. That
// made the allow case impossible to observe through scope(), since a permitted
// caller and a rejected one both came back as the same PermissionDenied. The
// predicate is where the decision actually lives, so it is what the test asserts
// on. See worktime_scope_internal_test.go.
func callerMayActFor(callerUserID, employeeID int64) bool {
	return employeeID <= 0 || employeeID == callerUserID
}

func (a *WorkTimeAPI) scope(ctx context.Context, employeeID int64) (businessID, employee int64, err error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, 0, onekiterrors.PermissionDenied()
	}

	if !callerMayActFor(authUser.UserID, employeeID) {
		return 0, 0, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}

	return businessID, authUser.UserID, nil
}

// actor returns the signed-in user, which is recorded as logged_by so a day
// filled in by a manager is never mistaken for the person's own entry.
func actor(ctx context.Context) int64 {
	if authUser, ok := authPkg.GetAuthUser(ctx); ok {
		return authUser.UserID
	}
	return 0
}

// mapServiceError turns a domain rule into the status a caller should see.
// Without this every rule reads as a 500, and a locked day looks like an outage.
func mapServiceError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidWorkMode),
		errors.Is(err, application.ErrWorkTypeTooDeep):
		return onekiterrors.New(http.StatusBadRequest, "Invalid work type")
	case errors.Is(err, application.ErrProjectRequired):
		return onekiterrors.New(http.StatusBadRequest, "Project is required")
	case errors.Is(err, application.ErrWorkDayNotFound):
		return onekiterrors.New(http.StatusNotFound, "Work day not found")
	case errors.Is(err, application.ErrWorkTypeNotFound):
		return onekiterrors.New(http.StatusNotFound, "Work type not found")
	case errors.Is(err, application.ErrDayLocked):
		return onekiterrors.New(http.StatusConflict, "Work day is locked")
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	default:
		return workspaceInternalError(ctx, "Worktime operation failed", err)
	}
}

func toContractWorkDay(day domain.WorkDay) *flowtimev1.WorkDay {
	return &flowtimev1.WorkDay{
		Id:             day.ID,
		EmployeeId:     day.EmployeeID,
		UserId:         ptrInt64(day.UserID),
		Date:           day.Date.Format("2006-01-02"),
		Timezone:       day.Timezone,
		WorkMode:       day.WorkMode,
		StartTime:      ptrString(day.StartTime),
		EndTime:        ptrString(day.EndTime),
		BreakMinutes:   day.BreakMinutes,
		Status:         day.Status,
		ClockedMinutes: day.ClockedMinutes,
		LoggedMinutes:  day.LoggedMinutes,
		GapMinutes:     day.GapMinutes,
		LoggedBy:       ptrInt64(day.LoggedBy),
		ApprovedBy:     ptrInt64(day.ApprovedBy),
		ApprovedAt:     ptrString(formatTime(day.ApprovedAt)),
		ReopenedBy:     ptrInt64(day.ReopenedBy),
		ReopenedAt:     ptrString(formatTime(day.ReopenedAt)),
		CreatedAt:      formatTime(day.CreatedAt),
		UpdatedAt:      formatTime(day.UpdatedAt),
	}
}

func toContractLine(line domain.WorkDayLine) *flowtimev1.WorkDayLine {
	return &flowtimev1.WorkDayLine{
		Id:              line.ID,
		WorkDayId:       line.WorkDayID,
		ProjectId:       ptrInt64(line.ProjectID),
		TaskId:          ptrInt64(line.TaskID),
		WorkTypeId:      ptrInt64(line.WorkTypeID),
		WorkMode:        ptrString(line.WorkMode),
		Description:     line.Description,
		DurationMinutes: line.DurationMinutes,
		IsBillable:      line.IsBillable,
		Status:          line.Status,
		ClientKey:       line.ClientKey,
		LoggedBy:        ptrInt64(line.LoggedBy),
		CreatedAt:       formatTime(line.CreatedAt),
		UpdatedAt:       formatTime(line.UpdatedAt),
	}
}

func toContractWorkType(workType domain.WorkType) *flowtimev1.WorkType {
	return &flowtimev1.WorkType{
		Id:           workType.ID,
		ParentId:     ptrInt64(workType.ParentID),
		Name:         workType.Name,
		Description:  workType.Description,
		HrCategoryId: ptrInt64(workType.HRCategoryID),
		SortOrder:    workType.SortOrder,
		Archived:     workType.Archived,
		CreatedAt:    formatTime(workType.CreatedAt),
		UpdatedAt:    formatTime(workType.UpdatedAt),
	}
}

// formatTime renders a timestamp, or an empty string for the zero value, so an
// unset approval does not serialise as year one.
func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

// The contract marks these fields nullable, so the generated Go type uses
// pointers. Zero and empty map to nil rather than to 0 and "": an unapproved day
// has no approver, and serialising one as user 0 would invent a person.
func ptrString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func ptrInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func (a *WorkTimeAPI) GetWorkDay(ctx context.Context, req *flowtimev1.GetWorkDayRequest) (*flowtimev1.GetWorkDayResponse, error) {
	businessID, employeeID, err := a.scope(ctx, req.GetEmployeeId())
	if err != nil {
		return nil, err
	}

	day, lines, found, err := a.Service.GetWorkDay(ctx, businessID, employeeID, req.GetDate())
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	// A day nobody has filled in yet is an empty response, not a 404: the client
	// renders a blank day for that date either way.
	if !found {
		return &flowtimev1.GetWorkDayResponse{Lines: []*flowtimev1.WorkDayLine{}}, nil
	}

	out := make([]*flowtimev1.WorkDayLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, toContractLine(line))
	}
	return &flowtimev1.GetWorkDayResponse{WorkDay: toContractWorkDay(day), Lines: out}, nil
}

func (a *WorkTimeAPI) UpsertWorkDay(ctx context.Context, req *flowtimev1.UpsertWorkDayRequest) (*flowtimev1.UpsertWorkDayResponse, error) {
	businessID, employeeID, err := a.scope(ctx, req.GetEmployeeId())
	if err != nil {
		return nil, err
	}

	day, err := a.Service.UpsertWorkDay(ctx, domain.WorkDayInput{
		BusinessID:   businessID,
		EmployeeID:   employeeID,
		Date:         req.GetDate(),
		Timezone:     req.GetTimezone(),
		WorkMode:     req.GetWorkMode(),
		StartTime:    req.GetStartTime(),
		EndTime:      req.GetEndTime(),
		BreakMinutes: req.GetBreakMinutes(),
		LoggedBy:     actor(ctx),
	})
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	return &flowtimev1.UpsertWorkDayResponse{WorkDay: toContractWorkDay(day)}, nil
}

func (a *WorkTimeAPI) ListWorkDays(ctx context.Context, req *flowtimev1.ListWorkDaysRequest) (*flowtimev1.ListWorkDaysResponse, error) {
	businessID, employeeID, err := a.scope(ctx, req.GetEmployeeId())
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	days, total, err := a.Service.ListWorkDays(ctx, businessID, employeeID, req.GetFromDate(), req.GetToDate(), limit, offset)
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}

	out := make([]*flowtimev1.WorkDay, 0, len(days))
	for _, day := range days {
		out = append(out, toContractWorkDay(day))
	}
	return &flowtimev1.ListWorkDaysResponse{
		WorkDays:   out,
		Pagination: buildPagination(page, limit, total, len(out)),
	}, nil
}

// ListWorkDaysForWorkspace lists work days across a workspace's roster, for
// the team view. Unlike scope(), this is not about the caller's own day: the
// caller is asking to see other people's, which is exactly what scope() now
// refuses to trust an argument for. Authorization here is Service.
// RequireWorkspaceRole, checked inside ListWorkDaysForWorkspace itself against
// the workspace being asked about, not against a claimed employee identity.
func (a *WorkTimeAPI) ListWorkDaysForWorkspace(ctx context.Context, req *flowtimev1.ListWorkDaysForWorkspaceRequest) (*flowtimev1.ListWorkDaysForWorkspaceResponse, error) {
	businessID, callerID, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	days, total, err := a.Service.ListWorkDaysForWorkspace(
		ctx, callerID, req.GetWorkspaceId(), businessID, req.GetFromDate(), req.GetToDate(), limit, offset,
	)
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}

	out := make([]*flowtimev1.WorkDay, 0, len(days))
	for _, day := range days {
		out = append(out, toContractWorkDay(day))
	}
	return &flowtimev1.ListWorkDaysForWorkspaceResponse{
		WorkDays:   out,
		Pagination: buildPagination(page, limit, total, len(out)),
	}, nil
}

func (a *WorkTimeAPI) CreateWorkDayLine(ctx context.Context, req *flowtimev1.CreateWorkDayLineRequest) (*flowtimev1.CreateWorkDayLineResponse, error) {
	businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	line, err := a.Service.CreateWorkDayLine(ctx, domain.WorkDayLineInput{
		BusinessID:      businessID,
		WorkDayID:       req.GetWorkDayId(),
		ProjectID:       req.GetProjectId(),
		TaskID:          req.GetTaskId(),
		WorkTypeID:      req.GetWorkTypeId(),
		WorkMode:        req.GetWorkMode(),
		Description:     req.GetDescription(),
		DurationMinutes: req.GetDurationMinutes(),
		IsBillable:      req.GetIsBillable(),
		ClientKey:       req.GetClientKey(),
		LoggedBy:        actor(ctx),
	})
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	return &flowtimev1.CreateWorkDayLineResponse{Line: toContractLine(line)}, nil
}

func (a *WorkTimeAPI) UpdateWorkDayLine(ctx context.Context, req *flowtimev1.UpdateWorkDayLineRequest) (*flowtimev1.UpdateWorkDayLineResponse, error) {
	businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	line, err := a.Service.UpdateWorkDayLine(ctx, domain.WorkDayLineInput{
		ID:              req.GetId(),
		BusinessID:      businessID,
		ProjectID:       req.GetProjectId(),
		TaskID:          req.GetTaskId(),
		WorkTypeID:      req.GetWorkTypeId(),
		WorkMode:        req.GetWorkMode(),
		Description:     req.GetDescription(),
		DurationMinutes: req.GetDurationMinutes(),
		IsBillable:      req.GetIsBillable(),
		LoggedBy:        actor(ctx),
	})
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	return &flowtimev1.UpdateWorkDayLineResponse{Line: toContractLine(line)}, nil
}

func (a *WorkTimeAPI) DeleteWorkDayLine(ctx context.Context, req *flowtimev1.DeleteWorkDayLineRequest) (*flowtimev1.DeleteWorkDayLineResponse, error) {
	businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	removed, err := a.Service.DeleteWorkDayLine(ctx, req.GetId(), businessID, actor(ctx))
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	// Reporting 404 rather than success on a foreign id: a caller that deleted
	// nothing should not be told it worked.
	if !removed {
		return nil, onekiterrors.New(http.StatusNotFound, "time entry not found")
	}
	return &flowtimev1.DeleteWorkDayLineResponse{Message: "deleted"}, nil
}

func (a *WorkTimeAPI) ListWorkTypes(ctx context.Context, req *flowtimev1.ListWorkTypesRequest) (*flowtimev1.ListWorkTypesResponse, error) {
	businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	types, err := a.Service.ListWorkTypes(ctx, businessID, req.GetIncludeArchived())
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}

	out := make([]*flowtimev1.WorkType, 0, len(types))
	for _, workType := range types {
		out = append(out, toContractWorkType(workType))
	}
	return &flowtimev1.ListWorkTypesResponse{WorkTypes: out}, nil
}

func (a *WorkTimeAPI) CreateWorkType(ctx context.Context, req *flowtimev1.CreateWorkTypeRequest) (*flowtimev1.CreateWorkTypeResponse, error) {
	businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	workType, err := a.Service.CreateWorkType(ctx, domain.WorkTypeInput{
		BusinessID:   businessID,
		ParentID:     req.GetParentId(),
		Name:         req.GetName(),
		Description:  req.GetDescription(),
		HRCategoryID: req.GetHrCategoryId(),
		SortOrder:    req.GetSortOrder(),
	})
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	return &flowtimev1.CreateWorkTypeResponse{WorkType: toContractWorkType(workType)}, nil
}

func (a *WorkTimeAPI) UpdateWorkType(ctx context.Context, req *flowtimev1.UpdateWorkTypeRequest) (*flowtimev1.UpdateWorkTypeResponse, error) {
	businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	workType, err := a.Service.UpdateWorkType(ctx, domain.WorkTypeInput{
		ID:           req.GetId(),
		BusinessID:   businessID,
		ParentID:     req.GetParentId(),
		Name:         req.GetName(),
		Description:  req.GetDescription(),
		HRCategoryID: req.GetHrCategoryId(),
		SortOrder:    req.GetSortOrder(),
		Archived:     req.GetArchived(),
	})
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	return &flowtimev1.UpdateWorkTypeResponse{WorkType: toContractWorkType(workType)}, nil
}

func (a *WorkTimeAPI) DeleteWorkType(ctx context.Context, req *flowtimev1.DeleteWorkTypeRequest) (*flowtimev1.DeleteWorkTypeResponse, error) {
	businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	removed, err := a.Service.DeleteWorkType(ctx, req.GetId(), businessID)
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	if !removed {
		return nil, onekiterrors.New(http.StatusNotFound, "work type not found")
	}
	return &flowtimev1.DeleteWorkTypeResponse{Message: "deleted"}, nil
}
