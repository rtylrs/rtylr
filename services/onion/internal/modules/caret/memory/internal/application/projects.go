package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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
	messages, err := s.repository.GetSessionMessages(ctx, session)
	if err != nil {
		return "", err
	}
	memories, err := s.repository.AssistantMemories(ctx, session)
	if err != nil {
		return "", err
	}

	// Each section has its own budget, so a long chat can never push the
	// profile, preferences or the newest turns out of the prompt.
	var output strings.Builder
	output.WriteString(clipEnd(strings.TrimSpace(profile), profileBudgetBytes))
	if saved := renderMemories(memories); saved != "" {
		output.WriteString("\n\nSaved user preferences:\n" + saved)
	}
	compaction := readCompaction(chat)
	history := orderedHistory(messages)
	transcript, omitted := transcriptWindow(uncoveredMessages(history, compaction.ThroughID), transcriptBudgetBytes)
	if compaction.Summary != "" {
		output.WriteString("\n\nSummary of the earlier part of this chat (older messages are not shown):\n" + clipEnd(compaction.Summary, summaryBudgetBytes))
	}
	if transcript != "" {
		heading := "\n\nConversation so far, oldest first:\n"
		if omitted > 0 {
			heading = fmt.Sprintf("\n\nConversation so far, oldest first (%d older message(s) are not shown yet):\n", omitted)
		}
		output.WriteString(heading + transcript)
	}
	if metadata.ProjectID != "" {
		project, err := s.repository.ProjectConversationContext(ctx, metadata.ProjectID, session)
		if err != nil {
			return "", err
		}
		if project = strings.TrimSpace(project); project != "" {
			output.WriteString("\n\nProject context shared across this project's chats:\n" + clipEnd(project, projectBudgetBytes))
		}
	}
	return output.String(), nil
}

// CompactionWork is the next batch of history to fold into a chat's summary.
type CompactionWork struct {
	PreviousSummary string
	PreviousThrough int
	Transcript      string
	NewThrough      int
}

// PendingCompaction reports whether a chat's unsummarized history has
// outgrown the context window, and if so which turns to summarize next.
func (s *Service) PendingCompaction(ctx context.Context, session string, userID int) (CompactionWork, bool, error) {
	chat, err := s.repository.GetChatSession(ctx, session, userID)
	if err != nil {
		return CompactionWork{}, false, err
	}
	messages, err := s.repository.GetSessionMessages(ctx, session)
	if err != nil {
		return CompactionWork{}, false, err
	}
	compaction := readCompaction(chat)
	ordered := append([]domain.ConversationMessage(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	batch := compactionSplit(uncoveredMessages(ordered, compaction.ThroughID))
	if len(batch) == 0 {
		return CompactionWork{}, false, nil
	}
	turns := make([]string, len(batch))
	for i, message := range batch {
		turns[i] = renderTurn(message)
	}
	return CompactionWork{
		PreviousSummary: compaction.Summary,
		PreviousThrough: compaction.ThroughID,
		Transcript:      strings.Join(turns, "\n\n"),
		NewThrough:      batch[len(batch)-1].ID,
	}, true, nil
}

// SaveCompaction stores a new rolling summary. It is a no-op when another
// compaction already moved the summary on, so concurrent replies cannot
// overwrite a newer summary with an older one.
func (s *Service) SaveCompaction(ctx context.Context, session string, userID int, work CompactionWork, summary string) error {
	summary = strings.TrimSpace(summary)
	if summary == "" || work.NewThrough <= work.PreviousThrough {
		return errors.New("compaction summary is empty")
	}
	chat, err := s.repository.GetChatSession(ctx, session, userID)
	if err != nil {
		return err
	}
	if readCompaction(chat).ThroughID != work.PreviousThrough {
		return nil
	}
	metadata := map[string]any{}
	if strings.TrimSpace(chat.Metadata) != "" && chat.Metadata != "null" {
		if err := json.Unmarshal([]byte(chat.Metadata), &metadata); err != nil {
			return err
		}
	}
	metadata[compactedThroughKey] = work.NewThrough
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = s.repository.UpdateChatSession(ctx, session, userID, chat.Title, clipEnd(summary, summaryBudgetBytes), string(encoded))
	return err
}
