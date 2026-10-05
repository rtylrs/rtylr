package aiquery

import (
	"context"

	"github.com/voxire/onion/internal/modules/aiquery/internal/application"
	"github.com/voxire/onion/internal/modules/aiquery/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/tenant"
)

// Module exposes tenant-scoped analytical reads without granting write access.
type Module struct {
	service *application.Service
}

func NewModule(_ context.Context) *Module {
	return &Module{service: application.NewService(infrastructure.NewExecutor(), tenant.ResolveBusinessID, infrastructure.BuildSQL)}
}

// SchemaPrompt returns the allowlisted view/column schema for the LLM prompt.
func (m *Module) SchemaPrompt() string {
	return m.service.SchemaPrompt()
}

// RunJSON validates an LLM-produced query spec and runs it for the caller's
// business (resolved and access-checked from businessID), returning rows and
// the compiled SQL.
func (m *Module) RunJSON(ctx context.Context, businessID, userID int64, specJSON string) ([]map[string]any, string, error) {
	return m.service.RunJSON(ctx, businessID, userID, specJSON)
}
