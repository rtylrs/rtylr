# Website Lead Routing and Newsletter Ingestion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every voxire.com form delivers to info@voxire.com and the agreed team addresses through the onion backend, the blog newsletter forms leave Web3Forms, the auto-reply speaks for the team, and the lead badge matches the real budget bands.

**Architecture:** The `websitelead` module in onion gains a `/newsletter` endpoint backed by a new `voxire_newsletter_subscribers` table and reuses its detached-goroutine notification pattern. Recipient lists move from Go literals into `AppConfig`, parsed from two env vars with compiled-in defaults. voxire.com's two newsletter components switch from a browser call to Web3Forms to the existing axios client, and the contact page loses its dead Web3Forms branch.

**Tech Stack:** Go 1.26, chi, sqlc (MySQL, `database/sql`), Resend via `internal/platform/email`, Go `testing`. Next.js 16 pages router, TypeScript strict, axios, pnpm.

**Spec:** `docs/superpowers/specs/2026-09-21-website-leads-routing-design.md` (this worktree)

## Global Constraints

- Two repositories. Onion work happens in `/Users/abedamouneh/Development/onion-worktrees/website-leads` on branch `feat/website-leads-routing`. voxire.com work happens in `/Users/abedamouneh/Development/voxire` on `main`. Re-check `git branch --show-current` before every commit in voxire: another agent has switched that checkout's branch mid-session twice today.
- Commit messages: short imperative subject, no conventional-commit prefix, no em dashes, no `Co-Authored-By`. One logical change per commit.
- No em dashes anywhere, including email copy, comments and SQL comments. Restructure or use a colon.
- Go: `make lint`, `make vet`, `make test`, `make json-tags-check`, `make mutation-validation-check` must pass before the onion branch is pushed. JSON tags are snake_case.
- Onion `.env.example` is the operator index: every new `os.Getenv("X")` in non-test code must be documented there, checked by `python3 scripts/check_env_example.py`.
- voxire: pnpm only for dependency changes (none expected). `npm run typecheck` and `pnpm build` must pass. All user-facing strings go through i18n; no new strings are expected.
- Lead notification default recipients, exact: `info@voxire.com`, `sales@voxire.com`, `mohammad.saab@voxire.com`, `mohamad.homsi@voxire.com`, `amounehabed@gmail.com`. Newsletter notification default: `info@voxire.com`.
- Company WhatsApp is `+961 3 940 708`, digits `9613940708`.
- Rollout order is onion first, voxire second. Do not push the voxire commits until `POST https://api.voxire.com/newsletter` with `{}` returns `400 {"message":"email is required"}`.

---

## File map

Onion, created:

- `internal/platform/config/notify_emails.go`: `ParseEmailList` and the two default lists.
- `internal/platform/config/notify_emails_test.go`
- `internal/modules/websitelead/internal/domain/subscriber.go`: `SubscriberInput`, allowed sources and languages.
- `internal/modules/websitelead/internal/application/budget.go`: `budgetTier`.
- `internal/modules/websitelead/internal/application/budget_test.go`
- `internal/modules/websitelead/internal/application/newsletter.go`: `Subscribe` and its notification.
- `internal/modules/websitelead/internal/application/newsletter_test.go`
- `internal/modules/websitelead/internal/api/newsletter_handler.go`
- `internal/modules/websitelead/internal/api/newsletter_handler_test.go`
- `internal/platform/database/schemas/voxire/newsletter_subscribers.sql`

Onion, modified:

- `internal/platform/config/config.go`: two new `AppConfig` fields, populated in `LoadAndValidate`.
- `.env.example`: two variables under the Email section.
- `internal/modules/websitelead/internal/application/service.go`: `Notify` struct, `NewService` signature, `sendEmail` seam, recipient use, badge refactor, team-voice copy.
- `internal/modules/websitelead/internal/application/service_test.go`: mock gains `CreateSubscriber`, seam-based assertions.
- `internal/modules/websitelead/internal/infrastructure/repository.go`: `CreateSubscriber`.
- `internal/modules/websitelead/sql/schema.sql`, `sql/queries/websitelead.sql`: new table and query. Generated files under `internal/infrastructure/persistence/sqlc/` are regenerated, never hand-edited.
- `internal/modules/websitelead/internal/api/routes.go`: register the newsletter route.
- `internal/modules/websitelead/module.go`: `NewModule` takes `Notify`; `/newsletter` anonymous.
- `cmd/api/main.go:680`: pass the config lists.

voxire, modified:

- `src/components/NewsletterDrop.tsx`, `src/components/NewsletterReportBanner.tsx`: post to the API.
- `pages/contact.tsx`: delete the Web3Forms branch.
- `.env.local.example`: delete the Web3Forms section.

voxire, deleted:

- `src/lib/web3forms.ts`

---

### Task 1: Recipient lists in configuration

**Files:**
- Create: `internal/platform/config/notify_emails.go`
- Create: `internal/platform/config/notify_emails_test.go`
- Modify: `internal/platform/config/config.go:81-233` (struct) and `:296-330` (population)
- Modify: `.env.example:140-142`

**Interfaces:**
- Produces: `config.ParseEmailList(raw string, fallback []string) []string`, `config.DefaultLeadNotifyEmails []string`, `config.DefaultNewsletterNotifyEmails []string`, `AppConfig.LeadNotifyEmails []string`, `AppConfig.NewsletterNotifyEmails []string`.

- [ ] **Step 1: Write the failing test**

`internal/platform/config/notify_emails_test.go`:

```go
package config

import (
	"reflect"
	"testing"
)

func TestParseEmailList(t *testing.T) {
	fallback := []string{"info@voxire.com"}
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"unset uses fallback", "", fallback},
		{"whitespace only uses fallback", " \t ", fallback},
		{"single", "a@voxire.com", []string{"a@voxire.com"}},
		{"trims and lowercases", " A@Voxire.com , b@voxire.com ", []string{"a@voxire.com", "b@voxire.com"}},
		{"drops empties and duplicates", "a@voxire.com,,a@voxire.com,", []string{"a@voxire.com"}},
		{"keeps order of first appearance", "b@voxire.com,a@voxire.com,b@voxire.com", []string{"b@voxire.com", "a@voxire.com"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseEmailList(tc.raw, fallback)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParseEmailList(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseEmailListReturnsCopyOfFallback(t *testing.T) {
	fallback := []string{"info@voxire.com"}
	got := ParseEmailList("", fallback)
	got[0] = "changed"
	if fallback[0] != "info@voxire.com" {
		t.Fatal("fallback slice was mutated through the returned list")
	}
}

func TestDefaultNotifyLists(t *testing.T) {
	wantLead := []string{
		"info@voxire.com",
		"sales@voxire.com",
		"mohammad.saab@voxire.com",
		"mohamad.homsi@voxire.com",
		"amounehabed@gmail.com",
	}
	if !reflect.DeepEqual(DefaultLeadNotifyEmails, wantLead) {
		t.Fatalf("DefaultLeadNotifyEmails = %v", DefaultLeadNotifyEmails)
	}
	if !reflect.DeepEqual(DefaultNewsletterNotifyEmails, []string{"info@voxire.com"}) {
		t.Fatalf("DefaultNewsletterNotifyEmails = %v", DefaultNewsletterNotifyEmails)
	}
}

func TestLoadAndValidateReadsNotifyLists(t *testing.T) {
	setLicenseConfigEnv(t)
	t.Setenv("LEAD_NOTIFY_EMAILS", "x@voxire.com, y@voxire.com")
	t.Setenv("NEWSLETTER_NOTIFY_EMAILS", "")
	cfg, err := LoadAndValidate()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.LeadNotifyEmails, []string{"x@voxire.com", "y@voxire.com"}) {
		t.Fatalf("LeadNotifyEmails = %v", cfg.LeadNotifyEmails)
	}
	if !reflect.DeepEqual(cfg.NewsletterNotifyEmails, DefaultNewsletterNotifyEmails) {
		t.Fatalf("NewsletterNotifyEmails = %v", cfg.NewsletterNotifyEmails)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd /Users/abedamouneh/Development/onion-worktrees/website-leads && go test ./internal/platform/config/ -run 'TestParseEmailList|TestDefaultNotifyLists|TestLoadAndValidateReadsNotifyLists' -v`
Expected: FAIL to compile with `undefined: ParseEmailList`.

- [ ] **Step 3: Write the parser and defaults**

`internal/platform/config/notify_emails.go`:

```go
package config

import "strings"

// DefaultLeadNotifyEmails receives website lead notifications when
// LEAD_NOTIFY_EMAILS is unset. A misconfigured deploy must never mean a lead
// is stored and nobody is told.
var DefaultLeadNotifyEmails = []string{
	"info@voxire.com",
	"sales@voxire.com",
	"mohammad.saab@voxire.com",
	"mohamad.homsi@voxire.com",
	"amounehabed@gmail.com",
}

// DefaultNewsletterNotifyEmails receives newsletter signup notifications when
// NEWSLETTER_NOTIFY_EMAILS is unset. Subscribers are lower priority than
// leads, so only the shared inbox hears about them by default.
var DefaultNewsletterNotifyEmails = []string{"info@voxire.com"}

// ParseEmailList splits a comma-separated list of addresses, trims and
// lower-cases each one, drops empties and duplicates while keeping first
// appearance order, and returns a copy of fallback when nothing survives.
func ParseEmailList(raw string, fallback []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, part := range strings.Split(raw, ",") {
		addr := strings.ToLower(strings.TrimSpace(part))
		if addr == "" {
			continue
		}
		if _, dup := seen[addr]; dup {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	if len(out) == 0 {
		return append([]string(nil), fallback...)
	}
	return out
}
```

In `internal/platform/config/config.go`, add to `AppConfig` directly after the `ResendKey string` field (line 91):

```go
	LeadNotifyEmails       []string
	NewsletterNotifyEmails []string
```

In `LoadAndValidate`, inside the `cfg := &AppConfig{ ... }` literal, directly after the `ResendKey: os.Getenv("RESEND_API_KEY"),` line:

```go
		LeadNotifyEmails:       ParseEmailList(os.Getenv("LEAD_NOTIFY_EMAILS"), DefaultLeadNotifyEmails),
		NewsletterNotifyEmails: ParseEmailList(os.Getenv("NEWSLETTER_NOTIFY_EMAILS"), DefaultNewsletterNotifyEmails),
```

Run `gofmt -l internal/platform/config/` afterwards; struct field alignment will need the whole block reformatted, and `gofmt -w` does it.

- [ ] **Step 4: Document the variables**

In `.env.example`, replace the Email section so it reads:

```
# ---- Email (optional) -----------------------------------------------
RESEND_API_KEY=
# Comma-separated recipients of website lead notifications (/contact, /quote).
# Unset means the compiled-in Voxire team list.
LEAD_NOTIFY_EMAILS=
# Comma-separated recipients of newsletter signup notifications (/newsletter).
# Unset means info@voxire.com.
NEWSLETTER_NOTIFY_EMAILS=
```

- [ ] **Step 5: Run the tests and the env gate**

Run: `go test ./internal/platform/config/ -v 2>&1 | tail -20 && python3 scripts/check_env_example.py`
Expected: all config tests PASS, env check exits 0 with no missing variables.

- [ ] **Step 6: Commit**

```bash
git add internal/platform/config/notify_emails.go internal/platform/config/notify_emails_test.go internal/platform/config/config.go .env.example
git commit -m "Read lead and newsletter notification recipients from configuration"
```

---

### Task 2: Inject recipients into the lead service and make sending testable

**Files:**
- Modify: `internal/modules/websitelead/internal/application/service.go:16-42` (imports, `notifyEmails`, `Service`, `NewService`) and `:224-233` (the `emailPkg.Send` call in `sendNotification`)
- Modify: `internal/modules/websitelead/internal/application/service_test.go`
- Modify: `internal/modules/websitelead/module.go`
- Modify: `cmd/api/main.go:680`

**Interfaces:**
- Consumes: `config.GlobalConfig.LeadNotifyEmails`, `config.GlobalConfig.NewsletterNotifyEmails` from Task 1.
- Produces: `application.Notify{LeadEmails, NewsletterEmails []string}`, `application.NewService(ctx, db, notify Notify) *Service`, package var `sendEmail = emailPkg.Send` (test seam), `websitelead.NewModule(ctx, db, notify application.Notify)`.

- [ ] **Step 1: Write the failing test**

Replace the whole of `internal/modules/websitelead/internal/application/service_test.go` with:

```go
package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

type mockRepo struct {
	createID  int64
	createErr error
	createGot domain.LeadInput
}

func (m *mockRepo) Create(_ context.Context, input domain.LeadInput) (int64, error) {
	m.createGot = input
	return m.createID, m.createErr
}

// captureEmails swaps the sendEmail seam for a recorder and returns a waiter
// that blocks until n sends have happened or the deadline passes.
func captureEmails(t *testing.T, n int) (*[]emailPkg.SendParams, func()) {
	t.Helper()
	var mu sync.Mutex
	var got []emailPkg.SendParams
	done := make(chan struct{}, 16)
	orig := sendEmail
	sendEmail = func(_ context.Context, p emailPkg.SendParams) (string, error) {
		mu.Lock()
		got = append(got, p)
		mu.Unlock()
		done <- struct{}{}
		return "id", nil
	}
	t.Cleanup(func() { sendEmail = orig })
	wait := func() {
		for i := 0; i < n; i++ {
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatalf("expected %d emails, got %d", n, len(got))
			}
		}
	}
	return &got, wait
}

func testNotify() Notify {
	return Notify{
		LeadEmails:       []string{"leads@example.com", "second@example.com"},
		NewsletterEmails: []string{"news@example.com"},
	}
}

func TestSubmit_Delegates(t *testing.T) {
	// A successful submit fires two detached emails (notification and
	// auto-reply). Wait for both so no goroutine outlives the seam swap.
	_, wait := captureEmails(t, 2)
	m := &mockRepo{createID: 7}
	s := &Service{repo: m, notify: testNotify()}
	id, err := s.Submit(context.Background(), domain.LeadInput{Name: "Alice", Email: "alice@example.com"})
	if err != nil || id != 7 {
		t.Errorf("id=%d err=%v", id, err)
	}
	if m.createGot.Name != "Alice" {
		t.Errorf("repo got %+v", m.createGot)
	}
	wait()
}

func TestSubmit_PropagatesError(t *testing.T) {
	captureEmails(t, 0)
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}, notify: testNotify()}
	if _, err := s.Submit(context.Background(), domain.LeadInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestSubmit_NotificationGoesToConfiguredLeadList(t *testing.T) {
	got, wait := captureEmails(t, 2)
	s := &Service{repo: &mockRepo{createID: 42}, notify: testNotify()}
	_, err := s.Submit(context.Background(), domain.LeadInput{
		Name: "Rania", Email: "rania@example.com", Service: "Web Development", Budget: "$8K - $20K",
	})
	if err != nil {
		t.Fatal(err)
	}
	wait()
	var notification *emailPkg.SendParams
	for i := range *got {
		if strings.HasPrefix((*got)[i].Subject, "[Lead #42]") {
			notification = &(*got)[i]
		}
	}
	if notification == nil {
		t.Fatalf("no lead notification among %d emails", len(*got))
	}
	if strings.Join(notification.To, ",") != "leads@example.com,second@example.com" {
		t.Fatalf("notification To = %v", notification.To)
	}
	if notification.ReplyTo != "rania@example.com" {
		t.Fatalf("notification ReplyTo = %q", notification.ReplyTo)
	}
}

func TestAsyncLeadContextDetachesCancellationAndPreservesValues(t *testing.T) {
	type ctxKey struct{}
	key := ctxKey{}
	requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), key, "request-value"))
	cancel()

	taskCtx, taskCancel := asyncLeadContext(requestCtx)
	defer taskCancel()

	if err := taskCtx.Err(); err != nil {
		t.Fatalf("detached task context is already canceled: %v", err)
	}
	if got := taskCtx.Value(key); got != "request-value" {
		t.Fatalf("task context lost request value: %v", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/modules/websitelead/... 2>&1 | head -20`
Expected: FAIL to compile with `undefined: sendEmail`, `undefined: Notify`, `unknown field notify`.

- [ ] **Step 3: Add the seam, the struct and the injection**

In `service.go`, replace lines 16 to 42 (from `// notifyEmails receives` through the end of `NewService`) with:

```go
// Notify holds the recipient lists for the two internal notifications this
// module sends. Both come from configuration; see config.ParseEmailList.
type Notify struct {
	LeadEmails       []string
	NewsletterEmails []string
}

// sendEmail is the outbound seam. Tests replace it to capture what would have
// gone to Resend without sending anything.
var sendEmail = emailPkg.Send

type repository interface {
	Create(ctx context.Context, input domain.LeadInput) (int64, error)
}

type Service struct {
	repo   repository
	notify Notify
}

const asyncLeadTaskTimeout = 10 * time.Second

func asyncLeadContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), asyncLeadTaskTimeout)
}

func NewService(_ context.Context, db *sql.DB, notify Notify) *Service {
	return &Service{
		repo:   infrastructure.NewRepository(db),
		notify: notify,
	}
}
```

In `sendNotification`, change the send call:

```go
	_, emailErr := sendEmail(ctx, emailPkg.SendParams{
		To:      s.notify.LeadEmails,
		Subject: subject,
		HTML:    html,
		From:    "Voxire Leads <noreply@voxire.com>",
		ReplyTo: lead.Email,
	})
```

In `sendAutoReply`, change `emailPkg.Send(ctx, emailPkg.SendParams{` to `sendEmail(ctx, emailPkg.SendParams{`. The rest of that call is edited in Task 4.

Replace `internal/modules/websitelead/module.go` body:

```go
package websitelead

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/websitelead/internal/api"
	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
)

// Notify re-exports the recipient configuration so cmd/api does not import the
// module's internal packages.
type Notify = application.Notify

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, notify Notify) *Module {
	return &Module{
		service: application.NewService(ctx, db, notify),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/quote", "/contact", "/analytics/event"}
}
```

In `cmd/api/main.go` line 680, replace `WebsiteLead: websitelead.NewModule(ctx, dbclient.Pool()),` with:

```go
		WebsiteLead: websitelead.NewModule(ctx, dbclient.Pool(), websitelead.Notify{
			LeadEmails:       platformConfig.GlobalConfig.LeadNotifyEmails,
			NewsletterEmails: platformConfig.GlobalConfig.NewsletterNotifyEmails,
		}),
```

`platformConfig` is already the import alias used on line 681 for the user module.

- [ ] **Step 4: Run the tests**

Run: `go build ./... && go test ./internal/modules/websitelead/... -v 2>&1 | tail -20`
Expected: build succeeds, all websitelead tests PASS including `TestSubmit_NotificationGoesToConfiguredLeadList`.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/websitelead/internal/application/service.go internal/modules/websitelead/internal/application/service_test.go internal/modules/websitelead/module.go cmd/api/main.go
git commit -m "Inject website lead notification recipients from configuration"
```

---

### Task 3: Budget tier matches the real bands

**Files:**
- Create: `internal/modules/websitelead/internal/application/budget.go`
- Create: `internal/modules/websitelead/internal/application/budget_test.go`
- Modify: `internal/modules/websitelead/internal/application/service.go:62-86` (`priorityBadge`)

**Interfaces:**
- Produces: `budgetTier(budget string) string` returning exactly `"HOT"`, `"WARM"` or `"NEW"`.

- [ ] **Step 1: Write the failing test**

`internal/modules/websitelead/internal/application/budget_test.go`:

```go
package application

import "testing"

func TestBudgetTier(t *testing.T) {
	cases := []struct {
		budget string
		want   string
	}{
		// Contact form bands
		{"Under $2K", "NEW"},
		{"$2K - $8K", "WARM"},
		{"$8K - $20K", "HOT"},
		{"$20K+", "HOT"},
		{"Let's figure it out", "NEW"},
		// Quote form bands
		{"Under $2k", "NEW"},
		{"$2k-$5k", "WARM"},
		{"$5k-$15k", "HOT"},
		{"$15k+", "HOT"},
		{"Let's discuss", "NEW"},
		// Free text and edge cases
		{"", "NEW"},
		{"around 10,000 USD", "WARM"},
		{"25,000", "HOT"},
		{"$3k to $6k", "WARM"},
		{"no budget yet", "NEW"},
		{"4k", "NEW"},
	}
	for _, tc := range cases {
		t.Run(tc.budget, func(t *testing.T) {
			if got := budgetTier(tc.budget); got != tc.want {
				t.Fatalf("budgetTier(%q) = %q, want %q", tc.budget, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/modules/websitelead/internal/application/ -run TestBudgetTier -v 2>&1 | head`
Expected: FAIL to compile with `undefined: budgetTier`.

- [ ] **Step 3: Implement the tier function**

`internal/modules/websitelead/internal/application/budget.go`:

```go
package application

import (
	"regexp"
	"strconv"
	"strings"
)

// Budget bands differ between the contact form ("$8K - $20K") and the quote
// form ("$5k-$15k"), and free text arrives too. The tier is decided by the
// largest figure in the string, in thousands of dollars.
const (
	tierHot  = "HOT"
	tierWarm = "WARM"
	tierNew  = "NEW"

	hotFromThousands  = 15
	warmFromThousands = 5
)

// budgetFigure matches "20K", "5k", "15k+", "10,000", "25000".
var budgetFigure = regexp.MustCompile(`(?i)(\d{1,3}(?:,\d{3})+|\d+)\s*(k)?`)

func budgetTier(budget string) string {
	maxThousands := 0.0
	for _, m := range budgetFigure.FindAllStringSubmatch(budget, -1) {
		raw := strings.ReplaceAll(m[1], ",", "")
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		if m[2] != "" {
			// "20K" is twenty thousand.
			if n > maxThousands {
				maxThousands = n
			}
			continue
		}
		if n >= 1000 {
			// "25,000" is twenty five thousand.
			if n/1000 > maxThousands {
				maxThousands = n / 1000
			}
		}
		// A bare small number with no K and no thousands separator is not a
		// budget figure ("within 2 weeks"), so it is ignored.
	}
	switch {
	case maxThousands >= hotFromThousands:
		return tierHot
	case maxThousands >= warmFromThousands:
		return tierWarm
	default:
		return tierNew
	}
}
```

Then in `service.go` replace the whole `priorityBadge` function (the one with the `strings.Contains(budget, "10,000")` chain) with:

```go
// priorityBadge returns an HTML badge based on the budget signal.
func priorityBadge(lead domain.LeadInput) string {
	switch budgetTier(lead.Budget) {
	case tierHot:
		return `<span style="display:inline-block;background:#ef4444;color:#fff;font-size:11px;font-weight:700;padding:3px 10px;border-radius:20px;letter-spacing:0.05em;text-transform:uppercase;">HOT</span>`
	case tierWarm:
		return `<span style="display:inline-block;background:#f59e0b;color:#111;font-size:11px;font-weight:700;padding:3px 10px;border-radius:20px;letter-spacing:0.05em;text-transform:uppercase;">WARM</span>`
	default:
		return `<span style="display:inline-block;background:rgba(99,198,189,0.2);color:#63c6bd;font-size:11px;font-weight:700;padding:3px 10px;border-radius:20px;letter-spacing:0.05em;text-transform:uppercase;">NEW</span>`
	}
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/modules/websitelead/... -v 2>&1 | grep -E "^(=== RUN|--- FAIL|FAIL|ok|PASS)" | tail -30`
Expected: every `TestBudgetTier` subtest PASS, package `ok`.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/websitelead/internal/application/budget.go internal/modules/websitelead/internal/application/budget_test.go internal/modules/websitelead/internal/application/service.go
git commit -m "Derive the lead priority badge from the budget bands the forms offer"
```

---

### Task 4: Team-voice auto-reply and neutral WhatsApp prefill

**Files:**
- Modify: `internal/modules/websitelead/internal/application/service.go`: `sendNotification` WhatsApp button (around line 111) and the whole of `sendAutoReply`.
- Modify: `internal/modules/websitelead/internal/application/service_test.go`: add one test.

**Interfaces:**
- Consumes: `sendEmail` seam and `captureEmails` helper from Task 2.

- [ ] **Step 1: Write the failing test**

Append to `service_test.go`:

```go
func TestSubmit_AutoReplySpeaksForTheTeam(t *testing.T) {
	got, wait := captureEmails(t, 2)
	s := &Service{repo: &mockRepo{createID: 9}, notify: testNotify()}
	if _, err := s.Submit(context.Background(), domain.LeadInput{
		Name: "Rania", Email: "rania@example.com", Phone: "+961 70 123 456",
	}); err != nil {
		t.Fatal(err)
	}
	wait()
	var reply, notification *emailPkg.SendParams
	for i := range *got {
		switch {
		case (*got)[i].Subject == "We received your message - Voxire":
			reply = &(*got)[i]
		case strings.HasPrefix((*got)[i].Subject, "[Lead #9]"):
			notification = &(*got)[i]
		}
	}
	if reply == nil || notification == nil {
		t.Fatalf("expected a reply and a notification, got %d emails", len(*got))
	}
	if reply.From != "Voxire <info@voxire.com>" {
		t.Fatalf("reply From = %q", reply.From)
	}
	if reply.ReplyTo != "info@voxire.com" || strings.Join(reply.To, ",") != "rania@example.com" {
		t.Fatalf("reply routing To=%v ReplyTo=%q", reply.To, reply.ReplyTo)
	}
	for _, banned := range []string{"Abed", "\u2014"} {
		if strings.Contains(reply.HTML, banned) {
			t.Fatalf("reply still contains %q", banned)
		}
		if strings.Contains(notification.HTML, banned) {
			t.Fatalf("notification still contains %q", banned)
		}
	}
	for _, want := range []string{"Thanks, Rania.", "24 hours", "wa.me/9613940708"} {
		if !strings.Contains(reply.HTML, want) {
			t.Fatalf("reply missing %q", want)
		}
	}
	if !strings.Contains(notification.HTML, "this+is+Voxire+following+up") {
		t.Fatal("notification WhatsApp prefill is not the neutral team wording")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/modules/websitelead/internal/application/ -run TestSubmit_AutoReplySpeaksForTheTeam -v 2>&1 | tail -5`
Expected: FAIL on `reply From = "Abed at Voxire <info@voxire.com>"`.

- [ ] **Step 3: Rewrite the copy**

In `sendNotification`, change the WhatsApp quick action:

```go
		waBtn := ctaButton(
			fmt.Sprintf("https://wa.me/%s?text=Hi+%s%%2C+this+is+Voxire+following+up+on+your+request.", strings.TrimPrefix(cleanPhone, "+"), url.QueryEscape(lead.Name)),
			"WhatsApp",
			"#25d366", "#111118",
		)
```

Add `"net/url"` to the imports of `service.go`. Note that `url.QueryEscape` turns spaces into `+`, which is what `wa.me` expects in `text=`.

Also in `sendNotification`, the footer line separates "Voxire lead management" from "voxire.com" with the HTML em dash entity. Change it to a colon: `Voxire lead management: voxire.com`.

Replace the whole `sendAutoReply` function:

```go
// ─── Auto-reply to submitter ──────────────────────────────────────────────────

func (s *Service) sendAutoReply(ctx context.Context, lead domain.LeadInput) {
	var summaryRows strings.Builder
	if lead.Service != "" && lead.Service != "contact" {
		fmt.Fprintf(&summaryRows, `<tr><td style="padding:7px 0;color:rgba(0,0,0,0.45);font-size:13px;width:120px;">Service</td><td style="padding:7px 0;font-size:14px;color:#111;font-weight:500;">%s</td></tr>`,
			lead.Service)
	}
	if lead.Budget != "" {
		fmt.Fprintf(&summaryRows, `<tr><td style="padding:7px 0;color:rgba(0,0,0,0.45);font-size:13px;">Budget</td><td style="padding:7px 0;font-size:14px;color:#111;">%s</td></tr>`,
			lead.Budget)
	}

	summaryBlock := ""
	if summaryRows.Len() > 0 {
		summaryBlock = fmt.Sprintf(`
    <div style="margin:24px 0;padding:16px 20px;background:#f5f5f7;border-radius:10px;">
      <p style="font-size:12px;color:rgba(0,0,0,0.45);margin:0 0 8px 0;text-transform:uppercase;letter-spacing:0.06em;font-weight:500;">Your request summary</p>
      <table style="width:100%%;border-collapse:collapse;">%s</table>
    </div>`, summaryRows.String())
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<body style="margin:0;padding:0;background:#f0f0f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<div style="max-width:560px;margin:0 auto;padding:40px 16px;">

  <div style="background:#fff;border-radius:14px;padding:36px;box-shadow:0 2px 16px rgba(0,0,0,0.07);">

    <p style="font-size:20px;font-weight:700;color:#63c6bd;margin:0 0 28px 0;letter-spacing:-0.01em;">Voxire</p>

    <h2 style="font-size:22px;font-weight:700;color:#111;margin:0 0 12px 0;line-height:1.3;">
      Thanks, %s. We have your request.
    </h2>
    <p style="font-size:15px;color:#444;line-height:1.7;margin:0 0 20px 0;">
      One of the Voxire team will reply within <strong>24 hours</strong>, usually much sooner.
    </p>

    %s

    <p style="font-size:14px;color:#555;line-height:1.7;margin:0 0 28px 0;">
      If your project is time-sensitive, reach the team on WhatsApp:
    </p>

    <a href="https://wa.me/9613940708?text=Hi+Voxire%%2C+I+just+submitted+a+request+on+voxire.com"
       style="display:inline-block;background:#25d366;color:#fff;padding:12px 24px;border-radius:8px;text-decoration:none;font-size:14px;font-weight:600;">
      WhatsApp Voxire
    </a>

  </div>

  <p style="font-size:12px;color:rgba(0,0,0,0.3);text-align:center;margin-top:20px;">
    Voxire Digital Agency &middot; Beirut, Lebanon &middot;
    <a href="https://voxire.com" style="color:rgba(0,0,0,0.3);text-decoration:none;">voxire.com</a>
  </p>

</div>
</body>
</html>`,
		lead.Name,
		summaryBlock,
	)

	_, emailErr := sendEmail(ctx, emailPkg.SendParams{
		To:      []string{lead.Email},
		Subject: "We received your message - Voxire",
		HTML:    html,
		From:    "Voxire <info@voxire.com>",
		ReplyTo: "info@voxire.com",
	})
	if emailErr != nil {
		logger.Ctx(ctx).Warn("Failed to send auto-reply to lead",
			logger.Error(emailErr),
		)
	}
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/modules/websitelead/... -v 2>&1 | grep -E "^(--- FAIL|FAIL|ok)"`
Expected: only `ok` lines.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/websitelead/internal/application/service.go internal/modules/websitelead/internal/application/service_test.go
git commit -m "Send the website lead auto-reply in the team voice"
```

---

### Task 5: Newsletter subscriber table and repository

**Files:**
- Create: `internal/platform/database/schemas/voxire/newsletter_subscribers.sql`
- Create: `internal/modules/websitelead/internal/domain/subscriber.go`
- Modify: `internal/modules/websitelead/sql/schema.sql`
- Modify: `internal/modules/websitelead/sql/queries/websitelead.sql`
- Modify: `internal/modules/websitelead/internal/infrastructure/repository.go`
- Regenerate: `internal/modules/websitelead/internal/infrastructure/persistence/sqlc/*.go` via `make generate`

**Interfaces:**
- Produces: `domain.SubscriberInput{Email, Source, PostSlug, PostTitle, PostCategory, Language string}`, `domain.AllowedNewsletterSources []string`, `domain.IsAllowedNewsletterSource(s string) bool`, `domain.IsAllowedLanguage(s string) bool`, `(*Repository).CreateSubscriber(ctx, domain.SubscriberInput) (int64, error)`, sqlc `CreateNewsletterSubscriber`.

- [ ] **Step 1: Write the domain type and its test**

`internal/modules/websitelead/internal/domain/subscriber.go`:

```go
package domain

// SubscriberInput is one newsletter signup from voxire.com.
type SubscriberInput struct {
	Email        string
	Source       string
	PostSlug     string
	PostTitle    string
	PostCategory string
	Language     string
}

// AllowedNewsletterSources are the form placements on voxire.com that can
// produce a signup. Anything else is rejected at validation.
var AllowedNewsletterSources = []string{
	"intel_drop",
	"report_banner",
	"article_pdf_mid",
	"article_pdf_end",
}

func IsAllowedNewsletterSource(s string) bool {
	for _, allowed := range AllowedNewsletterSources {
		if s == allowed {
			return true
		}
	}
	return false
}

// IsAllowedLanguage accepts the two languages the site is served in.
func IsAllowedLanguage(s string) bool {
	return s == "en" || s == "ar"
}
```

`internal/modules/websitelead/internal/domain/subscriber_test.go`:

```go
package domain

import "testing"

func TestIsAllowedNewsletterSource(t *testing.T) {
	for _, ok := range []string{"intel_drop", "report_banner", "article_pdf_mid", "article_pdf_end"} {
		if !IsAllowedNewsletterSource(ok) {
			t.Errorf("%q should be allowed", ok)
		}
	}
	for _, bad := range []string{"", "web3forms", "INTEL_DROP", "article_pdf"} {
		if IsAllowedNewsletterSource(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestIsAllowedLanguage(t *testing.T) {
	if !IsAllowedLanguage("en") || !IsAllowedLanguage("ar") {
		t.Fatal("en and ar must be allowed")
	}
	if IsAllowedLanguage("") || IsAllowedLanguage("fr") || IsAllowedLanguage("EN") {
		t.Fatal("only lowercase en and ar are allowed")
	}
}
```

Run: `go test ./internal/modules/websitelead/internal/domain/ -v 2>&1 | tail -5`
Expected: PASS.

- [ ] **Step 2: Add the table to both schema locations**

Append to `internal/modules/websitelead/sql/schema.sql`:

```sql

CREATE TABLE IF NOT EXISTS voxire_newsletter_subscribers (
    id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    source        VARCHAR(50)  NOT NULL,
    post_slug     VARCHAR(255),
    post_title    VARCHAR(255),
    post_category VARCHAR(100),
    language      VARCHAR(5)   NOT NULL DEFAULT 'en',
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_newsletter_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

Create `internal/platform/database/schemas/voxire/newsletter_subscribers.sql` with the identical statement (no leading blank line). This is the copy `scripts/migrate.sh` applies in production; the module copy is the sqlc input.

- [ ] **Step 3: Add the query and regenerate**

Append to `internal/modules/websitelead/sql/queries/websitelead.sql`:

```sql

-- name: CreateNewsletterSubscriber :execresult
INSERT INTO voxire_newsletter_subscribers (email, source, post_slug, post_title, post_category, language)
VALUES (?, ?, ?, ?, ?, ?);
```

Run: `make generate 2>&1 | grep -E "websitelead|Failed" ; git status --short internal/modules/websitelead/internal/infrastructure/persistence/sqlc/`
Expected: the websitelead module regenerates successfully and `models.go` plus `websitelead.sql.go` show as modified with a `VoxireNewsletterSubscriber` struct and `CreateNewsletterSubscriber` method.

- [ ] **Step 4: Add the repository method**

Append to `internal/modules/websitelead/internal/infrastructure/repository.go`:

```go

func (r *Repository) CreateSubscriber(ctx context.Context, input domain.SubscriberInput) (int64, error) {
	result, err := r.queries.CreateNewsletterSubscriber(ctx, sqlc.CreateNewsletterSubscriberParams{
		Email:        input.Email,
		Source:       input.Source,
		PostSlug:     sql.NullString{String: input.PostSlug, Valid: input.PostSlug != ""},
		PostTitle:    sql.NullString{String: input.PostTitle, Valid: input.PostTitle != ""},
		PostCategory: sql.NullString{String: input.PostCategory, Valid: input.PostCategory != ""},
		Language:     input.Language,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert newsletter subscriber", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}
```

Run: `go build ./... && go vet ./internal/modules/websitelead/...`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add internal/platform/database/schemas/voxire/newsletter_subscribers.sql internal/modules/websitelead/sql internal/modules/websitelead/internal/infrastructure internal/modules/websitelead/internal/domain/subscriber.go internal/modules/websitelead/internal/domain/subscriber_test.go
git commit -m "Add the newsletter subscribers table and repository"
```

---

### Task 6: Subscribe service with its notification

**Files:**
- Create: `internal/modules/websitelead/internal/application/newsletter.go`
- Create: `internal/modules/websitelead/internal/application/newsletter_test.go`
- Modify: `internal/modules/websitelead/internal/application/service.go` (repository interface)
- Modify: `internal/modules/websitelead/internal/application/service_test.go` (mock)

**Interfaces:**
- Consumes: `domain.SubscriberInput`, `(*Repository).CreateSubscriber` from Task 5, `sendEmail` seam and `Notify.NewsletterEmails` from Task 2.
- Produces: `(*Service).Subscribe(ctx, domain.SubscriberInput) (int64, error)`.

- [ ] **Step 1: Write the failing test**

`internal/modules/websitelead/internal/application/newsletter_test.go`:

```go
package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
)

func TestSubscribe_Delegates(t *testing.T) {
	// A successful subscribe fires one detached notification. Wait for it so
	// no goroutine outlives the seam swap.
	_, wait := captureEmails(t, 1)
	m := &mockRepo{subscribeID: 3}
	s := &Service{repo: m, notify: testNotify()}
	id, err := s.Subscribe(context.Background(), domain.SubscriberInput{Email: "r@example.com", Source: "intel_drop", Language: "en"})
	if err != nil || id != 3 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if m.subscribeGot.Email != "r@example.com" {
		t.Fatalf("repo got %+v", m.subscribeGot)
	}
	wait()
}

func TestSubscribe_PropagatesError(t *testing.T) {
	captureEmails(t, 0)
	want := errors.New("db down")
	s := &Service{repo: &mockRepo{subscribeErr: want}, notify: testNotify()}
	if _, err := s.Subscribe(context.Background(), domain.SubscriberInput{Email: "r@example.com"}); !errors.Is(err, want) {
		t.Fatalf("expected sentinel, got %v", err)
	}
}

func TestSubscribe_NotifiesNewsletterListOnly(t *testing.T) {
	got, wait := captureEmails(t, 1)
	s := &Service{repo: &mockRepo{subscribeID: 11}, notify: testNotify()}
	_, err := s.Subscribe(context.Background(), domain.SubscriberInput{
		Email: "rania@example.com", Source: "article_pdf_mid", PostSlug: "seo-lebanon-2026",
		PostTitle: "SEO in Lebanon", PostCategory: "SEO", Language: "ar",
	})
	if err != nil {
		t.Fatal(err)
	}
	wait()
	if len(*got) != 1 {
		t.Fatalf("expected exactly one email, got %d", len(*got))
	}
	p := (*got)[0]
	if strings.Join(p.To, ",") != "news@example.com" {
		t.Fatalf("To = %v, want the newsletter list", p.To)
	}
	if p.Subject != "[Newsletter] rania@example.com via article_pdf_mid" {
		t.Fatalf("Subject = %q", p.Subject)
	}
	if p.From != "Voxire Website <noreply@voxire.com>" || p.ReplyTo != "rania@example.com" {
		t.Fatalf("From=%q ReplyTo=%q", p.From, p.ReplyTo)
	}
	for _, want := range []string{"SEO in Lebanon", "seo-lebanon-2026", "SEO", "ar", "#11"} {
		if !strings.Contains(p.HTML, want) {
			t.Fatalf("HTML missing %q", want)
		}
	}
	if strings.Contains(p.HTML, "\u2014") {
		t.Fatal("HTML contains an em dash")
	}
}
```

Extend `mockRepo` in `service_test.go`:

```go
type mockRepo struct {
	createID     int64
	createErr    error
	createGot    domain.LeadInput
	subscribeID  int64
	subscribeErr error
	subscribeGot domain.SubscriberInput
}

func (m *mockRepo) CreateSubscriber(_ context.Context, input domain.SubscriberInput) (int64, error) {
	m.subscribeGot = input
	return m.subscribeID, m.subscribeErr
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/modules/websitelead/internal/application/ 2>&1 | head -5`
Expected: FAIL to compile with `s.Subscribe undefined`.

- [ ] **Step 3: Implement Subscribe**

In `service.go`, extend the repository interface:

```go
type repository interface {
	Create(ctx context.Context, input domain.LeadInput) (int64, error)
	CreateSubscriber(ctx context.Context, input domain.SubscriberInput) (int64, error)
}
```

`internal/modules/websitelead/internal/application/newsletter.go`:

```go
package application

import (
	"context"
	"fmt"
	"html"
	"time"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

// Subscribe stores a newsletter signup and tells the team about it. The row
// is the source of truth; the notification runs detached like lead emails.
func (s *Service) Subscribe(ctx context.Context, input domain.SubscriberInput) (int64, error) {
	id, err := s.repo.CreateSubscriber(ctx, input)
	if err != nil {
		return 0, err
	}
	go func() {
		taskCtx, cancel := asyncLeadContext(ctx)
		defer cancel()
		s.sendSubscriberNotification(taskCtx, id, input)
	}()
	return id, nil
}

func (s *Service) sendSubscriberNotification(ctx context.Context, id int64, sub domain.SubscriberInput) {
	esc := html.EscapeString
	rows := fieldRow("Email", fmt.Sprintf(`<a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a>`, esc(sub.Email), esc(sub.Email)))
	rows += fieldRow("Source", esc(sub.Source))
	if sub.PostTitle != "" {
		rows += fieldRow("Post", esc(sub.PostTitle))
	}
	if sub.PostSlug != "" {
		rows += fieldRow("Slug", fmt.Sprintf(`<a href="https://voxire.com/blog/%s/" style="color:#63c6bd;text-decoration:none;">%s</a>`, esc(sub.PostSlug), esc(sub.PostSlug)))
	}
	if sub.PostCategory != "" {
		rows += fieldRow("Category", esc(sub.PostCategory))
	}
	rows += fieldRow("Language", esc(sub.Language))

	now := time.Now().UTC().Format("2 Jan 2006, 15:04 UTC")
	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<body style="margin:0;padding:0;background:#0d0d14;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<div style="max-width:620px;margin:0 auto;padding:40px 16px;">
  <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
    <span style="font-size:18px;font-weight:700;color:#63c6bd;letter-spacing:-0.01em;">Voxire</span>
    <span style="font-size:12px;color:rgba(240,240,246,0.35);">%s</span>
  </div>
  <div style="background:#111118;border-radius:14px;padding:32px;border:1px solid rgba(255,255,255,0.06);">
    <h2 style="font-size:20px;font-weight:700;margin:0 0 6px 0;color:#f0f0f6;">New newsletter subscriber</h2>
    <p style="font-size:13px;color:rgba(240,240,246,0.4);margin:0 0 28px 0;">Subscriber #%d via voxire.com</p>
    <table style="width:100%%;border-collapse:collapse;">%s</table>
  </div>
  <p style="font-size:12px;color:rgba(240,240,246,0.2);text-align:center;margin-top:24px;">Voxire newsletter: voxire.com</p>
</div>
</body>
</html>`, now, id, rows)

	_, err := sendEmail(ctx, emailPkg.SendParams{
		To:      s.notify.NewsletterEmails,
		Subject: fmt.Sprintf("[Newsletter] %s via %s", sub.Email, sub.Source),
		HTML:    body,
		From:    "Voxire Website <noreply@voxire.com>",
		ReplyTo: sub.Email,
	})
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to send newsletter subscriber notification",
			logger.Error(err),
			logger.Any("subscriber_id", id),
		)
	}
}
```

`fieldRow` already exists in `service.go` and is reused.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/modules/websitelead/... -v 2>&1 | grep -E "^(--- FAIL|FAIL|ok)"`
Expected: only `ok` lines.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/websitelead/internal/application/newsletter.go internal/modules/websitelead/internal/application/newsletter_test.go internal/modules/websitelead/internal/application/service.go internal/modules/websitelead/internal/application/service_test.go
git commit -m "Store newsletter signups and notify the configured inbox"
```

---

### Task 7: Newsletter HTTP handler and public route

**Files:**
- Create: `internal/modules/websitelead/internal/api/newsletter_handler.go`
- Create: `internal/modules/websitelead/internal/api/newsletter_handler_test.go`
- Modify: `internal/modules/websitelead/internal/api/routes.go`
- Modify: `internal/modules/websitelead/module.go` (`GetAnonymousRoutes`)

**Interfaces:**
- Consumes: `(*application.Service).Subscribe`, `domain.IsAllowedNewsletterSource`, `domain.IsAllowedLanguage`.
- Produces: `POST /newsletter` accepting `NewsletterRequest{email, source, post_slug, post_title, post_category, language}` and returning `NewsletterResponse{id, message}`.

- [ ] **Step 1: Write the failing test**

`internal/modules/websitelead/internal/api/newsletter_handler_test.go`:

```go
package api

import (
	"context"
	"strings"
	"testing"
)

func validNewsletter() *NewsletterRequest {
	return &NewsletterRequest{Email: "rania@example.com", Source: "intel_drop"}
}

func TestNewsletterHandler_Validate_HappyPathEverySource(t *testing.T) {
	h := &NewsletterHandler{}
	for _, src := range []string{"intel_drop", "report_banner", "article_pdf_mid", "article_pdf_end"} {
		req := validNewsletter()
		req.Source = src
		if err := h.Validate(context.Background(), req); err != nil {
			t.Errorf("source %q: unexpected error %v", src, err)
		}
	}
}

func TestNewsletterHandler_Validate_Rejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*NewsletterRequest)
		want   string
	}{
		{"missing email", func(r *NewsletterRequest) { r.Email = "" }, "email is required"},
		{"malformed email", func(r *NewsletterRequest) { r.Email = "not-an-email" }, "email is invalid"},
		{"email without tld", func(r *NewsletterRequest) { r.Email = "a@b" }, "email is invalid"},
		{"email too long", func(r *NewsletterRequest) { r.Email = strings.Repeat("a", 250) + "@x.com" }, "email is invalid"},
		{"missing source", func(r *NewsletterRequest) { r.Source = "" }, "source is required"},
		{"unknown source", func(r *NewsletterRequest) { r.Source = "web3forms" }, "source is invalid"},
		{"bad language", func(r *NewsletterRequest) { r.Language = "fr" }, "language is invalid"},
		{"slug too long", func(r *NewsletterRequest) { r.PostSlug = strings.Repeat("s", 256) }, "post_slug is too long"},
		{"title too long", func(r *NewsletterRequest) { r.PostTitle = strings.Repeat("t", 256) }, "post_title is too long"},
		{"category too long", func(r *NewsletterRequest) { r.PostCategory = strings.Repeat("c", 101) }, "post_category is too long"},
	}
	h := &NewsletterHandler{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validNewsletter()
			tc.mutate(req)
			err := h.Validate(context.Background(), req)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestNewsletterHandler_Validate_DefaultsLanguageAndTrims(t *testing.T) {
	h := &NewsletterHandler{}
	req := &NewsletterRequest{Email: "  Rania@Example.com ", Source: "intel_drop"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if req.Email != "rania@example.com" {
		t.Fatalf("email not normalised: %q", req.Email)
	}
	if req.Language != "en" {
		t.Fatalf("language not defaulted: %q", req.Language)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/modules/websitelead/internal/api/ 2>&1 | head -5`
Expected: FAIL to compile with `undefined: NewsletterHandler`.

- [ ] **Step 3: Write the handler**

`internal/modules/websitelead/internal/api/newsletter_handler.go`:

```go
package api

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func NewsletterRoutes() []string { return []string{"/newsletter"} }

type NewsletterHandler struct{ Service *application.Service }

type NewsletterRequest struct {
	Email        string `json:"email"`
	Source       string `json:"source"`
	PostSlug     string `json:"post_slug"`
	PostTitle    string `json:"post_title"`
	PostCategory string `json:"post_category"`
	Language     string `json:"language"`
}

type NewsletterResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// simpleEmail is deliberately loose: one @, something on both sides, a dot in
// the domain. Deliverability is Resend's problem, not the form's.
var simpleEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Validate normalises the request in place (trimmed lower-case email, default
// language) before checking it, so Serve stores exactly what was validated.
func (h *NewsletterHandler) Validate(_ context.Context, req *NewsletterRequest) error {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Source = strings.TrimSpace(req.Source)
	req.Language = strings.TrimSpace(req.Language)
	if req.Language == "" {
		req.Language = "en"
	}

	if req.Email == "" {
		return onekiterrors.New(http.StatusBadRequest, "email is required")
	}
	if len(req.Email) > 255 || !simpleEmail.MatchString(req.Email) {
		return onekiterrors.New(http.StatusBadRequest, "email is invalid")
	}
	if req.Source == "" {
		return onekiterrors.New(http.StatusBadRequest, "source is required")
	}
	if !domain.IsAllowedNewsletterSource(req.Source) {
		return onekiterrors.New(http.StatusBadRequest, "source is invalid")
	}
	if !domain.IsAllowedLanguage(req.Language) {
		return onekiterrors.New(http.StatusBadRequest, "language is invalid")
	}
	if len(req.PostSlug) > 255 {
		return onekiterrors.New(http.StatusBadRequest, "post_slug is too long")
	}
	if len(req.PostTitle) > 255 {
		return onekiterrors.New(http.StatusBadRequest, "post_title is too long")
	}
	if len(req.PostCategory) > 100 {
		return onekiterrors.New(http.StatusBadRequest, "post_category is too long")
	}
	return nil
}

func (h *NewsletterHandler) Serve(ctx context.Context, req *NewsletterRequest) (NewsletterResponse, error) {
	id, err := h.Service.Subscribe(ctx, domain.SubscriberInput{
		Email:        req.Email,
		Source:       req.Source,
		PostSlug:     req.PostSlug,
		PostTitle:    req.PostTitle,
		PostCategory: req.PostCategory,
		Language:     req.Language,
	})
	if err != nil {
		return NewsletterResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to subscribe")
	}
	return NewsletterResponse{ID: id, Message: "Subscribed"}, nil
}
```

In `routes.go`, add after the contact registration and before the analytics one:

```go
	http.RegisterPOSTWithValidationAndMiddleware[NewsletterRequest, NewsletterResponse](
		ctx, router, NewsletterRoutes(), &NewsletterHandler{Service: service},
		http.PerRouteRateLimit(10, time.Minute),
	)
```

In `module.go`, change `GetAnonymousRoutes`:

```go
func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/quote", "/contact", "/newsletter", "/analytics/event"}
}
```

- [ ] **Step 4: Run the tests and the repository gates**

Run: `go build ./... && go test ./internal/modules/websitelead/... ./internal/platform/config/ 2>&1 | grep -E "^(--- FAIL|FAIL|ok)" && make json-tags-check && make mutation-validation-check`
Expected: only `ok` lines, both checks exit 0.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/websitelead/internal/api/newsletter_handler.go internal/modules/websitelead/internal/api/newsletter_handler_test.go internal/modules/websitelead/internal/api/routes.go internal/modules/websitelead/module.go
git commit -m "Expose a public newsletter signup endpoint"
```

---

### Task 8: Full onion verification and push

**Files:** none new.

- [ ] **Step 1: Run every gate**

Run, from the worktree root:

```bash
make lint && make vet && make test 2>&1 | tail -15 && make json-tags-check && make mutation-validation-check && python3 scripts/check_env_example.py
```

Expected: lint clean (it runs with `--fix`, so `git status` must stay clean afterwards or the fixes need committing as their own commit "Apply lint fixes to website lead routing"), vet clean, `make test` ends with the SUCCESS line and no FAIL, both checks and the env check exit 0.

- [ ] **Step 2: Confirm the diff is the intended set**

Run: `git log --oneline origin/main..HEAD && git diff --stat origin/main..HEAD`
Expected: nine commits (spec, plan, Tasks 1 to 7) touching only the files in the file map plus regenerated sqlc output.

- [ ] **Step 3: Push and open the PR**

```bash
git push -u origin feat/website-leads-routing
gh pr create --title "Route website leads to the team and ingest newsletter signups" --body-file docs/superpowers/specs/2026-09-21-website-leads-routing-design.md
```

Then stop. The deploy and the `scripts/migrate.sh` run for `newsletter_subscribers.sql` are done by whoever owns onion deploys. Do not start Task 9 until step 1 of the rollout order in the spec passes:

```bash
curl -s -o /dev/stdout -w "\n%{http_code}\n" -X POST -H 'Content-Type: application/json' --data '{}' https://api.voxire.com/newsletter
```

Expected: `{"message":"email is required","statusCode":400}` and `400`.

---

### Task 9: Blog popup posts to the API

**Files:**
- Modify: `/Users/abedamouneh/Development/voxire/src/components/NewsletterDrop.tsx:1-6` (imports) and `:57-88` (`handleSubmit`)

**Interfaces:**
- Consumes: `api` default export from `@/src/lib/api` (axios instance, base URL resolves to `https://api.voxire.com/` in production); `useTranslation()` returns `{ t, lang }`.

- [ ] **Step 1: Confirm the branch**

Run: `cd /Users/abedamouneh/Development/voxire && git branch --show-current && git status --short`
Expected: `main`, clean apart from `.codex-handoff/`. If not on `main`, stop and report.

- [ ] **Step 2: Rewrite the imports and the submit handler**

Replace the import block at the top of `NewsletterDrop.tsx`:

```tsx
import { useState, useEffect, useRef } from "react";
import { X, ArrowRight, CheckCircle, Envelope } from "@phosphor-icons/react";
import { useTranslation } from "@/src/i18n";
import { gtagEvent } from "@/src/lib/gtag";
import api from "@/src/lib/api";
```

Change `const { t } = useTranslation();` to `const { t, lang } = useTranslation();`.

Replace `handleSubmit`:

```tsx
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) return;
    setStatus("sending");
    setErrorMsg("");

    try {
      await api.post("/newsletter", {
        email,
        source: "intel_drop",
        post_slug: postSlug,
        post_category: postCategory,
        language: lang,
      });
      setStatus("success");
      localStorage.setItem(SUBSCRIBED_KEY, "1");
      gtagEvent("newsletter_signup", { label: postSlug, method: "intel_drop" });
      setTimeout(() => setVisible(false), 4000);
    } catch {
      setStatus("error");
      setErrorMsg(t("newsletterDrop.errorMsg"));
    }
  };
```

- [ ] **Step 3: Typecheck**

Run: `npm run typecheck`
Expected: exits 0. If `lang` is not exposed by `useTranslation`, read `src/i18n/index.tsx:5` which documents `const { t, lang, setLang, dir } = useTranslation();` and fix the destructuring, not the hook.

- [ ] **Step 4: Commit**

```bash
git add src/components/NewsletterDrop.tsx
git commit -m "Send blog popup newsletter signups to the Voxire API"
```

---

### Task 10: Report banner posts to the API

**Files:**
- Modify: `/Users/abedamouneh/Development/voxire/src/components/NewsletterReportBanner.tsx:1-7` (imports) and the Web3Forms block inside `handleSubmit` (around lines 252-277)

**Interfaces:**
- Consumes: same as Task 9.

- [ ] **Step 1: Rewrite the imports**

```tsx
import { useState, useEffect } from "react";
import { ArrowRight, CheckCircle, FileArrowDown, DownloadSimple } from "@phosphor-icons/react";
import { useTranslation } from "@/src/i18n";
import { gtagEvent } from "@/src/lib/gtag";
import api from "@/src/lib/api";
import type { BlogPost } from "@/src/types";
```

Change `const { t } = useTranslation();` in the component to `const { t, lang } = useTranslation();`.

- [ ] **Step 2: Replace the lead-capture block**

Replace everything from the comment `/* Fire-and-forget lead capture via Web3Forms.` through the closing `}` of its `catch` with:

```tsx
    /* Fire-and-forget lead capture. If the call fails, the user already has their PDF. */
    try {
      await api.post("/newsletter", {
        email,
        source: hasPost ? `article_pdf_${variant}` : "report_banner",
        post_slug: post?.slug ?? "",
        post_title: post?.title ?? "",
        post_category: post?.category ?? "",
        language: lang,
      });
    } catch {
      /* Network or server error: PDF already delivered, swallow silently */
    }
```

`variant` is `"mid" | "end" | "listing"`. When `hasPost` is true the component is rendered with `variant="mid"` from `pages/blog/[slug].tsx:719`, so the source is `article_pdf_mid`; the `end` variant is not mounted today but stays valid. `listing` is only used without a post, which maps to `report_banner`.

- [ ] **Step 3: Typecheck**

Run: `npm run typecheck`
Expected: exits 0.

- [ ] **Step 4: Commit**

```bash
git add src/components/NewsletterReportBanner.tsx
git commit -m "Send blog report banner signups to the Voxire API"
```

---

### Task 11: Remove Web3Forms from the site

**Files:**
- Delete: `/Users/abedamouneh/Development/voxire/src/lib/web3forms.ts`
- Modify: `/Users/abedamouneh/Development/voxire/pages/contact.tsx:216-241`
- Modify: `/Users/abedamouneh/Development/voxire/.env.local.example`

- [ ] **Step 1: Delete the shared key module**

Run: `git rm src/lib/web3forms.ts && grep -rn "web3forms\|WEB3FORMS" pages src --include='*.ts' --include='*.tsx'`
Expected: the only remaining hits are in `pages/contact.tsx` and the comment in `pages/ar/contact.tsx:10`.

- [ ] **Step 2: Collapse the contact submit to one path**

In `pages/contact.tsx`, replace the block from `const web3formsKey = process.env.NEXT_PUBLIC_WEB3FORMS_KEY;` through the closing `}` of the `else` branch (the lines that end with `await api.post("/contact", { ...formData, message: enrichedMessage });` and `}`) with:

```tsx
    try {
      await api.post("/contact", { ...formData, message: enrichedMessage });
```

The `setStatus("success");` line and everything after it stay as they are inside the same `try`.

In `pages/ar/contact.tsx`, replace the doc comment lines 9 to 12 so the paragraph reads:

```
 * WhatsApp first, deliberately. The English page posts to api.voxire.com;
 * duplicating that here would add a second form for a market where WhatsApp
 * is how business actually starts. The full form stays one link away.
```

- [ ] **Step 3: Remove the env documentation**

Replace the entire content of `.env.local.example` with:

```
# No client-side secrets are needed to run the site. Forms post to
# api.voxire.com in production and to http://localhost:9999/ in development
# (see src/lib/api.ts), so they return errors locally unless onion is running.
```

- [ ] **Step 4: Typecheck and build**

Run: `npm run typecheck && pnpm build 2>&1 | tail -8`
Expected: typecheck exits 0, build ends with the build-info line and no error. Then:

```bash
grep -rl "api.web3forms.com" out/_next/static/chunks | wc -l
```

Expected: `0`.

- [ ] **Step 5: Commit in three pieces**

```bash
git add src/lib/web3forms.ts src/components  # records the deletion; components already committed
git commit -m "Delete the shared Web3Forms key module"
git add pages/contact.tsx pages/ar/contact.tsx
git commit -m "Submit the contact form through the Voxire API only"
git add .env.local.example
git commit -m "Drop the Web3Forms environment documentation"
```

If `git add src/components` stages nothing, that is expected; the first commit contains only the deletion.

---

### Task 12: Rollout verification

**Files:** none.

- [ ] **Step 1: Gate on the onion deploy**

Run the curl from Task 8 step 3. Do not continue until it returns `400` with `email is required`.

- [ ] **Step 2: Push voxire**

Run: `cd /Users/abedamouneh/Development/voxire && git branch --show-current && git log --oneline origin/main..main`
Expected: `main` and the five commits from Tasks 9 to 11. Then `git push origin main`.

- [ ] **Step 3: Wait for Cloudflare and inspect the live chunk**

```bash
for i in $(seq 1 40); do c=$(curl -s "https://voxire.com/build-info.json?n=$i" | grep -oE '"commit": "[0-9a-f]{8}' | cut -c12-); [ "$c" = "$(git rev-parse --short=8 HEAD)" ] && echo "deployed $c" && break; sleep 10; done
curl -s https://voxire.com/blog/ -o /tmp/vx-blog.html && grep -oE 'src="[^"]*_next/static/[^"]+\.js"' /tmp/vx-blog.html | sed 's/src=//; s/"//g' | sort -u | while read -r u; do case "$u" in http*) url="$u";; *) url="https://voxire.com$u";; esac; curl -s "$url"; done | grep -c "api.web3forms.com"
```

Expected: `deployed <sha>` then `0`.

- [ ] **Step 4: One real signup from each form**

Open a live blog post in the browser, scroll past 82%, subscribe with `newsletter-rollout-test@voxire.com`. Then on `/blog/` use the report banner with the same address. Confirm:

- each shows its success state,
- two `[Newsletter] newsletter-rollout-test@voxire.com via ...` emails arrive in info@voxire.com, one `intel_drop` and one `report_banner`,
- whoever has database access sees two rows in `voxire_newsletter_subscribers`.

Report the outcome with the two subjects seen. If either email does not arrive within ten minutes, check the onion logs for `Failed to send newsletter subscriber notification` before touching code.
