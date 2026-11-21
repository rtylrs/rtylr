package pager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/business"
	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/pager/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	serverhttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/testutil"
)

func TestJoinAttachesToTheOpenTicketOrCreatesIt(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)

		// The customer is first: the ticket is theirs to create.
		first := h.want(h.join(h.a, "42"), http.StatusOK, "")
		if first.text("number") != "42" || first.text("status") != domain.StatusWaiting || first.text("public_token") == "" {
			h.Fatalf("join = %v", first.body)
		}
		// A second phone follows the same ticket with the same token.
		if again := h.want(h.join(h.a, "42"), http.StatusOK, ""); again.text("public_token") != first.text("public_token") {
			h.Fatal("a second phone got a different token for the same number")
		}

		// The cashier is first: the customer attaches, typed with Arabic-Indic
		// digits, padding and lower case.
		staffTicket := h.create(h.a, "a7")
		attached := h.want(h.join(h.a, " A٧ "), http.StatusOK, "")
		if attached.text("number") != "A7" {
			h.Fatalf("number = %q, want the normalised A7", attached.text("number"))
		}
		// Entering an open number again returns the ticket that holds it.
		if again := h.create(h.a, "A7"); again["id"] != staffTicket["id"] {
			h.Fatalf("staff create of an open number made ticket %v, want the existing %v", again["id"], staffTicket["id"])
		}

		open := h.open(h.a)
		if len(open) != 2 {
			h.Fatalf("open tickets = %v, want exactly the two numbers", open)
		}
		if open[0]["number"] != "42" || open[0]["source"] != domain.SourceCustomer || open[0]["phone_linked"] != true {
			h.Fatalf("customer ticket = %v", open[0])
		}
		if open[1]["id"] != staffTicket["id"] || open[1]["source"] != domain.SourceStaff || open[1]["phone_linked"] != true {
			h.Fatalf("staff ticket after a customer joined = %v", open[1])
		}
	})
}

func TestStrictVenueJoinsOnlyNumbersTheStaffEntered(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, true)

		h.want(h.join(h.a, "42"), http.StatusNotFound, "ticket_not_found")
		if open := h.open(h.a); len(open) != 0 {
			h.Fatalf("a refused join left tickets behind: %v", open)
		}

		h.create(h.a, "42")
		h.want(h.join(h.a, "42"), http.StatusOK, "")
	})
}

func TestPublicRoutesAnswerOnlyForAnEnabledVenue(t *testing.T) {
	eachBackend(t, func(h *harness) {
		venuePath := "/pager/v1/public/venues/" + h.a.slug
		h.want(h.call(nil, http.MethodGet, venuePath, ""), http.StatusNotFound, "venue_not_found")
		h.want(h.join(h.a, "42"), http.StatusNotFound, "venue_not_found")
		h.want(h.call(nil, http.MethodGet, "/pager/v1/public/venues/no-such-place", ""), http.StatusNotFound, "venue_not_found")

		h.enable(h.a, false)
		found := h.want(h.call(nil, http.MethodGet, venuePath, ""), http.StatusOK, "")
		if found.text("name") != "Cedar Café" || found.body["enabled"] != true {
			h.Fatalf("venue = %v", found.body)
		}
		token := h.want(h.join(h.a, "42"), http.StatusOK, "").text("public_token")
		h.want(h.publicTicket(token), http.StatusOK, "")

		// Switching the pager off hides the tickets already handed out.
		h.want(h.call(h.a.owner, http.MethodPut, "/pager/v1/settings", `{"enabled":false}`), http.StatusOK, "")
		h.want(h.publicTicket(token), http.StatusNotFound, "ticket_not_found")
		h.want(h.call(nil, http.MethodPost, "/pager/v1/public/tickets/"+token+"/acknowledge", ""), http.StatusNotFound, "ticket_not_found")
	})
}

func TestJoinRefusesNumbersThatAreNotReceiptNumbers(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		for _, number := range []string{"   ", "12-3", "4 2", "#42", strings.Repeat("9", 13)} {
			h.want(h.join(h.a, number), http.StatusBadRequest, "invalid_number")
		}
		if open := h.open(h.a); len(open) != 0 {
			h.Fatalf("an invalid number was stored: %v", open)
		}
	})
}

// Business B's staff see and change nothing of business A's, and the business
// comes from who is signed in, never from the request.
func TestStaffRoutesAreScopedToTheCallersBusiness(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		ours := h.create(h.a, "5")
		// The same number is open at both venues at once.
		theirs := h.create(h.b, "5")

		ids := func(tickets []map[string]any) []any {
			out := make([]any, 0, len(tickets))
			for _, ticket := range tickets {
				out = append(out, ticket["id"])
			}
			return out
		}
		testutil.AssertTenantIsolation(h, ids(h.open(h.b)), []any{theirs["id"]}, []any{ours["id"]})
		testutil.AssertTenantIsolation(h, ids(h.open(h.a)), []any{ours["id"]}, []any{theirs["id"]})

		for _, action := range []string{"ready", "remind", "undo", "collect"} {
			h.want(h.move(h.b.member, ours["id"], action), http.StatusNotFound, "ticket_not_found")
		}
		if open := h.open(h.a); len(open) != 1 || open[0]["status"] != domain.StatusWaiting {
			h.Fatalf("business A's ticket after business B's attempts = %v", open)
		}

		// B's switches are B's alone.
		if settings := h.call(h.b.member, http.MethodGet, "/pager/v1/settings", ""); settings.body["enabled"] == true {
			h.Fatalf("business B reads business A's settings: %v", settings.body)
		}

		// No session, and a session whose membership no longer resolves.
		stranger := &authPkg.AuthUser{UserID: 1, BusinessID: h.a.businessID}
		for _, user := range []*authPkg.AuthUser{nil, stranger} {
			if r := h.call(user, http.MethodGet, "/pager/v1/tickets", ""); r.status != http.StatusForbidden {
				h.Fatalf("list without a membership = %d %v, want 403", r.status, r.body)
			}
			if r := h.move(user, ours["id"], "collect"); r.status != http.StatusForbidden {
				h.Fatalf("collect without a membership = %d %v, want 403", r.status, r.body)
			}
		}
	})
}

func TestOnlyOwnersAndAdminsChangeSettings(t *testing.T) {
	eachBackend(t, func(h *harness) {
		if r := h.call(h.a.member, http.MethodPut, "/pager/v1/settings", `{"enabled":true}`); r.status != http.StatusForbidden {
			h.Fatalf("member update = %d %v, want 403", r.status, r.body)
		}
		if r := h.want(h.call(h.a.member, http.MethodGet, "/pager/v1/settings", ""), http.StatusOK, ""); r.body["enabled"] == true {
			h.Fatalf("a refused update took effect: %v", r.body)
		}

		h.enable(h.a, true)
		// A switch left out of the request keeps its value.
		r := h.want(h.call(h.a.owner, http.MethodPut, "/pager/v1/settings", `{"strict":false}`), http.StatusOK, "")
		if r.body["enabled"] != true || r.body["strict"] == true {
			h.Fatalf("settings after a partial update = %v, want enabled and not strict", r.body)
		}
	})
}

// The cashier and several phones enter the same number at the same moment.
// Whoever wins, there is one open ticket and everyone holds its token.
func TestConcurrentCreateAndJoinShareOneTicket(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		const rounds, phones = 12, 6

		for round := range rounds {
			number := fmt.Sprintf("R%d", round)
			tokens := make([]string, phones)
			statuses := make([]int, phones+1)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := range phones {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					r := h.join(h.a, number)
					statuses[i], tokens[i] = r.status, r.text("public_token")
				}()
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				statuses[phones] = h.call(h.a.member, http.MethodPost, "/pager/v1/tickets", h.numberBody(number)).status
			}()
			close(start)
			wg.Wait()

			for i, status := range statuses {
				if status != http.StatusOK {
					h.Fatalf("round %d request %d answered %d", round, i, status)
				}
			}
			for _, token := range tokens {
				if token == "" || token != tokens[0] {
					h.Fatalf("round %d handed out different tokens for one number: %v", round, tokens)
				}
			}
		}

		open := h.open(h.a)
		if len(open) != rounds {
			h.Fatalf("%d open tickets for %d numbers: %v", len(open), rounds, open)
		}
		for _, ticket := range open {
			if ticket["phone_linked"] != true {
				h.Fatalf("ticket %v lost its phone link in the race", ticket)
			}
		}
	})
}

// The rule lives in the database, not only in the service: a second open
// ticket for a number is refused even when written directly.
func TestDatabaseRefusesASecondOpenTicketForANumber(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := moduleinfra.NewRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	businessID := now.UnixMicro()
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM pager_tickets WHERE business_id IN (?, ?)", businessID, businessID+1)
	})
	ticket := func(businessID int64, token string) domain.Ticket {
		return domain.Ticket{BusinessID: businessID, Number: "42", Source: domain.SourceStaff, PublicToken: token + fmt.Sprint(businessID), CreatedAt: now}
	}

	first, err := repo.CreateTicket(ctx, ticket(businessID, "first"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := repo.CreateTicket(ctx, ticket(businessID, "second")); !errors.Is(err, domain.ErrNumberTaken) {
		t.Fatalf("second open ticket for the number: err = %v, want ErrNumberTaken", err)
	}
	// Ready is still open.
	if err := repo.MarkReady(ctx, businessID, first.ID, now, domain.OpenSince(now)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateTicket(ctx, ticket(businessID, "third")); !errors.Is(err, domain.ErrNumberTaken) {
		t.Fatalf("open ticket next to a ready one: err = %v, want ErrNumberTaken", err)
	}
	// Another business is free to use the number, and so is this one once the
	// ticket is collected.
	if _, err := repo.CreateTicket(ctx, ticket(businessID+1, "other")); err != nil {
		t.Fatalf("same number at another business: %v", err)
	}
	if err := repo.Collect(ctx, businessID, first.ID, domain.OpenSince(now)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateTicket(ctx, ticket(businessID, "fourth")); err != nil {
		t.Fatalf("number of a collected ticket: %v", err)
	}
}

func TestPublicRepliesNeverCarryTheTicketID(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		staffTicket := h.create(h.a, "42")
		id := fmt.Sprint(staffTicket["id"])

		keys := func(r reply) []string {
			out := make([]string, 0, len(r.body))
			for key := range r.body {
				out = append(out, key)
			}
			sort.Strings(out)
			return out
		}
		allowed := func(name string, r reply, fields ...string) {
			h.Helper()
			for _, key := range keys(r) {
				if !slices.Contains(fields, key) {
					h.Errorf("%s reply carries %q: %v", name, key, r.body)
				}
			}
		}

		joined := h.want(h.join(h.a, "42"), http.StatusOK, "")
		allowed("join", joined, "public_token", "number", "status")
		token := joined.text("public_token")
		// 32 base64url characters carry 192 random bits.
		if len(token) != 32 {
			h.Fatalf("token %q is not 32 characters", token)
		}
		other := h.want(h.join(h.a, "43"), http.StatusOK, "").text("public_token")
		if other == token {
			h.Fatal("two tickets share a token")
		}

		followed := h.want(h.publicTicket(token), http.StatusOK, "")
		allowed("ticket", followed, "number", "status", "venue_name", "ahead", "ready_at", "reminded_at", "acknowledged_at", "server_time")
		if followed.text("venue_name") != "Cedar Café" || followed.text("server_time") != "2026-11-19T12:00:00.000Z" {
			h.Fatalf("ticket = %v", followed.body)
		}
		allowed("venue", h.call(nil, http.MethodGet, "/pager/v1/public/venues/"+h.a.slug, ""), "name", "logo_url", "enabled", "strict")

		// The id is not a way in, and neither is a token nobody was given.
		h.want(h.publicTicket(id), http.StatusNotFound, "ticket_not_found")
		h.want(h.publicTicket(strings.Repeat("A", 32)), http.StatusNotFound, "ticket_not_found")
	})
}

func TestStaffMovesATicketThroughItsStatuses(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		token := h.want(h.join(h.a, "42"), http.StatusOK, "").text("public_token")
		id := h.open(h.a)[0]["id"]
		acknowledge := func() reply {
			return h.call(nil, http.MethodPost, "/pager/v1/public/tickets/"+token+"/acknowledge", "")
		}

		// Nothing to remind, undo or silence while the order is being made.
		h.want(h.move(h.a.member, id, "remind"), http.StatusConflict, "invalid_transition")
		h.want(h.move(h.a.member, id, "undo"), http.StatusConflict, "invalid_transition")
		h.want(acknowledge(), http.StatusOK, "")
		if early := h.publicTicket(token); early.text("acknowledged_at") != "" {
			h.Fatalf("a waiting ticket was acknowledged: %v", early.body)
		}

		h.clock.advance(time.Minute)
		ready := h.want(h.move(h.a.member, id, "ready"), http.StatusOK, "").ticket()
		if ready["status"] != domain.StatusReady || ready["ready_at"] != "2026-11-19T12:01:00.000Z" {
			h.Fatalf("ready = %v", ready)
		}
		h.want(h.move(h.a.member, id, "ready"), http.StatusConflict, "invalid_transition")

		// Silencing the bell is not collecting the order.
		h.clock.advance(time.Second)
		h.want(acknowledge(), http.StatusOK, "")
		silenced := h.want(h.publicTicket(token), http.StatusOK, "")
		if silenced.text("status") != domain.StatusReady || silenced.text("acknowledged_at") != "2026-11-19T12:01:01.000Z" {
			h.Fatalf("after acknowledge = %v", silenced.body)
		}

		// Calling again rings every phone again.
		h.clock.advance(time.Second)
		reminded := h.want(h.move(h.a.member, id, "remind"), http.StatusOK, "").ticket()
		if reminded["reminded_at"] != "2026-11-19T12:01:02.000Z" || reminded["acknowledged_at"] != nil {
			h.Fatalf("after remind = %v", reminded)
		}

		undone := h.want(h.move(h.a.member, id, "undo"), http.StatusOK, "").ticket()
		if undone["status"] != domain.StatusWaiting || undone["ready_at"] != nil || undone["reminded_at"] != nil {
			h.Fatalf("after undo = %v", undone)
		}

		collected := h.want(h.move(h.a.member, id, "collect"), http.StatusOK, "").ticket()
		if collected["status"] != domain.StatusCollected {
			h.Fatalf("after collect = %v", collected)
		}
		for _, action := range []string{"ready", "remind", "undo", "collect"} {
			h.want(h.move(h.a.member, id, action), http.StatusConflict, "invalid_transition")
		}
		if h.want(h.publicTicket(token), http.StatusOK, "").text("status") != domain.StatusCollected {
			h.Fatal("the phone was not told the order was collected")
		}

		// Collecting frees the number for the next order.
		if next := h.create(h.a, "42"); next["id"] == id {
			h.Fatal("the collected ticket was handed out again")
		}
		h.want(h.move(h.a.member, 999999999, "ready"), http.StatusNotFound, "ticket_not_found")
	})
}

func TestAheadCountsTheWaitingTicketsEnteredEarlier(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		// Another venue's queue is not this venue's queue.
		h.create(h.b, "1")
		var tokens []string
		for _, number := range []string{"10", "11", "12"} {
			tokens = append(tokens, h.want(h.join(h.a, number), http.StatusOK, "").text("public_token"))
		}
		ahead := func(token string) any { return h.want(h.publicTicket(token), http.StatusOK, "").body["ahead"] }

		// A count of zero is left out of the reply.
		if ahead(tokens[0]) != nil || ahead(tokens[2]) != float64(2) {
			h.Fatalf("ahead = %v and %v, want none and 2", ahead(tokens[0]), ahead(tokens[2]))
		}
		h.want(h.move(h.a.member, h.open(h.a)[0]["id"], "ready"), http.StatusOK, "")
		if ahead(tokens[2]) != float64(1) || ahead(tokens[0]) != nil {
			h.Fatalf("after the first is ready: ahead = %v and %v, want 1 and none", ahead(tokens[2]), ahead(tokens[0]))
		}
	})
}

func TestOpenTicketsExpireAfterSixHours(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a, false)
		token := h.want(h.join(h.a, "42"), http.StatusOK, "").text("public_token")
		id := h.open(h.a)[0]["id"]

		h.clock.advance(domain.OpenLifetime)
		if h.want(h.publicTicket(token), http.StatusOK, "").text("status") != domain.StatusWaiting || len(h.open(h.a)) != 1 {
			h.Fatal("a ticket exactly six hours old is not open any more")
		}

		h.clock.advance(time.Millisecond)
		if got := h.want(h.publicTicket(token), http.StatusOK, "").text("status"); got != domain.StatusExpired {
			h.Fatalf("status past six hours = %q, want expired", got)
		}
		if open := h.open(h.a); len(open) != 0 {
			h.Fatalf("an expired ticket is still on the board: %v", open)
		}
		for _, action := range []string{"ready", "collect"} {
			h.want(h.move(h.a.member, id, action), http.StatusConflict, "invalid_transition")
		}

		// Tomorrow's order 42 is a new ticket; yesterday's phone stays expired.
		next := h.want(h.join(h.a, "42"), http.StatusOK, "")
		if next.text("public_token") == token || next.text("status") != domain.StatusWaiting {
			h.Fatalf("join after expiry = %v, want a new waiting ticket", next.body)
		}
		if got := h.want(h.publicTicket(token), http.StatusOK, "").text("status"); got != domain.StatusExpired {
			h.Fatalf("the old ticket reads %q after its number was reused", got)
		}
		if open := h.open(h.a); len(open) != 1 || open[0]["id"] == id {
			h.Fatalf("board after the number was reused = %v", open)
		}
	})
}

func TestRateLimitsComeFromTheContract(t *testing.T) {
	h := newHarness(t, nil)
	h.enable(h.a, false)
	h.enable(h.b, false)
	const phone = "203.0.113.9:5000"
	venuePath := "/pager/v1/public/venues/" + h.a.slug

	for i := range 60 {
		if r := h.callFrom(phone, nil, http.MethodGet, venuePath, ""); r.status != http.StatusOK {
			t.Fatalf("request %d = %d, want 200", i+1, r.status)
		}
	}
	if r := h.callFrom(phone, nil, http.MethodGet, venuePath, ""); r.status != http.StatusTooManyRequests {
		t.Fatalf("61st request = %d, want 429 from @meta(rate_limit, 60/m)", r.status)
	}
	// The budget is per client and per venue, and staff routes carry none.
	if r := h.call(nil, http.MethodGet, venuePath, ""); r.status != http.StatusOK {
		t.Fatalf("another client = %d, want 200", r.status)
	}
	if r := h.callFrom(phone, nil, http.MethodGet, "/pager/v1/public/venues/"+h.b.slug, ""); r.status != http.StatusOK {
		t.Fatalf("same client at another venue = %d, want 200", r.status)
	}
	for range 130 {
		if r := h.callFrom(phone, h.a.member, http.MethodGet, "/pager/v1/tickets", ""); r.status != http.StatusOK {
			t.Fatalf("staff board poll = %d, want 200", r.status)
		}
	}
}

// The anonymous list is written by hand next to a generated route table, so
// it is checked against the mounted routes behind the real bearer gate.
func TestOnlyTheCustomerRoutesAreAnonymous(t *testing.T) {
	h := newHarness(t, nil)
	h.enable(h.a, false)
	token := h.want(h.join(h.a, "42"), http.StatusOK, "").text("public_token")
	policy := serverhttp.NewRouteAccessPolicy()
	policy.RegisterAnonymousRoutes((&Module{}).GetAnonymousRoutes()...)
	gate := serverhttp.GetAuthMiddlewareFunc(policy, nil)(h.mux)

	for route, anonymous := range map[string]bool{
		"GET /pager/v1/public/venues/" + h.a.slug:                 true,
		"POST /pager/v1/public/venues/" + h.a.slug + "/join":      true,
		"GET /pager/v1/public/tickets/" + token:                   true,
		"POST /pager/v1/public/tickets/" + token + "/acknowledge": true,
		"GET /pager/v1/tickets":                                   false,
		"POST /pager/v1/tickets":                                  false,
		"POST /pager/v1/tickets/7/ready":                          false,
		"POST /pager/v1/tickets/7/remind":                         false,
		"POST /pager/v1/tickets/7/undo":                           false,
		"POST /pager/v1/tickets/7/collect":                        false,
		"GET /pager/v1/settings":                                  false,
		"PUT /pager/v1/settings":                                  false,
	} {
		method, target, _ := strings.Cut(route, " ")
		req := httptest.NewRequest(method, target, strings.NewReader(`{"number":"42"}`))
		req.Header.Set("Content-Type", "application/json")
		if _, pattern := h.mux.Handler(req); pattern == "" {
			t.Errorf("%s is not a mounted route", route)
		}
		rec := httptest.NewRecorder()
		gate.ServeHTTP(rec, req)
		if anonymous && rec.Code != http.StatusOK {
			t.Errorf("%s without a session = %d %s, want 200", route, rec.Code, rec.Body.String())
		}
		if !anonymous && rec.Code != http.StatusUnauthorized {
			t.Errorf("%s without a session = %d %s, want 401", route, rec.Code, rec.Body.String())
		}
	}
}

// The venue is the business the public menu resolves for the same slug, read
// through the business module rather than a copy of its query.
func TestVenueIsTheBusinessBehindTheMenuSlug(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := time.Now().UnixMicro()
	slug := fmt.Sprintf("pager-test-%d", businessID)
	if _, err := db.Exec("INSERT INTO businesses (id, name, owner_id, slug) VALUES (?, ?, ?, ?)", businessID, "Pager Test "+slug, businessID, slug); err != nil {
		t.Fatalf("seed business: %v", err)
	}
	t.Cleanup(func() {
		for _, statement := range []string{
			"DELETE FROM pager_tickets WHERE business_id = ?",
			"DELETE FROM pager_settings WHERE business_id = ?",
			"DELETE FROM businesses WHERE id = ?",
		} {
			if _, err := db.Exec(statement, businessID); err != nil {
				t.Errorf("clean up: %v", err)
			}
		}
	})

	module := NewModule(ctx, db, business.NewModule(ctx, db))
	h := &harness{T: t, mux: http.NewServeMux()}
	module.RegisterAPI(h.mux)
	enabled := true
	if _, err := module.service.UpdateSettings(ctx, businessID, &enabled, nil); err != nil {
		t.Fatalf("enable: %v", err)
	}

	// The menu lower-cases and trims the slug; so does the pager.
	found := h.want(h.call(nil, http.MethodGet, "/pager/v1/public/venues/"+strings.ToUpper(slug), ""), http.StatusOK, "")
	if found.text("name") != "Pager Test "+slug {
		t.Fatalf("venue = %v", found.body)
	}
	token := h.want(h.call(nil, http.MethodPost, "/pager/v1/public/venues/"+slug+"/join", `{"number":"42"}`), http.StatusOK, "").text("public_token")
	if got := h.want(h.publicTicket(token), http.StatusOK, "").text("venue_name"); got != "Pager Test "+slug {
		t.Fatalf("venue_name = %q", got)
	}

	// A deactivated business has no pager, whatever its settings say.
	if _, err := db.Exec("UPDATE businesses SET is_active = FALSE WHERE id = ?", businessID); err != nil {
		t.Fatal(err)
	}
	h.want(h.call(nil, http.MethodGet, "/pager/v1/public/venues/"+slug, ""), http.StatusNotFound, "venue_not_found")
	h.want(h.publicTicket(token), http.StatusNotFound, "ticket_not_found")
}
