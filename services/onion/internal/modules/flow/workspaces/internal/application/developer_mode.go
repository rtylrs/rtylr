package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrPersonalProjectDeveloperMode rejects switching the My Tasks bucket, which
// is never a software project.
var ErrPersonalProjectDeveloperMode = errors.New("the My Tasks bucket cannot be a developer project")

type auditRecorder interface {
	RecordAudit(ctx context.Context, event audit.Event) error
}

// WithAuditRecorder wires the activity-log recorder.
func (s *Service) WithAuditRecorder(recorder auditRecorder) *Service {
	s.audit = recorder
	return s
}

// SetProjectDeveloperMode turns developer features on or off. Turning them off
// only hides them: nothing stored for the project is deleted.
func (s *Service) SetProjectDeveloperMode(ctx context.Context, projectID, callerID, businessID int64, enabled bool) (domain.Project, error) {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.Project{}, err
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.IsPersonal {
		return domain.Project{}, ErrPersonalProjectDeveloperMode
	}
	if p.DeveloperMode == enabled {
		return p, nil
	}
	if err := s.repo.UpdateProjectDeveloperMode(ctx, projectID, businessID, enabled); err != nil {
		return domain.Project{}, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)

	action, description := "developer_mode_disabled", "Turned off developer features for "+p.Name
	if enabled {
		action, description = "developer_mode_enabled", "Turned on developer features for "+p.Name
	}
	s.recordAudit(ctx, audit.Event{
		BusinessID: businessID, ActorID: callerID, EntityType: "flow_project", EntityID: projectID,
		Action: action, Description: description,
		OldValues: map[string]bool{"developer_mode": p.DeveloperMode},
		NewValues: map[string]bool{"developer_mode": enabled},
	})
	return s.repo.GetProjectV2ByID(ctx, projectID, businessID)
}

// recordAudit logs a failure instead of returning it: the change has committed.
func (s *Service) recordAudit(ctx context.Context, event audit.Event) {
	if s.audit == nil || event.BusinessID <= 0 || event.EntityID <= 0 {
		return
	}
	if err := s.audit.RecordAudit(ctx, event); err != nil {
		logger.Ctx(ctx).Error("failed to record flow audit entry",
			logger.String("action", event.Action), logger.Int("business_id", int(event.BusinessID)), logger.Error(err))
	}
}
