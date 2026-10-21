package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
)

// The attempt budget is process-global state, so every invocation needs its own
// user id or a -count=2 run inherits the first one's exhausted budget.
var verifyEmailTestUserSeq atomic.Uint64

func nextVerifyEmailTestUser() int64 {
	return 1_000_000 + int64(verifyEmailTestUserSeq.Add(1))
}

// verifyAttemptProbe answers every verification lookup with no rows and counts
// the queries, so a test can tell "the code was rejected before the database"
// from "the database rejected the code".
type verifyAttemptProbe struct{ selects int }

type verifyAttemptConnector struct{ probe *verifyAttemptProbe }

func (c verifyAttemptConnector) Connect(context.Context) (driver.Conn, error) {
	return &verifyAttemptConn{probe: c.probe}, nil
}
func (c verifyAttemptConnector) Driver() driver.Driver { return verifyAttemptDriver{c.probe} }

type verifyAttemptDriver struct{ probe *verifyAttemptProbe }

func (d verifyAttemptDriver) Open(string) (driver.Conn, error) {
	return &verifyAttemptConn{probe: d.probe}, nil
}

type verifyAttemptConn struct{ probe *verifyAttemptProbe }

func (c *verifyAttemptConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported by the probe driver")
}
func (c *verifyAttemptConn) Close() error              { return nil }
func (c *verifyAttemptConn) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }
func (c *verifyAttemptConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &verifyAttemptTx{}, nil
}

func (c *verifyAttemptConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	c.probe.selects++

	return &verifyAttemptRows{}, nil
}

func (c *verifyAttemptConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

type verifyAttemptTx struct{}

func (verifyAttemptTx) Commit() error   { return nil }
func (verifyAttemptTx) Rollback() error { return nil }

type verifyAttemptRows struct{}

func (verifyAttemptRows) Columns() []string { return []string{"id"} }
func (verifyAttemptRows) Close() error      { return nil }
func (verifyAttemptRows) Next([]driver.Value) error {
	return io.EOF
}

func newVerifyAttemptService(t *testing.T) (*Service, *verifyAttemptProbe) {
	t.Helper()

	t.Setenv("EMAIL_VERIFICATION_SECRET", "test-verification-secret")

	probe := &verifyAttemptProbe{}
	db := sql.OpenDB(verifyAttemptConnector{probe: probe})
	t.Cleanup(func() { _ = db.Close() })

	return &Service{repo: infrastructure.NewRepository(db)}, probe
}

// A code minted before codes were bound to an account cannot be retried into
// working, so the caller is told to request a new one rather than to try again.
func TestVerifyEmailCodeReportsSupersededLegacyCode(t *testing.T) {
	svc, probe := newVerifyAttemptService(t)

	err := svc.VerifyEmailCode(context.Background(), "123456")
	if !errors.Is(err, ErrEmailCodeSuperseded) {
		t.Fatalf("VerifyEmailCode(legacy) = %v, want ErrEmailCodeSuperseded", err)
	}
	if probe.selects != 0 {
		t.Fatalf("a legacy code still reached the database %d time(s)", probe.selects)
	}
}

// Every other rejection is one opaque answer, so the anonymous endpoint cannot
// be used to learn which accounts have a pending code.
func TestVerifyEmailCodeCollapsesOtherFailures(t *testing.T) {
	svc, _ := newVerifyAttemptService(t)

	for _, token := range []string{"", "not-a-token", "1042.01234", "0.0123456789", "1042.012345678a"} {
		if err := svc.VerifyEmailCode(context.Background(), token); !errors.Is(err, ErrInvalidEmailCode) {
			t.Errorf("VerifyEmailCode(%q) = %v, want ErrInvalidEmailCode", token, err)
		}
	}
}

// Past the per-user budget the endpoint stops before the database, so a wrong
// code cannot be retried indefinitely behind the per-IP rate limit.
func TestVerifyEmailCodeStopsAfterTheAttemptBudget(t *testing.T) {
	svc, probe := newVerifyAttemptService(t)

	userID := nextVerifyEmailTestUser()
	for i := 0; i < 5; i++ {
		token := fmt.Sprintf("%d.%010d", userID, i+1)
		if err := svc.VerifyEmailCode(context.Background(), token); !errors.Is(err, ErrInvalidEmailCode) {
			t.Fatalf("attempt %d: VerifyEmailCode = %v, want ErrInvalidEmailCode", i+1, err)
		}
	}
	if probe.selects != 5 {
		t.Fatalf("%d database probes for 5 attempts, want 5", probe.selects)
	}

	overCap := fmt.Sprintf("%d.9999999999", userID)
	if err := svc.VerifyEmailCode(context.Background(), overCap); !errors.Is(err, ErrTooManyEmailCodeAttempts) {
		t.Fatalf("over the cap: VerifyEmailCode = %v, want ErrTooManyEmailCodeAttempts", err)
	}
	if probe.selects != 5 {
		t.Fatalf("the over-cap attempt still reached the database (%d probes)", probe.selects)
	}
}
