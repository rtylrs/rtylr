package api

import (
	"context"
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
		req.GetFrom(), req.GetTo(), int64(limit), offset,
	)
	if err != nil {
		if err == domain.ErrUnauthorized {
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
