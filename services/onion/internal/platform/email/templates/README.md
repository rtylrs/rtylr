# email-templates

Transactional HTML email templates for rtylr, plus their assets. They live in
the `onion` backend at `internal/platform/email/templates` and are embedded into
the binary at build time via the `//go:embed` directives in
`internal/platform/email/render.go`.

## Layout

`base.html` is the shared layout and lives in the root. Every body template
lives under `templates/<domain>/` so the repo stays navigable as it grows:

```
base.html                              Outer layout — <head>, styles, header logo, footer
templates/welcome/welcome.html         Sent on signup
templates/auth/password_reset.html     Sent on password-reset request
templates/team/invitation.html         Sent when a user is invited to a business
templates/team/access_request.html     Sent when a user requests app access
templates/billing/invoice.html         Sent with an invoice
assets/                                Images / static files referenced by absolute URL from the HTML
```

New templates go under the matching domain folder in `templates/` (`auth/`,
`team/`, `billing/`, `welcome/`, …) — create a new folder when a genuinely new
domain appears. The Go renderer addresses a body template by its path under
`templates/` without the extension, e.g. `Render("auth/password_reset", data)`.

## Template contract

`base.html` is always parsed first and renders the chrome. It pulls in the named
template's body via a `body` block and titles the page from `.Subject`:

```html
<title>{{.Subject}}</title>
...
<div class="body">{{template "body" .}}</div>
```

Every other file **must** define exactly one `body` block:

```html
{{define "body"}}
  <h1>...</h1>
  ...
{{end}}
```

The Go renderer (`onion/internal/platform/email/render.go`) executes
`base.html` + `<domain>/<name>.html` together and passes a per-template data
struct. The fields each template expects:

| Template               | Data fields                                 |
| ---------------------- | ------------------------------------------- |
| `welcome/welcome`      | `Name`                                      |
| `team/invitation`      | `InviterName`, `BusinessName` (either may be empty), `InviteLink`, `ValidDays` |
| `team/access_request`  | `RequesterName`, `RequesterEmail`, `App`, `TeamLink` |
| `auth/password_reset`  | (see `email.go` call site)                  |
| `billing/invoice`      | (see `email.go` call site)                  |

`.Subject` is available to every template (used by the base layout).

## Look and buttons

The layout follows rtylr.com: a flat white card with a hairline border on a
soft grey page, ink (`#131313`) type and a black button, in DM Sans with system
fallbacks. Clients that support `prefers-color-scheme` get a dark variant with
the light logo.

Buttons go through the base layout's `button` template so they render as a
table cell with an inline fill, which survives clients that strip `<style>`:

```html
{{template "button" (action .InviteLink "Accept invitation")}}
```

`action` pairs the link with its label; the link is escaped and unsafe schemes
are filtered like any other `href`.

## Adding a template

1. Add `<domain>/<name>.html` here with a single `{{define "body"}}...{{end}}`
   block (create the domain folder if it's new).
2. In `onion`, add a `Render("<domain>/<name>", data)` call in
   `internal/platform/email/email.go` with the matching data struct, plus a case
   in `internal/platform/email/render_test.go`.

## Assets

HTML email clients don't run JS and have spotty support for embedded data URIs,
so images are referenced by **absolute hosted URL**, not inlined. Keep source
images under `assets/` for reference; host the served copies on the CDN and point
the HTML at those URLs. The `//go:embed` patterns in
`internal/platform/email/render.go` only embed HTML — if a template ever needs a
binary asset embedded into the binary, widen those patterns accordingly.

## Editing

These render in email clients (Outlook, Gmail, Apple Mail), not browsers — keep
to inline-safe, table-friendly HTML and the styles already defined in
`base.html`. Test a change by rendering it through `onion` before shipping.
