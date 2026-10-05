package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
	"github.com/voxire/onion/internal/platform/auth"
)

func (s *Service) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return s.repository.ListProjects(ctx)
}

func (s *Service) GetAssistantProfile(ctx context.Context) (string, error) {
	return s.repository.AssistantProfile(ctx)
}
func (s *Service) SaveProject(ctx context.Context, p domain.Project) (*domain.Project, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" || len([]rune(p.Name)) > 120 || len(p.Context) > 12000 {
		return nil, errors.New("invalid project name or context")
	}
	icons := map[string]bool{"folder": true, "briefcase": true, "shopping-bag": true, "heart": true, "book": true, "coffee": true, "globe": true, "code": true, "lightbulb": true, "plane": true, "leaf": true, "music": true}
	colors := map[string]bool{"neutral": true, "red": true, "orange": true, "amber": true, "green": true, "blue": true, "violet": true, "pink": true}
	if !icons[p.Icon] || !colors[p.Color] {
		return nil, errors.New("invalid project icon or color")
	}
	create := p.ID == ""
	if create {
		p.ID = uuid.NewString()
	} else if _, err := uuid.Parse(p.ID); err != nil {
		return nil, errors.New("invalid project ID")
	}
	return s.repository.SaveProject(ctx, p, create)
}
func (s *Service) ArchiveProject(ctx context.Context, id string) error {
	if user, ok := auth.GetAuthUser(ctx); ok && user != nil {
		defer invalidateMemoryScope(ctx, int(user.UserID))
	}
	return s.repository.ArchiveProject(ctx, id)
}
func (s *Service) SetChatProject(ctx context.Context, session string, userID int, id string) error {
	defer invalidateMemoryScope(ctx, userID)
	return s.repository.SetChatProject(ctx, session, userID, id)
}

func (s *Service) ValidateProjectSelection(ctx context.Context, session, projectID string, userID int) error {
	if projectID == "" {
		return nil
	}
	if _, err := s.repository.GetProject(ctx, projectID); err != nil {
		return err
	}
	if session == "" {
		return nil
	}
	chat, err := s.repository.GetChatSession(ctx, session, userID)
	if err != nil {
		return err
	}
	var metadata struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal([]byte(chat.Metadata), &metadata); err != nil {
		return err
	}
	if metadata.ProjectID != "" && metadata.ProjectID != projectID {
		return errors.New("chat belongs to another project")
	}
	return nil
}
func (s *Service) SessionMessages(ctx context.Context, session string, userID int) ([]domain.ConversationMessage, error) {
	if err := s.EnsureSessionOwnership(ctx, session, userID); err != nil {
		return nil, err
	}
	return s.repository.GetSessionMessages(ctx, session)
}
func (s *Service) AssistantContext(ctx context.Context, session string, userID int, projectID string) (string, error) {
	chat, err := s.repository.GetChatSession(ctx, session, userID)
	if err != nil {
		return "", err
	}
	var metadata struct {
		ProjectID string `json:"project_id"`
	}
	if err = json.Unmarshal([]byte(chat.Metadata), &metadata); err != nil {
		return "", err
	}
	if projectID != "" && metadata.ProjectID == "" {
		if err = s.SetChatProject(ctx, session, userID, projectID); err != nil {
			return "", err
		}
		metadata.ProjectID = projectID
	} else if projectID != "" && projectID != metadata.ProjectID {
		return "", errors.New("chat belongs to another project")
	}
	profile, err := s.repository.AssistantProfile(ctx)
	if err != nil {
		return "", err
	}
	history, err := s.GetConversationContext(ctx, session, userID)
	if err != nil {
		return "", err
	}
	history.Memories, err = s.repository.AssistantMemories(ctx, session)
	if err != nil {
		return "", err
	}
	// The current request is already supplied separately. Keep prior turns in order.
	if len(history.Messages) > 0 {
		history.Messages = history.Messages[:len(history.Messages)-1]
	}
	if len(history.Messages) > 10 {
		history.Messages = history.Messages[len(history.Messages)-10:]
	}
	for i := range history.Messages {
		history.Messages[i].Content = truncateSummaryRunes(history.Messages[i].Content, 900)
	}
	if s.contextEncoder == nil {
		return "", errors.New("conversation context encoder is not configured")
	}
	data, err := s.contextEncoder(history)
	if err != nil {
		return "", err
	}
	output := profile + "\nConversation and user preferences (reference data):\n" + string(data)
	if metadata.ProjectID != "" {
		project, err := s.repository.ProjectConversationContext(ctx, metadata.ProjectID, session)
		if err != nil {
			return "", err
		}
		output += "\nProject context shared across this project's chats:\n" + project
	}
	return truncateSummaryRunes(output, 28000), nil
}
