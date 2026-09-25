package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/wsbus"
	"github.com/voxire/onion/internal/testutil"
)

// sentinelEvent terminates a collection window. It is published on the client's
// own topic after the operation under test, and the hub fans out from a single
// loop in publish order into a per-client FIFO buffer, so the sentinel coming
// back proves every event published before it has already been delivered. That
// is what makes "nothing was published" a deterministic assertion instead of a
// sleep long enough to feel safe.
const sentinelEvent = "test.sentinel"

// readTimeout bounds every socket read so a missing event fails the test
// instead of hanging the package.
const readTimeout = 5 * time.Second

// startHub installs a real wsbus hub as the process default for one test, with
// the same self-scoped `user:<id>` rule main.go wires in production: a client
// may only ever subscribe to its own topic. Assertions therefore run against
// the wire a browser sees, not a stub.
func startHub(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	wsbus.SetDefault(ctx, wsbus.New(func(_ context.Context, userID int64, topic string) bool {
		scope, id, ok := wsbus.ParseTopic(topic)
		return ok && scope == "user" && id == userID
	}))

	return ctx
}

// wsFrame mirrors the hub's server-to-client envelope.
type wsFrame struct {
	Op    string       `json:"op"`
	Event *wsbus.Event `json:"event,omitempty"`
	Error string       `json:"error,omitempty"`
}

// wsClient is one browser: a real websocket connected to the real hub and
// subscribed to its own user topic.
//
// Frames are pulled off the socket by a single goroutine into a channel, and
// the assertions wait on the channel. A read is never given a deadline of its
// own: cancelling one closes the whole connection, which would make the first
// "nothing arrived" check destroy the client it was checking.
type wsClient struct {
	t      *testing.T
	conn   *websocket.Conn
	frames chan wsFrame
	userID int64
	topic  string
}

// dialUser connects a client for userID and leaves it subscribed to
// `user:<userID>`, ready to record.
func dialUser(ctx context.Context, t *testing.T, userID int64) *wsClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsbus.Default().ServeWS(w, r, userID)
	}))
	t.Cleanup(srv.Close)

	dialCtx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	conn, resp, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("websocket.Dial for user %d: %v", userID, err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })

	c := &wsClient{
		t:      t,
		conn:   conn,
		frames: make(chan wsFrame, 64),
		userID: userID,
		topic:  fmt.Sprintf("user:%d", userID),
	}
	go c.pump(ctx)
	c.subscribe(ctx)

	return c
}

// pump moves frames off the socket until the connection or the test ends.
func (c *wsClient) pump(ctx context.Context) {
	defer close(c.frames)

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}

		var frame wsFrame
		if unmarshalErr := json.Unmarshal(data, &frame); unmarshalErr != nil {
			return
		}

		select {
		case c.frames <- frame:
		case <-ctx.Done():
			return
		}
	}
}

// subscribe joins the client's own topic and blocks until the subscription is
// provably live.
//
// The ack is not enough on its own: the hub writes it straight to the client's
// send buffer while the registration is still queued on its run loop, so an
// event published right after an ack can race ahead of the subscription. A
// sentinel that comes back proves the registration landed.
func (c *wsClient) subscribe(ctx context.Context) {
	c.t.Helper()

	writeCtx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	req := fmt.Sprintf(`{"op":"subscribe","topics":[%q]}`, c.topic)
	if err := c.conn.Write(writeCtx, websocket.MessageText, []byte(req)); err != nil {
		c.t.Fatalf("subscribe write for user %d: %v", c.userID, err)
	}

	deadline := time.Now().Add(readTimeout)
	for time.Now().Before(deadline) {
		wsbus.Publish(wsbus.Event{Topic: c.topic, Type: sentinelEvent})
		if c.awaitSentinel(200 * time.Millisecond) {
			c.drain()
			return
		}
	}

	c.t.Fatalf("user %d never became subscribed to %s", c.userID, c.topic)
}

// awaitSentinel reports whether a sentinel arrives within wait. Anything else
// read in the meantime is discarded, which is only ever the subscribe ack or a
// duplicate sentinel from an earlier attempt.
func (c *wsClient) awaitSentinel(wait time.Duration) bool {
	deadline := time.Now().Add(wait)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false
		}
		frame, ok := c.read(remaining)
		if !ok {
			return false
		}
		if frame.Event != nil && frame.Event.Type == sentinelEvent {
			return true
		}
	}
}

// drain clears anything already buffered, so a collection window only sees
// events published after it opened.
func (c *wsClient) drain() {
	for {
		if _, ok := c.read(100 * time.Millisecond); !ok {
			return
		}
	}
}

// read returns the next frame, or false if none arrives within wait.
func (c *wsClient) read(wait time.Duration) (wsFrame, bool) {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case frame, ok := <-c.frames:
		return frame, ok
	case <-timer.C:
		return wsFrame{}, false
	}
}

// recorded returns every event delivered to this client since the last drain,
// closing the window with a sentinel.
func (c *wsClient) recorded() []wsbus.Event {
	c.t.Helper()

	wsbus.Publish(wsbus.Event{Topic: c.topic, Type: sentinelEvent})

	events := make([]wsbus.Event, 0, 4)
	deadline := time.Now().Add(readTimeout)
	for time.Now().Before(deadline) {
		frame, ok := c.read(time.Until(deadline))
		if !ok {
			c.t.Fatalf("user %d never received the closing sentinel", c.userID)
		}
		if frame.Event == nil {
			continue
		}
		if frame.Event.Type == sentinelEvent {
			return events
		}
		events = append(events, *frame.Event)
	}

	c.t.Fatalf("user %d never received the closing sentinel", c.userID)
	return nil
}

// countType counts events of a given type, ignoring the rest.
func countType(events []wsbus.Event, eventType string) int {
	n := 0
	for _, e := range events {
		if e.Type == eventType {
			n++
		}
	}
	return n
}

// assignmentNotification finds the notification raised for a given task, if any.
func assignmentNotification(ctx context.Context, t *testing.T, svc *Service, userID, taskID int64) (domain.Notification, bool) {
	t.Helper()

	notifs, _, err := svc.ListNotifications(ctx, userID, 100, 0)
	if err != nil {
		t.Fatalf("ListNotifications(%d): %v", userID, err)
	}
	for _, n := range notifs {
		if n.TaskID != nil && *n.TaskID == taskID {
			return n, true
		}
	}

	return domain.Notification{}, false
}

// A reassignment is the case that was silent: task events reach `task:<id>` and
// `project:<id>`, and a client on My tasks or Today subscribes to neither. Both
// sides need telling: the previous assignee has to drop the task from their
// list and the new one has to pick it up.
func TestUpdateTaskReassignmentPublishesToBothUserTopics(t *testing.T) {
	db := testutil.OpenDB(t)
	initCache(t)
	ctx := startHub(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "reassigned work", cacheTestUserA)

	previous := dialUser(ctx, t, cacheTestUserA)
	next := dialUser(ctx, t, cacheTestUserB)

	newAssignee := cacheTestUserB
	if err := svc.UpdateTask(ctx, domain.UpdateTaskInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		AssigneeID: &newAssignee,
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	previousEvents := previous.recorded()
	if countType(previousEvents, taskUpdatedEvent) != 1 {
		t.Fatalf("previous assignee got %d task events, want 1: %+v", countType(previousEvents, taskUpdatedEvent), previousEvents)
	}
	if previousEvents[0].EntityID != task.ID {
		t.Errorf("previous assignee's event entity id = %d, want %d", previousEvents[0].EntityID, task.ID)
	}
	if previousEvents[0].ActorID != cacheTestUserA {
		t.Errorf("previous assignee's event actor id = %d, want %d", previousEvents[0].ActorID, cacheTestUserA)
	}

	nextEvents := next.recorded()
	if countType(nextEvents, taskUpdatedEvent) != 1 {
		t.Errorf("new assignee got %d task events, want 1: %+v", countType(nextEvents, taskUpdatedEvent), nextEvents)
	}
	// The inbox half of the same change rides the topic the frontend already
	// handles, so the badge updates without waiting on the task branch.
	if countType(nextEvents, "notification.created") != 1 {
		t.Errorf("new assignee got %d notification events, want 1: %+v", countType(nextEvents, "notification.created"), nextEvents)
	}
}

// The scope decision, pinned: only an assignment change reaches a user topic.
// Status and title edits on a task someone already holds stay on the task and
// project topics, because sending every task event to the assignee is chattier
// for a case that is not silent today.
func TestUpdateTaskWithoutReassignmentPublishesToNoUserTopic(t *testing.T) {
	db := testutil.OpenDB(t)
	initCache(t)
	ctx := startHub(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "before rename", cacheTestUserA)

	assignee := dialUser(ctx, t, cacheTestUserA)
	bystander := dialUser(ctx, t, cacheTestUserB)

	if err := svc.UpdateTask(ctx, domain.UpdateTaskInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		Title:      "after rename",
		Status:     "in_progress",
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if events := assignee.recorded(); len(events) != 0 {
		t.Errorf("a rename reached the assignee's user topic: %+v", events)
	}
	if events := bystander.recorded(); len(events) != 0 {
		t.Errorf("a rename reached an uninvolved user's topic: %+v", events)
	}
	if _, found := assignmentNotification(ctx, t, svc, cacheTestUserB, task.ID); found {
		t.Error("a rename created an assignment notification")
	}
}

// Being handed a task produced no inbox row at all before this: nothing on the
// assignee-change path called CreateNotification.
func TestUpdateTaskReassignmentNotifiesTheNewAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "needs an owner", cacheTestUserA)

	newAssignee := cacheTestUserB
	if err := svc.UpdateTask(ctx, domain.UpdateTaskInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		Title:      "renamed on assignment",
		AssigneeID: &newAssignee,
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	notification, found := assignmentNotification(ctx, t, svc, cacheTestUserB, task.ID)
	if !found {
		t.Fatal("the new assignee has no notification for the task they were just given")
	}
	if notification.Type != assignedNotification {
		t.Errorf("notification type = %q, want %q", notification.Type, assignedNotification)
	}
	// The same call renamed the task, so the row has to name what the assignee
	// will actually see on their board.
	if notification.TaskTitle != "renamed on assignment" {
		t.Errorf("notification task title = %q, want %q", notification.TaskTitle, "renamed on assignment")
	}
	if notification.ActorID == nil || *notification.ActorID != cacheTestUserA {
		t.Errorf("notification actor = %v, want %d", notification.ActorID, cacheTestUserA)
	}

	// The previous assignee is told through the realtime event, not the inbox:
	// losing a task is not news worth a badge.
	if _, found := assignmentNotification(ctx, t, svc, cacheTestUserA, task.ID); found {
		t.Error("the previous assignee got an inbox row for a task taken off them")
	}
}

// Assigning work to yourself is not news. The realtime event still goes out,
// because it keeps the list correct, but the inbox stays quiet.
func TestUpdateTaskSelfAssignmentSkipsTheNotification(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "self assigned", cacheTestUserB)

	newAssignee := cacheTestUserA
	if err := svc.UpdateTask(ctx, domain.UpdateTaskInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		AssigneeID: &newAssignee,
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if _, found := assignmentNotification(ctx, t, svc, cacheTestUserA, task.ID); found {
		t.Error("assigning a task to yourself created an inbox row")
	}
}
