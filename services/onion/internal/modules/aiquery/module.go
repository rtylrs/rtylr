package aiquery

import (
	"context"

	"github.com/voxire/onion/internal/modules/aiquery/internal/application"
)

// Module exposes the safe NL-to-SQL query engine. It registers no HTTP routes of
// its own — the echo (Caret) module consumes it for Analyst mode via primitive
// methods, so echo never depends on this module's internal types.
type Module struct {
	service *application.Service
}

func NewModule(_ context.Context) *Module {
	return &Module{service: application.NewService()}
}

// SchemaPrompt returns the allowlisted view/column schema for the LLM prompt.
func (m *Module) SchemaPrompt() string {
	return m.service.SchemaPrompt()
}

// RunJSON validates an LLM-produced query spec and runs it for the caller's
// tenant (resolved from businessID), returning rows and the compiled SQL.
func (m *Module) RunJSON(ctx context.Context, businessID, userID int64, specJSON string) ([]map[string]any, string, error) {
	return m.service.RunJSON(ctx, businessID, userID, specJSON)
}
