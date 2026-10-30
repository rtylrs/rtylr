package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/api"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/application"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/snips"
	hubmemoryv1 "github.com/voxire/voxie/gen/go/hub/memory/v1"
)

type Module struct {
	memoryService *application.Service
	repository    *infrastructure.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repository := infrastructure.NewRepository(db)
	memoryService := application.NewService(repository)

	return &Module{
		memoryService: memoryService,
		repository:    repository,
	}
}

func (m *Module) WithIdentityReaders(defaultBusiness func(context.Context, int64) (int64, error), profile func(context.Context, int64, int64) (string, string, string, string, error)) *Module {
	m.repository.WithIdentityReaders(defaultBusiness, profile)
	return m
}

func (m *Module) DeleteSession(ctx context.Context, session string, userID int) error {
	return m.memoryService.DeleteSession(ctx, session, userID)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubmemoryv1.RegisterMemoryServiceServer(
		api.NewMemoryAPI(m.memoryService),
		hubmemoryv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub memory service: %w", err))
	}
}

func (m *Module) CreateSession(ctx context.Context, userID int, title string) (string, error) {
	session, err := m.memoryService.CreateSession(ctx, userID, title)
	if err != nil {
		return "", err
	}
	return session.SessionID, nil
}

// ThreadMessage is one message of a chat, as the buddy reads it.
type ThreadMessage struct {
	Role, Content string
	CreatedAt     time.Time
}

// BuddyThread returns the person's buddy chat, creating it on first use.
func (m *Module) BuddyThread(ctx context.Context, businessID int64, userID int) (string, error) {
	return m.memoryService.EnsureBuddyThread(ctx, businessID, userID, "Caret Buddy")
}

// BuddyLastPost is when the buddy last wrote in the person's buddy chat.
func (m *Module) BuddyLastPost(ctx context.Context, sessionID string, userID int) (time.Time, bool, error) {
	return m.memoryService.BuddyLastPost(ctx, sessionID, userID)
}

// RecentMessages returns the newest messages of a chat the person owns.
func (m *Module) RecentMessages(ctx context.Context, sessionID string, userID, limit int) ([]ThreadMessage, error) {
	messages, err := m.memoryService.RecentMessages(ctx, sessionID, userID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]ThreadMessage, 0, len(messages))
	for _, message := range messages {
		out = append(out, ThreadMessage{Role: message.Role, Content: message.Content, CreatedAt: message.CreatedAt})
	}
	return out, nil
}

func (m *Module) StoreMessage(ctx context.Context, sessionID string, userID int, role, content string, tokens int) (any, error) {
	return m.memoryService.StoreMessage(ctx, sessionID, userID, role, content, tokens)
}

// EnsureSessionOwnership is used by backend integrations that accept a
// client-provided session ID. It keeps ownership verification in the memory
// module instead of making every caller duplicate repository rules.
func (m *Module) EnsureSessionOwnership(ctx context.Context, sessionID string, userID int) error {
	return m.memoryService.EnsureSessionOwnership(ctx, sessionID, userID)
}

func (m *Module) SummarizeSession(ctx context.Context, sessionID string, userID int) error {
	return m.memoryService.SummarizeSession(ctx, sessionID, userID)
}

func (m *Module) AssistantContext(ctx context.Context, session string, userID int, projectID string) (string, error) {
	return m.memoryService.AssistantContext(ctx, session, userID, projectID)
}

func (m *Module) GetAssistantProfile(ctx context.Context) (string, error) {
	return m.memoryService.GetAssistantProfile(ctx)
}

func (m *Module) ValidateProjectSelection(ctx context.Context, session, project string, userID int) error {
	return m.memoryService.ValidateProjectSelection(ctx, session, project, userID)
}
func (m *Module) StoreMessageWithImages(ctx context.Context, session string, userID int, role, content string, tokens int, images []snips.ImageInput) (any, error) {
	return m.memoryService.StoreMessageWithImages(ctx, session, userID, role, content, tokens, images)
}

func (m *Module) StoreMessageWithTranscript(ctx context.Context, session string, userID int, role, content string, tokens int, blocks json.RawMessage, responseMs int64) (any, error) {
	return m.memoryService.StoreMessageWithTranscript(ctx, session, userID, role, content, tokens, blocks, responseMs)
}

// PendingCompaction returns the chat history that has outgrown the context
// window: the current summary, the turns to fold into it, and the message
// range they cover. ok is false when nothing needs compacting.
func (m *Module) PendingCompaction(ctx context.Context, session string, userID int) (previousSummary, transcript string, previousThrough, newThrough int, ok bool, err error) {
	work, ok, err := m.memoryService.PendingCompaction(ctx, session, userID)
	return work.PreviousSummary, work.Transcript, work.PreviousThrough, work.NewThrough, ok, err
}

// SaveCompaction stores the summary covering messages through newThrough.
func (m *Module) SaveCompaction(ctx context.Context, session string, userID, previousThrough, newThrough int, summary string) error {
	return m.memoryService.SaveCompaction(ctx, session, userID, application.CompactionWork{PreviousThrough: previousThrough, NewThrough: newThrough}, summary)
}

// PendingChatTitle returns the first exchange of a chat that should get a
// written title. ok is false when it has one or the user renamed it.
func (m *Module) PendingChatTitle(ctx context.Context, session string, userID int) (currentTitle, firstPrompt, firstReply string, ok bool, err error) {
	work, ok, err := m.memoryService.PendingChatTitle(ctx, session, userID)
	return work.CurrentTitle, work.FirstPrompt, work.FirstReply, ok, err
}

// SaveChatTitle stores a written title unless the chat was renamed meanwhile.
func (m *Module) SaveChatTitle(ctx context.Context, session string, userID int, expectedTitle, title string) error {
	return m.memoryService.SaveChatTitle(ctx, session, userID, expectedTitle, title)
}

// RecordReplyFeedback stores a verdict on a reply in one of the user's chats
// and returns how many verdicts are waiting to be learned from.
func (m *Module) RecordReplyFeedback(ctx context.Context, session string, userID int, messageID, rating, reason string) (int, error) {
	pending, err := m.memoryService.RecordReplyFeedback(ctx, session, userID, messageID, rating, reason)
	switch {
	case errors.Is(err, application.ErrChatSessionNotFound):
		return 0, onekiterrors.New(http.StatusNotFound, "reply not found")
	case errors.Is(err, application.ErrInvalidFeedback):
		return 0, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return pending, err
}

// PendingLessons returns the user's current lessons and their unlearned
// feedback, rendered as reference text for a distillation prompt.
func (m *Module) PendingLessons(ctx context.Context, userID int) (lessons []string, feedback string, feedbackIDs []int64, err error) {
	work, err := m.memoryService.PendingLessonWork(ctx, userID)
	if err != nil || len(work.Feedback) == 0 {
		return nil, "", nil, err
	}
	var text strings.Builder
	for i, item := range work.Feedback {
		feedbackIDs = append(feedbackIDs, item.ID)
		verdict := map[string]string{"up": "thumbs up", "down": "thumbs down", "correction": "the person corrected Caret in their next message"}[item.Rating]
		fmt.Fprintf(&text, "Feedback %d: %s\n", i+1, verdict)
		if item.Reason != "" {
			fmt.Fprintf(&text, "Their words: %s\n", item.Reason)
		}
		fmt.Fprintf(&text, "They had asked: %s\nCaret replied: %s\n\n", item.Prompt, item.Reply)
	}
	return work.Lessons, text.String(), feedbackIDs, nil
}

// SaveLessons replaces the user's lessons and marks their feedback learned.
func (m *Module) SaveLessons(ctx context.Context, userID int, lessons []string, feedbackIDs []int64) error {
	return m.memoryService.SaveLessons(ctx, userID, lessons, feedbackIDs)
}

// PendingChatLearning returns the part of a chat not yet read for durable
// facts, once enough of it has accumulated, and the facts already known.
func (m *Module) PendingChatLearning(ctx context.Context, session string, userID int) (transcript string, known []string, throughID int, ok bool, err error) {
	work, ok, err := m.memoryService.PendingChatLearning(ctx, session, userID)
	return work.Transcript, work.Known, work.ThroughID, ok, err
}

// SaveLearnedFacts stores facts learned from a chat; personal facts follow
// the person, business facts stay with this business.
func (m *Module) SaveLearnedFacts(ctx context.Context, session string, userID int, personal, business []string, throughID int) error {
	facts := make([]application.LearnedFact, 0, len(personal)+len(business))
	for _, value := range personal {
		facts = append(facts, application.LearnedFact{Personal: true, Value: value})
	}
	for _, value := range business {
		facts = append(facts, application.LearnedFact{Value: value})
	}
	return m.memoryService.SaveLearnedFacts(ctx, session, userID, facts, throughID)
}

// LearnToolRoute remembers which tool served a kind of request.
func (m *Module) LearnToolRoute(ctx context.Context, userID int, tool string, words []string) error {
	return m.memoryService.LearnToolRoute(ctx, userID, tool, words)
}

// ToolRoutes returns the user's learned request-to-tool routes.
func (m *Module) ToolRoutes(ctx context.Context, userID int) (map[string][][]string, error) {
	return m.memoryService.ToolRoutes(ctx, userID)
}
