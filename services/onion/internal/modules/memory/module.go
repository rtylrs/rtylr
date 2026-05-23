package memory

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/memory/internal/application"
	"github.com/voxire/onion/internal/modules/memory/internal/routes"
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

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.memoryService)
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

func (m *Module) SummarizeSession(ctx context.Context, sessionID string, userID int) error {
	return m.memoryService.SummarizeSession(ctx, sessionID, userID)
}
