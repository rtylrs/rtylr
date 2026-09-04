package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowtimev2 "github.com/voxire/voxie/gen/go/flow/time/v2"
)

// WorkTimeAPIV2 serves flow.time.v2. listTeamWorkDays replaces the workspace-
// scoped team view: the roster is the caller's business team, and who may see
// whose days is governed by the business's configured visibility policy (W1),
// enforced server-side.
type WorkTimeAPIV2 struct {
	Service *application.Service
}

func NewWorkTimeAPIV2(service *application.Service) *WorkTimeAPIV2 {
	return &WorkTimeAPIV2{Service: service}
}

func toContractWorkDayV2(d domain.WorkDay) *flowtimev2.WorkDay {
	out := &flowtimev2.WorkDay{
		Id:             d.ID,
		EmployeeId:     d.EmployeeID,
		Date:           d.Date.Format("2006-01-02"),
		Timezone:       d.Timezone,
		WorkMode:       d.WorkMode,
		StartTime:      &d.StartTime,
		EndTime:        &d.EndTime,
		BreakMinutes:   d.BreakMinutes,
		Status:         d.Status,
		ClockedMinutes: d.ClockedMinutes,
		LoggedMinutes:  d.LoggedMinutes,
		GapMinutes:     d.GapMinutes,
	}
	if d.UserID != 0 {
		out.UserId = &d.UserID
	}
	return out
}

// ListTeamWorkDays is the v2 team view. Rosters resolve through hr_employees
// against the business team; the visibility policy decides which rows the
// caller sees.
func (a *WorkTimeAPIV2) ListTeamWorkDays(ctx context.Context, req *flowtimev2.ListTeamWorkDaysRequest) (*flowtimev2.ListTeamWorkDaysResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	days, total, err := a.Service.ListTeamWorkDays(
		ctx, authUser.UserID, businessID,
		req.GetFromDate(), req.GetToDate(), int64(limit), offset,
	)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("ListTeamWorkDays failed", logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list team work days")
	}

	out := make([]*flowtimev2.WorkDay, 0, len(days))
	for _, day := range days {
		out = append(out, toContractWorkDayV2(day))
	}
	return &flowtimev2.ListTeamWorkDaysResponse{
		WorkDays:   out,
		Pagination: buildPagination(page, limit, total, len(out)),
	}, nil
}

// scope resolves the tenant for a v2 time call.
//
// Mirrors the v1 helper deliberately, including the rule that a caller may only
// name themselves as the employee: callerMayActFor is where that decision lives,
// and asking to see someone else's day goes through ListTeamWorkDays, which
// checks the business visibility policy instead of trusting an argument.
func (a *WorkTimeAPIV2) scope(
	ctx context.Context,
	employeeID int64,
) (ownerID, businessID, employee int64, err error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, 0, 0, onekiterrors.PermissionDenied()
	}
	if !callerMayActFor(authUser.UserID, employeeID) {
		return 0, 0, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, 0, 0, onekiterrors.PermissionDenied()
	}
	ownerID, ownerErr := tenant.ResolveOwnerID(ctx, businessID, authUser.UserID)
	if ownerErr != nil {
		return 0, 0, 0, workspaceInternalError(ctx, "ResolveOwnerID failed", ownerErr)
	}
	return ownerID, businessID, authUser.UserID, nil
}

func toContractLineV2(line domain.WorkDayLine) *flowtimev2.WorkDayLine {
	return &flowtimev2.WorkDayLine{
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

// v2's WorkType carries no timestamps: a work type is a picker entry, and
// nothing in the UI showed when one was created.
func toContractWorkTypeV2(workType domain.WorkType) *flowtimev2.WorkType {
	return &flowtimev2.WorkType{
		Id:           workType.ID,
		ParentId:     ptrInt64(workType.ParentID),
		Name:         workType.Name,
		Description:  workType.Description,
		HrCategoryId: ptrInt64(workType.HRCategoryID),
		SortOrder:    workType.SortOrder,
		Archived:     workType.Archived,
	}
}

func (a *WorkTimeAPIV2) GetWorkDay(ctx context.Context, req *flowtimev2.GetWorkDayRequest) (*flowtimev2.GetWorkDayResponse, error) {
	ownerID, businessID, employeeID, err := a.scope(ctx, req.GetEmployeeId())
	if err != nil {
		return nil, err
	}

	day, lines, found, err := a.Service.GetWorkDay(ctx, ownerID, businessID, employeeID, req.GetDate())
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	// A day nobody has filled in yet is an empty response, not a 404: the client
	// renders a blank day for that date either way.
	if !found {
		return &flowtimev2.GetWorkDayResponse{Lines: []*flowtimev2.WorkDayLine{}}, nil
	}

	out := make([]*flowtimev2.WorkDayLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, toContractLineV2(line))
	}
	return &flowtimev2.GetWorkDayResponse{WorkDay: toContractWorkDayV2(day), Lines: out}, nil
}

func (a *WorkTimeAPIV2) UpsertWorkDay(ctx context.Context, req *flowtimev2.UpsertWorkDayRequest) (*flowtimev2.UpsertWorkDayResponse, error) {
	ownerID, businessID, employeeID, err := a.scope(ctx, req.GetEmployeeId())
	if err != nil {
		return nil, err
	}

	day, err := a.Service.UpsertWorkDay(ctx, domain.WorkDayInput{
		OwnerID:      ownerID,
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
	return &flowtimev2.UpsertWorkDayResponse{WorkDay: toContractWorkDayV2(day)}, nil
}

func (a *WorkTimeAPIV2) ListWorkDays(ctx context.Context, req *flowtimev2.ListWorkDaysRequest) (*flowtimev2.ListWorkDaysResponse, error) {
	ownerID, businessID, employeeID, err := a.scope(ctx, req.GetEmployeeId())
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	days, total, err := a.Service.ListWorkDays(ctx, ownerID, businessID, employeeID, req.GetFromDate(), req.GetToDate(), limit, offset)
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}

	out := make([]*flowtimev2.WorkDay, 0, len(days))
	for _, day := range days {
		out = append(out, toContractWorkDayV2(day))
	}
	return &flowtimev2.ListWorkDaysResponse{
		WorkDays:   out,
		Pagination: buildPagination(page, limit, total, len(out)),
	}, nil
}

func (a *WorkTimeAPIV2) CreateWorkDayLine(ctx context.Context, req *flowtimev2.CreateWorkDayLineRequest) (*flowtimev2.CreateWorkDayLineResponse, error) {
	_, businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	line, err := a.Service.CreateWorkDayLine(ctx, domain.WorkDayLineInput{
		OwnerID:         businessID,
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
	return &flowtimev2.CreateWorkDayLineResponse{Line: toContractLineV2(line)}, nil
}

func (a *WorkTimeAPIV2) UpdateWorkDayLine(ctx context.Context, req *flowtimev2.UpdateWorkDayLineRequest) (*flowtimev2.UpdateWorkDayLineResponse, error) {
	_, businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	line, err := a.Service.UpdateWorkDayLine(ctx, domain.WorkDayLineInput{
		ID:              req.GetId(),
		OwnerID:         businessID,
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
	return &flowtimev2.UpdateWorkDayLineResponse{Line: toContractLineV2(line)}, nil
}

func (a *WorkTimeAPIV2) DeleteWorkDayLine(ctx context.Context, req *flowtimev2.DeleteWorkDayLineRequest) (*flowtimev2.DeleteWorkDayLineResponse, error) {
	_, businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	removed, err := a.Service.DeleteWorkDayLine(ctx, req.GetId(), businessID, businessID, actor(ctx))
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	// 404 rather than success on a foreign id: a caller that deleted nothing
	// should not be told it worked.
	if !removed {
		return nil, onekiterrors.New(http.StatusNotFound, "time entry not found")
	}
	return &flowtimev2.DeleteWorkDayLineResponse{Message: "deleted"}, nil
}

func (a *WorkTimeAPIV2) ListWorkTypes(ctx context.Context, req *flowtimev2.ListWorkTypesRequest) (*flowtimev2.ListWorkTypesResponse, error) {
	ownerID, businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	types, err := a.Service.ListWorkTypes(ctx, ownerID, businessID, req.GetIncludeArchived())
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}

	out := make([]*flowtimev2.WorkType, 0, len(types))
	for _, workType := range types {
		out = append(out, toContractWorkTypeV2(workType))
	}
	return &flowtimev2.ListWorkTypesResponse{WorkTypes: out}, nil
}

func (a *WorkTimeAPIV2) CreateWorkType(ctx context.Context, req *flowtimev2.CreateWorkTypeRequest) (*flowtimev2.CreateWorkTypeResponse, error) {
	ownerID, businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	workType, err := a.Service.CreateWorkType(ctx, domain.WorkTypeInput{
		OwnerID:      ownerID,
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
	return &flowtimev2.CreateWorkTypeResponse{WorkType: toContractWorkTypeV2(workType)}, nil
}

func (a *WorkTimeAPIV2) UpdateWorkType(ctx context.Context, req *flowtimev2.UpdateWorkTypeRequest) (*flowtimev2.UpdateWorkTypeResponse, error) {
	ownerID, businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	workType, err := a.Service.UpdateWorkType(ctx, domain.WorkTypeInput{
		ID:           req.GetId(),
		OwnerID:      ownerID,
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
	return &flowtimev2.UpdateWorkTypeResponse{WorkType: toContractWorkTypeV2(workType)}, nil
}

func (a *WorkTimeAPIV2) DeleteWorkType(ctx context.Context, req *flowtimev2.DeleteWorkTypeRequest) (*flowtimev2.DeleteWorkTypeResponse, error) {
	ownerID, businessID, _, err := a.scope(ctx, 0)
	if err != nil {
		return nil, err
	}

	removed, err := a.Service.DeleteWorkType(ctx, req.GetId(), ownerID, businessID)
	if err != nil {
		return nil, mapServiceError(ctx, err)
	}
	if !removed {
		return nil, onekiterrors.New(http.StatusNotFound, "work type not found")
	}
	return &flowtimev2.DeleteWorkTypeResponse{Message: "deleted"}, nil
}
