package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/aiquery/internal/domain"
	"github.com/voxire/onion/internal/modules/aiquery/internal/infrastructure"
)

type executor interface {
	Run(ctx context.Context, query string, args []any) ([]map[string]any, error)
}

type Service struct {
	exec executor
}

func NewService() *Service {
	return &Service{exec: infrastructure.NewExecutor()}
}

// Result is a structured query outcome the AI layer can summarize for the user.
type Result struct {
	View string           `json:"view"`
	SQL  string           `json:"sql"`
	Rows []map[string]any `json:"rows"`
}

// RunSpec validates the spec, compiles it to a parameterized query bound to
// ownerID, and executes it on the read replica. The returned SQL is safe to log
// (no raw user input — identifiers come from the allowlist, values are bound).
func (s *Service) RunSpec(ctx context.Context, ownerID int64, spec domain.QuerySpec) (Result, error) {
	query, args, err := domain.BuildSQL(spec, ownerID)
	if err != nil {
		return Result{}, err
	}

	rows, err := s.exec.Run(ctx, query, args)
	if err != nil {
		return Result{}, err
	}

	return Result{View: spec.View, SQL: query, Rows: rows}, nil
}

// Schema exposes the allowlisted views/columns for prompting the LLM.
func (s *Service) Schema() map[string][]string {
	return domain.Schema()
}
