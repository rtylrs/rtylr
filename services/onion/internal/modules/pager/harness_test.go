package pager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/pager/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/testutil"
)

// eachBackend runs a test against the in-memory repository and, when DB_HOST
// points at a migrated database, against the real queries too.
func eachBackend(t *testing.T, run func(h *harness)) {
	t.Helper()
	t.Run("memory", func(t *testing.T) {
		run(newHarness(t, nil))
	})
	t.Run("mysql", func(t *testing.T) {
		run(newHarness(t, testutil.OpenDB(t)))
	})
}

// venue is a business in a test, with a signed-in owner and a plain member.
type venue struct {
	businessID int64
	slug       string
	owner      *authPkg.AuthUser
	member     *authPkg.AuthUser
}

// harness is the module served over HTTP for one test, which it embeds.
type harness struct {
	*testing.T
	mux        *http.ServeMux
	clock      *testClock
	businesses *fakeBusinesses
	db         *sql.DB
	requests   atomic.Int64
	a, b       venue
}

var harnessSeq atomic.Int64

// newHarness serves the module over HTTP for two businesses nobody else uses.
func newHarness(t *testing.T, db *sql.DB) *harness {
	t.Helper()
	h := &harness{
		T:          t,
		mux:        http.NewServeMux(),
		clock:      &testClock{now: time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)},
		businesses: &fakeBusinesses{bySlug: map[string]int64{}, names: map[int64]string{}},
		db:         db,
	}
	base := time.Now().UnixMicro() + harnessSeq.Add(10)
	members := fakeScopes{}
	h.a = h.newVenue(base, "cedar-cafe", "Cedar Café", members)
	h.b = h.newVenue(base+1, "olive-grill", "Olive Grill", members)
	// After testutil.OpenDB, so staff requests resolve against these members.
	t.Cleanup(tenant.SetScopeResolver(members))

	var service *application.Service
	if db == nil {
		service = application.NewService(newMemoryRepository(), venues{h.businesses}, h.clock.Now)
	} else {
		service = application.NewService(moduleinfra.NewRepository(db), venues{h.businesses}, h.clock.Now)
		t.Cleanup(func() {
			for _, statement := range []string{
				"DELETE FROM pager_tickets WHERE business_id IN (?, ?)",
				"DELETE FROM pager_settings WHERE business_id IN (?, ?)",
			} {
				if _, err := db.Exec(statement, h.a.businessID, h.b.businessID); err != nil {
					t.Errorf("clean up: %v", err)
				}
			}
		})
	}
	(&Module{service: service}).RegisterAPI(h.mux)
	return h
}

func (h *harness) newVenue(businessID int64, slug, name string, members fakeScopes) venue {
	h.businesses.bySlug[slug] = businessID
	h.businesses.names[businessID] = name
	owner := &authPkg.AuthUser{UserID: businessID*10 + 1, BusinessID: businessID, EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}}
	member := &authPkg.AuthUser{UserID: businessID*10 + 2, BusinessID: businessID, EffectiveAccess: &authorization.EffectiveAccess{}}
	members[owner.UserID], members[member.UserID] = businessID, businessID
	return venue{businessID: businessID, slug: slug, owner: owner, member: member}
}

type reply struct {
	status int
	body   map[string]any
}

// code is the stable error code of a failed reply.
func (r reply) code() string {
	code, _ := r.body["code"].(string)
	return code
}

func (r reply) text(key string) string {
	value, _ := r.body[key].(string)
	return value
}

// ticket is the ticket a staff reply carries.
func (r reply) ticket() map[string]any {
	ticket, _ := r.body["ticket"].(map[string]any)
	return ticket
}

// call sends one request. user is the signed-in caller, or nil for a customer.
// Every request comes from its own address so rate limits stay out of the way.
func (h *harness) call(user *authPkg.AuthUser, method, path, body string) reply {
	h.Helper()
	n := h.requests.Add(1)
	return h.callFrom(fmt.Sprintf("10.%d.%d.%d:4000", n>>16&255, n>>8&255, n&255), user, method, path, body)
}

func (h *harness) callFrom(remoteAddr string, user *authPkg.AuthUser, method, path, body string) reply {
	h.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	if user != nil {
		req = req.WithContext(authPkg.WithAuthUser(req.Context(), user))
	}
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	out := reply{status: rec.Code}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out.body); err != nil {
			h.Fatalf("%s %s answered %d with a body that is not JSON: %s", method, path, rec.Code, rec.Body.String())
		}
	}
	return out
}

// want fails the test unless the reply has the status and, for an error, the code.
func (h *harness) want(r reply, status int, code string) reply {
	h.Helper()
	if r.status != status || r.code() != code {
		h.Fatalf("got %d %q %v, want %d %q", r.status, r.code(), r.body, status, code)
	}
	return r
}

// enable turns a venue's pager on.
func (h *harness) enable(v venue, strict bool) {
	h.Helper()
	h.want(h.call(v.owner, http.MethodPut, "/pager/v1/settings", fmt.Sprintf(`{"enabled":true,"strict":%t}`, strict)), http.StatusOK, "")
}

func (h *harness) join(v venue, number string) reply {
	h.Helper()
	return h.call(nil, http.MethodPost, "/pager/v1/public/venues/"+v.slug+"/join", h.numberBody(number))
}

// create enters a number from the staff board and returns the ticket.
func (h *harness) create(v venue, number string) map[string]any {
	h.Helper()
	return h.want(h.call(v.member, http.MethodPost, "/pager/v1/tickets", h.numberBody(number)), http.StatusOK, "").ticket()
}

func (h *harness) numberBody(number string) string {
	h.Helper()
	body, err := json.Marshal(map[string]string{"number": number})
	if err != nil {
		h.Fatal(err)
	}
	return string(body)
}

func (h *harness) move(user *authPkg.AuthUser, id any, action string) reply {
	h.Helper()
	return h.call(user, http.MethodPost, fmt.Sprintf("/pager/v1/tickets/%v/%s", id, action), "")
}

func (h *harness) publicTicket(token string) reply {
	h.Helper()
	return h.call(nil, http.MethodGet, "/pager/v1/public/tickets/"+token, "")
}

// open lists a venue's open tickets as its staff see them.
func (h *harness) open(v venue) []map[string]any {
	h.Helper()
	r := h.want(h.call(v.member, http.MethodGet, "/pager/v1/tickets", ""), http.StatusOK, "")
	list, _ := r.body["tickets"].([]any)
	out := make([]map[string]any, 0, len(list))
	for _, item := range list {
		ticket, _ := item.(map[string]any)
		out = append(out, ticket)
	}
	return out
}

type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// fakeBusinesses stands in for the business module, with its slug handling.
type fakeBusinesses struct {
	bySlug map[string]int64
	names  map[int64]string
}

func (f *fakeBusinesses) PublicBusinessBySlug(_ context.Context, slug string) (int64, string, string, error) {
	id, ok := f.bySlug[strings.ToLower(strings.TrimSpace(slug))]
	if !ok {
		return 0, "", "", sql.ErrNoRows
	}
	return id, f.names[id], "", nil
}

func (f *fakeBusinesses) PublicBusinessByID(_ context.Context, businessID int64) (string, string, error) {
	name, ok := f.names[businessID]
	if !ok {
		return "", "", sql.ErrNoRows
	}
	return name, "", nil
}

// fakeScopes maps a user to the one business they belong to.
type fakeScopes map[int64]int64

func (f fakeScopes) ActiveBusinessOwner(context.Context, int64) (int64, error) {
	return 0, errors.New("not used by the pager")
}

func (f fakeScopes) UserBusiness(_ context.Context, businessID, userID int64) (int64, error) {
	if own, ok := f[userID]; ok && own == businessID {
		return businessID, nil
	}
	return 0, errors.New("not a member of this business")
}
