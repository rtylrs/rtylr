package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/api"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/application"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/infrastructure"
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
