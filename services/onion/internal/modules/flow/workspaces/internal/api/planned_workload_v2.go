package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtimev2 "github.com/voxire/voxie/gen/go/flow/time/v2"
)

// parseWorkingDays reads "1,2,3,4,5" (ISO weekdays, 7 = Sunday). Empty means
// the default Monday to Friday.
func parseWorkingDays(value string) ([]time.Weekday, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var days []time.Weekday
	for _, part := range strings.Split(value, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > 7 {
			return nil, errors.New("working_days must be ISO weekdays 1 to 7")
		}
		days = append(days, time.Weekday(n%7))
	}
	return days, nil
}

func plannedDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func (a *WorkTimeAPIV2) ListPlannedWorkload(ctx context.Context, req *flowtimev2.ListPlannedWorkloadRequest) (*flowtimev2.ListPlannedWorkloadResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	from, fromErr := time.Parse("2006-01-02", req.GetFromDate())
	to, toErr := time.Parse("2006-01-02", req.GetToDate())
	if fromErr != nil || toErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, domain.ErrInvalidPlannedRange.Error())
	}
	workingDays, err := parseWorkingDays(req.GetWorkingDays())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "working_days must be ISO weekdays 1 to 7")
	}
	days, tasks, people, err := a.Service.PlannedWorkload(ctx, caller.UserID, businessID, from, to, workingDays)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPlannedRange) {
			return nil, onekiterrors.New(http.StatusBadRequest, domain.ErrInvalidPlannedRange.Error())
		}
		return nil, workspaceInternalError(ctx, "ListPlannedWorkload failed", err)
	}
	outDays := make([]*flowtimev2.PlannedWorkloadDay, 0, len(days))
	used := map[int64]bool{}
	for _, d := range days {
		outDays = append(outDays, &flowtimev2.PlannedWorkloadDay{
			UserId:         d.UserID,
			Date:           d.Date.Format("2006-01-02"),
			PlannedMinutes: clampInt32(d.Minutes),
			TaskIds:        d.TaskIDs,
		})
		for _, id := range d.TaskIDs {
			used[id] = true
		}
	}
	assigneeIDs := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		if used[t.ID] {
			assigneeIDs = append(assigneeIDs, t.AssigneeID)
		}
	}
	names, err := a.Service.AssigneeNames(ctx, businessID, assigneeIDs)
	if err != nil {
		return nil, workspaceInternalError(ctx, "ListPlannedWorkload names failed", err)
	}
	outTasks := make([]*flowtimev2.PlannedWorkloadTask, 0, len(used))
	for _, t := range tasks {
		if !used[t.ID] {
			continue
		}
		outTasks = append(outTasks, &flowtimev2.PlannedWorkloadTask{
			Id:              t.ID,
			Identifier:      "T-" + strconv.FormatInt(t.ID, 10),
			Title:           t.Title,
			ProjectId:       t.ProjectID,
			AssigneeId:      t.AssigneeID,
			AssigneeName:    names[t.AssigneeID],
			StartDate:       plannedDate(t.StartDate),
			DueDate:         plannedDate(t.DueDate),
			EstimateMinutes: clampInt32(t.EstimateMinutes),
		})
	}
	if people == nil {
		people = []int64{}
	}
	return &flowtimev2.ListPlannedWorkloadResponse{
		Days: outDays, Tasks: outTasks, Everyone: len(people) == 0, People: people,
	}, nil
}
