package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// LinkPreviewer reads a link's title (and PR/issue state) once, when added.
type LinkPreviewer interface {
	Preview(ctx context.Context, link domain.ClassifiedLink) (title, state string)
}

// WithLinkPreviewer wires title fetching. Without it links keep their URL label.
func (s *Service) WithLinkPreviewer(p LinkPreviewer) *Service {
	s.linkPreviewer = p
	return s
}

// CreateTaskLink adds a URL to a task, or returns the link the task already
// has for that URL. The caller must already hold member access to the task.
func (s *Service) CreateTaskLink(ctx context.Context, taskID, businessID, userID int64, rawURL string) (domain.TaskLink, error) {
	classified, err := domain.ClassifyLink(rawURL)
	if err != nil {
		return domain.TaskLink{}, err
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil || task.BusinessID != businessID {
		return domain.TaskLink{}, domain.ErrUnauthorized
	}
	if existing, found, err := s.repo.FindTaskLinkByURL(ctx, taskID, businessID, classified.URL); err != nil {
		return domain.TaskLink{}, err
	} else if found {
		return existing, nil
	}

	link := domain.TaskLink{
		TaskID: taskID, BusinessID: businessID, URL: classified.URL,
		Kind: classified.Kind, Ref: classified.Ref, AddedBy: userID,
	}
	if s.linkPreviewer != nil {
		link.Title, link.State = s.linkPreviewer.Preview(ctx, classified)
	}
	id, err := s.repo.CreateTaskLink(ctx, link)
	if err != nil {
		return domain.TaskLink{}, err
	}
	s.logLinkActivity(ctx, taskID, userID, "link_added", "", linkLabel(link))
	return s.repo.GetTaskLink(ctx, id, businessID)
}

func (s *Service) ListTaskLinks(ctx context.Context, taskID, businessID int64) ([]domain.TaskLink, error) {
	return s.repo.ListTaskLinks(ctx, taskID, businessID)
}

// TaskLinkTaskID resolves the task a link belongs to, for authorization.
func (s *Service) TaskLinkTaskID(ctx context.Context, linkID, businessID int64) (int64, error) {
	link, err := s.repo.GetTaskLink(ctx, linkID, businessID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}
	return link.TaskID, nil
}

func (s *Service) DeleteTaskLink(ctx context.Context, linkID, businessID, userID int64) error {
	link, err := s.repo.GetTaskLink(ctx, linkID, businessID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteTaskLink(ctx, linkID, businessID); err != nil {
		return err
	}
	s.logLinkActivity(ctx, link.TaskID, userID, "link_removed", linkLabel(link), "")
	return nil
}

func (s *Service) logLinkActivity(ctx context.Context, taskID, userID int64, action, from, to string) {
	if err := s.repo.LogActivity(ctx, taskID, userID, action, from, to); err != nil {
		logger.Ctx(ctx).Error("task link: failed to log activity", logger.String("action", action), logger.Error(err))
	}
}

func linkLabel(l domain.TaskLink) string {
	if l.Title != "" {
		return l.Title
	}
	if l.Ref != "" && l.Kind != domain.LinkKindGeneric {
		return l.Ref
	}
	return l.URL
}
