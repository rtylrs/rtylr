package api

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/business/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubbusinessv1 "github.com/voxire/voxie/gen/go/hub/business/v1"
)

// The caller these tests authenticate as: one business, one user.
const (
	callerBusinessID int64 = 7100
	callerUserID     int64 = 7200
)

// execStub stands in for MySQL. The slug update is scoped by a predicate in
// SQL, so the row count the driver reports is the only outcome a caller can
// observe, and recording the statement shows which tenant the write addressed.
type execStub struct {
	mu   sync.Mutex
	stmt string
	args []driver.Value
	rows int64
}

type execConn struct{ stub *execStub }

func (s *execStub) Open(string) (driver.Conn, error) { return &execConn{stub: s}, nil }

func (c *execConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *execConn) Close() error                        { return nil }
func (c *execConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func (c *execConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.stub.mu.Lock()
	defer c.stub.mu.Unlock()
	c.stub.stmt = query
	c.stub.args = make([]driver.Value, 0, len(args))
	for _, a := range args {
		c.stub.args = append(c.stub.args, a.Value)
	}
	return execResult{rows: c.stub.rows}, nil
}

type execResult struct{ rows int64 }

func (r execResult) LastInsertId() (int64, error) { return 0, nil }
func (r execResult) RowsAffected() (int64, error) { return r.rows, nil }

var execStubSeq atomic.Int64

func newStubDB(t *testing.T, rows int64) (*sql.DB, *execStub) {
	t.Helper()

	stub := &execStub{rows: rows}
	name := fmt.Sprintf("business-slug-stub-%d", execStubSeq.Add(1))
	sql.Register(name, stub)

	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, stub
}

// assertScopedToCaller requires both the business id and the user id to be
// bound. With the owner left out of the statement the update renames whatever
// row the business id names, which is the cross-tenant write.
func assertScopedToCaller(t *testing.T, stub *execStub) {
	t.Helper()

	stub.mu.Lock()
	stmt, args := stub.stmt, append([]driver.Value(nil), stub.args...)
	stub.mu.Unlock()

	var ids []int64
	for _, a := range args {
		if v, ok := a.(int64); ok {
			ids = append(ids, v)
		}
	}
	if len(ids) != 2 || ids[0] != callerBusinessID || ids[1] != callerUserID {
		t.Fatalf("statement %q bound %v, want the caller business %d and user %d", stmt, ids, callerBusinessID, callerUserID)
	}
}

func assertPermissionDenied(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("cross-tenant slug update reported success")
	}
	var apiErr *onekiterrors.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an api error, got %T: %v", err, err)
	}
	if apiErr.HTTPStatusCode() != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", apiErr.HTTPStatusCode())
	}
}

func TestSetSlug_CrossTenantUpdateIsDenied(t *testing.T) {
	db, stub := newStubDB(t, 0) // the owner predicate rejected the write
	api := NewBusinessAPI(application.NewService(context.Background(), moduleinfra.NewRepository(db)))

	_, err := api.SetSlug(deviceCtx(callerUserID, callerBusinessID), &hubbusinessv1.SetSlugRequest{Slug: "hijacked"})

	assertPermissionDenied(t, err)
	assertScopedToCaller(t, stub)
}

func TestSetSlug_OwnBusinessUpdateSucceeds(t *testing.T) {
	db, stub := newStubDB(t, 1)
	api := NewBusinessAPI(application.NewService(context.Background(), moduleinfra.NewRepository(db)))

	resp, err := api.SetSlug(deviceCtx(callerUserID, callerBusinessID), &hubbusinessv1.SetSlugRequest{Slug: "voxire"})
	if err != nil {
		t.Fatalf("SetSlug: %v", err)
	}
	if resp.GetSlug() != "voxire" {
		t.Fatalf("slug = %q, want %q", resp.GetSlug(), "voxire")
	}
	assertScopedToCaller(t, stub)
}

func TestSetSlug_UnauthenticatedIsDenied(t *testing.T) {
	db, _ := newStubDB(t, 1)
	api := NewBusinessAPI(application.NewService(context.Background(), moduleinfra.NewRepository(db)))

	_, err := api.SetSlug(context.Background(), &hubbusinessv1.SetSlugRequest{Slug: "voxire"})

	assertPermissionDenied(t, err)
}

func TestSetSlugHandler_CrossTenantUpdateIsDenied(t *testing.T) {
	db, stub := newStubDB(t, 0)
	handler := &SetSlugHandler{Service: application.NewService(context.Background(), moduleinfra.NewRepository(db))}

	_, err := handler.Serve(deviceCtx(callerUserID, callerBusinessID), &SetSlugRequest{Slug: "hijacked"})

	assertPermissionDenied(t, err)
	assertScopedToCaller(t, stub)
}

// A body naming another business or owner changes nothing: the target comes
// from the authenticated device, so the foreign ids never reach the statement.
func TestSetSlugHandler_IgnoresBusinessAndOwnerFromBody(t *testing.T) {
	db, stub := newStubDB(t, 1)
	handler := &SetSlugHandler{Service: application.NewService(context.Background(), moduleinfra.NewRepository(db))}

	var req SetSlugRequest
	body := `{"slug":"voxire","business_id":992001,"owner_id":991001}`
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if _, err := handler.Serve(deviceCtx(callerUserID, callerBusinessID), &req); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	assertScopedToCaller(t, stub)
}
