package application

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/voxire/onion/internal/modules/aiquery/internal/domain"
	"github.com/voxire/onion/internal/modules/aiquery/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/tenant"
)

type executor interface {
	Run(ctx context.Context, query string, args []any) ([]map[string]any, error)
}

// businessScopeResolver maps the caller's context to the business id that owns
// their rows. It is a field so the fail-closed path can be exercised without a
// database; production always uses tenant.ResolveBusinessID.
type businessScopeResolver func(ctx context.Context, businessID, userID int64) (int64, error)

type Service struct {
	exec         executor
	resolveScope businessScopeResolver
}

func NewService() *Service {
	return &Service{
		exec:         infrastructure.NewExecutor(),
		resolveScope: tenant.ResolveBusinessID,
	}
}

// Result is a structured query outcome the AI layer can summarize for the user.
type Result struct {
	View string           `json:"view"`
	SQL  string           `json:"sql"`
	Rows []map[string]any `json:"rows"`
}

// RunSpec validates the spec, compiles it to a parameterized query bound to
// businessID, and executes it on the read replica. The returned SQL is safe to
// log (no raw user input — identifiers come from the allowlist, values are
// bound).
func (s *Service) RunSpec(ctx context.Context, businessID int64, spec domain.QuerySpec) (Result, error) {
	query, args, err := domain.BuildSQL(spec, businessID)
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

// RunJSON parses an LLM-produced query spec (JSON), resolves the caller's
// canonical business id, and executes it — returning rows and the compiled SQL.
// The spec is validated against the allowlist and bound to the business before
// anything runs.
//
// tenant.ResolveBusinessID verifies the caller actually has access to the
// business and fails closed with tenant.ErrNoBusinessScope when it cannot, so
// an unverifiable request never reaches the database.
func (s *Service) RunJSON(ctx context.Context, businessID, userID int64, specJSON string) ([]map[string]any, string, error) {
	var spec domain.QuerySpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		return nil, "", fmt.Errorf("invalid query spec JSON: %w", err)
	}
	scopedBusinessID, err := s.resolveScope(ctx, businessID, userID)
	if err != nil {
		return nil, "", err
	}
	res, err := s.RunSpec(ctx, scopedBusinessID, spec)
	if err != nil {
		return nil, "", err
	}
	return res.Rows, res.SQL, nil
}

// SchemaPrompt renders the allowlisted views and columns as a compact,
// deterministic schema string to embed in the spec-generation prompt.
func (s *Service) SchemaPrompt() string {
	schema := domain.Schema()
	views := make([]string, 0, len(schema))
	for v := range schema {
		views = append(views, v)
	}
	sort.Strings(views)

	var b strings.Builder
	for _, v := range views {
		cols := append([]string(nil), schema[v]...)
		sort.Strings(cols)
		b.WriteString(v)
		b.WriteString("(")
		b.WriteString(strings.Join(cols, ", "))
		b.WriteString(")\n")
	}
	return b.String()
}
