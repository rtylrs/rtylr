package echo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy"
	"github.com/voxire/onion/internal/modules/caret/internal/api"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/modules/caret/internal/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

var aiStreamRateLimit = serverHttp.AIUserRateLimit(5, time.Minute)

type Module struct {
	buddies        *buddy.Module
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

	// Sub-agents call the same executors as the main reply (reads only); the
	// list is resolved at call time because it includes the agent executor.
	agents := &application.SubAgentExecutor{AI: snipsService}
	find := &application.BusinessFindExecutor{}
	toolExecutors = append(toolExecutors, agents, find)
	agents.Executors = func() []application.ToolExecutor { return toolExecutors }
	find.Executors = agents.Executors

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
	learner, _ := m.memoryService.(application.LessonLearner)
	history, _ := m.memoryService.(handlers.DecisionHistory)
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
			Learner:  learner,
			AI:       m.snipsService,
			History:  history,
		},
		hubcaretv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register caret runtime service: %w", err))
	}
	if m.buddies != nil {
		if err := hubcaretv1.RegisterBuddyServiceServer(&handlers.BuddyAPI{Store: m.buddies, Executors: m.toolExecutors}, hubcaretv1.WithMux(mux)); err != nil {
			panic(fmt.Errorf("register buddy service: %w", err))
		}
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
