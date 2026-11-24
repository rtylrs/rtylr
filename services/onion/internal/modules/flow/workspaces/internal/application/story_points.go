package application

import (
	"context"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// SetProjectPointsScale picks the scale the project's story points use. Points
// already on tasks are kept; a value off the new scale stays until edited.
func (s *Service) SetProjectPointsScale(ctx context.Context, projectID, callerID, businessID int64, scale string) (domain.Project, error) {
	if !domain.IsValidPointsScale(scale) {
		return domain.Project{}, domain.ErrInvalidPointsScale
	}
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.Project{}, err
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.PointsScale == scale {
		return p, nil
	}
	if err := s.repo.UpdateProjectPointsScale(ctx, projectID, businessID, scale); err != nil {
		return domain.Project{}, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	s.recordAudit(ctx, audit.Event{
		BusinessID: businessID, ActorID: callerID, EntityType: "flow_project", EntityID: projectID,
		Action: "points_scale_changed", Description: "Changed the story points scale of " + p.Name,
		OldValues: map[string]string{"points_scale": p.PointsScale},
		NewValues: map[string]string{"points_scale": scale},
	})
	return s.repo.GetProjectV2ByID(ctx, projectID, businessID)
}

// ValidateStoryPoints checks points against the project's scale. Zero (a
// clear) is always valid; a task outside a project uses Fibonacci.
func (s *Service) ValidateStoryPoints(ctx context.Context, businessID, projectID int64, points int32) error {
	if points == 0 {
		return nil
	}
	if points < 0 {
		return domain.ErrInvalidStoryPoints
	}
	scale := domain.PointsScaleFibonacci
	if projectID > 0 {
		p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
		if err != nil {
			return err
		}
		scale = p.PointsScale
	}
	if !domain.IsValidStoryPoints(scale, points) {
		return domain.ErrInvalidStoryPoints
	}
	return nil
}

// SetTaskStoryPoints sets a task's points, checked against its project's
// scale. Zero clears them.
func (s *Service) SetTaskStoryPoints(ctx context.Context, businessID, taskID, projectID int64, points int32) error {
	if err := s.ValidateStoryPoints(ctx, businessID, projectID, points); err != nil {
		return err
	}
	if points == 0 {
		return s.repo.SetTaskStoryPoints(ctx, businessID, taskID, nil)
	}
	return s.repo.SetTaskStoryPoints(ctx, businessID, taskID, &points)
}
