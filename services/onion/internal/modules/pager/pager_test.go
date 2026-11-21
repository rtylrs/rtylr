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

func TestScanOpensAnUnlabelledTicketTheBoardShowsAtOnce(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)

		scan := h.want(h.scan(h.a), http.StatusOK, "")
		word := wordOf(scan.body)
		if _, number, ok := domain.ParseMatchWord(word); !ok || number != 0 || scan.body["code"] != nil ||
			scan.text("status") != domain.StatusWaiting {
			h.Fatalf("scan = %v", scan.body)
		}
		if color, _ := scan.body["color"].(float64); int(color) != domain.MatchColor(word) {
			h.Fatalf("scan colour = %v, want %d for %s", scan.body["color"], domain.MatchColor(word), word)
		}

		// The very next board poll has the phone, before anyone touches it.
		open := h.open(h.a)
		if len(open) != 1 || wordOf(open[0]) != word || open[0]["label"] != nil ||
			open[0]["status"] != domain.StatusWaiting || open[0]["created_at"] != "2026-11-19T12:00:00.000Z" {
			h.Fatalf("board after a scan = %v", open)
		}
		// The phone shows the word and no queue position until it has a label.
		followed := h.want(h.publicTicket(scan.text("public_token")), http.StatusOK, "")
		if wordOf(followed.body) != word || followed.text("label") != "" || followed.body["ahead"] != nil {
			h.Fatalf("phone before labelling = %v", followed.body)
		}

		// Every phone waiting to be labelled shows its own word.
		words := map[string]bool{word: true}
		for range 40 {
			next := wordOf(h.want(h.scan(h.a), http.StatusOK, "").body)
			if words[next] {
				h.Fatalf("word %s handed to two waiting phones", next)
			}
			words[next] = true
		}
	})
}

// Once every word is on a waiting phone, the next phones get the words again
// with a number, still one phone per word and number.
func TestScanNumbersTheWordsOnceEveryWordIsTaken(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		seen := map[string]bool{}
		for i := range len(domain.MatchWords) + 10 {
			word := wordOf(h.want(h.scan(h.a), http.StatusOK, "").body)
			_, number, ok := domain.ParseMatchWord(word)
			wantNumber := 0
			if i >= len(domain.MatchWords) {
				wantNumber = 2
			}
			if !ok || seen[word] || number != wantNumber {
				h.Fatalf("scan %d shows %q (number %d), want number %d once and only once", i+1, word, number, wantNumber)
			}
			seen[word] = true
		}

		// Labelling a phone frees its word for the next phone, without a number.
		labelledWord := ""
		for _, ticket := range h.open(h.a) {
			if word := wordOf(ticket); !strings.ContainsAny(word, "0123456789") {
				labelledWord = word
				h.want(h.label(h.a.member, ticket["id"], "42"), http.StatusOK, "")
				break
			}
		}
		if next := wordOf(h.want(h.scan(h.a), http.StatusOK, "").body); next != labelledWord {
			h.Fatalf("after labelling %s the next phone shows %s", labelledWord, next)
		}
	})
}

// Phones that scanned before the words keep their two character code until
// they are labelled or expire.
func TestTicketsFromBeforeTheWordsStillShowTheirCode(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		token := h.legacyTicket(h.a, "A7")

		open := h.open(h.a)
		color, _ := open[0]["color"].(float64)
		if len(open) != 1 || open[0]["code"] != "A7" || open[0]["word"] != nil || int(color) != domain.MatchColor("A7") {
			h.Fatalf("board with an old ticket = %v", open)
		}
		phone := h.want(h.publicTicket(token), http.StatusOK, "")
		if phone.text("code") != "A7" || phone.body["word"] != nil || phone.text("status") != domain.StatusWaiting {
			h.Fatalf("old phone = %v", phone.body)
		}

		// New phones get words beside it, and it is labelled as before.
		h.scanned(h.a)
		h.want(h.label(h.a.member, open[0]["id"], "42"), http.StatusOK, "")
		if phone := h.want(h.publicTicket(token), http.StatusOK, ""); phone.text("label") != "42" || phone.body["code"] != nil {
			h.Fatalf("old phone after labelling = %v", phone.body)
		}
	})
}

func TestCashierLabelsAPhoneWithANumberOrAName(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		token, ticket := h.scanned(h.a)

		// Either mode, the default: Arabic-Indic digits are a number.
		labelled := h.want(h.label(h.a.member, ticket["id"], " ٤٢ "), http.StatusOK, "").ticket()
		if labelled["label"] != "42" || labelled["label_kind"] != domain.LabelNumber || labelled["word"] != nil {
			h.Fatalf("labelled = %v", labelled)
		}
		phone := h.want(h.publicTicket(token), http.StatusOK, "")
		if phone.text("label") != "42" || phone.text("label_kind") != domain.LabelNumber || phone.body["word"] != nil {
			h.Fatalf("phone after labelling = %v", phone.body)
		}

		// A typo is fixed by labelling again while the order is waiting.
		fixed := h.want(h.label(h.a.member, ticket["id"], "  Sara   Haddad "), http.StatusOK, "").ticket()
		if fixed["label"] != "Sara Haddad" || fixed["label_kind"] != domain.LabelName {
			h.Fatalf("relabelled = %v", fixed)
		}

		// Names repeat; an open number belongs to one ticket.
		_, second := h.scanned(h.a)
		h.want(h.label(h.a.member, second["id"], "Sara Haddad"), http.StatusOK, "")
		_, third := h.scanned(h.a)
		h.want(h.label(h.a.member, third["id"], "17"), http.StatusOK, "")
		_, fourth := h.scanned(h.a)
		h.want(h.label(h.a.member, fourth["id"], "17"), http.StatusConflict, "label_taken")
		// The refused ticket is still waiting for a label, word and all.
		for _, open := range h.open(h.a) {
			if open["id"] == fourth["id"] && (open["label"] != nil || open["word"] == nil) {
				h.Fatalf("a refused label changed the ticket: %v", open)
			}
		}

		for _, bad := range []string{"   ", strings.Repeat("x", 41), "Sara\u0000"} {
			h.want(h.label(h.a.member, fourth["id"], bad), http.StatusBadRequest, "invalid_label")
		}

		// Number mode refuses a name; name mode keeps digits as a name, which
		// may repeat an open number.
		h.labelMode(h.a, domain.LabelModeNumber)
		h.want(h.label(h.a.member, fourth["id"], "Sara"), http.StatusBadRequest, "invalid_label")
		h.want(h.label(h.a.member, fourth["id"], "a18"), http.StatusOK, "")
		h.labelMode(h.a, domain.LabelModeName)
		named := h.want(h.label(h.a.member, fourth["id"], "17"), http.StatusOK, "").ticket()
		if named["label"] != "17" || named["label_kind"] != domain.LabelName {
			h.Fatalf("name mode label = %v", named)
		}

		// Once ready, the label stays as it is.
		h.want(h.move(h.a.member, third["id"], "ready"), http.StatusOK, "")
		h.want(h.label(h.a.member, third["id"], "18"), http.StatusConflict, "invalid_transition")
		h.want(h.label(h.a.member, 999999999, "18"), http.StatusNotFound, "ticket_not_found")
	})
}

func TestOnlyALabelledTicketCanBeReady(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		_, ticket := h.scanned(h.a)
		h.want(h.move(h.a.member, ticket["id"], "ready"), http.StatusConflict, "invalid_transition")
		// A phone scanned by mistake can still be dismissed.
		h.want(h.move(h.a.member, ticket["id"], "collect"), http.StatusOK, "")
		if open := h.open(h.a); len(open) != 0 {
			h.Fatalf("a dismissed phone is still on the board: %v", open)
		}
	})
}

func TestUnlabelledTicketsExpireAfterFifteenMinutes(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		token, ticket := h.scanned(h.a)
		_, kept := h.labelled(h.a, "42")

		h.clock.advance(domain.UnlabelledLifetime)
		if len(h.open(h.a)) != 2 || h.want(h.publicTicket(token), http.StatusOK, "").text("status") != domain.StatusWaiting {
			h.Fatal("an unlabelled ticket exactly 15 minutes old is not open any more")
		}

		h.clock.advance(time.Millisecond)
		if got := h.want(h.publicTicket(token), http.StatusOK, "").text("status"); got != domain.StatusExpired {
			h.Fatalf("unlabelled status past 15 minutes = %q, want expired", got)
		}
		if open := h.open(h.a); len(open) != 1 || open[0]["id"] != kept {
			h.Fatalf("board past 15 minutes = %v, want only the labelled order", open)
		}
		h.want(h.label(h.a.member, ticket["id"], "43"), http.StatusConflict, "invalid_transition")
		h.want(h.move(h.a.member, ticket["id"], "collect"), http.StatusConflict, "invalid_transition")

		// Its word is free for the next phone, and its phone stays expired.
		h.want(h.scan(h.a), http.StatusOK, "")
		if got := h.want(h.publicTicket(token), http.StatusOK, "").text("status"); got != domain.StatusExpired {
			h.Fatalf("the expired phone reads %q after the next scan", got)
		}
	})
}

func TestScanCapsTheOpenUnlabelledTickets(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		h.enable(h.b)
		for i := range domain.MaxOpenUnlabelledTickets {
			if r := h.scan(h.a); r.status != http.StatusOK {
				h.Fatalf("scan %d = %d %v", i+1, r.status, r.body)
			}
		}
		h.want(h.scan(h.a), http.StatusTooManyRequests, "too_many_open_tickets")
		// Another venue has its own room.
		h.want(h.scan(h.b), http.StatusOK, "")

		// Labelling a phone makes room for the next.
		h.want(h.label(h.a.member, h.open(h.a)[0]["id"], "1"), http.StatusOK, "")
		h.want(h.scan(h.a), http.StatusOK, "")
		h.want(h.scan(h.a), http.StatusTooManyRequests, "too_many_open_tickets")

		// So does time.
		h.clock.advance(domain.UnlabelledLifetime + time.Second)
		h.want(h.scan(h.a), http.StatusOK, "")
	})
}

func TestPublicRoutesAnswerOnlyForAnEnabledVenue(t *testing.T) {
	eachBackend(t, func(h *harness) {
		venuePath := "/pager/v1/public/venues/" + h.a.slug
		h.want(h.call(nil, http.MethodGet, venuePath, ""), http.StatusNotFound, "venue_not_found")
		h.want(h.scan(h.a), http.StatusNotFound, "venue_not_found")
		h.want(h.call(nil, http.MethodGet, "/pager/v1/public/venues/no-such-place", ""), http.StatusNotFound, "venue_not_found")
		h.want(h.call(nil, http.MethodPost, "/pager/v1/public/venues/no-such-place/scan", ""), http.StatusNotFound, "venue_not_found")

		h.enable(h.a)
		found := h.want(h.call(nil, http.MethodGet, venuePath, ""), http.StatusOK, "")
		if found.text("name") != "Cedar Café" || found.body["enabled"] != true {
			h.Fatalf("venue = %v", found.body)
		}
		token := h.want(h.scan(h.a), http.StatusOK, "").text("public_token")
		h.want(h.publicTicket(token), http.StatusOK, "")

		// Switching the pager off hides the tickets already handed out.
		h.want(h.call(h.a.owner, http.MethodPut, "/pager/v1/settings", `{"enabled":false}`), http.StatusOK, "")
		h.want(h.publicTicket(token), http.StatusNotFound, "ticket_not_found")
		h.want(h.call(nil, http.MethodPost, "/pager/v1/public/tickets/"+token+"/acknowledge", ""), http.StatusNotFound, "ticket_not_found")
		if open := h.open(h.a); len(open) != 1 {
			h.Fatalf("board after the pager was switched off = %v", open)
		}
	})
}

// Business B's staff see and change nothing of business A's, and the business
// comes from who is signed in, never from the request.
func TestStaffRoutesAreScopedToTheCallersBusiness(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		h.enable(h.b)
		_, oursID := h.labelled(h.a, "5")
		// The same number is open at both venues at once.
		_, theirsID := h.labelled(h.b, "5")
		_, unlabelled := h.scanned(h.a)
		ours := map[string]any{"id": oursID}
		theirs := map[string]any{"id": theirsID}

		ids := func(tickets []map[string]any) []any {
			out := make([]any, 0, len(tickets))
			for _, ticket := range tickets {
				out = append(out, ticket["id"])
			}
			return out
		}
		testutil.AssertTenantIsolation(h, ids(h.open(h.b)), []any{theirs["id"]}, []any{ours["id"], unlabelled["id"]})
		testutil.AssertTenantIsolation(h, ids(h.open(h.a)), []any{ours["id"], unlabelled["id"]}, []any{theirs["id"]})

		for _, action := range []string{"ready", "remind", "undo", "collect"} {
			h.want(h.move(h.b.member, ours["id"], action), http.StatusNotFound, "ticket_not_found")
		}
		// B's cashier cannot label a phone that scanned at A.
		h.want(h.label(h.b.member, unlabelled["id"], "9"), http.StatusNotFound, "ticket_not_found")
		h.want(h.label(h.b.member, ours["id"], "9"), http.StatusNotFound, "ticket_not_found")
		open := h.open(h.a)
		if len(open) != 2 || open[0]["status"] != domain.StatusWaiting || open[0]["label"] != "5" || open[1]["label"] != nil {
			h.Fatalf("business A's tickets after business B's attempts = %v", open)
		}

		// B's settings are B's alone.
		h.labelMode(h.a, domain.LabelModeName)
		if settings := h.call(h.b.member, http.MethodGet, "/pager/v1/settings", ""); settings.text("label_mode") != domain.LabelModeEither {
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
			if r := h.label(user, unlabelled["id"], "9"); r.status != http.StatusForbidden {
				h.Fatalf("label without a membership = %d %v, want 403", r.status, r.body)
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

		if r := h.call(h.a.member, http.MethodPut, "/pager/v1/settings", `{"label_mode":"name"}`); r.status != http.StatusForbidden {
			h.Fatalf("member label mode update = %d %v, want 403", r.status, r.body)
		}
		if r := h.want(h.call(h.a.member, http.MethodGet, "/pager/v1/settings", ""), http.StatusOK, ""); r.text("label_mode") != domain.LabelModeEither {
			h.Fatalf("default settings = %v, want label mode either", r.body)
		}

		h.enable(h.a)
		// A setting left out of the request keeps its value.
		r := h.want(h.call(h.a.owner, http.MethodPut, "/pager/v1/settings", `{"label_mode":"number"}`), http.StatusOK, "")
		if r.body["enabled"] != true || r.text("label_mode") != domain.LabelModeNumber {
			h.Fatalf("settings after a partial update = %v, want enabled in number mode", r.body)
		}
		h.want(h.call(h.a.owner, http.MethodPut, "/pager/v1/settings", `{"label_mode":"strict"}`), http.StatusBadRequest, "invalid_label_mode")
		if r := h.call(h.a.member, http.MethodGet, "/pager/v1/settings", ""); r.text("label_mode") != domain.LabelModeNumber {
			h.Fatalf("a refused mode took effect: %v", r.body)
		}
	})
}

// A rush of phones scanning at the same moment each get their own word, and
// two cashiers typing the same number on two phones leave it on one ticket.
func TestConcurrentScansAndLabelsKeepCodesAndNumbersUnique(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		const phones = 60

		codes := make([]string, phones)
		statuses := make([]int, phones)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := range phones {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				r := h.scan(h.a)
				statuses[i], codes[i] = r.status, wordOf(r.body)
			}()
		}
		close(start)
		wg.Wait()
		seen := map[string]bool{}
		for i, code := range codes {
			if statuses[i] != http.StatusOK || code == "" || seen[code] {
				h.Fatalf("scan %d = %d %q; codes %v", i, statuses[i], code, codes)
			}
			seen[code] = true
		}

		open := h.open(h.a)
		if len(open) != phones {
			h.Fatalf("%d tickets on the board for %d phones", len(open), phones)
		}
		for round := range 10 {
			first, second := open[2*round]["id"], open[2*round+1]["id"]
			number := fmt.Sprintf("R%d", round)
			results := make([]reply, 2)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i, id := range []any{first, second} {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					results[i] = h.label(h.a.member, id, number)
				}()
			}
			close(start)
			wg.Wait()
			ok := 0
			for _, r := range results {
				switch {
				case r.status == http.StatusOK:
					ok++
				case r.code() != "label_taken":
					h.Fatalf("round %d: %d %v", round, r.status, r.body)
				}
			}
			if ok != 1 {
				h.Fatalf("round %d: %d tickets took number %s", round, ok, number)
			}
		}
	})
}

// The rules live in the database, not only in the service: a second open
// unlabelled ticket with a word, or a second open ticket with a number, is
// refused even when written directly.
func TestDatabaseRefusesDuplicateOpenCodesAndNumbers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := moduleinfra.NewRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	openSince, unlabelledSince := domain.OpenSince(now), domain.UnlabelledSince(now)
	businessID := now.UnixMicro()
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM pager_tickets WHERE business_id IN (?, ?)", businessID, businessID+1)
	})
	seq := 0
	scan := func(businessID int64, word string) (domain.Ticket, error) {
		seq++
		return repo.CreateTicket(ctx, domain.Ticket{BusinessID: businessID, Code: word, PublicToken: fmt.Sprintf("t%d-%d", businessID, seq), CreatedAt: now})
	}

	first, err := scan(businessID, "panda")
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if _, err := scan(businessID, "panda"); !errors.Is(err, domain.ErrCodeTaken) {
		t.Fatalf("second open ticket with the word: err = %v, want ErrCodeTaken", err)
	}
	if _, err := scan(businessID, "panda2"); err != nil {
		t.Fatalf("the word with a number: %v", err)
	}
	if _, err := scan(businessID, "panda2"); !errors.Is(err, domain.ErrCodeTaken) {
		t.Fatalf("second open ticket with the word and number: err = %v, want ErrCodeTaken", err)
	}
	if _, err := scan(businessID+1, "panda"); err != nil {
		t.Fatalf("same word at another business: %v", err)
	}
	// Labelling frees the word.
	if err := repo.Label(ctx, businessID, first.ID, "42", domain.LabelNumber, openSince, unlabelledSince); err != nil {
		t.Fatal(err)
	}
	second, err := scan(businessID, "panda")
	if err != nil {
		t.Fatalf("word of a labelled ticket: %v", err)
	}

	if err := repo.Label(ctx, businessID, second.ID, "42", domain.LabelNumber, openSince, unlabelledSince); !errors.Is(err, domain.ErrLabelTaken) {
		t.Fatalf("second open ticket with the number: err = %v, want ErrLabelTaken", err)
	}
	if err := repo.Label(ctx, businessID, second.ID, "42", domain.LabelName, openSince, unlabelledSince); err != nil {
		t.Fatalf("a name equal to an open number: %v", err)
	}
	// Ready is still open; collected frees the number.
	if err := repo.MarkReady(ctx, businessID, first.ID, now, openSince); err != nil {
		t.Fatal(err)
	}
	third, err := scan(businessID, "lion")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Label(ctx, businessID, third.ID, "42", domain.LabelNumber, openSince, unlabelledSince); !errors.Is(err, domain.ErrLabelTaken) {
		t.Fatalf("number of a ready ticket: err = %v, want ErrLabelTaken", err)
	}
	if err := repo.Collect(ctx, businessID, first.ID, openSince, unlabelledSince); err != nil {
		t.Fatal(err)
	}
	if err := repo.Label(ctx, businessID, third.ID, "42", domain.LabelNumber, openSince, unlabelledSince); err != nil {
		t.Fatalf("number of a collected ticket: %v", err)
	}
}

func TestPublicRepliesNeverCarryTheTicketID(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)

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

		scanned := h.want(h.scan(h.a), http.StatusOK, "")
		allowed("scan", scanned, "public_token", "word", "word_number", "color", "status")
		token := scanned.text("public_token")
		// 32 base64url characters carry 192 random bits.
		if len(token) != 32 {
			h.Fatalf("token %q is not 32 characters", token)
		}
		other := h.want(h.scan(h.a), http.StatusOK, "").text("public_token")
		if other == token {
			h.Fatal("two tickets share a token")
		}
		id := fmt.Sprint(h.open(h.a)[0]["id"])
		h.want(h.label(h.a.member, id, "42"), http.StatusOK, "")

		followed := h.want(h.publicTicket(token), http.StatusOK, "")
		allowed("ticket", followed, "label", "label_kind", "word", "word_number", "code", "color", "status", "venue_name", "ahead", "ready_at", "reminded_at", "acknowledged_at", "server_time")
		if followed.text("venue_name") != "Cedar Café" || followed.text("server_time") != "2026-11-19T12:00:00.000Z" {
			h.Fatalf("ticket = %v", followed.body)
		}
		allowed("venue", h.call(nil, http.MethodGet, "/pager/v1/public/venues/"+h.a.slug, ""), "name", "logo_url", "enabled")

		// The id is not a way in, and neither is a token nobody was given.
		h.want(h.publicTicket(id), http.StatusNotFound, "ticket_not_found")
		h.want(h.publicTicket(strings.Repeat("A", 32)), http.StatusNotFound, "ticket_not_found")
	})
}

func TestStaffMovesATicketThroughItsStatuses(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		token, id := h.labelled(h.a, "42")
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
		_, next := h.scanned(h.a)
		h.want(h.label(h.a.member, next["id"], "42"), http.StatusOK, "")
		h.want(h.move(h.a.member, 999999999, "ready"), http.StatusNotFound, "ticket_not_found")
	})
}

func TestAheadCountsTheWaitingTicketsEnteredEarlier(t *testing.T) {
	eachBackend(t, func(h *harness) {
		h.enable(h.a)
		h.enable(h.b)
		// Another venue's queue is not this venue's queue, and a phone the
		// cashier has not labelled yet is not in the queue at all.
		h.labelled(h.b, "1")
		var tokens []string
		for i, number := range []string{"10", "11", "12"} {
			token, _ := h.labelled(h.a, number)
			tokens = append(tokens, token)
			if i == 0 {
				h.scanned(h.a)
			}
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
		h.enable(h.a)
		token, id := h.labelled(h.a, "42")

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
		nextToken, _ := h.labelled(h.a, "42")
		if nextToken == token {
			h.Fatal("tomorrow's phone got yesterday's token")
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
	h.enable(h.a)
	h.enable(h.b)
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
	h.enable(h.a)
	token := h.want(h.scan(h.a), http.StatusOK, "").text("public_token")
	policy := serverhttp.NewRouteAccessPolicy()
	policy.RegisterAnonymousRoutes((&Module{}).GetAnonymousRoutes()...)
	gate := serverhttp.GetAuthMiddlewareFunc(policy, nil)(h.mux)

	for route, anonymous := range map[string]bool{
		"GET /pager/v1/public/venues/" + h.a.slug:                 true,
		"POST /pager/v1/public/venues/" + h.a.slug + "/scan":      true,
		"GET /pager/v1/public/tickets/" + token:                   true,
		"POST /pager/v1/public/tickets/" + token + "/acknowledge": true,
		"GET /pager/v1/public/push-key":                           true,
		"POST /pager/v1/public/tickets/" + token + "/push":        true,
		"GET /pager/v1/tickets":                                   false,
		"POST /pager/v1/tickets/7/label":                          false,
		"POST /pager/v1/tickets/7/ready":                          false,
		"POST /pager/v1/tickets/7/remind":                         false,
		"POST /pager/v1/tickets/7/undo":                           false,
		"POST /pager/v1/tickets/7/collect":                        false,
		"GET /pager/v1/settings":                                  false,
		"PUT /pager/v1/settings":                                  false,
	} {
		method, target, _ := strings.Cut(route, " ")
		req := httptest.NewRequest(method, target, strings.NewReader(`{"label":"42",`+validSubscriptionFields+`}`))
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
	token := h.want(h.call(nil, http.MethodPost, "/pager/v1/public/venues/"+slug+"/scan", ""), http.StatusOK, "").text("public_token")
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
