package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/platform/database/dbclient"
)

// scriptedQuery is one canned answer for a matching statement. Rows are only
// consulted for statements that return rows.
type scriptedQuery struct {
	match   func(query string) bool
	columns []string
	rows    [][]driver.Value
	err     error
}

type recordedStatement struct {
	query string
	args  []driver.Value
}

// fakeDatabase is a database/sql driver that answers from a script and records
// every statement, so tests can assert not just that a call failed but that it
// reached the database with nothing written.
type fakeDatabase struct {
	mu         sync.Mutex
	script     []scriptedQuery
	statements []recordedStatement
}

func (f *fakeDatabase) respond(query string, args []driver.Value) *scriptedQuery {
	for i := range f.script {
		if f.script[i].match == nil || f.script[i].match(query) {
			return &f.script[i]
		}
	}
	return nil
}

func (f *fakeDatabase) record(query string, args []driver.Value) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statements = append(f.statements, recordedStatement{query: query, args: args})
}

// Writes returns every recorded statement that is not a plain read.
func (f *fakeDatabase) Writes() []recordedStatement {
	f.mu.Lock()
	defer f.mu.Unlock()
	var writes []recordedStatement
	for _, statement := range f.statements {
		if !isReadStatement(statement.query) {
			writes = append(writes, statement)
		}
	}
	return writes
}

func (f *fakeDatabase) Queries() []recordedStatement {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedStatement(nil), f.statements...)
}

// isReadStatement reports whether a statement only reads. sqlc prefixes its
// queries with a `-- name: ...` comment line, so those are stripped first.
func isReadStatement(query string) bool {
	for _, line := range strings.Split(query, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		return strings.HasPrefix(strings.ToUpper(line), "SELECT")
	}
	return false
}

func valuesOf(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	return values
}

type fakeConnector struct{ db *fakeDatabase }

func (c fakeConnector) Connect(context.Context) (driver.Conn, error) { return &fakeConn{db: c.db}, nil }
func (c fakeConnector) Driver() driver.Driver                        { return fakeDriver{} }

type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) { return nil, driver.ErrBadConn }

type fakeConn struct{ db *fakeDatabase }

func (c *fakeConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *fakeConn) Close() error                        { return nil }
func (c *fakeConn) Begin() (driver.Tx, error)           { return fakeTx{}, nil }

type fakeTx struct{}

func (fakeTx) Commit() error   { return nil }
func (fakeTx) Rollback() error { return nil }

func (c *fakeConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	values := valuesOf(args)
	c.db.record(query, values)
	answer := c.db.respond(query, values)
	if answer == nil {
		return driver.RowsAffected(1), nil
	}
	if answer.err != nil {
		return nil, answer.err
	}
	return driver.RowsAffected(1), nil
}

func (c *fakeConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	values := valuesOf(args)
	c.db.record(query, values)
	answer := c.db.respond(query, values)
	if answer == nil {
		return &fakeRows{columns: []string{"1"}, rows: [][]driver.Value{{int64(1)}}}, nil
	}
	if answer.err != nil {
		return nil, answer.err
	}
	return &fakeRows{columns: answer.columns, rows: answer.rows}, nil
}

type fakeRows struct {
	columns []string
	rows    [][]driver.Value
	next    int
}

func (r *fakeRows) Columns() []string { return r.columns }
func (r *fakeRows) Close() error      { return nil }

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.next >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.next])
	r.next++
	return nil
}

// newFakeDB builds a *sql.DB backed by the scripted driver and points the
// package-level pool at it, so tenant.ResolveBusinessID — which reads
// dbclient.Pool() rather than taking a handle — sees the same script.
func newFakeDB(t *testing.T, script []scriptedQuery) (*sql.DB, *fakeDatabase) {
	t.Helper()
	fake := &fakeDatabase{script: script}
	db := sql.OpenDB(fakeConnector{db: fake})
	t.Cleanup(func() { _ = db.Close() })
	t.Cleanup(dbclient.SetPoolForTesting(db))
	return db, fake
}

func containsAny(fragments ...string) func(string) bool {
	return func(query string) bool {
		for _, fragment := range fragments {
			if strings.Contains(query, fragment) {
				return true
			}
		}
		return false
	}
}
