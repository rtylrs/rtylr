package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/aiquery/internal/domain"
	"github.com/voxire/onion/internal/modules/aiquery/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/tenant"
)

type recordingExecutor struct {
	calls int
	query string
	args  []any
	rows  []map[string]any
	err   error
}

func (e *recordingExecutor) Run(_ context.Context, query string, args []any) ([]map[string]any, error) {
	e.calls++
	e.query = query
	e.args = args
	return e.rows, e.err
}

func newTestService(resolve businessScopeResolver, exec executor) *Service {
	return &Service{exec: exec, resolveScope: resolve, compile: infrastructure.BuildSQL}
}

func fixedResolver(businessID int64) businessScopeResolver {
	return func(context.Context, int64, int64) (int64, error) { return businessID, nil }
}

// The exploit as reported: a member of business B1 asks for invoice totals and
// must only ever get B1's rows back, whatever the spec says.
func TestRunJSON_BindsTheCallersBusiness(t *testing.T) {
	exec := &recordingExecutor{rows: []map[string]any{{"invoice_number": "B1-1"}}}
	svc := newTestService(fixedResolver(7001), exec)

	rows, sql, err := svc.RunJSON(context.Background(), 7001, 555,
		`{"prompt":"total invoices by status","view":"ai_invoices","columns":["status"],"aggregations":[{"func":"sum","col":"total","as":"total"}],"group_by":["status"]}`)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "business_id = ?") {
		t.Fatalf("query is not business scoped: %s", sql)
	}
	if strings.Contains(sql, "owner_id") {
		t.Fatalf("query is owner scoped: %s", sql)
	}
	if len(exec.args) != 1 || exec.args[0] != int64(7001) {
		t.Fatalf("expected only the caller's business bound, got %v", exec.args)
	}
	if len(rows) != 1 {
		t.Fatalf("expected the caller's rows, got %v", rows)
	}
}

// The resolved business is the scope, not whatever the request carried: a token
// with no business falls back to the user's default, and that id is the filter.
func TestRunJSON_ScopesByResolvedBusinessNotRequestBusiness(t *testing.T) {
	exec := &recordingExecutor{}
	svc := newTestService(fixedResolver(7002), exec)

	if _, _, err := svc.RunJSON(context.Background(), 0, 555,
		`{"view":"ai_invoices","columns":["invoice_number"]}`); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(exec.args) != 1 || exec.args[0] != int64(7002) {
		t.Fatalf("expected the resolved business, got %v", exec.args)
	}
}

// Fails closed: an unverifiable business scope must not degrade into an owner id
// or an unscoped query, and must never reach the database.
func TestRunJSON_UnresolvableBusinessScopeNeverQueries(t *testing.T) {
	exec := &recordingExecutor{}
	svc := newTestService(func(context.Context, int64, int64) (int64, error) {
		return 0, tenant.ErrNoBusinessScope
	}, exec)

	_, _, err := svc.RunJSON(context.Background(), 7001, 555,
		`{"view":"ai_invoices","columns":["invoice_number"]}`)
	if !errors.Is(err, tenant.ErrNoBusinessScope) {
		t.Fatalf("expected ErrNoBusinessScope, got %v", err)
	}
	if exec.calls != 0 {
		t.Fatalf("expected no query to run, got %d: %s", exec.calls, exec.query)
	}
}

func TestRunJSON_RejectsSpecThatNamesBusinessScope(t *testing.T) {
	for _, spec := range []string{
		`{"view":"ai_invoices","where":[{"col":"business_id","op":"=","value":7002}]}`,
		`{"view":"ai_invoices","columns":["business_id","invoice_number"]}`,
		`{"view":"ai_employees","aggregations":[{"func":"avg","col":"salary","as":"avg"}],"group_by":["business_id"]}`,
	} {
		exec := &recordingExecutor{}
		svc := newTestService(fixedResolver(7001), exec)

		if _, _, err := svc.RunJSON(context.Background(), 7001, 555, spec); err == nil {
			t.Errorf("spec %s: expected rejection", spec)
		}
		if exec.calls != 0 {
			t.Errorf("spec %s: expected no query to run, got %d", spec, exec.calls)
		}
	}
}

// A zero-scope resolution must not compile into a query, even when the
// resolver itself hands back 0 without an error.
func TestRunSpec_ZeroBusinessScopeIsRejected(t *testing.T) {
	exec := &recordingExecutor{}
	svc := newTestService(nil, exec)

	if _, err := svc.RunSpec(context.Background(), 0, domain.QuerySpec{View: "ai_invoices"}); !errors.Is(err, domain.ErrNoBusinessScope) {
		t.Fatalf("expected ErrNoBusinessScope, got %v", err)
	}
	if exec.calls != 0 {
		t.Fatalf("expected no query to run, got %d", exec.calls)
	}
}
