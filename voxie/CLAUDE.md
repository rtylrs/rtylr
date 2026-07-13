# CLAUDE.md — voxie Contracts

## 1. PROJECT IDENTITY

- **Product:** RTYLR — commerce operating system
- **Repo role:** API contracts. Single source of truth for all API request/response types shared between the Go backend (`onion`) and the TypeScript frontend (`rtylr`). Neither repo writes its own type definitions for cross-boundary types — they import from here.
- **Schema language:** [onekit](https://github.com/1homsi/onekit)'s `.onk` — every service has been fully migrated off Protocol Buffers; `proto/`, `buf`, and the old `corezio/onekit` protoc plugins no longer exist in this repo.
- **Built by:** Voxire — Abed El-Fattah Amouneh & Mohammad Homsi
- **Business email:** info@voxire.com
- **WhatsApp:** +961 3 940 708

---

## 2. GIT WORKFLOW — RULES

- Feature branches are the **default**. Prefix: `feat/`, `fix/`, `chore/`, `refactor/`
- **Direct push to `main`** is acceptable ONLY for: trivial fixes to `.onk` comments or formatting. Structural or type-breaking changes always go through a branch.
- **No `.md` files** in commits except `CLAUDE.md` itself.
- **Atomic commits** — one logical unit per commit. A new field and its comments belong together.
- Commit message format: short imperative, no conventional prefix. `add staff user payload` not `feat: add staff user payload`.
- **`gen/go` and `gen/ts` are both committed.** CI (`.github/workflows/contracts.yml`) regenerates them from `api/*.onk` on every push to `main` and fails PRs on drift — never hand-edit either. `rtylr` and `pos` consume `gen/ts` directly from the submodule; there is no separate frontend generation step anymore.
- After committing to `voxie`, update the submodule ref in both `rtylr` and `onion` repos.

### Definition of Done

An `.onk` change is only complete when ALL of the following are true:

- `onek check .` passes with zero errors
- All new fields have a comment describing their purpose
- `make generate` was run and `gen/go` + `gen/ts` are committed alongside the change (CI will also catch drift, but don't rely on it)
- Submodule refs updated in `onion` and `packages` (shared by `rtylr` and `pos`)
- If a new domain or structural pattern was introduced, it is documented in `rtylr/docs/`

### PR Pre-Merge Checklist

Run through this before merging any contract change:

- [ ] `onek check .` passes with zero output
- [ ] All new fields have inline comments
- [ ] `make generate` run and `gen/go` + `gen/ts` committed (CI enforces this on push to `main`, but PRs must already be clean)
- [ ] Submodule ref updated in `onion` — no local regeneration needed there anymore, `gen/go` ships pre-built
- [ ] Submodule ref updated in `packages` (the `voxie` submodule lives directly at `packages/voxie`) — no local generation needed there either, `rtylr` and `pos` both pick it up via the shared `packages` submodule
- [ ] `rtylr/docs/` updated if the change affects the API contract

### When You Introduce a New Pattern

If you create a new `.onk` convention (a new naming pattern, a new use of `oneof`, a new enum structure, etc.), document it in this CLAUDE.md before the commit is made.

### Updating Documentation

**When you add a new domain, new messages, or make breaking changes that affect the API contract, update `rtylr/docs/`** to reflect the new types available and any migration notes.

---

## 3. TECH STACK

| Component | Technology | Version | Notes |
|---|---|---|---|
| Schema language | [onekit](https://github.com/1homsi/onekit) `.onk` | pinned tag (see `Makefile` `ONK_VERSION`) | Messages, enums, services, and HTTP/validation decorators (`@get`, `@required`, `@gt`, etc.) in one flat file per concern (`models.onk`, `service.onk`) |
| Build tool | `onek` CLI (`github.com/1homsi/onekit/cmd/onek`) | pinned (`ONK_VERSION`) | `onek check .` validates the whole schema tree; `onek build .` (via `make generate`) generates Go types/validation/server/client and TypeScript types/client, one package per schema directory |
| Go module | `github.com/voxire/voxie` | — | Imported in `onion` via `replace`; `gen/go` ships pre-built, no local generation needed in `onion`. Generated Go has zero external dependencies (pure stdlib) |
| TS package | `@voxire/voxie` (in the `packages` repo) | — | `packages/voxie` carries `voxie` as its own submodule and re-exports `gen/ts` directly via a wildcard `exports` map (`./*` → `./gen/ts/*.ts`) — no per-file export entry needed. `rtylr` and `pos` both consume it through the shared `packages` submodule, no local generation needed |

### Adding Dependencies

The `onek` CLI is a build-time tool only — generated Go/TypeScript code has zero runtime dependencies of its own (Go side is pure stdlib: `net/http`, `encoding/json`, etc.; TS side uses only `fetch`). Don't add anything voxie's own schema/codegen tooling doesn't require — application logic still belongs in `onion` and the frontend apps.

---

## 4. ARCHITECTURE & PATTERNS

### Repository Layout

```
voxie/
├── api/                         Source of truth — edit these files
│   ├── common/                  Shared cross-app contracts
│   │   ├── auth/v1/             login, register, verify_email, reset_password, google_auth, invite_accept, ...
│   │   ├── oauth/v1/            OAuth token issuing
│   │   └── pagination/v1/
│   ├── crm/{capability}/v1/
│   ├── erp/{capability}/v1/
│   ├── finance/{capability}/v1/
│   ├── flow/{capability}/v1/
│   ├── hr/{capability}/v1/
│   ├── hub/{capability}/{subdomain}/v1/
│   ├── insights/{capability}/v1/
│   └── pos/{capability}/{subdomain}/v1/
├── gen/go/                     Generated Go output — committed, regenerated by `make generate` / CI
├── gen/ts/                     Generated TypeScript output — committed, regenerated by `make generate` / CI
├── onekit.toml                 onek config — module path, generate targets (go-server, go-client, ts-client) and their `out` dirs
├── go.mod                      no proto/onekit-runtime deps — generated Go is pure stdlib
├── package.json                 devDependencies: typescript only (codegen tooling, no runtime deps)
├── scripts/generate_validation_metadata.py   legacy proto-era scanner, now a no-op (see "Known gaps" in README.md) — no .onk-aware equivalent exists yet
├── .github/workflows/contracts.yml   generate + build/vet + typecheck + drift-check + auto-commit CI
└── Makefile                    clean / install-onek / install-js / generate / typecheck targets
```

### `.onk` Conventions

- **Directory layout mirrors HTTP path:** a service under `api/hub/team/members/v1/` gets base path `/api/hub/team/members/v1` inferred automatically (kebab-casing multi-word segments, e.g. `time_entry` → `time-entry`) — no explicit `base_path:` unless the route is genuinely irregular.
- **Two files per service directory:** `models.onk` (messages + enums) and `service.onk` (the service + its RPCs).
- **Field naming:** `snake_case` in `.onk`. Generated Go uses PascalCase struct fields with `json:"snake_case"` tags; generated TypeScript keeps `snake_case` to match the wire JSON exactly.
- **Message naming:** PascalCase. e.g., `StaffUserPayload`, `ListTeamRequest`, `InviteUserResponse`.
- **Version suffix:** always `v1` in the directory path.
- **Payload vs Request/Response:** Use `Payload` for reusable data types, `Request`/`Response` for operation-specific envelopes.
- **All new fields must have a `///` doc comment** describing their purpose (drop pure boilerplate like "X is the X value for this contract" — comments should carry real information).
- **Reusable types:** payload-only messages live in the local `models.onk` beside the capability that owns them.
- **Shared contracts:** cross-app contracts live under `api/common/{domain}/v1/`, not under an app root.
- **Cross-directory references:** reference another directory's message/enum by its plain name — no import statement. `onek` resolves it automatically as long as the name isn't ambiguous (declared identically in more than one *other* directory); if it is, rename one side (see `crm.customer.v1.Customer` vs the POS-specific `PosCustomer` in `pos.customer.v1` for a real example of this).
- **`int64`/`uint64` wire encoding defaults to a JSON string** (matches proto's old `INT64_ENCODING_STRING`) — only add `@encode(number)` when a field genuinely needs to serialize as a plain JSON number.
- **Validation decorators:** `@required`, `@gt(N)`/`@gte(N)`/`@lt(N)`/`@lte(N)` (numeric bounds, stackable), `@pattern("regex")`, `@email`, `@uri`, `@min_items(N)`/`@max_items(N)` (repeated fields), `@in("a", "b", ...)`.
- **Optionality:** `field: T?` for an optional field; add `@nullable` when JSON `null` (not just field absence) is a meaningful, distinct value.
- **Request-body binding:** a whole request-as-body POST/PUT/PATCH needs no annotation. A request with a nested body sub-message (mirroring the old proto `body: "body"` HTTP option) uses `@body("field_name")` on the RPC declaration.
- **No cross-field validation yet:** a constraint like "exactly one of these fields is required" (the old proto `buf.validate.message.oneof`) has no `.onk` decorator equivalent. Leave the fields as plain optionals and add a hand-written check in the handler (see `flow.task.v1.ListTasksRequest` for the one existing case).
- **No field-reservation mechanism:** unlike proto's `reserved`, `.onk` doesn't track removed field numbers (there are no field numbers at all — fields are matched by JSON key). Never delete or rename a field a live consumer still reads; deprecate it in a comment instead (see below) and never reuse the old name for something semantically different.

### Generation Ownership

Both Go and TypeScript are owned and committed by `voxie` itself — neither consumer regenerates anything locally anymore:

- **Go (`gen/go`):** `make generate` (or CI on push to `main`) runs `onek build .`, which generates types, field validation, an HTTP server (`RegisterXServiceServer`), and an HTTP client (`NewXServiceClient`) per schema directory. `onion` consumes the submodule at whatever commit is checked out.
- **TypeScript (`gen/ts`):** the same `onek build .` run also generates TS types and a typed `fetch`-based client per service (self-contained — no runtime npm dependencies). `packages/voxie` (in the `packages` repo, shared by `rtylr` and `pos`) carries `voxie` as its own submodule and re-exports `gen/ts` directly via a wildcard `exports` map — no per-file export entry needed when a new schema file is added.
- **Frontend validation metadata** (`gen/ts/validation/metadata.ts`) is currently always empty — see "Known gaps" in `README.md`.

### How Consumers Import These Types

Import paths include the capability path:

**Go backend (`onion`):**

```go
import membersv1 "github.com/voxire/voxie/gen/go/api/hub/team/members/v1"

return handlers.ListUsersResponse{
    Users: []*membersv1.StaffUserPayload{ ... },
}, nil
```

**TypeScript frontends (`rtylr`, `pos`) via `@voxire/voxie`:**

```ts
import type { StaffUserPayload } from '@voxire/voxie/api/hub/team/members/v1/types';
import type { LoginRequest } from '@voxire/voxie/api/common/auth/v1/types';
import { FloorServiceClient } from '@voxire/voxie/api/pos/floor/v1/client';
```

### Deprecating and Removing Fields

**Never delete a field a live consumer still reads.** This breaks the JSON wire format. Follow this process:

1. **Comment the field as deprecated:** `/// Deprecated: use new_field_name instead. Remove after YYYY-MM-DD.`
2. Coordinate removal across `onion` (stop writing) and `rtylr` (stop reading) before actually deleting the field from `.onk`.
3. Once removed, never reuse the old field name for something with a different meaning — a stale client could still send/expect the old semantics.

### Enum Evolution Rules

Enums are the most dangerous part of a schema to change. Follow these rules:

- **Adding a new enum value is backwards compatible.** Document new values clearly.
- **Never rename an existing enum value.** The name is part of the JSON serialisation — renaming it is a breaking change equivalent to renaming a field.
- **Never remove an enum value** a live consumer might still send or expect. Leave it in place with a `/// Deprecated` comment instead.

### Optionality

Every field is present-or-absent in JSON; there are no proto3-style implicit zero values to worry about. Use `field: T?` whenever "the field was not provided" is a meaningful, distinct state from any concrete value of `T` (e.g. `false`, `0`, `""`), and add `@nullable` when an explicit JSON `null` also needs to be distinguishable from absence. Document in a comment why a field is optional/nullable.

### When to Introduce a Breaking Change

Create a new `v2` directory (e.g. `api/{domain}/v2/`) only for genuinely breaking changes that cannot be handled via deprecation:

| Change | Breaking? | Action |
|---|---|---|
| Add a new field | No | Add to v1 |
| Mark a field deprecated | No | Add comment, keep the field |
| Rename a field | **Yes** | Create v2 |
| Rename a message | **Yes** | Create v2 |
| Change a field's type | **Yes** | Create v2 |
| Change semantic meaning of a field | **Yes** | Create v2 |
| Add a new enum value | No | Add to v1 |
| Rename/remove an enum value | **Yes** | Create v2 |

### TypeScript Exports Are Automatic

`packages/voxie/package.json` re-exports `gen/ts` via a wildcard `exports` map (`"./*": "./gen/ts/*.ts"`) — there is no per-file export allowlist to maintain. Any path under `gen/ts/` is importable from `@voxire/voxie/{path}` the moment it's generated and committed. `make generate` always runs `rm -rf gen/go gen/ts` first, so deleting an `.onk` file and regenerating removes its output cleanly — there's no stale-export risk to guard against manually.

Order of operations when adding a new service:
1. Create `api/common/{domain}/v1/` for shared contracts or `api/{app}/{capability}/{subdomain}/v1/` for app-owned contracts, with `models.onk` + `service.onk`.
2. Run `onek check .`.
3. Run `make generate` and commit the `.onk` source together with the regenerated `gen/go` and `gen/ts`.
4. In `onion`: bump the `voxie` submodule pointer — no generation step needed there.
5. In `packages` (shared by `rtylr` and `pos`): bump the `voxie` submodule pointer at `packages/voxie`. No local generation step or export-path edit needed there either.

---

## 5. CONTRACT DOMAINS

| Domain | Package examples | Covers |
|---|---|---|
| `common` | `common.auth.v1`, `common.oauth.v1`, `common.pagination.v1` | Shared auth (login, register, verify email, reset password, Google auth, invite accept), OAuth token issuing, pagination |
| `hub` | `hub.team.members.v1`, `hub.business.v1` | Team, billing, config, permissions, API keys, webhooks, notifications, security |
| `erp` | `erp.product.v1`, `erp.inventory.v1` | Products, inventory, suppliers, purchase orders |
| `hr` | `hr.staff.v1`, `hr.employees.payroll.runs.v1` | Staff, departments, shifts, payroll, time off |
| `crm` | `crm.customer.v1`, `crm.loyalty.v1` | Customers, loyalty, campaigns, segments |
| `finance` | `finance.invoices.v1`, `finance.expenses.v1` | Invoices, expenses, quotes, bills, bank accounts, journal, reports |
| `flow` | `flow.workspace.v1`, `flow.task.v1` | Workspaces, projects, tasks, labels |
| `pos` | `pos.orders.v1`, `pos.orders.payments.v1` | Orders, tables, sessions, terminals, floors, KDS, combos, modifiers, gift cards, loyalty, receipts |
| `insights` | `insights.dashboard.v1`, `insights.product.v1` | Analytics deep-dive |
| `booking` | `booking.appointment.v1`, `booking.calendar.v1` | Appointments, calendars, resources, waitlists |
| `console` | `console.admin.v1` | Internal admin tooling |
| `support` | `support.inbox.v1`, `support.knowledge.v1` | Support inbox, canned responses, knowledge base |

---

## 6. CODE QUALITY RULES — NON-NEGOTIABLE

1. **Never hand-edit `gen/go` or `gen/ts`.** Both are generated and committed — regenerate with `make generate`, never patch the output directly.
2. **All new message fields must have a `///` comment** describing their purpose.
3. **Field names are snake_case** in `.onk` — no exceptions.
4. **Never remove or rename an existing field** a live consumer reads without a deprecation plan (see section 4).
5. **Never rename an existing field or message** without creating a v2 directory.
6. **Run `make generate` and commit `gen/go` + `gen/ts`** after `.onk` changes.
7. **Run `onek check .` before committing.**
8. **New services follow** `api/common/{domain}/v1/` or `api/{app}/{capability}/{subdomain}/v1/`, each with `models.onk` + `service.onk`.
9. **Never hand-add entries to `packages/voxie/package.json`'s `exports` map.** It's a wildcard (`./*` → `./gen/ts/*.ts`) — new schema files are importable automatically once `gen/ts` is regenerated and committed.

---

## 7. SECURITY RULES

- **No secrets in `.onk` definitions.** Schema files define types, not values. Never embed API keys, tokens, or credentials as default field values or comments.
- **Sensitive fields must be documented.** If a message carries sensitive data (passwords, tokens, PII), add a comment on the field flagging this: `/// Sensitive: do not log this field.`
- **Field renaming is a security concern too.** A renamed field could cause consumers to silently stop reading it, leading to auth or permission checks being skipped. Treat renames with the same care as breaking changes.
- **Deprecate, don't delete, auth-related fields.** If you are removing a field used in authentication or authorisation, coordinate the deprecation window carefully with both `onion` and `rtylr` to avoid a security gap.

---

## 8. KNOWN ISSUES

| Issue | Location | Priority |
|---|---|---|
| Go module is declared as `v0.0.0` — no semantic versioning enforced | `go.mod` | Low |
| Frontend validation metadata generator has no `.onk`-aware equivalent — always emits empty metadata | `scripts/generate_validation_metadata.py` | Medium — see README.md "Known gaps" |
| No `.onk` decorator for cross-field constraints (e.g. "exactly one of these fields required") | `flow.task.v1.ListTasksRequest` | Low — one known occurrence, handled with a manual check |

---

## 9. SESSION STARTUP CHECKLIST

1. `git pull origin main` on the `voxie` repo
2. Check if `onion` and `packages` (the `voxie` submodule at `packages/voxie`) have pending submodule ref updates
3. Re-read this CLAUDE.md
4. Confirm `onek` is installed: `onek --version` (or `make install-onek`)

---

## 10. ENVIRONMENT SETUP

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| onek | pinned (`ONK_VERSION` in Makefile) | `make install-onek` |
| pnpm | — | `make install-js` |

### Generate Consumer Code

```bash
# In voxie/ — regenerates gen/go and gen/ts, and commits both
make generate
```

There is nothing to run in `onion`, `packages`, `rtylr`, or `pos` — they all consume the committed output directly from the submodule.

### Add a new service

1. Create `api/common/{domain}/v1/` for shared contracts or `api/{app}/{capability}/{subdomain}/v1/` for app-owned contracts, with `models.onk` + `service.onk`.
2. Run `onek check .`.
3. Run `make generate` and commit the `.onk` files together with the regenerated `gen/go` and `gen/ts`.
4. In `onion`: `cd voxie && git pull && cd ..` → `git add voxie` → import from the generated capability path under `gen/go/api/`. No local generation step needed.
5. In `packages`: `cd voxie && git pull && cd ..` → `git add voxie` → import from `@voxire/voxie/api/{capability path}`. No local generation step or export-path edit needed.

### Submodule update workflow

```bash
# In onion/
cd voxie && git pull origin main && cd ..
git add voxie
git commit -m "update voxie contracts"

# In packages/ (consumed by both rtylr and pos)
cd voxie && git pull origin main && cd ..
git add voxie
git commit -m "update voxie contracts"
```
