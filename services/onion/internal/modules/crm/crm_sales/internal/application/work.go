package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
)

// LogActivity records a touch on a lead or deal and, optionally, its
// follow-up task.
func (s *Service) LogActivity(ctx context.Context, businessID, actorID int64, in domain.ActivityInput) (domain.Activity, *domain.Task, error) {
	if err := in.Normalize(s.now()); err != nil {
		return domain.Activity{}, nil, err
	}
	if in.FollowUp != nil {
		if err := s.validateAssignee(ctx, businessID, in.FollowUp.AssignedTo); err != nil {
			return domain.Activity{}, nil, err
		}
	}
	return s.repo.LogActivity(ctx, businessID, actorID, in)
}

func (s *Service) ListLeadActivities(ctx context.Context, businessID, leadID int64, cursor string, limit int) (domain.ActivityPage, error) {
	if _, err := s.repo.GetLead(ctx, businessID, leadID); err != nil {
		return domain.ActivityPage{}, err
	}
	return s.repo.ListLeadActivities(ctx, businessID, leadID, cursor, pageSize(limit))
}

func (s *Service) ListDealActivities(ctx context.Context, businessID, dealID int64, cursor string, limit int) (domain.ActivityPage, error) {
	deal, err := s.repo.GetDeal(ctx, businessID, dealID)
	if err != nil {
		return domain.ActivityPage{}, err
	}
	return s.repo.ListDealActivities(ctx, businessID, deal, cursor, pageSize(limit))
}

func (s *Service) DeleteActivity(ctx context.Context, businessID, id int64) error {
	return s.repo.DeleteActivity(ctx, businessID, id)
}

func (s *Service) ListTasks(ctx context.Context, businessID int64, f domain.TaskFilter, day domain.CallerDay) (domain.TaskPage, error) {
	if f.Scope == "" {
		f.Scope = domain.TasksOpen
	}
	if f.Scope != domain.TasksOpen && f.Scope != domain.TasksCompleted {
		return domain.TaskPage{}, &domain.ValidationError{Message: "unknown task scope"}
	}
	if !f.Due.IsValid() {
		return domain.TaskPage{}, &domain.ValidationError{Message: "unknown due window"}
	}
	f.Limit = pageSize(f.Limit)
	return s.repo.ListTasks(ctx, businessID, f, day.Date)
}

func (s *Service) CountTasks(ctx context.Context, businessID int64, assignedTo *int64, day domain.CallerDay) (domain.TaskCounts, error) {
	return s.repo.CountTasks(ctx, businessID, assignedTo, day.Date)
}

// CreateTask schedules work on a lead, a deal or neither; unassigned tasks go
// to their creator.
func (s *Service) CreateTask(ctx context.Context, businessID, actorID int64, in domain.TaskInput) (domain.Task, error) {
	if err := in.Normalize(); err != nil {
		return domain.Task{}, err
	}
	if in.AssignedTo == nil {
		in.AssignedTo = &actorID
	}
	if err := s.validateAssignee(ctx, businessID, in.AssignedTo); err != nil {
		return domain.Task{}, err
	}
	return s.repo.CreateTask(ctx, businessID, actorID, in)
}

func (s *Service) UpdateTask(ctx context.Context, businessID, id int64, in domain.TaskInput) (domain.Task, error) {
	if err := in.Normalize(); err != nil {
		return domain.Task{}, err
	}
	if err := s.validateAssignee(ctx, businessID, in.AssignedTo); err != nil {
		return domain.Task{}, err
	}
	return s.repo.UpdateTask(ctx, businessID, id, in)
}

func (s *Service) CompleteTask(ctx context.Context, businessID, id int64, done bool) (domain.Task, error) {
	return s.repo.CompleteTask(ctx, businessID, id, done, s.now())
}

func (s *Service) DeleteTask(ctx context.Context, businessID, id int64) error {
	return s.repo.DeleteTask(ctx, businessID, id)
}

// ListPipelines returns the business's pipelines, creating the default one on
// first use.
func (s *Service) ListPipelines(ctx context.Context, businessID int64) ([]domain.Pipeline, error) {
	pipelines, err := s.repo.ListPipelines(ctx, businessID)
	if err != nil || len(pipelines) > 0 {
		return pipelines, err
	}
	if err := s.repo.EnsureDefaultPipeline(ctx, businessID); err != nil {
		return nil, err
	}
	return s.repo.ListPipelines(ctx, businessID)
}

func (s *Service) CreatePipeline(ctx context.Context, businessID int64, in domain.PipelineInput) (domain.Pipeline, error) {
	if err := in.Normalize(); err != nil {
		return domain.Pipeline{}, err
	}
	if _, err := s.ListPipelines(ctx, businessID); err != nil {
		return domain.Pipeline{}, err
	}
	id, err := s.repo.CreatePipeline(ctx, businessID, in)
	if err != nil {
		return domain.Pipeline{}, err
	}
	return s.repo.GetPipeline(ctx, businessID, id)
}

func (s *Service) UpdatePipeline(ctx context.Context, businessID, id int64, in domain.PipelineInput) (domain.Pipeline, error) {
	if err := in.Normalize(); err != nil {
		return domain.Pipeline{}, err
	}
	if err := s.repo.UpdatePipeline(ctx, businessID, id, in); err != nil {
		return domain.Pipeline{}, err
	}
	return s.repo.GetPipeline(ctx, businessID, id)
}

func (s *Service) DeletePipeline(ctx context.Context, businessID, id int64) error {
	return s.repo.DeletePipeline(ctx, businessID, id)
}
