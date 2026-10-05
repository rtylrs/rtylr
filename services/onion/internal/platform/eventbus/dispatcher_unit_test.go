package eventbus_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/eventbus"
)

func TestDispatcherDistributesFactBeforeMarkingIt(t *testing.T) {
	ctx := context.Background()
	state := &dispatchDBState{}
	db := sql.OpenDB(dispatchConnector{state: state})
	t.Cleanup(func() { _ = db.Close() })

	subscriber := &recordingSubscriber{}
	dispatched, err := eventbus.NewDispatcher(db, subscriber).DispatchBatch(ctx)
	if err != nil {
		t.Fatalf("DispatchBatch: %v", err)
	}
	if dispatched != 1 || state.marks != 1 {
		t.Fatalf("dispatched=%d marks=%d, want 1 each", dispatched, state.marks)
	}
	if len(subscriber.events) != 1 {
		t.Fatalf("subscriber received %d events, want 1", len(subscriber.events))
	}
	event := subscriber.events[0]
	if event.ID != 71 || event.BusinessID != 19 || event.Type != "order_created" || event.SubjectID != 202 {
		t.Fatalf("subscriber received wrong event: %+v", event)
	}
}

func TestDispatcherLeavesFactUndispatchedWhenConsumerFails(t *testing.T) {
	ctx := context.Background()
	state := &dispatchDBState{}
	db := sql.OpenDB(dispatchConnector{state: state})
	t.Cleanup(func() { _ = db.Close() })

	subscriber := &recordingSubscriber{err: errors.New("consumer unavailable")}
	dispatched, err := eventbus.NewDispatcher(db, subscriber).DispatchBatch(ctx)
	if err != nil {
		t.Fatalf("DispatchBatch should retry a consumer failure on the next pass: %v", err)
	}
	if dispatched != 0 || state.marks != 0 {
		t.Fatalf("dispatched=%d marks=%d after consumer failure, want 0", dispatched, state.marks)
	}
}

type recordingSubscriber struct {
	events []eventbus.Event
	err    error
}

func (s *recordingSubscriber) DispatchEvent(_ context.Context, event eventbus.Event) error {
	s.events = append(s.events, event)
	return s.err
}

type dispatchDBState struct{ marks int }

type dispatchConnector struct{ state *dispatchDBState }

func (c dispatchConnector) Connect(context.Context) (driver.Conn, error) {
	return &dispatchConn{state: c.state}, nil
}

func (c dispatchConnector) Driver() driver.Driver { return dispatchDriver{} }

type dispatchDriver struct{}

func (dispatchDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type dispatchConn struct{ state *dispatchDBState }

func (c *dispatchConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not supported") }
func (c *dispatchConn) Close() error                        { return nil }
func (c *dispatchConn) Begin() (driver.Tx, error)           { return nil, errors.New("not supported") }

func (c *dispatchConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if !strings.Contains(query, "FROM platform_events") {
		return nil, errors.New("unexpected query")
	}
	return &dispatchRows{values: []driver.Value{
		int64(71), int64(19), "order_created", "order", int64(202), []byte(`{"order_id":202}`), time.Date(2026, 10, 4, 8, 0, 0, 0, time.UTC),
	}}, nil
}

func (c *dispatchConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if !strings.Contains(query, "SET dispatched_at = NOW()") {
		return nil, errors.New("unexpected exec")
	}
	c.state.marks++
	return driver.RowsAffected(1), nil
}

type dispatchRows struct {
	values []driver.Value
	read   bool
}

func (*dispatchRows) Columns() []string {
	return []string{"id", "business_id", "event_type", "subject_type", "subject_id", "payload", "occurred_at"}
}
func (*dispatchRows) Close() error { return nil }
func (r *dispatchRows) Next(dest []driver.Value) error {
	if r.read {
		return io.EOF
	}
	r.read = true
	copy(dest, r.values)
	return nil
}
