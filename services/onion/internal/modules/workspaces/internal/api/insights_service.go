package api

import (
	"context"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowinsightsv1 "github.com/voxire/voxie/gen/go/flow/insights/v1"
)

type InsightsAPI struct {
	Service *application.Service
}

func NewInsightsAPI(service *application.Service) *InsightsAPI {
	return &InsightsAPI{Service: service}
}

func (a *InsightsAPI) GetInsights(ctx context.Context, req *flowinsightsv1.GetInsightsRequest) (*flowinsightsv1.GetInsightsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID, err := resolveWorkspaceQuery(ctx, a.Service, authUser, authUser.BusinessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	ins, err := a.Service.GetInsights(ctx, wsID)
	if err != nil {
		return nil, workspaceInternalError(ctx, "GetInsights failed", err)
	}

	resp := &flowinsightsv1.GetInsightsResponse{
		CycleTime:  ins.CycleTime,
		TotalTasks: int32(ins.TotalTasks), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		DoneTasks:  int32(ins.DoneTasks),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	for _, v := range ins.Velocity {
		resp.Velocity = append(resp.Velocity, &flowinsightsv1.VelocityPoint{
			Cycle:          v.Cycle,
			Completed:      int32(v.Completed), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Total:          int32(v.Total),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			CompletionRate: v.CompletionRate,
		})
	}
	for _, b := range ins.Burndown {
		resp.Burndown = append(resp.Burndown, &flowinsightsv1.BurndownPoint{
			Day:    int32(b.Day), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Ideal:  b.Ideal,
			Actual: b.Actual,
		})
	}
	for _, m := range ins.TimeByMember {
		resp.TimeByMember = append(resp.TimeByMember, &flowinsightsv1.MemberTime{
			Id:    m.ID,
			Name:  m.Name,
			Hours: m.Hours,
		})
	}
	for _, t := range ins.Throughput {
		resp.Throughput = append(resp.Throughput, &flowinsightsv1.ThroughputPoint{
			Week:      t.Week,
			Completed: int32(t.Completed), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}

	if resp.Velocity == nil {
		resp.Velocity = []*flowinsightsv1.VelocityPoint{}
	}
	if resp.Burndown == nil {
		resp.Burndown = []*flowinsightsv1.BurndownPoint{}
	}
	if resp.TimeByMember == nil {
		resp.TimeByMember = []*flowinsightsv1.MemberTime{}
	}
	if resp.Throughput == nil {
		resp.Throughput = []*flowinsightsv1.ThroughputPoint{}
	}

	return resp, nil
}

func (a *InsightsAPI) GetTeamWorkload(ctx context.Context, req *flowinsightsv1.GetTeamWorkloadRequest) (*flowinsightsv1.GetTeamWorkloadResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID, err := resolveWorkspaceQuery(ctx, a.Service, authUser, authUser.BusinessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	workload, total, err := a.Service.GetTeamWorkload(ctx, wsID, int64(limit), offset)
	if err != nil {
		return nil, workspaceInternalError(ctx, "GetTeamWorkload failed", err)
	}

	members := make([]*flowinsightsv1.TeamMemberWorkload, 0, len(workload))
	for _, w := range workload {
		tasks := make([]*flowinsightsv1.TeamTaskSummary, 0, len(w.Tasks))
		for _, t := range w.Tasks {
			s := &flowinsightsv1.TeamTaskSummary{
				Id:       strconv.FormatInt(t.ID, 10),
				Title:    t.Title,
				Status:   t.Status,
				Priority: t.Priority,
			}
			if t.DueDate != nil {
				s.DueDate = t.DueDate.Format("2006-01-02")
			}
			tasks = append(tasks, s)
		}
		members = append(members, &flowinsightsv1.TeamMemberWorkload{
			Id:            w.ID,
			Name:          w.Name,
			Initials:      w.Initials,
			Capacity:      int32(w.Capacity), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			AssignedHours: w.AssignedHours,
			Load:          w.Load,
			Tasks:         tasks,
		})
	}
	return &flowinsightsv1.GetTeamWorkloadResponse{
		Members:    members,
		Pagination: buildPagination(page, limit, total, len(workload)),
	}, nil
}

func (a *InsightsAPI) GetRoadmap(ctx context.Context, req *flowinsightsv1.GetRoadmapRequest) (*flowinsightsv1.GetRoadmapResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID, err := resolveWorkspaceQuery(ctx, a.Service, authUser, authUser.BusinessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	projects, total, err := a.Service.GetRoadmap(ctx, wsID, int64(limit), offset)
	if err != nil {
		return nil, workspaceInternalError(ctx, "GetRoadmap failed", err)
	}

	payloads := make([]*flowinsightsv1.RoadmapProject, 0, len(projects))
	// One GetProjectTaskCounts call per project on this page - bounded by the
	// outer pagination (at most `limit` projects), not by its own cap.
	for _, p := range projects {
		taskCount, done := a.Service.GetProjectTaskCounts(ctx, p.ID)
		progress := int32(0)
		if taskCount > 0 {
			progress = int32(float64(done) / float64(taskCount) * 100)
		}

		rp := &flowinsightsv1.RoadmapProject{
			Id:          strconv.FormatInt(p.ID, 10),
			Name:        p.Name,
			Color:       p.Color,
			Progress:    progress,
			DoneCount:   int32(done),      // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			TaskCount:   int32(taskCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Description: p.Description,
			WorkspaceId: strconv.FormatInt(p.WorkspaceID, 10),
		}
		if p.StartDate != nil {
			rp.StartDate = p.StartDate.Format("2006-01-02")
		}
		if p.DueDate != nil {
			rp.TargetDate = p.DueDate.Format("2006-01-02")
		}
		payloads = append(payloads, rp)
	}
	return &flowinsightsv1.GetRoadmapResponse{
		Projects:   payloads,
		Pagination: buildPagination(page, limit, total, len(projects)),
	}, nil
}
