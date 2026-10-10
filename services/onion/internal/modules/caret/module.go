package echo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/api"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/modules/caret/internal/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

var aiStreamRateLimit = serverHttp.AIUserRateLimit(5, time.Minute)

type Module struct {
	snipsService   *application.Service
	memoryService  application.MemoryService
	quotaService   handlers.QuotaService
	analystService application.AnalystService
	toolExecutors  []application.ToolExecutor
}

func NewModule(ctx context.Context, memoryService application.MemoryService, quotaService handlers.QuotaService, analystService application.AnalystService, toolExecutors ...application.ToolExecutor) *Module {
	if err := application.ValidateToolRegistry(); err != nil {
		panic(fmt.Errorf("invalid Caret tool registry: %w", err))
	}
	snipsService, err := application.NewService(ctx)
	if err != nil {
		snipsService = nil
	}
	if analystService != nil {
		toolExecutors = append([]application.ToolExecutor{
			application.NewBusinessSnapshotToolExecutor(analystService),
			application.NewAnalyticsQueryToolExecutor(snipsService, analystService),
		}, toolExecutors...)
	}
	if profile, ok := memoryService.(application.AssistantProfileReader); ok {
		toolExecutors = append([]application.ToolExecutor{application.NewAssistantToolExecutor(profile, nil)}, toolExecutors...)
	}

	return &Module{
		snipsService:   snipsService,
		memoryService:  memoryService,
		quotaService:   quotaService,
		analystService: analystService,
		toolExecutors:  toolExecutors,
	}
}

func (m *Module) GetSnipsService() *application.Service {
	return m.snipsService
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubcaretv1.RegisterCaretServiceServer(
		api.NewCaretAPI(),
		hubcaretv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub caret service: %w", err))
	}
	if err := hubcaretv1.RegisterCaretRuntimeServiceServer(
		&handlers.RuntimeAPI{
			Tools:    &handlers.ListToolsHandler{},
			Status:   &handlers.ProviderStatusHandler{SnipsService: m.snipsService},
			Decision: &handlers.ToolDecisionHandler{Executors: m.toolExecutors, RequirePendingApproval: true},
		},
		hubcaretv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register caret runtime service: %w", err))
	}
	// The reply stream is Server-Sent Events, which the JSON contract cannot
	// carry, so it stays a raw handler behind the per-user AI rate limit.
	for _, path := range handlers.GenerateTextStreamRoutes() {
		stream := &handlers.GenerateTextStreamHandler{
			AiService:      m.snipsService,
			MemoryService:  m.memoryService,
			QuotaService:   m.quotaService,
			AnalystService: m.analystService,
			ToolExecutors:  m.toolExecutors,
		}
		mux.Handle("POST "+path, aiStreamRateLimit(stream))
	}
}
