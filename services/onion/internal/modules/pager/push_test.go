package pager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

// fakePush records what would have gone to the push services. gone lists the
// endpoints the push service no longer knows; fail makes every send fail.
type fakePush struct {
	mu   sync.Mutex
	sent []sentPush
	gone map[string]bool
	fail bool
}

type sentPush struct {
	endpoint string
	topic    string
	payload  map[string]string
}

func (f *fakePush) PublicKey() string { return "test-vapid-key" }

func (f *fakePush) Send(_ context.Context, sub domain.PushSubscription, payload []byte, topic string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var body map[string]string
	if err := json.Unmarshal(payload, &body); err != nil {
		return err
	}
	f.sent = append(f.sent, sentPush{endpoint: sub.Endpoint, topic: topic, payload: body})
	if f.gone[sub.Endpoint] {
		return domain.ErrPushGone
	}
	if f.fail {
		return errors.New("push service unreachable")
	}
	return nil
}

// take returns what was sent since the last call.
func (f *fakePush) take() []sentPush {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.sent
	f.sent = nil
	return out
}

// enablePush turns push on with pushes sent before the staff reply returns,
// so a test sees them without waiting.
func (h *harness) enablePush() *fakePush {
	push := &fakePush{gone: map[string]bool{}}
	h.service.EnablePush(push, func(send func()) { send() })
	return push
}

func endpoint(name string) string {
	return "https://fcm.googleapis.com/fcm/send/" + name
}

var (
	testP256dh = base64.RawURLEncoding.EncodeToString(append([]byte{4}, make([]byte, 64)...))
	testAuth   = base64.RawURLEncoding.EncodeToString(make([]byte, 16))
	// validSubscriptionFields is the body of an acceptable subscription
	// without its braces.
	validSubscriptionFields = fmt.Sprintf(`"endpoint":%q,"p256dh":%q,"auth":%q`, endpoint("anonymous"), testP256dh, testAuth)
)

func (h *harness) subscribe(token, endpoint, p256dh, auth string) reply {
	h.Helper()
	body, err := json.Marshal(map[string]string{"endpoint": endpoint, "p256dh": p256dh, "auth": auth})
	if err != nil {
		h.Fatal(err)
	}
	return h.call(nil, http.MethodPost, "/pager/v1/public/tickets/"+token+"/push", string(body))
}

// pushRows counts the stored subscriptions of a venue's tickets.
func (h *harness) pushRows(v venue) int {
	h.Helper()
	if h.db == nil {
		h.memory.mu.Lock()
		defer h.memory.mu.Unlock()
		count := 0
		for _, row := range h.memory.pushes {
			if row.sub.BusinessID == v.businessID {
				count++
			}
		}
		return count
	}
	var count int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM pager_push_subscriptions WHERE business_id = ?", v.businessID).Scan(&count); err != nil {
		h.Fatal(err)
	}
	return count
}

// joinAndFind joins a number as a customer and returns the token and the
// staff's id for the ticket.
func (h *harness) joinAndFind(v venue, number string) (string, any) {
	h.Helper()
	token := h.want(h.join(v, number), http.StatusOK, "").text("public_token")
	for _, ticket := range h.open(v) {
		if ticket["number"] == number {
			return token, ticket["id"]
		}
	}
	h.Fatalf("ticket %s is not on the board", number)
	return "", nil
}

func TestPushStaysOffWithoutKeys(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		key := h.want(h.call(nil, http.MethodGet, "/pager/v1/public/push-key", ""), http.StatusOK, "")
		if key.text("public_key") != "" {
			h.Fatalf("push key without configuration = %q, want none", key.text("public_key"))
		}
		token, id := h.joinAndFind(h.a, "42")
		// A browser that subscribes anyway is told yes and nothing is kept.
		h.want(h.subscribe(token, endpoint("a"), testP256dh, testAuth), http.StatusOK, "")
		if rows := h.pushRows(h.a); rows != 0 {
			h.Fatalf("stored %d subscriptions with push off", rows)
		}
		h.want(h.move(h.a.member, id, "ready"), http.StatusOK, "")
		h.want(h.move(h.a.member, id, "remind"), http.StatusOK, "")
		h.want(h.move(h.a.member, id, "collect"), http.StatusOK, "")
	})
}

func TestPushRingsOnReadyAndOnCallAgain(t *testing.T) {
	eachBackend(t, func(h *harness) {
		push := h.enablePush()
		h.enable(h.a, false)
		key := h.want(h.call(nil, http.MethodGet, "/pager/v1/public/push-key", ""), http.StatusOK, "")
		if key.text("public_key") != "test-vapid-key" {
			h.Fatalf("push key = %q", key.text("public_key"))
		}
		token, id := h.joinAndFind(h.a, "42")
		h.want(h.subscribe(token, endpoint("a"), testP256dh, testAuth), http.StatusOK, "")
		h.want(h.subscribe(token, endpoint("b"), testP256dh, testAuth), http.StatusOK, "")
		// The same browser subscribing again on a reload is still one row.
		h.want(h.subscribe(token, endpoint("a"), testP256dh, testAuth), http.StatusOK, "")
		if rows := h.pushRows(h.a); rows != 2 {
			h.Fatalf("stored %d subscriptions, want 2", rows)
		}
		if sent := push.take(); len(sent) != 0 {
			h.Fatalf("subscribing sent %v", sent)
		}

		h.want(h.move(h.a.member, id, "ready"), http.StatusOK, "")
		sent := push.take()
		if len(sent) != 2 {
			h.Fatalf("ready sent %d pushes, want one per browser", len(sent))
		}
		for _, p := range sent {
			want := map[string]string{"kind": "ready", "number": "42", "venue": "Cedar Café", "token": token}
			if fmt.Sprint(p.payload) != fmt.Sprint(want) {
				h.Fatalf("ready payload = %v, want %v", p.payload, want)
			}
			if p.topic == "" || p.topic != sent[0].topic {
				h.Fatalf("topic = %q, want one topic per ticket", p.topic)
			}
		}

		h.want(h.move(h.a.member, id, "remind"), http.StatusOK, "")
		sent = push.take()
		if len(sent) != 2 || sent[0].payload["kind"] != "remind" {
			h.Fatalf("call again sent %v, want a remind push per browser", sent)
		}

		// Undo is not news for the customer.
		h.want(h.move(h.a.member, id, "undo"), http.StatusOK, "")
		if sent := push.take(); len(sent) != 0 {
			h.Fatalf("undo sent %v", sent)
		}
	})
}

func TestPushFailureDoesNotFailTheStaffAction(t *testing.T) {
	h := newHarness(t, nil)
	push := h.enablePush()
	push.fail = true
	h.enable(h.a, false)
	token, id := h.joinAndFind(h.a, "42")
	h.want(h.subscribe(token, endpoint("a"), testP256dh, testAuth), http.StatusOK, "")
	h.want(h.move(h.a.member, id, "ready"), http.StatusOK, "")
	if sent := push.take(); len(sent) != 1 {
		t.Fatalf("sent %d pushes, want the one attempt", len(sent))
	}
	// An unreachable push service is not a reason to forget the browser.
	if rows := h.pushRows(h.a); rows != 1 {
		t.Fatalf("stored %d subscriptions after a failed push, want 1", rows)
	}
}

func TestPushSubscribeNeedsAnOpenTicketAtAnEnabledVenue(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enablePush()
		h.enable(h.a, false)
		token, id := h.joinAndFind(h.a, "42")

		h.want(h.subscribe(strings.Repeat("A", 32), endpoint("a"), testP256dh, testAuth), http.StatusNotFound, "ticket_not_found")
		h.want(h.subscribe("not-a-token", endpoint("a"), testP256dh, testAuth), http.StatusNotFound, "ticket_not_found")
		h.want(h.subscribe(fmt.Sprint(id), endpoint("a"), testP256dh, testAuth), http.StatusNotFound, "ticket_not_found")

		h.want(h.call(h.a.owner, http.MethodPut, "/pager/v1/settings", `{"enabled":false}`), http.StatusOK, "")
		h.want(h.subscribe(token, endpoint("a"), testP256dh, testAuth), http.StatusNotFound, "ticket_not_found")
		h.enable(h.a, false)

		h.want(h.move(h.a.member, id, "collect"), http.StatusOK, "")
		h.want(h.subscribe(token, endpoint("a"), testP256dh, testAuth), http.StatusNotFound, "ticket_not_found")

		expiring, _ := h.joinAndFind(h.a, "43")
		h.clock.advance(domain.OpenLifetime + time.Minute)
		h.want(h.subscribe(expiring, endpoint("a"), testP256dh, testAuth), http.StatusNotFound, "ticket_not_found")

		if rows := h.pushRows(h.a); rows != 0 {
			h.Fatalf("stored %d subscriptions for tickets nobody may follow", rows)
		}
	})
}

func TestPushRefusesEndpointsOutsideThePushServices(t *testing.T) {
	h := newHarness(t, nil)
	h.enablePush()
	h.enable(h.a, false)
	token, _ := h.joinAndFind(h.a, "42")

	for _, bad := range []struct{ name, endpoint, p256dh, auth string }{
		{"plain http", "http://fcm.googleapis.com/fcm/send/x", testP256dh, testAuth},
		{"internal host", "https://169.254.169.254/latest/meta-data", testP256dh, testAuth},
		{"localhost", "https://localhost/fcm/send/x", testP256dh, testAuth},
		{"lookalike suffix", "https://fcm.googleapis.com.evil.example/fcm/send/x", testP256dh, testAuth},
		{"lookalike prefix", "https://evilfcm.googleapis.com/fcm/send/x", testP256dh, testAuth},
		{"bare suffix host", "https://push.services.mozilla.com/wpush/v2/x", testP256dh, testAuth},
		{"port", "https://fcm.googleapis.com:8443/fcm/send/x", testP256dh, testAuth},
		{"credentials", "https://user:pass@fcm.googleapis.com/fcm/send/x", testP256dh, testAuth},
		{"not a url", "fcm.googleapis.com/fcm/send/x", testP256dh, testAuth},
		{"short key", endpoint("x"), base64.RawURLEncoding.EncodeToString(make([]byte, 65)), testAuth},
		{"bad base64 key", endpoint("x"), "!!!!", testAuth},
		{"short auth", endpoint("x"), testP256dh, base64.RawURLEncoding.EncodeToString(make([]byte, 8))},
	} {
		if r := h.subscribe(token, bad.endpoint, bad.p256dh, bad.auth); r.status != http.StatusBadRequest {
			t.Errorf("%s: got %d %v, want 400", bad.name, r.status, r.body)
		}
	}
	if rows := h.pushRows(h.a); rows != 0 {
		t.Fatalf("stored %d refused subscriptions", rows)
	}

	for _, good := range []string{
		"https://fcm.googleapis.com/fcm/send/abc",
		"https://updates.push.services.mozilla.com/wpush/v2/abc",
		"https://wns2-par02p.notify.windows.com/w/?token=abc",
		"https://web.push.apple.com/abc",
	} {
		h.want(h.subscribe(token, good, testP256dh, testAuth), http.StatusOK, "")
	}
	// Keys padded the way some browsers report them are accepted too.
	h.want(h.subscribe(token, endpoint("padded"), testP256dh+"=", testAuth+"=="), http.StatusOK, "")
}

func TestPushForgetsGoneBrowsersAndClosedTickets(t *testing.T) {
	eachBackend(t, func(h *harness) {
		push := h.enablePush()
		h.enable(h.a, false)
		token, id := h.joinAndFind(h.a, "42")
		h.want(h.subscribe(token, endpoint("gone"), testP256dh, testAuth), http.StatusOK, "")
		h.want(h.subscribe(token, endpoint("kept"), testP256dh, testAuth), http.StatusOK, "")
		push.gone[endpoint("gone")] = true

		h.want(h.move(h.a.member, id, "ready"), http.StatusOK, "")
		if sent := push.take(); len(sent) != 2 {
			h.Fatalf("ready sent %d pushes, want 2", len(sent))
		}
		if rows := h.pushRows(h.a); rows != 1 {
			h.Fatalf("%d subscriptions left after a 410, want 1", rows)
		}
		h.want(h.move(h.a.member, id, "remind"), http.StatusOK, "")
		if sent := push.take(); len(sent) != 1 || sent[0].endpoint != endpoint("kept") {
			h.Fatalf("call again sent %v, want only the browser still subscribed", sent)
		}

		h.want(h.move(h.a.member, id, "collect"), http.StatusOK, "")
		if rows := h.pushRows(h.a); rows != 0 {
			h.Fatalf("%d subscriptions left after collect", rows)
		}
	})
}

func TestPushSweepsSubscriptionsOfTicketsPastTheirLifetime(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enablePush()
		h.enable(h.a, false)
		h.enable(h.b, false)
		old, _ := h.joinAndFind(h.a, "42")
		h.want(h.subscribe(old, endpoint("old"), testP256dh, testAuth), http.StatusOK, "")

		h.clock.advance(domain.OpenLifetime + time.Minute)
		// Any later subscription, at any venue, clears what has expired.
		fresh, _ := h.joinAndFind(h.b, "7")
		h.want(h.subscribe(fresh, endpoint("fresh"), testP256dh, testAuth), http.StatusOK, "")
		if rows := h.pushRows(h.a); rows != 0 {
			h.Fatalf("%d subscriptions of an expired ticket left", rows)
		}
		if rows := h.pushRows(h.b); rows != 1 {
			h.Fatalf("%d fresh subscriptions, want 1", rows)
		}
	})
}

func TestPushLimitsBrowsersPerTicket(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enablePush()
		h.enable(h.a, false)
		token, _ := h.joinAndFind(h.a, "42")
		for i := range domain.MaxPushSubscriptionsPerTicket {
			h.want(h.subscribe(token, endpoint(fmt.Sprint(i)), testP256dh, testAuth), http.StatusOK, "")
		}
		h.want(h.subscribe(token, endpoint("one-more"), testP256dh, testAuth), http.StatusTooManyRequests, "too_many_push_subscriptions")
		if rows := h.pushRows(h.a); rows != domain.MaxPushSubscriptionsPerTicket {
			h.Fatalf("stored %d subscriptions, want %d", rows, domain.MaxPushSubscriptionsPerTicket)
		}
	})
}
