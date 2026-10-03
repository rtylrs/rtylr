package memory

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/api"
	"github.com/voxire/onion/internal/modules/memory/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/snips"
	hubmemoryv1 "github.com/voxire/voxie/gen/go/hub/memory/v1"
)

type Module struct {
	memoryService *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	memoryService := application.NewService(ctx, db)

	return &Module{
		memoryService: memoryService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.memoryService
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubmemoryv1.RegisterMemoryServiceServer(
		api.NewMemoryAPI(m.memoryService),
		hubmemoryv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub memory service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterPOSTWithValidation[api.SessionRequest, api.SessionResponse](
		ctx,
		router,
		[]string{"/memory/sessions"},
		&api.SessionHandler{MemoryService: m.memoryService},
	)
	serverHttp.RegisterPOSTWithValidation[api.MemoryRequest, api.MemoryResponse](
		ctx,
		router,
		[]string{"/memory/store"},
		&api.MemoryHandler{MemoryService: m.memoryService},
	)
	serverHttp.RegisterPOSTWithValidation[api.ContextRequest, api.ContextResponse](
		ctx,
		router,
		[]string{"/memory/context"},
		&api.ContextHandler{MemoryService: m.memoryService},
	)
	serverHttp.RegisterPOSTWithValidation[api.SearchRequest, api.SearchResponse](
		ctx,
		router,
		[]string{"/memory/search"},
		&api.SearchHandler{MemoryService: m.memoryService},
	)
	serverHttp.RegisterPOSTWithValidation[api.StatsRequest, api.StatsResponse](
		ctx,
		router,
		[]string{"/memory/stats"},
		&api.StatsHandler{MemoryService: m.memoryService},
	)
	serverHttp.RegisterPOSTWithValidation[api.SummarizeRequest, api.SummarizeResponse](
		ctx,
		router,
		[]string{"/memory/summarize"},
		&api.SummarizeHandler{MemoryService: m.memoryService},
	)
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
