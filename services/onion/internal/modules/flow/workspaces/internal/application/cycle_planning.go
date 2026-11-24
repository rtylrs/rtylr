package application

import (
	"context"
	"sort"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// CycleMinutesPerDay is the working day length capacity is measured in. Flow
// has no per-business working hours setting yet, so it is the common 8 hours.
const CycleMinutesPerDay = 8 * 60

// CycleLeaveSource reads approved leave for the capacity view. HR owns leave;
// Flow only asks which of a cycle's dates each user is away on, through this
// port, and never reads HR tables itself.
type CycleLeaveSource interface {
	// ApprovedLeaveDates returns, per user id, the YYYY-MM-DD dates in
	// [from, to] covered by an approved leave request.
	ApprovedLeaveDates(ctx context.Context, businessID int64, userIDs []int64, from, to time.Time) (map[int64]map[string]bool, error)
}

// SetCycleLeaveSource wires the HR leave reader. Without one, capacity counts
// working days only and says so (leave_applied false).
func (s *Service) SetCycleLeaveSource(source CycleLeaveSource) {
	if source != nil {
		s.cycleLeave = source
	}
}

// CycleProjectID returns the project of a cycle in the business, or
// domain.ErrCycleNotFound.
func (s *Service) CycleProjectID(ctx context.Context, businessID, cycleID int64) (int64, error) {
	c, err := s.repo.GetCycleV2(ctx, businessID, cycleID)
	if err != nil {
		return 0, err
	}
	if c.ProjectID == nil {
		return 0, domain.ErrCycleNotFound
	}
	return *c.ProjectID, nil
}

func (s *Service) GetCycleV2(ctx context.Context, businessID, cycleID int64) (domain.Cycle, error) {
	return s.repo.GetCycleV2(ctx, businessID, cycleID)
}

func (s *Service) UpdateCycleV2(ctx context.Context, in domain.UpdateCycleV2Input) error {
	return s.repo.UpdateCycleV2(ctx, in)
}

// DeleteCycle removes a cycle. Its tasks return to the project's backlog.
func (s *Service) DeleteCycle(ctx context.Context, businessID, cycleID int64) error {
	return s.repo.DeleteCycle(ctx, businessID, cycleID)
}

// StartCycle starts an upcoming cycle and returns it as stored.
func (s *Service) StartCycle(ctx context.Context, in domain.StartCycleInput) (domain.Cycle, error) {
	if in.Today.IsZero() {
		in.Today = time.Now().UTC().Truncate(24 * time.Hour)
	}
	if err := s.repo.StartCycle(ctx, in); err != nil {
		return domain.Cycle{}, err
	}
	return s.repo.GetCycleV2(ctx, in.BusinessID, in.ID)
}

// CompleteCycle completes a cycle, settles its unfinished tasks and returns the
// completed cycle.
func (s *Service) CompleteCycle(ctx context.Context, in domain.CompleteCycleInput) (domain.Cycle, error) {
	if len(in.Decisions) > domain.MaxCycleTaskBatch {
		return domain.Cycle{}, domain.ErrInvalidCycleRequest
	}
	if _, err := s.repo.CompleteCycle(ctx, in); err != nil {
		return domain.Cycle{}, err
	}
	return s.repo.GetCycleV2(ctx, in.BusinessID, in.ID)
}

// MoveCycleTask places a task in a cycle or the backlog and returns where it
// landed.
func (s *Service) MoveCycleTask(ctx context.Context, in domain.MoveCycleTaskInput) (domain.MovedCycleTask, error) {
	return s.repo.MoveCycleTask(ctx, in)
}

// AttachTaskCycles fills CycleID, CycleName and PlanRank on tasks read from
// flow_tasks, which has no cycle column.
func (s *Service) AttachTaskCycles(ctx context.Context, tasks []domain.Task) error {
	return s.repo.AttachTaskCycles(ctx, tasks)
}

// AttachTaskMemberships fills what a task read from flow_tasks carries in
// side tables or columns the generated reads skip: its milestone, its story
// points and its cycle with plan rank.
func (s *Service) AttachTaskMemberships(ctx context.Context, tasks []domain.Task) error {
	if err := s.AttachMilestoneIDs(ctx, tasks); err != nil {
		return err
	}
	if err := s.attachTaskChecksStates(ctx, tasks); err != nil {
		return err
	}
	if err := s.repo.AttachTaskStoryPoints(ctx, tasks); err != nil {
		return err
	}
	return s.AttachTaskCycles(ctx, tasks)
}

// SetTaskCycle moves a task to the end of cycleID, or to the backlog when
// cycleID is zero. It is the task panel's write; the Backlog screen uses
// MoveCycleTask with a position.
func (s *Service) SetTaskCycle(ctx context.Context, businessID, taskID, cycleID int64) error {
	_, err := s.repo.MoveCycleTask(ctx, domain.MoveCycleTaskInput{
		BusinessID: businessID, TaskID: taskID, CycleID: cycleID,
	})
	return err
}

// GetCycleCapacity works out each person's available and committed time for a
// cycle. People are the project's members plus anyone assigned a cycle task.
// Available time is working days minus approved leave days, read through the
// HR leave port; when that read is not possible the figure counts working days
// only and LeaveApplied is false.
func (s *Service) GetCycleCapacity(ctx context.Context, businessID, cycleID int64) (domain.CycleCapacity, error) {
	cycle, err := s.repo.GetCycleV2(ctx, businessID, cycleID)
	if err != nil {
		return domain.CycleCapacity{}, err
	}
	loads, err := s.repo.CycleMemberLoads(ctx, businessID, cycleID)
	if err != nil {
		return domain.CycleCapacity{}, err
	}
	out := domain.CycleCapacity{MinutesPerDay: CycleMinutesPerDay}

	names := map[int64]string{}
	var order []int64
	seen := map[int64]bool{}
	add := func(userID int64, name string) {
		if userID <= 0 || seen[userID] {
			return
		}
		seen[userID] = true
		order = append(order, userID)
		if name != "" {
			names[userID] = name
		}
	}
	if cycle.ProjectID != nil {
		members, _, memberErr := s.repo.ListProjectMembers(ctx, *cycle.ProjectID, 500, 0)
		if memberErr != nil {
			return domain.CycleCapacity{}, memberErr
		}
		for _, m := range members {
			add(m.UserID, m.Name)
		}
	}
	byUser := map[int64]domain.CycleMemberLoad{}
	for _, load := range loads {
		if load.UserID == 0 {
			out.UnassignedMinutes = int(load.Minutes)
			out.UnassignedCount = load.TaskCount
			continue
		}
		byUser[load.UserID] = load
		add(load.UserID, "")
	}
	var missing []int64
	for _, id := range order {
		if names[id] == "" {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		resolved, nameErr := s.repo.AssigneeNames(ctx, businessID, missing)
		if nameErr == nil {
			for id, name := range resolved {
				names[id] = name
			}
		}
	}

	var workingDays []time.Time
	if !cycle.StartDate.IsZero() && !cycle.EndDate.IsZero() {
		workingDays = domain.WorkingDaysBetween(cycle.StartDate, cycle.EndDate)
	}
	var leave map[int64]map[string]bool
	if s.cycleLeave != nil && len(order) > 0 && len(workingDays) > 0 {
		dates, leaveErr := s.cycleLeave.ApprovedLeaveDates(ctx, businessID, order, cycle.StartDate, cycle.EndDate)
		if leaveErr == nil {
			leave = dates
			out.LeaveApplied = true
		}
	}

	for _, id := range order {
		leaveDays := 0
		for _, day := range workingDays {
			if leave[id][day.Format("2006-01-02")] {
				leaveDays++
			}
		}
		load := byUser[id]
		out.Members = append(out.Members, domain.CycleMemberCapacity{
			UserID:           id,
			Name:             names[id],
			WorkingDays:      len(workingDays),
			LeaveDays:        leaveDays,
			AvailableMinutes: (len(workingDays) - leaveDays) * CycleMinutesPerDay,
			CommittedMinutes: int(load.Minutes),
			TaskCount:        load.TaskCount,
		})
	}
	sort.SliceStable(out.Members, func(i, j int) bool {
		return out.Members[i].CommittedMinutes > out.Members[j].CommittedMinutes
	})
	return out, nil
}

// AttachTaskDependencies fills Dependencies (the ids of the tasks blocking
// each task) on a page of tasks, so a list can draw dependency arrows (T-342).
func (s *Service) AttachTaskDependencies(ctx context.Context, businessID int64, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]int64, len(tasks))
	for i := range tasks {
		ids[i] = tasks[i].ID
	}
	byTask, err := s.repo.ListTaskDependencies(ctx, ids, businessID)
	if err != nil {
		return err
	}
	for i := range tasks {
		tasks[i].Dependencies = byTask[tasks[i].ID]
	}
	return nil
}
