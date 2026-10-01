package tenant

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
)

// ownerStubDriver is a database/sql driver whose every query fails. It stands
// in for an unreachable database so the lookup-error path is exercised without
// a live server. It is test-only and registers under a private name.
type ownerStubDriver struct{}

type ownerStubConn struct{}

type ownerStubStmt struct{}

type ownerStubResult struct{}

// ownerStubMode selects the stub's behaviour per test.
type ownerStubMode int

const (
	// ownerStubQueryError makes every query fail with a driver error.
	ownerStubQueryError ownerStubMode = iota
	// ownerStubNoRows returns an empty result set, which database/sql reports
	// as sql.ErrNoRows.
	ownerStubNoRows
	// ownerStubZeroOwner returns a row whose owner_id is 0.
	ownerStubZeroOwner
	// ownerStubOwner returns a row with a valid positive owner_id.
	ownerStubOwner
)

var (
	ownerStubCurrentMode ownerStubMode
	ownerStubOwnerValue  int64 = 7
	ownerStubRowsErr           = errors.New("owner stub: connection refused")
)

func (ownerStubDriver) Open(string) (driver.Conn, error) { return ownerStubConn{}, nil }

func (ownerStubConn) Prepare(string) (driver.Stmt, error) { return ownerStubStmt{}, nil }
func (ownerStubConn) Close() error                        { return nil }
func (ownerStubConn) Begin() (driver.Tx, error)           { return nil, ownerStubRowsErr }

func (ownerStubStmt) Close() error  { return nil }
func (ownerStubStmt) NumInput() int { return 1 }
func (ownerStubStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, ownerStubRowsErr
}
func (ownerStubStmt) Query([]driver.Value) (driver.Rows, error) {
	switch ownerStubCurrentMode {
	case ownerStubNoRows:
		return &ownerStubRows{}, nil
	case ownerStubZeroOwner:
		return &ownerStubRows{values: [][]driver.Value{{int64(0)}}}, nil
	case ownerStubOwner:
		return &ownerStubRows{values: [][]driver.Value{{ownerStubOwnerValue}}}, nil
	default:
		return nil, ownerStubRowsErr
	}
}

type ownerStubRows struct {
	values [][]driver.Value
	pos    int
}

func (r *ownerStubRows) Columns() []string { return []string{"owner_id"} }
func (r *ownerStubRows) Close() error      { return nil }
func (r *ownerStubRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.pos])
	r.pos++
	return nil
}

func (ownerStubResult) LastInsertId() (int64, error) { return 0, ownerStubRowsErr }
func (ownerStubResult) RowsAffected() (int64, error) { return 0, ownerStubRowsErr }

func init() {
	sql.Register("owner_stub", ownerStubDriver{})
}

// ownerStubDB is a non-nil *sql.DB handle driven by the given stub mode.
func ownerStubDB(t *testing.T, mode ownerStubMode) *sql.DB {
	t.Helper()
	previous := ownerStubCurrentMode
	ownerStubCurrentMode = mode
	t.Cleanup(func() { ownerStubCurrentMode = previous })

	db, err := sql.Open("owner_stub", "")
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestResolveOwnerIDFailsClosed asserts that no failure path can hand a caller
// a usable owner: never the fallback user (an unrelated user's id) and never
// the shared 0 bucket. Every path must return a non-nil error.
func TestResolveOwnerIDFailsClosed(t *testing.T) {
	const fallbackUserID = 4242

	tests := []struct {
		name       string
		businessID int64
		db         func(*testing.T) *sql.DB
		wantErr    error
	}{
		{
			name:       "no business in scope",
			businessID: 0,
			db:         func(t *testing.T) *sql.DB { return ownerStubDB(t, ownerStubOwner) },
			wantErr:    ErrNoBusinessScope,
		},
		{
			name:       "negative business id",
			businessID: -1,
			db:         func(t *testing.T) *sql.DB { return ownerStubDB(t, ownerStubOwner) },
			wantErr:    ErrNoBusinessScope,
		},
		{
			name:       "nil database pool",
			businessID: 7,
			db:         func(*testing.T) *sql.DB { return nil },
			wantErr:    ErrOwnerLookupFailed,
		},
		{
			name:       "unreachable database",
			businessID: 7,
			db:         func(t *testing.T) *sql.DB { return ownerStubDB(t, ownerStubQueryError) },
			wantErr:    ErrOwnerLookupFailed,
		},
		{
			name:       "business row missing",
			businessID: 7,
			db:         func(t *testing.T) *sql.DB { return ownerStubDB(t, ownerStubNoRows) },
			wantErr:    ErrOwnerLookupFailed,
		},
		{
			name:       "business row carries a zero owner",
			businessID: 7,
			db:         func(t *testing.T) *sql.DB { return ownerStubDB(t, ownerStubZeroOwner) },
			wantErr:    ErrOwnerLookupFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerID, err := resolveOwnerID(context.Background(), tt.db(t), tt.businessID, fallbackUserID)

			if err == nil {
				t.Fatalf("resolveOwnerID() error = nil, want %v (owner id returned: %d)", tt.wantErr, ownerID)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("resolveOwnerID() error = %v, want %v", err, tt.wantErr)
			}
			if ownerID != 0 {
				t.Errorf("resolveOwnerID() owner id = %d, want 0 on a failure path", ownerID)
			}
			if ownerID == fallbackUserID {
				t.Error("resolveOwnerID() returned the fallback user as owner; that is the fail-open")
			}
		})
	}
}

// TestResolveOwnerIDReturnsVerifiedOwnerOnSuccess pins the success path: a
// verified active business with a positive owner returns that owner and no
// error. The fail-closed hardening must not break it.
func TestResolveOwnerIDReturnsVerifiedOwnerOnSuccess(t *testing.T) {
	const wantOwner = 7

	db := ownerStubDB(t, ownerStubOwner)
	ownerID, err := resolveOwnerID(context.Background(), db, 42, 4242)
	if err != nil {
		t.Fatalf("resolveOwnerID() error = %v, want nil", err)
	}
	if ownerID != wantOwner {
		t.Fatalf("resolveOwnerID() owner id = %d, want %d", ownerID, wantOwner)
	}
}

func TestResolveOwnerIDNeverReturnsUsableOwnerOnFailure(t *testing.T) {
	// A failed lookup must not be distinguishable from a success by the id
	// alone; callers key off a positive id.
	ownerID, err := resolveOwnerID(context.Background(), nil, 7, 99)
	if err == nil {
		t.Fatal("expected an error for a nil pool")
	}
	if ownerID > 0 {
		t.Fatalf("owner id = %d, must not be usable after a failed lookup", ownerID)
	}
}

func TestResolveBusinessIDFailsClosedWithoutDatabase(t *testing.T) {
	resolved, err := resolveBusinessID(context.Background(), nil, 42, 7)

	if resolved != 0 {
		t.Fatalf("resolved business id = %d, want 0", resolved)
	}
	if !errors.Is(err, ErrNoBusinessScope) {
		t.Fatalf("error = %v, want ErrNoBusinessScope", err)
	}
}

func TestResolveBusinessIDRequiresAuthenticatedUser(t *testing.T) {
	resolved, err := resolveBusinessID(context.Background(), nil, 42, 0)

	if resolved != 0 {
		t.Fatalf("resolved business id = %d, want 0", resolved)
	}
	if !errors.Is(err, ErrNoBusinessScope) {
		t.Fatalf("error = %v, want ErrNoBusinessScope", err)
	}
}

func TestBusinessAccessQueryChecksActiveBusinessMembership(t *testing.T) {
	query := strings.ToLower(businessAccessQuery)

	for _, expected := range []string{
		"coalesce(b.is_active, true) = true",
		"b.owner_id = ?",
		"from authorization_business_members abm",
		"abm.user_id = ?",
		"abm.business_id = b.id",
		"abm.status = 'active'",
	} {
		if !strings.Contains(query, expected) {
			t.Errorf("business access query missing %q", expected)
		}
	}
}
