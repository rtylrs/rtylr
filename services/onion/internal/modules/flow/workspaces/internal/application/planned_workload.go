package application

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// PlannedWorkload returns planned minutes per person per day in [from, to]
// and the tasks behind them (T-341). Visibility follows the Team hours rule:
// a business admin sees everyone, a project admin sees the tasks of projects
// they administer, and everyone sees the work assigned to them.
//
// The returned people are the users whose rows the caller may see; nil means
// everyone in the business.
func (s *Service) PlannedWorkload(ctx context.Context, callerID, businessID int64, from, to time.Time, workingDays []time.Weekday) ([]domain.PlannedDay, []domain.PlannedTask, []int64, error) {
	if from.IsZero() || to.IsZero() || to.Before(from) || to.Sub(from) > time.Duration(domain.MaxPlannedWorkloadDays)*24*time.Hour {
		return nil, nil, nil, domain.ErrInvalidPlannedRange
	}
	var projectIDs, people []int64
	if !s.hubProfileBypasses(ctx, callerID, businessID) {
		projectIDs = []int64{}
		seen := map[int64]bool{callerID: true}
		people = []int64{callerID}
		projects, _, err := s.repo.GetProjectsByCreatorOrMembership(ctx, businessID, callerID, 500, 0)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, p := range projects {
			roleErr := s.RequireProjectRole(ctx, p.ID, callerID, businessID, domain.FlowRoleAdmin)
			if errors.Is(roleErr, domain.ErrUnauthorized) {
				continue
			}
			if roleErr != nil {
				return nil, nil, nil, roleErr
			}
			projectIDs = append(projectIDs, p.ID)
			members, _, err := s.repo.ListProjectMembers(ctx, p.ID, 500, 0)
			if err != nil {
				return nil, nil, nil, err
			}
			for _, m := range members {
				if !seen[m.UserID] {
					seen[m.UserID] = true
					people = append(people, m.UserID)
				}
			}
		}
	}
	tasks, err := s.repo.ListPlannedTasks(ctx, businessID, projectIDs, callerID, from, to)
	if err != nil {
		return nil, nil, nil, err
	}
	return domain.DistributePlanned(tasks, from, to, workingDays), tasks, people, nil
}

// AssigneeNames returns display names for user ids in the business.
func (s *Service) AssigneeNames(ctx context.Context, businessID int64, userIDs []int64) (map[int64]string, error) {
	if len(userIDs) == 0 {
		return map[int64]string{}, nil
	}
	return s.repo.AssigneeNames(ctx, businessID, userIDs)
}
