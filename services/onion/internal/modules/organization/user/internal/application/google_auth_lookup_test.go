package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/netip"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
)

// errGoogleUserLookup stands in for a lookup failure that is neither nil nor
// sql.ErrNoRows (connection reset, context deadline, lock wait timeout).
var errGoogleUserLookup = errors.New("injected user lookup failure")

// googleLookupFailureDriver fails every read and counts writes, so the test can
// prove that provisioning was never attempted.
type googleLookupFailureDriver struct {
	execs int
}

func (d *googleLookupFailureDriver) Open(string) (driver.Conn, error) {
	return &googleLookupFailureConn{driver: d}, nil
}

type googleLookupFailureConn struct {
	driver *googleLookupFailureDriver
}

func (c *googleLookupFailureConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported by the fake driver")
}

func (c *googleLookupFailureConn) Close() error { return nil }

func (c *googleLookupFailureConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported by the fake driver")
}

func (c *googleLookupFailureConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errGoogleUserLookup
}

func (c *googleLookupFailureConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	c.driver.execs++

	return nil, errGoogleUserLookup
}

type googleLookupFailureConnector struct {
	driver *googleLookupFailureDriver
}

func (c googleLookupFailureConnector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.Open("")
}

func (c googleLookupFailureConnector) Driver() driver.Driver { return c.driver }

// Only sql.ErrNoRows means "no such user". Treating every other lookup error as
// a miss sent linkGoogleUser into the provisioning branch, which tried to
// create an account for an email that already exists.
func TestGoogleAuthLookupFailureDoesNotProvision(t *testing.T) {
	fake := &googleLookupFailureDriver{}
	db := sql.OpenDB(googleLookupFailureConnector{driver: fake})
	t.Cleanup(func() { _ = db.Close() })

	svc := &Service{repo: infrastructure.NewRepository(db)}

	_, err := svc.linkGoogleUser(context.Background(), &GoogleUser{
		Email:         "Existing@Example.com",
		EmailVerified: true,
	}, "", netip.Addr{}, "")
	if !errors.Is(err, errGoogleUserLookup) {
		t.Fatalf("linkGoogleUser = %v, want the lookup error", err)
	}
	if fake.execs != 0 {
		t.Fatalf("the lookup failure fell through to %d write(s); provisioning for an existing email would fail or duplicate the account", fake.execs)
	}
}
