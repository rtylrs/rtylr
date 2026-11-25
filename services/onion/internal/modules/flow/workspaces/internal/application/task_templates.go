package application

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// ListTaskTemplates returns the project's templates. Any project viewer may
// read them, since the New task form uses them.
func (s *Service) ListTaskTemplates(ctx context.Context, businessID, projectID, callerID int64) ([]domain.TaskTemplate, error) {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, err
	}
	return s.repo.ListTaskTemplates(ctx, businessID, projectID)
}

// SetTaskTemplate saves or, for a blank body, removes the project's template
// for a task type. Project admins only, like the project's other settings.
func (s *Service) SetTaskTemplate(ctx context.Context, businessID, projectID, callerID int64, taskType, body string) (*domain.TaskTemplate, error) {
	if !domain.IsValidTaskType(taskType) {
		return nil, domain.ErrInvalidTaskType
	}
	if len(body) > domain.MaxTaskTemplateLength {
		return nil, domain.ErrTaskTemplateTooLong
	}
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, err
	}
	if strings.TrimSpace(body) == "" {
		return nil, s.repo.DeleteTaskTemplate(ctx, businessID, projectID, taskType)
	}
	if err := s.repo.UpsertTaskTemplate(ctx, businessID, projectID, taskType, body, callerID); err != nil {
		return nil, err
	}
	templates, err := s.repo.ListTaskTemplates(ctx, businessID, projectID)
	if err != nil {
		return nil, err
	}
	for i := range templates {
		if templates[i].TaskType == taskType {
			return &templates[i], nil
		}
	}
	return nil, nil
}
