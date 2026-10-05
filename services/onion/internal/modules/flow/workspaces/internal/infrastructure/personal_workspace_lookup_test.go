package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
)

// errPersonalWorkspaceLookup stands in for a transient lookup failure (pool
// exhaustion, statement timeout) that is neither nil nor sql.ErrNoRows.
var errPersonalWorkspaceLookup = errors.New("injected personal-workspace lookup failure")

// lookupFailureDriver fails every read of the idempotency lookup and counts
// writes, so the test can prove the INSERT was never attempted.
type lookupFailureDriver struct {
	execs int
}

func (d *lookupFailureDriver) Open(string) (driver.Conn, error) {
	return &lookupFailureConn{driver: d}, nil
}

type lookupFailureConn struct {
	driver *lookupFailureDriver
}

func (c *lookupFailureConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported by the fake driver")
}

func (c *lookupFailureConn) Close() error { return nil }

func (c *lookupFailureConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported by the fake driver")
}

func (c *lookupFailureConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errPersonalWorkspaceLookup
}

func (c *lookupFailureConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	c.driver.execs++

	return nil, errPersonalWorkspaceLookup
}

type lookupFailureConnector struct {
	driver *lookupFailureDriver
}

func (c lookupFailureConnector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.Open("")
}

func (c lookupFailureConnector) Driver() driver.Driver { return c.driver }

// CreatePersonalWorkspace is called on every registration and login hook and is
// documented as idempotent. flow_workspaces carries no unique key, so treating
// a transient lookup failure as "no workspace yet" creates a second 'Personal'
// workspace for the same owner+business. The lookup error has to stop the write.
func TestCreatePersonalWorkspaceLookupFailureDoesNotInsert(t *testing.T) {
	fake := &lookupFailureDriver{}
	db := sql.OpenDB(lookupFailureConnector{driver: fake})
	t.Cleanup(func() { _ = db.Close() })

	repo := &Repository{queries: sqlc.New(db), pool: db}

	if _, err := repo.CreatePersonalWorkspace(context.Background(), 7, 7); !errors.Is(err, errPersonalWorkspaceLookup) {
		t.Fatalf("CreatePersonalWorkspace = %v, want the lookup error", err)
	}
	if fake.execs != 0 {
		t.Fatalf("the lookup failure fell through to %d write(s); a duplicate personal workspace would be created", fake.execs)
	}
}
