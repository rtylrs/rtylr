package api

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// inviteStore holds the rows the employee-invite path reads and writes so the
// real repository, service and handler run against it without a MySQL server.
// It answers exactly the statements hr_core issues on that path and fails on
// anything else, so an unexpected query shows up as a test failure instead of
// silently returning no rows.
type inviteStore struct {
	mu        sync.Mutex
	invites   map[string]*inviteStoreInvite
	employees map[int64]*inviteStoreEmployee
	accounts  map[int64]string
	consumed  map[string]int
}

type inviteStoreInvite struct {
	ownerID    int64
	businessID int64
	employeeID int64
	expiresAt  time.Time
	consumed   bool
}

type inviteStoreEmployee struct {
	businessID int64
	email      string
	status     string
	userID     int64
}

func newInviteStore() *inviteStore {
	return &inviteStore{
		invites:   map[string]*inviteStoreInvite{},
		employees: map[int64]*inviteStoreEmployee{},
		accounts:  map[int64]string{},
		consumed:  map[string]int{},
	}
}

func (s *inviteStore) addAccount(userID int64, email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[userID] = email
}

func (s *inviteStore) addInvited(businessID, employeeID int64, email, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.employees[employeeID] = &inviteStoreEmployee{businessID: businessID, email: email, status: "invited"}
	s.invites[token] = &inviteStoreInvite{
		ownerID:    businessID,
		businessID: businessID,
		employeeID: employeeID,
		expiresAt:  time.Now().Add(72 * time.Hour),
	}
}

func (s *inviteStore) tokenConsumed(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, ok := s.invites[token]
	return !ok || invite.consumed
}

// tokenConsumptions counts the successful consume statements for a token, so a
// replay can be told apart from the single legitimate use.
func (s *inviteStore) tokenConsumptions(token string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.consumed[token]
}

func (s *inviteStore) employee(employeeID int64) (inviteStoreEmployee, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	employee, ok := s.employees[employeeID]
	if !ok {
		return inviteStoreEmployee{}, false
	}
	return *employee, true
}

// unlinkAccount releases an account from whatever employee row holds it, which
// models the only real recovery from a failed acceptance.
func (s *inviteStore) unlinkAccount(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, employee := range s.employees {
		if employee.userID == userID {
			employee.userID = 0
		}
	}
}

type inviteStoreState struct {
	invites   map[string]*inviteStoreInvite
	employees map[int64]*inviteStoreEmployee
}

func (s *inviteStore) snapshot() *inviteStoreState {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := &inviteStoreState{
		invites:   make(map[string]*inviteStoreInvite, len(s.invites)),
		employees: make(map[int64]*inviteStoreEmployee, len(s.employees)),
	}
	for token, invite := range s.invites {
		copied := *invite
		state.invites[token] = &copied
	}
	for id, employee := range s.employees {
		copied := *employee
		state.employees[id] = &copied
	}
	return state
}

func (s *inviteStore) restore(state *inviteStoreState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invites = state.invites
	s.employees = state.employees
}

// --- driver plumbing ---

type inviteStoreConnector struct{ store *inviteStore }

func (c inviteStoreConnector) Connect(context.Context) (driver.Conn, error) {
	return &inviteStoreConn{store: c.store}, nil
}

func (c inviteStoreConnector) Driver() driver.Driver { return inviteStoreDriver{} }

type inviteStoreDriver struct{}

func (inviteStoreDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("invite store connections come from the connector")
}

type inviteStoreConn struct{ store *inviteStore }

func (c *inviteStoreConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("invite store answers queries directly")
}

func (c *inviteStoreConn) Close() error { return nil }

func (c *inviteStoreConn) Begin() (driver.Tx, error) {
	return &inviteStoreTx{store: c.store, snapshot: c.store.snapshot()}, nil
}

type inviteStoreTx struct {
	store    *inviteStore
	snapshot *inviteStoreState
}

func (tx *inviteStoreTx) Commit() error { return nil }

func (tx *inviteStoreTx) Rollback() error {
	tx.store.restore(tx.snapshot)
	return nil
}

func inviteStoreArg(args []driver.NamedValue, index int) int64 {
	if index >= len(args) {
		return 0
	}
	value, _ := args[index].Value.(int64)
	return value
}

func inviteStoreStringArg(args []driver.NamedValue, index int) string {
	if index >= len(args) {
		return ""
	}
	value, _ := args[index].Value.(string)
	return value
}

func (c *inviteStoreConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "FROM hr_employee_invites"):
		return c.store.queryInvite(inviteStoreStringArg(args, 0))
	case strings.Contains(query, "FROM users"):
		return c.store.queryAccount(inviteStoreArg(args, 0))
	case strings.Contains(query, "FROM hr_employees") && strings.Contains(query, "user_id = ?"):
		return c.store.queryEmployeeIDByUser(inviteStoreArg(args, 0), inviteStoreArg(args, 1))
	case strings.Contains(query, "FROM hr_employees"):
		return c.store.queryEmployee(inviteStoreArg(args, 0), inviteStoreArg(args, 1))
	}
	return nil, fmt.Errorf("invite store: unexpected query %q", query)
}

func (c *inviteStoreConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	switch {
	case strings.Contains(query, "UPDATE hr_employee_invites"):
		return c.store.consumeInvite(inviteStoreStringArg(args, 0))
	case strings.Contains(query, "UPDATE hr_employees"):
		return c.store.linkEmployee(inviteStoreArg(args, 0), inviteStoreArg(args, 1), inviteStoreArg(args, 2))
	}
	return nil, fmt.Errorf("invite store: unexpected statement %q", query)
}

func (s *inviteStore) queryInvite(token string) (driver.Rows, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	columns := []string{"id", "owner_id", "business_id", "employee_id", "token", "expires_at", "consumed_at", "created_at"}
	invite, ok := s.invites[token]
	if !ok || invite.consumed {
		return &inviteStoreRows{columns: columns}, nil
	}
	now := time.Now().UTC()
	return &inviteStoreRows{columns: columns, rows: [][]driver.Value{{
		int64(1), invite.ownerID, invite.businessID, invite.employeeID, token, invite.expiresAt, nil, now,
	}}}, nil
}

func (s *inviteStore) queryAccount(userID int64) (driver.Rows, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	columns := []string{"email"}
	email, ok := s.accounts[userID]
	if !ok {
		return &inviteStoreRows{columns: columns}, nil
	}
	return &inviteStoreRows{columns: columns, rows: [][]driver.Value{{email}}}, nil
}

func (s *inviteStore) queryEmployee(businessID, employeeID int64) (driver.Rows, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	columns := []string{
		"id", "owner_id", "business_id", "user_id", "department_id", "location_id", "work_schedule_id",
		"first_name", "last_name", "email", "phone", "employment_type", "status",
		"salary", "hourly_rate", "overtime_rate", "start_date", "end_date",
		"address", "emergency_contact_name", "emergency_contact_phone", "notes",
		"clock_in_pin_hash", "created_at", "updated_at",
	}
	employee, ok := s.employees[employeeID]
	if !ok || employee.businessID != businessID {
		return &inviteStoreRows{columns: columns}, nil
	}
	now := time.Now().UTC()
	var userID driver.Value
	if employee.userID != 0 {
		userID = employee.userID
	}
	return &inviteStoreRows{columns: columns, rows: [][]driver.Value{{
		employeeID, businessID, businessID, userID, nil, nil, nil,
		"Invited", "Employee", employee.email, nil, "full_time", employee.status,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil,
		nil, now, now,
	}}}, nil
}

func (s *inviteStore) queryEmployeeIDByUser(businessID, userID int64) (driver.Rows, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	columns := []string{"id"}
	if userID == 0 {
		return &inviteStoreRows{columns: columns}, nil
	}
	var match int64
	for id, employee := range s.employees {
		if employee.businessID == businessID && employee.userID == userID && (match == 0 || id < match) {
			match = id
		}
	}
	if match == 0 {
		return &inviteStoreRows{columns: columns}, nil
	}
	return &inviteStoreRows{columns: columns, rows: [][]driver.Value{{match}}}, nil
}

func (s *inviteStore) consumeInvite(token string) (driver.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invite, ok := s.invites[token]
	if !ok || invite.consumed {
		return driver.RowsAffected(0), nil
	}
	invite.consumed = true
	s.consumed[token]++
	return driver.RowsAffected(1), nil
}

func (s *inviteStore) linkEmployee(userID, employeeID, businessID int64) (driver.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	employee, ok := s.employees[employeeID]
	if !ok || employee.businessID != businessID || employee.status != "invited" {
		return driver.RowsAffected(0), nil
	}
	for id, other := range s.employees {
		if id != employeeID && other.userID == userID {
			return nil, &mysql.MySQLError{
				Number:  1062,
				Message: fmt.Sprintf("Duplicate entry '%d' for key 'hr_employees.uniq_employee_user_id'", userID),
			}
		}
	}
	employee.userID = userID
	employee.status = "active"
	return driver.RowsAffected(1), nil
}

type inviteStoreRows struct {
	columns []string
	rows    [][]driver.Value
	pos     int
}

func (r *inviteStoreRows) Columns() []string { return r.columns }

func (r *inviteStoreRows) Close() error { return nil }

func (r *inviteStoreRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

func openInviteStore(t *testing.T, store *inviteStore) *sql.DB {
	t.Helper()
	db := sql.OpenDB(inviteStoreConnector{store: store})
	t.Cleanup(func() { _ = db.Close() })
	return db
}
